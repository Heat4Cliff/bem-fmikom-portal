package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/storage"
	"github.com/gin-gonic/gin"
)

type Documents struct {
	service  *service.DocumentService
	files    storage.DocumentStorage
	maxBytes int64
}

func NewDocuments(s *service.DocumentService, files storage.DocumentStorage, maxBytes int64) *Documents {
	return &Documents{service: s, files: files, maxBytes: maxBytes}
}
func (h *Documents) Register(engine *gin.Engine, auth middleware.CurrentUserLoader) {
	g := engine.Group("/api/v1/admin")
	g.Use(middleware.RequireAuth(auth), middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala))
	g.GET("/documents", h.list)
	g.POST("/documents", h.upload)
	g.GET("/documents/:id", h.get)
	g.PATCH("/documents/:id", h.update)
	g.DELETE("/documents/:id", h.delete)
	g.GET("/documents/:id/download", h.download)
}

func (h *Documents) list(c *gin.Context) {
	u, _ := middleware.CurrentUser(c)
	f := service.DocumentFilters{Page: documentQueryInt(c, "page", 1), Limit: documentQueryInt(c, "limit", 20), Search: c.Query("search"), Category: strings.ToUpper(c.Query("category")), FileType: c.Query("file_type")}
	if raw := c.Query("period_id"); raw != "" {
		v, e := strconv.ParseInt(raw, 10, 64)
		if e != nil || v < 1 {
			documentError(c, service.ErrDocumentInvalid)
			return
		}
		f.PeriodID = &v
	}
	if raw := c.Query("division_id"); raw != "" {
		v, e := strconv.ParseInt(raw, 10, 64)
		if e != nil || v < 1 {
			documentError(c, service.ErrDocumentInvalid)
			return
		}
		f.DivisionID = &v
	}
	if raw := c.Query("date_from"); raw != "" {
		v, e := time.Parse("2006-01-02", raw)
		if e != nil {
			documentError(c, service.ErrDocumentInvalid)
			return
		}
		f.DateFrom = &v
	}
	if raw := c.Query("date_to"); raw != "" {
		v, e := time.Parse("2006-01-02", raw)
		if e != nil {
			documentError(c, service.ErrDocumentInvalid)
			return
		}
		v = v.AddDate(0, 0, 1)
		f.DateTo = &v
	}
	r, err := h.service.List(c.Request.Context(), u, f)
	if err != nil {
		documentError(c, err)
		return
	}
	listSuccess(c, "Arsip dokumen berhasil diambil", r)
}

func (h *Documents) get(c *gin.Context) {
	u, _ := middleware.CurrentUser(c)
	id, ok := documentID(c)
	if !ok {
		return
	}
	d, err := h.service.Get(c.Request.Context(), u, id)
	if err != nil {
		documentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Dokumen berhasil diambil", d))
}
func (h *Documents) update(c *gin.Context) {
	u, _ := middleware.CurrentUser(c)
	id, ok := documentID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var in struct {
		Name     string `json:"name"`
		Category string `json:"category"`
	}
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if dec.Decode(&in) != nil {
		documentError(c, service.ErrDocumentInvalid)
		return
	}
	d, err := h.service.Update(c.Request.Context(), u, id, in.Name, strings.ToUpper(in.Category))
	if err != nil {
		documentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Metadata dokumen berhasil diperbarui", d))
}
func (h *Documents) delete(c *gin.Context) {
	u, _ := middleware.CurrentUser(c)
	id, ok := documentID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), u, id); err != nil {
		documentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Dokumen berhasil dihapus", gin.H{"id": id}))
}
func (h *Documents) download(c *gin.Context) {
	u, _ := middleware.CurrentUser(c)
	id, ok := documentID(c)
	if !ok {
		return
	}
	reader, d, err := h.service.Download(c.Request.Context(), u, id)
	if err != nil {
		documentError(c, err)
		return
	}
	defer reader.Close()
	info, err := reader.Stat()
	if err != nil {
		documentError(c, err)
		return
	}
	c.Header("Content-Type", d.MimeType)
	c.Header("Content-Disposition", serviceDocumentDisposition(d.OriginalFilename))
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, reader)
}

