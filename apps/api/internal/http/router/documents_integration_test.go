package router

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func runDocumentIntegration(t *testing.T, ctx context.Context, db *sql.DB, engine *gin.Engine, adminToken, bphToken, kepalaAToken, kepalaBToken string, periodID, divisionAID, divisionBID int64, storagePath string, logger *bytes.Buffer) {
	t.Helper()
	pdf := []byte("%PDF-1.7\nphase8 private archive integration\n%%EOF")
	upload := documentMultipart(t, periodID, divisionAID, "SK", "Surat Keputusan BEM", "..\\private evidence.pdf", pdf)
	created := performDocumentUpload(engine, adminToken, upload)
	if created.Code != http.StatusCreated {
		t.Fatalf("valid document upload: status=%d body=%s", created.Code, created.Body.String())
	}
	var envelope struct {
		Data struct {
			ID       int64  `json:"id"`
			Name     string `json:"name"`
			Filename string `json:"original_filename"`
			Hash     string `json:"sha256"`
		} `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &envelope); err != nil || envelope.Data.ID < 1 || envelope.Data.Filename != "private evidence.pdf" {
		t.Fatalf("document response invalid or filename unsafe: %v %s", err, created.Body.String())
	}
	id := envelope.Data.ID
	if strings.Contains(created.Body.String(), "storage_key") || strings.Contains(created.Body.String(), ".staging") {
		t.Fatal("private storage key leaked in API response")
	}
	got := getWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id)+"/download", adminToken)
	if got.Code != http.StatusOK || !bytes.Equal(got.Body.Bytes(), pdf) {
		t.Fatalf("authorized download lost bytes: status=%d body=%q", got.Code, got.Body.Bytes())
	}
	if !strings.Contains(strings.ToLower(got.Header().Get("Content-Disposition")), "attachment") || got.Header().Get("X-Content-Type-Options") != "nosniff" || !strings.Contains(got.Header().Get("Cache-Control"), "no-store") {
		t.Fatalf("unsafe download headers: %v", got.Header())
	}
	if got := getWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), kepalaAToken); got.Code != http.StatusOK {
		t.Fatalf("KB-A could not access own division document: %d %s", got.Code, got.Body.String())
	}
	if got := getWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), kepalaBToken); got.Code != http.StatusNotFound {
		t.Fatalf("KB-B cross-division detail should be concealed: %d", got.Code)
	}
	if got := getWithToken(engine, "/api/v1/admin/documents?division_id="+documentIDString(divisionBID), kepalaAToken); got.Code != http.StatusForbidden {
		t.Fatalf("KB-A cross-division list should be denied: %d", got.Code)
	}
	if got := getWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id)+"/download", bphToken); got.Code != http.StatusOK || !bytes.Equal(got.Body.Bytes(), pdf) {
		t.Fatalf("BPH cross-division download denied: %d", got.Code)
	}
	if got := getWithToken(engine, "/api/v1/admin/documents?period_id="+documentIDString(periodID)+"&category=SK&search=Surat", adminToken); got.Code != http.StatusOK || !strings.Contains(got.Body.String(), "Surat Keputusan BEM") {
		t.Fatalf("document filtering failed: %d %s", got.Code, got.Body.String())
	}
	if got := contentRequest(engine, http.MethodPatch, "/api/v1/admin/documents/"+documentIDString(id), `{"name":"Keputusan Terbit","category":"LPJ"}`, adminToken); got.Code != http.StatusOK || !strings.Contains(got.Body.String(), "Keputusan Terbit") {
		t.Fatalf("metadata update failed: %d %s", got.Code, got.Body.String())
	}
	var storedHash, storedFilename, storageKey string
	if err := db.QueryRowContext(ctx, "SELECT sha256,original_filename,storage_key FROM documents WHERE id=?", id).Scan(&storedHash, &storedFilename, &storageKey); err != nil {
		t.Fatal(err)
	}
	expectedHash := sha256.Sum256(pdf)
	if storedHash != envelope.Data.Hash || storedHash != hex.EncodeToString(expectedHash[:]) || storedFilename != "private evidence.pdf" || strings.Contains(storageKey, "private evidence") {
		t.Fatalf("stored metadata/hash unsafe: %s %s %s", storedHash, storedFilename, storageKey)
	}
	checkAdditionalDocumentFormats(t, engine, adminToken, periodID, divisionAID)
	var uploadLogs, downloadLogs int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM activity_logs WHERE entity_type='DOCUMENT' AND entity_id=? AND action='UPLOAD'", documentIDString(id)).Scan(&uploadLogs); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM activity_logs WHERE entity_type='DOCUMENT' AND entity_id=? AND action='DOWNLOAD'", documentIDString(id)).Scan(&downloadLogs); err != nil {
		t.Fatal(err)
	}
	if uploadLogs != 1 || downloadLogs < 2 {
		t.Fatalf("document activity evidence upload=%d download=%d", uploadLogs, downloadLogs)
	}
	var updateLogs int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM activity_logs WHERE entity_type='DOCUMENT' AND entity_id=? AND action='UPDATE'", documentIDString(id)).Scan(&updateLogs); err != nil {
		t.Fatal(err)
	}
	if updateLogs != 1 {
		t.Fatalf("document metadata update audit count=%d", updateLogs)
	}
	if strings.Contains(logger.String(), "private evidence.pdf") || strings.Contains(logger.String(), ".staging/") {
		t.Fatal("filename or private storage path appeared in request logs")
	}
	bad := documentMultipart(t, periodID, divisionAID, "OTHER", "Spoofed file", "spoof.pdf", []byte("not a pdf"))
	if got := performDocumentUpload(engine, adminToken, bad); got.Code != http.StatusUnprocessableEntity {
		t.Fatalf("spoofed PDF signature status=%d", got.Code)
	}
	badMIME := documentMultipartWithMIME(t, periodID, divisionAID, "OTHER", "MIME mismatch", "mismatch.pdf", pdf, "application/msword")
	if got := performDocumentUpload(engine, adminToken, badMIME); got.Code != http.StatusUnprocessableEntity {
		t.Fatalf("MIME/extension mismatch status=%d", got.Code)
	}
	if got := postDocumentWithoutToken(engine, upload); got.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated archive upload status=%d", got.Code)
	}
	if got := getWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), "invalid-token"); got.Code != http.StatusUnauthorized {
		t.Fatalf("invalid auth status=%d", got.Code)
	}
	if got := deleteWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), kepalaBToken); got.Code != http.StatusNotFound {
		t.Fatalf("cross-division delete should be concealed: %d", got.Code)
	}
	var objectKey string
	if err := db.QueryRowContext(ctx, "SELECT storage_key FROM documents WHERE id=?", id).Scan(&objectKey); err != nil {
		t.Fatal(err)
	}
	objectPath := filepath.Join(storagePath, filepath.FromSlash(objectKey))
	if err := os.Remove(objectPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(objectPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(objectPath, ".keep"), []byte("force storage delete failure"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := deleteWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), adminToken); got.Code != http.StatusInternalServerError {
		t.Fatalf("storage deletion failure status=%d", got.Code)
	}
	var pending bool
	if err := db.QueryRowContext(ctx, "SELECT delete_pending FROM documents WHERE id=?", id).Scan(&pending); err != nil || pending {
		t.Fatalf("failed storage delete was not restored: pending=%v err=%v", pending, err)
	}
	if got := getWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), adminToken); got.Code != http.StatusOK {
		t.Fatalf("failed physical delete hid document: %d", got.Code)
	}
	if err := os.Remove(filepath.Join(objectPath, ".keep")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(objectPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(objectPath, pdf, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TRIGGER phase8_test_fail_document_delete BEFORE INSERT ON activity_logs FOR EACH ROW SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='injected archive audit failure'`); err != nil {
		t.Fatal(err)
	}
	if got := deleteWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), adminToken); got.Code != http.StatusInternalServerError {
		t.Fatalf("finalization failure status=%d", got.Code)
	}
	var pendingAfterCommitFailure bool
	if err := db.QueryRowContext(ctx, "SELECT delete_pending FROM documents WHERE id=?", id).Scan(&pendingAfterCommitFailure); err != nil || !pendingAfterCommitFailure {
		t.Fatalf("failed finalization did not leave retryable pending row: pending=%v err=%v", pendingAfterCommitFailure, err)
	}
	if _, err := db.ExecContext(ctx, "DROP TRIGGER phase8_test_fail_document_delete"); err != nil {
		t.Fatal(err)
	}
	if got := deleteWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), adminToken); got.Code != http.StatusOK {
		t.Fatalf("retrying pending deletion failed: %d %s", got.Code, got.Body.String())
	}
	if got := deleteWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), adminToken); got.Code != http.StatusNotFound {
		t.Fatalf("idempotent metadata deletion unexpectedly recreated document: %d", got.Code)
	}
	if got := getWithToken(engine, "/api/v1/admin/documents/"+documentIDString(id), adminToken); got.Code != http.StatusNotFound {
		t.Fatalf("deleted document remains visible: %d", got.Code)
	}
	var deletedCount, deleteLogs int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents WHERE id=?", id).Scan(&deletedCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM activity_logs WHERE entity_type='DOCUMENT' AND entity_id=? AND action='DELETE'", documentIDString(id)).Scan(&deleteLogs); err != nil {
		t.Fatal(err)
	}
	if deletedCount != 0 || deleteLogs != 1 {
		t.Fatalf("delete finalization metadata=%d logs=%d", deletedCount, deleteLogs)
	}
}

