package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/storage"
)

var (
	ErrMediaInvalid              = errors.New("media validation failed")
	ErrMediaForbidden            = errors.New("media upload forbidden")
	ErrMediaNotFound             = errors.New("media not found")
	ErrMediaUnsupported          = errors.New("image format is not supported")
	ErrMediaCorrupt              = errors.New("image data is corrupt or unreadable")
	ErrMediaDimensions           = errors.New("image dimensions exceed safe limits")
	ErrMediaTooLarge             = errors.New("image exceeds configured size limit")
	ErrMediaAnimationUnsupported = errors.New("animated GIF is not supported")
)

type MediaPurpose string

const (
	MediaArticleCover      MediaPurpose = "ARTICLE_COVER"
	MediaAnnouncementImage MediaPurpose = "ANNOUNCEMENT_IMAGE"
	MediaGalleryPreview    MediaPurpose = "GALLERY_PREVIEW"
	MediaMemberPhoto       MediaPurpose = "MEMBER_PHOTO"
	MediaLandingHero       MediaPurpose = "LANDING_HERO"
	MediaLandingBranding   MediaPurpose = "LANDING_BRANDING"
	maxMediaDimension                   = 8000
	maxMediaPixels                      = 12_000_000
)

type MediaAsset struct {
	ID               int64        `json:"id"`
	Purpose          MediaPurpose `json:"purpose"`
	OriginalFilename string       `json:"original_filename"`
	MIMEType         string       `json:"mime_type"`
	OriginalMIMEType string       `json:"original_mime_type"`
	StoredMIMEType   string       `json:"stored_mime_type"`
	FileExtension    string       `json:"file_extension"`
	SizeBytes        int64        `json:"size_bytes"`
	Width            int          `json:"width"`
	Height           int          `json:"height"`
	PublicURL        string       `json:"public_url"`
	CreatedAt        time.Time    `json:"created_at"`
}

type MediaService struct {
	db      *sql.DB
	storage storage.PublicMediaStorage
	maxSize int64
	log     *repository.ActivityLogs
}

func NewMediaService(db *sql.DB, files storage.PublicMediaStorage, maxSize int64) *MediaService {
	return &MediaService{db: db, storage: files, maxSize: maxSize, log: repository.NewActivityLogs(db)}
}

func (s *MediaService) Upload(ctx context.Context, actor domain.User, purpose MediaPurpose, name, declaredMIME string, source io.ReadSeeker) (MediaAsset, error) {
	if !mediaPurposeAllowed(actor.Role, purpose) {
		return MediaAsset{}, ErrMediaForbidden
	}
	name = safeMediaFilename(name)
	ext := strings.ToLower(filepath.Ext(name))
	if name == "" || s.maxSize < 1 {
		return MediaAsset{}, ErrMediaUnsupported
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return MediaAsset{}, ErrMediaCorrupt
	}
	input, err := io.ReadAll(io.LimitReader(source, s.maxSize+1))
	if err != nil {
		return MediaAsset{}, ErrMediaCorrupt
	}
	processed, err := processMediaImage(input, ext, declaredMIME, s.maxSize)
	if err != nil {
		return MediaAsset{}, err
	}
	keyID, err := randomMediaKey()
	if err != nil {
		return MediaAsset{}, err
	}
	key := keyID + processed.storedExt
	object, err := s.storage.Save(ctx, key, bytes.NewReader(processed.storedBytes), s.maxSize)
	if err != nil {
		return MediaAsset{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		_ = s.storage.Delete(ctx, key)
		return MediaAsset{}, fmt.Errorf("begin media metadata transaction: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO media_assets(uploaded_by,purpose,original_filename,original_mime_type,storage_key,mime_type,file_extension,size_bytes,width,height,sha256) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, actor.ID, purpose, name, acceptedImageTypes[ext], key, processed.storedMIME, processed.storedExt, object.SizeBytes, processed.width, processed.height, object.SHA256)
	if err != nil {
		_ = s.storage.Delete(ctx, key)
		return MediaAsset{}, fmt.Errorf("persist media metadata: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		_ = s.storage.Delete(ctx, key)
		return MediaAsset{}, fmt.Errorf("read media identifier: %w", err)
	}
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &actor.ID, Action: "UPLOAD", EntityType: "MEDIA_ASSET", EntityID: fmt.Sprint(id), Details: map[string]any{"purpose": purpose, "media_id": id, "size_bytes": object.SizeBytes, "original_mime_type": acceptedImageTypes[ext], "stored_mime_type": processed.storedMIME}}); err != nil {
		_ = s.storage.Delete(ctx, key)
		return MediaAsset{}, err
	}
	if err := tx.Commit(); err != nil {
		_ = s.storage.Delete(ctx, key)
		return MediaAsset{}, fmt.Errorf("commit media metadata: %w", err)
	}
	return MediaAsset{ID: id, Purpose: purpose, OriginalFilename: name, MIMEType: processed.storedMIME, OriginalMIMEType: acceptedImageTypes[ext], StoredMIMEType: processed.storedMIME, FileExtension: processed.storedExt, SizeBytes: object.SizeBytes, Width: processed.width, Height: processed.height, PublicURL: fmt.Sprintf("/api/v1/public/media/%d", id), CreatedAt: time.Now().UTC()}, nil
}

func (s *MediaService) Open(ctx context.Context, id int64) (storage.DocumentReader, MediaAsset, error) {
	if id < 1 {
		return nil, MediaAsset{}, ErrMediaNotFound
	}
	var asset MediaAsset
	var key string
	err := s.db.QueryRowContext(ctx, `SELECT id,purpose,original_filename,original_mime_type,storage_key,mime_type,file_extension,size_bytes,width,height,created_at FROM media_assets WHERE id=?`, id).Scan(&asset.ID, &asset.Purpose, &asset.OriginalFilename, &asset.OriginalMIMEType, &key, &asset.MIMEType, &asset.FileExtension, &asset.SizeBytes, &asset.Width, &asset.Height, &asset.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, MediaAsset{}, ErrMediaNotFound
	}
	if err != nil {
		return nil, MediaAsset{}, err
	}
	reader, err := s.storage.Open(ctx, key)
	if err != nil {
		return nil, MediaAsset{}, err
	}
	asset.PublicURL = fmt.Sprintf("/api/v1/public/media/%d", id)
	asset.StoredMIMEType = asset.MIMEType
	return reader, asset, nil
}

func mediaPurposeAllowed(role domain.Role, purpose MediaPurpose) bool {
	switch role {
	case domain.RoleSuperAdmin, domain.RoleBPH:
		return validMediaPurpose(purpose)
	case domain.RoleKepala:
		return purpose == MediaArticleCover || purpose == MediaGalleryPreview
	default:
		return false
	}
}

func validMediaPurpose(purpose MediaPurpose) bool {
	switch purpose {
	case MediaArticleCover, MediaAnnouncementImage, MediaGalleryPreview, MediaMemberPhoto, MediaLandingHero, MediaLandingBranding:
		return true
	default:
		return false
	}
}

func validMediaDimensions(width, height int) bool {
	return width > 0 && height > 0 && width <= maxMediaDimension && height <= maxMediaDimension && width <= maxMediaPixels/height
}

func randomMediaKey() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate media key: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

func safeMediaFilename(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	var out strings.Builder
	for _, r := range name {
		if unicode.IsControl(r) || r == '/' || r == '\\' || r == ':' {
			continue
		}
		out.WriteRune(r)
	}
	value := strings.TrimSpace(out.String())
	if value == "" || value == "." || value == ".." || len(value) > 255 {
		return ""
	}
	return value
}