func (h *Documents) upload(c *gin.Context) {
	u, _ := middleware.CurrentUser(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxBytes+(256<<10))
	reader, err := c.Request.MultipartReader()
	if err != nil {
		documentError(c, service.ErrDocumentInvalid)
		return
	}
	fields := map[string]string{}
	var stage, filename string
	var fileMIME string
	var object storage.DocumentObject
	defer func() {
		if stage != "" {
			_ = h.files.Delete(context.Background(), stage)
		}
	}()
	for {
		part, e := reader.NextPart()
		if errors.Is(e, io.EOF) {
			break
		}
		if e != nil {
			var over *http.MaxBytesError
			if errors.As(e, &over) {
				c.JSON(http.StatusRequestEntityTooLarge, FailureCode("REQUEST_TOO_LARGE", "Ukuran dokumen melebihi batas", nil))
				return
			}
			documentError(c, service.ErrDocumentInvalid)
			return
		}
		name := part.FormName()
		if name == "file" {
			if stage != "" {
				part.Close()
				documentError(c, service.ErrDocumentInvalid)
				return
			}
			filename = part.FileName()
			fileMIME, _, _ = mime.ParseMediaType(part.Header.Get("Content-Type"))
			id, e := randomDocumentID()
			if e != nil {
				part.Close()
				documentError(c, e)
				return
			}
			stage = ".staging/" + id + ".part"
			object, e = h.files.Save(c.Request.Context(), stage, part, h.maxBytes)
			part.Close()
			if e != nil {
				stage = ""
				var over *http.MaxBytesError
				if errors.Is(e, storage.ErrSizeLimit) || errors.As(e, &over) {
					c.JSON(http.StatusRequestEntityTooLarge, FailureCode("REQUEST_TOO_LARGE", "Ukuran dokumen melebihi batas", nil))
					return
				}
				documentError(c, e)
				return
			}
			continue
		}
		if name == "" || len(name) > 40 {
			part.Close()
			documentError(c, service.ErrDocumentInvalid)
			return
		}
		value, e := io.ReadAll(io.LimitReader(part, 4097))
		part.Close()
		if e != nil || len(value) > 4096 {
			documentError(c, service.ErrDocumentInvalid)
			return
		}
		if _, exists := fields[name]; exists {
			documentError(c, service.ErrDocumentInvalid)
			return
		}
		fields[name] = string(value)
	}
	parseID := func(k string) (int64, bool) { v, e := strconv.ParseInt(fields[k], 10, 64); return v, e == nil && v > 0 }
	period, pok := parseID("period_id")
	division, dok := parseID("division_id")
	ext := strings.ToLower(filepath.Ext(strings.ReplaceAll(filename, "\\", "/")))
	if filename == "" || !pok || !dok || !validMultipartCategory(fields["category"]) || strings.TrimSpace(fields["name"]) == "" {
		documentError(c, service.ErrDocumentInvalid)
		return
	}
	keyID := strings.TrimSuffix(strings.TrimPrefix(stage, ".staging/"), ".part")
	d, err := h.service.Upload(c.Request.Context(), u, service.DocumentUpload{PeriodID: period, DivisionID: division, Name: fields["name"], Category: strings.ToUpper(fields["category"]), Filename: filename, Extension: ext, MIMEType: fileMIME, StorageKey: stage, StorageID: keyID, SHA256: object.SHA256, SizeBytes: object.SizeBytes})
	if err != nil {
		documentError(c, err)
		return
	}
	stage = ""
	c.JSON(http.StatusCreated, Success("Dokumen berhasil diarsipkan", d))
}

func documentID(c *gin.Context) (int64, bool) {
	id, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil || id < 1 {
		c.JSON(http.StatusBadRequest, FailureCode("VALIDATION_ERROR", "ID tidak valid", nil))
		return 0, false
	}
	return id, true
}
func documentError(c *gin.Context, err error) {
	status, code, msg := http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan server"
	switch {
	case errors.Is(err, service.ErrDocumentInvalid):
		status, code, msg = http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Data dokumen tidak valid"
	case errors.Is(err, service.ErrDocumentForbidden):
		status, code, msg = http.StatusForbidden, "FORBIDDEN", "Akses ditolak"
	case errors.Is(err, service.ErrDocumentNotFound):
		status, code, msg = http.StatusNotFound, "NOT_FOUND", "Dokumen tidak ditemukan"
	case errors.Is(err, service.ErrDocumentConflict):
		status, code, msg = http.StatusConflict, "CONFLICT", "Dokumen tidak dapat diubah"
	}
	c.JSON(status, FailureCode(code, msg, nil))
}
func randomDocumentID() (string, error) {
	b := make([]byte, 16)
	if _, e := rand.Read(b); e != nil {
		return "", fmt.Errorf("generate storage identifier")
	}
	return hex.EncodeToString(b), nil
}
func validMultipartCategory(v string) bool {
	v = strings.ToUpper(v)
	return v == "SK" || v == "PROPOSAL" || v == "LPJ" || v == "OTHER"
}
func serviceDocumentDisposition(filename string) string {
	return "attachment; filename*=UTF-8''" + urlQueryEscape(filename)
}

func documentQueryInt(c *gin.Context, key string, fallback int) int {
	value := c.Query(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}
func urlQueryEscape(s string) string {
	var b strings.Builder
	const hex = "0123456789ABCDEF"
	for _, v := range []byte(s) {
		if v >= 'a' && v <= 'z' || v >= 'A' && v <= 'Z' || v >= '0' && v <= '9' || strings.ContainsRune("-_.~", rune(v)) {
			b.WriteByte(v)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[v>>4])
			b.WriteByte(hex[v&15])
		}
	}
	return b.String()
}
