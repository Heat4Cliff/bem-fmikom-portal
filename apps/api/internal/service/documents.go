package service

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"
	"unicode"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/storage"
)

var (
	ErrDocumentInvalid   = errors.New("document validation failed")
	ErrDocumentForbidden = errors.New("document access forbidden")
	ErrDocumentNotFound  = errors.New("document not found")
	ErrDocumentConflict  = errors.New("document conflict")
)

const documentNameMax = 240

type Document struct {
	ID               int64     `json:"id"`
	PeriodID         int64     `json:"period_id"`
	PeriodName       string    `json:"period_name"`
	DivisionID       int64     `json:"division_id"`
	DivisionName     string    `json:"division_name"`
	UploadedBy       int64     `json:"uploaded_by"`
	UploaderName     string    `json:"uploader_name"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	OriginalFilename string    `json:"original_filename"`
	MimeType         string    `json:"mime_type"`
	FileExtension    string    `json:"file_extension"`
	SizeBytes        int64     `json:"size_bytes"`
	SHA256           string    `json:"sha256"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type DocumentFilters struct {
	Page, Limit                int
	Search, Category, FileType string
	PeriodID, DivisionID       *int64
	DateFrom, DateTo           *time.Time
}

type DocumentService struct {
	db       *sql.DB
	storage  storage.DocumentStorage
	log      *repository.ActivityLogs
	maxBytes int64
}

func NewDocumentService(db *sql.DB, files storage.DocumentStorage, maxBytes int64) *DocumentService {
	return &DocumentService{db: db, storage: files, log: repository.NewActivityLogs(db), maxBytes: maxBytes}
}

func (s *DocumentService) List(ctx context.Context, user domain.User, f DocumentFilters) (Result[Document], error) {
	if !documentRole(user.Role) || (user.Role == domain.RoleKepala && (user.DivisionID == nil || *user.DivisionID < 1)) {
		return Result[Document]{}, ErrDocumentForbidden
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Category != "" && !validDocumentCategory(f.Category) {
		return Result[Document]{}, ErrDocumentInvalid
	}
	if f.FileType != "" {
		f.FileType = strings.ToLower(strings.TrimPrefix(f.FileType, "."))
		if !validDocumentExtension("." + f.FileType) {
			return Result[Document]{}, ErrDocumentInvalid
		}
	}
	if user.Role == domain.RoleKepala {
		if f.DivisionID != nil && *f.DivisionID != *user.DivisionID {
			return Result[Document]{}, ErrDocumentForbidden
		}
		f.DivisionID = user.DivisionID
	}
	where := []string{"d.delete_pending=FALSE"}
	args := []any{}
	if f.PeriodID != nil {
		where = append(where, "d.period_id=?")
		args = append(args, *f.PeriodID)
	}
	if f.DivisionID != nil {
		where = append(where, "d.division_id=?")
		args = append(args, *f.DivisionID)
	}
	if f.Category != "" {
		where = append(where, "d.category=?")
		args = append(args, f.Category)
	}
	if f.FileType != "" {
		where = append(where, "d.file_extension=?")
		args = append(args, "."+f.FileType)
	}
	if f.DateFrom != nil {
		where = append(where, "d.created_at>=?")
		args = append(args, f.DateFrom.UTC())
	}
	if f.DateTo != nil {
		where = append(where, "d.created_at<?")
		args = append(args, f.DateTo.UTC())
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		where = append(where, "(LOCATE(?,d.name)>0 OR LOCATE(?,d.original_filename)>0)")
		args = append(args, q, q)
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM documents d WHERE "+clause, args...).Scan(&total); err != nil {
		return Result[Document]{}, err
	}
	args = append(args, f.Limit, (f.Page-1)*f.Limit)
	rows, err := s.db.QueryContext(ctx, `SELECT d.id,d.period_id,p.name,d.division_id,v.name,d.uploaded_by,u.name,d.name,d.category,d.original_filename,d.mime_type,d.file_extension,d.size_bytes,d.sha256,d.created_at,d.updated_at FROM documents d JOIN periods p ON p.id=d.period_id JOIN divisions v ON v.id=d.division_id JOIN users u ON u.id=d.uploaded_by WHERE `+clause+` ORDER BY d.created_at DESC,d.id DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return Result[Document]{}, err
	}
	defer rows.Close()
	items := []Document{}
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.PeriodID, &d.PeriodName, &d.DivisionID, &d.DivisionName, &d.UploadedBy, &d.UploaderName, &d.Name, &d.Category, &d.OriginalFilename, &d.MimeType, &d.FileExtension, &d.SizeBytes, &d.SHA256, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return Result[Document]{}, err
		}
		items = append(items, d)
	}
	if err := rows.Err(); err != nil {
		return Result[Document]{}, err
	}
	pages := (total + f.Limit - 1) / f.Limit
	return Result[Document]{Items: items, Page: Page{Page: f.Page, Limit: f.Limit, Total: total, TotalPages: pages}}, nil
}

func (s *DocumentService) Get(ctx context.Context, user domain.User, id int64) (Document, error) {
	d, _, err := s.get(ctx, id, false)
	if err != nil {
		return Document{}, err
	}
	if !documentCanAccess(user, d.DivisionID) {
		return Document{}, ErrDocumentNotFound
	}
	return d, nil
}

func (s *DocumentService) Upload(ctx context.Context, user domain.User, in DocumentUpload) (Document, error) {
	if !documentCanAccess(user, in.DivisionID) {
		return Document{}, ErrDocumentForbidden
	}
	if user.Role == domain.RoleKepala {
		in.DivisionID = *user.DivisionID
	}
	if in.PeriodID < 1 || in.DivisionID < 1 || !validDocumentCategory(in.Category) || strings.TrimSpace(in.Name) == "" || runeLen(strings.TrimSpace(in.Name)) > documentNameMax || in.SizeBytes < 1 || in.SizeBytes > s.maxBytes || len(in.SHA256) != 64 || !validDocumentExtension(in.Extension) || !strings.EqualFold(strings.TrimSpace(in.MIMEType), canonicalDocumentMIME(in.Extension)) {
		return Document{}, ErrDocumentInvalid
	}
	if err := validateDocumentFile(s.storage, ctx, in.StorageKey, in.Extension); err != nil {
		return Document{}, ErrDocumentInvalid
	}
	var exists int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM divisions WHERE id=? AND period_id=?", in.DivisionID, in.PeriodID).Scan(&exists); err != nil {
		return Document{}, err
	}
	if exists != 1 {
		return Document{}, ErrDocumentInvalid
	}
	key := fmt.Sprintf("documents/%s/%s.bin", time.Now().UTC().Format("2006/01"), in.StorageID)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Document{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO documents(period_id,division_id,uploaded_by,name,category,original_filename,storage_key,mime_type,file_extension,size_bytes,sha256) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, in.PeriodID, in.DivisionID, user.ID, strings.TrimSpace(in.Name), in.Category, safeDocumentFilename(in.Filename), key, canonicalDocumentMIME(in.Extension), in.Extension, in.SizeBytes, in.SHA256)
	if err != nil {
		return Document{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Document{}, err
	}
	if err := s.storage.Promote(ctx, in.StorageKey, key); err != nil {
		return Document{}, err
	}
	promoted := true
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "UPLOAD", EntityType: "DOCUMENT", EntityID: fmt.Sprint(id), Details: map[string]any{"document_id": id, "category": in.Category, "period_id": in.PeriodID, "division_id": in.DivisionID, "size_bytes": in.SizeBytes, "sha256": in.SHA256}}); err != nil {
		_ = s.storage.Delete(context.Background(), key)
		return Document{}, err
	}
	if err := tx.Commit(); err != nil {
		if promoted {
			_ = s.storage.Delete(context.Background(), key)
		}
		return Document{}, err
	}
	d, _, err := s.get(ctx, id, false)
	return d, err
}

type DocumentUpload struct {
	PeriodID, DivisionID                                                         int64
	Name, Category, Filename, Extension, MIMEType, StorageKey, StorageID, SHA256 string
	SizeBytes                                                                    int64
}

func (s *DocumentService) Update(ctx context.Context, user domain.User, id int64, name, category string) (Document, error) {
	if strings.TrimSpace(name) == "" || runeLen(strings.TrimSpace(name)) > documentNameMax || !validDocumentCategory(category) {
		return Document{}, ErrDocumentInvalid
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Document{}, err
	}
	defer tx.Rollback()
	d, _, err := s.getTx(ctx, tx, id, true)
	if err != nil {
		return Document{}, err
	}
	if !documentCanAccess(user, d.DivisionID) {
		return Document{}, ErrDocumentNotFound
	}
	if _, err = tx.ExecContext(ctx, "UPDATE documents SET name=?,category=? WHERE id=? AND delete_pending=FALSE", strings.TrimSpace(name), category, id); err != nil {
		return Document{}, err
	}
	if err = s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "UPDATE", EntityType: "DOCUMENT", EntityID: fmt.Sprint(id), Details: map[string]any{"document_id": id}}); err != nil {
		return Document{}, err
	}
	if err = tx.Commit(); err != nil {
		return Document{}, err
	}
	return s.Get(ctx, user, id)
}

func (s *DocumentService) Delete(ctx context.Context, user domain.User, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	d, _, err := s.getTx(ctx, tx, id, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !documentCanAccess(user, d.DivisionID) {
		tx.Rollback()
		return ErrDocumentNotFound
	}
	if _, err = tx.ExecContext(ctx, "UPDATE documents SET delete_pending=TRUE WHERE id=?", id); err != nil {
		tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if err = s.storage.Delete(ctx, d.storageKey); err != nil {
		_, _ = s.db.ExecContext(context.Background(), "UPDATE documents SET delete_pending=FALSE WHERE id=?", id)
		return err
	}
	tx, err = s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "DELETE", EntityType: "DOCUMENT", EntityID: fmt.Sprint(id), Details: map[string]any{"document_id": id, "division_id": d.DivisionID}}); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM documents WHERE id=? AND delete_pending=TRUE", id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *DocumentService) Download(ctx context.Context, user domain.User, id int64) (storage.DocumentReader, Document, error) {
	d, key, err := s.get(ctx, id, false)
	if err != nil {
		return nil, Document{}, err
	}
	if !documentCanAccess(user, d.DivisionID) {
		return nil, Document{}, ErrDocumentNotFound
	}
	r, err := s.storage.Open(ctx, key)
	if err != nil {
		return nil, Document{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		r.Close()
		return nil, Document{}, err
	}
	if err = s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "DOWNLOAD", EntityType: "DOCUMENT", EntityID: fmt.Sprint(id), Details: map[string]any{"document_id": id, "division_id": d.DivisionID}}); err != nil {
		tx.Rollback()
		r.Close()
		return nil, Document{}, err
	}
	if err = tx.Commit(); err != nil {
		r.Close()
		return nil, Document{}, err
	}
	return r, d, nil
}

type documentRow struct {
	Document
	storageKey string
}

func (s *DocumentService) get(ctx context.Context, id int64, lock bool) (Document, string, error) {
	var d documentRow
	err := s.db.QueryRowContext(ctx, `SELECT d.id,d.period_id,p.name,d.division_id,v.name,d.uploaded_by,u.name,d.name,d.category,d.original_filename,d.mime_type,d.file_extension,d.size_bytes,d.sha256,d.created_at,d.updated_at,d.storage_key FROM documents d JOIN periods p ON p.id=d.period_id JOIN divisions v ON v.id=d.division_id JOIN users u ON u.id=d.uploaded_by WHERE d.id=? AND d.delete_pending=FALSE`, id).Scan(&d.ID, &d.PeriodID, &d.PeriodName, &d.DivisionID, &d.DivisionName, &d.UploadedBy, &d.UploaderName, &d.Name, &d.Category, &d.OriginalFilename, &d.MimeType, &d.FileExtension, &d.SizeBytes, &d.SHA256, &d.CreatedAt, &d.UpdatedAt, &d.storageKey)
	if errors.Is(err, sql.ErrNoRows) {
		return Document{}, "", ErrDocumentNotFound
	}
	return d.Document, d.storageKey, err
}
func (s *DocumentService) getTx(ctx context.Context, tx *sql.Tx, id int64, lock bool) (documentRow, string, error) {
	var d documentRow
	err := tx.QueryRowContext(ctx, `SELECT d.id,d.period_id,p.name,d.division_id,v.name,d.uploaded_by,u.name,d.name,d.category,d.original_filename,d.mime_type,d.file_extension,d.size_bytes,d.sha256,d.created_at,d.updated_at,d.storage_key FROM documents d JOIN periods p ON p.id=d.period_id JOIN divisions v ON v.id=d.division_id JOIN users u ON u.id=d.uploaded_by WHERE d.id=?`+map[bool]string{true: " FOR UPDATE", false: ""}[lock], id).Scan(&d.ID, &d.PeriodID, &d.PeriodName, &d.DivisionID, &d.DivisionName, &d.UploadedBy, &d.UploaderName, &d.Name, &d.Category, &d.OriginalFilename, &d.MimeType, &d.FileExtension, &d.SizeBytes, &d.SHA256, &d.CreatedAt, &d.UpdatedAt, &d.storageKey)
	if errors.Is(err, sql.ErrNoRows) {
		return d, "", ErrDocumentNotFound
	}
	return d, d.storageKey, err
}

func documentRole(role domain.Role) bool {
	return role == domain.RoleSuperAdmin || role == domain.RoleBPH || role == domain.RoleKepala
}
func documentCanAccess(user domain.User, division int64) bool {
	return documentRole(user.Role) && CanAccessDivision(user.Role, user.DivisionID, division)
}
func validDocumentCategory(v string) bool {
	return v == "SK" || v == "PROPOSAL" || v == "LPJ" || v == "OTHER"
}
func validDocumentExtension(v string) bool {
	switch strings.ToLower(v) {
	case ".pdf", ".doc", ".docx", ".xls", ".xlsx":
		return true
	}
	return false
}
func canonicalDocumentMIME(ext string) string {
	switch strings.ToLower(ext) {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}
	return "application/octet-stream"
}
func safeDocumentFilename(name string) string {
	base := path.Base(strings.ReplaceAll(name, "\\", "/"))
	var b strings.Builder
	for _, r := range base {
		if unicode.IsControl(r) || r == '/' || r == '\\' {
			continue
		}
		b.WriteRune(r)
	}
	v := strings.TrimSpace(b.String())
	if v == "" || v == "." {
		return "document"
	}
	rs := []rune(v)
	if len(rs) > 255 {
		v = string(rs[:255])
	}
	return v
}

func validateDocumentFile(store storage.DocumentStorage, ctx context.Context, key, ext string) error {
	r, err := store.Open(ctx, key)
	if err != nil {
		return err
	}
	defer r.Close()
	info, err := r.Stat()
	if err != nil {
		return err
	}
	if info.Size() < 1 {
		return ErrDocumentInvalid
	}
	var head [1024]byte
	n, _ := r.ReadAt(head[:], 0)
	data := head[:n]
	switch strings.ToLower(ext) {
	case ".pdf":
		if !strings.HasPrefix(string(data), "%PDF-") {
			return ErrDocumentInvalid
		}
	case ".doc", ".xls":
		signature := []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}
		if len(data) < len(signature) || string(data[:8]) != string(signature) {
			return ErrDocumentInvalid
		}
	case ".docx", ".xlsx":
		zr, err := zip.NewReader(r, info.Size())
		if err != nil || len(zr.File) > 2048 {
			return ErrDocumentInvalid
		}
		types, main := false, false
		total := uint64(0)
		for _, f := range zr.File {
			if f.UncompressedSize64 > 32<<20 {
				return ErrDocumentInvalid
			}
			total += f.UncompressedSize64
			if total > 64<<20 {
				return ErrDocumentInvalid
			}
			if f.Name == "[Content_Types].xml" {
				types = true
				rc, e := f.Open()
				if e != nil {
					return ErrDocumentInvalid
				}
				buf, e := io.ReadAll(io.LimitReader(rc, 1<<20))
				rc.Close()
				if e != nil {
					return ErrDocumentInvalid
				}
				var content struct{ XMLName xml.Name }
				if xml.Unmarshal(buf, &content) != nil {
					return ErrDocumentInvalid
				}
				if (ext == ".docx" && strings.Contains(string(buf), "wordprocessingml.document")) || (ext == ".xlsx" && strings.Contains(string(buf), "spreadsheetml.sheet")) {
					main = true
				}
			}
			if ext == ".docx" && f.Name == "word/document.xml" {
				main = true
			}
			if ext == ".xlsx" && f.Name == "xl/workbook.xml" {
				main = true
			}
		}
		if !types || !main {
			return ErrDocumentInvalid
		}
	default:
		return ErrDocumentInvalid
	}
	return nil
}