type multipartDocument struct {
	body        []byte
	contentType string
}

func documentMultipart(t *testing.T, periodID, divisionID int64, category, name, filename string, content []byte) multipartDocument {
	ext := strings.ToLower(filepath.Ext(filename))
	fileMIME := map[string]string{".pdf": "application/pdf", ".doc": "application/msword", ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document", ".xls": "application/vnd.ms-excel", ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}[ext]
	return documentMultipartWithMIME(t, periodID, divisionID, category, name, filename, content, fileMIME)
}

func documentMultipartWithMIME(t *testing.T, periodID, divisionID int64, category, name, filename string, content []byte, fileMIME string) multipartDocument {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, value := range map[string]string{"period_id": documentIDString(periodID), "division_id": documentIDString(divisionID), "category": category, "name": name} {
		if err := w.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "file", "filename": filename}))
	header.Set("Content-Type", fileMIME)
	file, err := w.CreatePart(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = file.Write(content); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	return multipartDocument{body: b.Bytes(), contentType: w.FormDataContentType()}
}
func performDocumentUpload(engine http.Handler, token string, m multipartDocument) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/documents", bytes.NewReader(m.body))
	req.Header.Set("Content-Type", m.contentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func checkAdditionalDocumentFormats(t *testing.T, engine http.Handler, token string, periodID, divisionID int64) {
	t.Helper()
	magic := []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}
	office := append(magic, make([]byte, 504)...)
	for _, item := range []struct {
		ext  string
		data []byte
	}{
		{ext: ".doc", data: office}, {ext: ".xls", data: office},
		{ext: ".docx", data: minimalOfficeZip(t, "word/document.xml", "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml")},
		{ext: ".xlsx", data: minimalOfficeZip(t, "xl/workbook.xml", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml")},
	} {
		result := performDocumentUpload(engine, token, documentMultipart(t, periodID, divisionID, "OTHER", "Format "+item.ext, "valid"+item.ext, item.data))
		if result.Code != http.StatusCreated {
			t.Fatalf("valid %s upload status=%d response=%s", item.ext, result.Code, result.Body.String())
		}
		var payload struct {
			Data struct {
				ID int64 `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(result.Body.Bytes(), &payload); err != nil || payload.Data.ID < 1 {
			t.Fatalf("valid %s response id: %v %s", item.ext, err, result.Body.String())
		}
		if deleted := deleteWithToken(engine, "/api/v1/admin/documents/"+documentIDString(payload.Data.ID), token); deleted.Code != http.StatusOK {
			t.Fatalf("cleanup valid %s document: %d %s", item.ext, deleted.Code, deleted.Body.String())
		}
	}
}

func minimalOfficeZip(t *testing.T, mainPath, mimeType string) []byte {
	t.Helper()
	var out bytes.Buffer
	zw := zip.NewWriter(&out)
	entries := map[string]string{"[Content_Types].xml": `<?xml version="1.0"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Override PartName="/` + mainPath + `" ContentType="` + mimeType + `"/></Types>`, mainPath: `<?xml version="1.0"?><document/>`}
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func postDocumentWithoutToken(engine http.Handler, m multipartDocument) *httptest.ResponseRecorder {
	return performDocumentUpload(engine, "", m)
}
func deleteWithToken(engine http.Handler, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func documentIDString(value int64) string { return strconv.FormatInt(value, 10) }
