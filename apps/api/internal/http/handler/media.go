package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/storage"
	"github.com/gin-gonic/gin"
)

type Media struct {
	service *service.MediaService
	files   storage.PublicMediaStorage
	maxSize int64
}

func NewMedia(s *service.MediaService, files storage.PublicMediaStorage, maxSize int64) *Media {
	return &Media{service: s, files: files, maxSize: maxSize}
}

func (h *Media) Register(engine *gin.Engine, auth middleware.CurrentUserLoader) {
	engine.POST("/api/v1/admin/media", middleware.RequireAuth(auth), h.upload)
	engine.GET("/api/v1/public/media/:id", h.read)
}

func (h *Media) upload(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxSize+(64<<10))
	if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			c.JSON(http.StatusRequestEntityTooLarge, FailureCode("MEDIA_TOO_LARGE", mediaSizeLimitMessage(h.maxSize), nil))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Formulir unggah gambar tidak valid", nil))
		return
	}
	defer func() {
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
	}()
	files := c.Request.MultipartForm.File["file"]
	if len(files) != 1 || len(c.Request.MultipartForm.File) != 1 || len(c.Request.MultipartForm.Value) != 1 {
		c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Unggah tepat satu file dengan purpose yang valid", nil))
		return
	}
	purposeValues := c.Request.MultipartForm.Value["purpose"]
	if len(purposeValues) != 1 {
		c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Tujuan gambar tidak valid", nil))
		return
	}
	if files[0].Size < 1 || files[0].Size > h.maxSize {
		c.JSON(http.StatusRequestEntityTooLarge, FailureCode("MEDIA_TOO_LARGE", mediaSizeLimitMessage(h.maxSize), nil))
		return
	}
	file, err := files[0].Open()
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "File gambar tidak dapat dibaca", nil))
		return
	}
	defer file.Close()
	asset, err := h.service.Upload(c.Request.Context(), user, service.MediaPurpose(purposeValues[0]), files[0].Filename, files[0].Header.Get("Content-Type"), file)
	if err != nil {
		mediaError(c, err, h.maxSize)
		return
	}
	c.JSON(http.StatusCreated, Success("Gambar berhasil diunggah", asset))
}

func (h *Media) read(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		mediaError(c, service.ErrMediaNotFound, h.maxSize)
		return
	}
	reader, asset, err := h.service.Open(c.Request.Context(), id)
	if err != nil {
		mediaError(c, err, h.maxSize)
		return
	}
	defer reader.Close()
	info, err := reader.Stat()
	if err != nil {
		mediaError(c, err, h.maxSize)
		return
	}
	c.Header("Content-Type", asset.MIMEType)
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, reader)
}

func mediaError(c *gin.Context, err error, maxBytes int64) {
	switch {
	case errors.Is(err, service.ErrMediaForbidden):
		c.JSON(http.StatusForbidden, FailureCode("FORBIDDEN", "Akses unggah media ditolak", nil))
	case errors.Is(err, service.ErrMediaUnsupported):
		c.JSON(http.StatusUnprocessableEntity, FailureCode("MEDIA_FORMAT_UNSUPPORTED", "Format gambar belum didukung.", nil))
	case errors.Is(err, service.ErrMediaAnimationUnsupported):
		c.JSON(http.StatusUnprocessableEntity, FailureCode("MEDIA_ANIMATION_UNSUPPORTED", "GIF animasi belum didukung. Gunakan GIF statis atau format lain.", nil))
	case errors.Is(err, service.ErrMediaCorrupt):
		c.JSON(http.StatusUnprocessableEntity, FailureCode("MEDIA_IMAGE_INVALID", "File gambar rusak atau tidak dapat dibaca.", nil))
	case errors.Is(err, service.ErrMediaDimensions):
		c.JSON(http.StatusUnprocessableEntity, FailureCode("MEDIA_DIMENSIONS_INVALID", "Dimensi gambar melebihi batas aman.", nil))
	case errors.Is(err, service.ErrMediaTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, FailureCode("MEDIA_TOO_LARGE", mediaSizeLimitMessage(maxBytes), nil))
	case errors.Is(err, service.ErrMediaInvalid):
		c.JSON(http.StatusUnprocessableEntity, FailureCode("MEDIA_IMAGE_INVALID", "File gambar rusak atau tidak dapat dibaca.", nil))
	case errors.Is(err, service.ErrMediaNotFound):
		c.JSON(http.StatusNotFound, FailureCode("NOT_FOUND", "Media tidak ditemukan", nil))
	case errors.Is(err, storage.ErrSizeLimit):
		c.JSON(http.StatusRequestEntityTooLarge, FailureCode("MEDIA_TOO_LARGE", "Ukuran gambar melebihi batas.", nil))
	default:
		c.JSON(http.StatusInternalServerError, FailureCode("INTERNAL_ERROR", "Terjadi kesalahan server", nil))
	}
}

func mediaSizeLimitMessage(maxBytes int64) string {
	mebibytes := strconv.FormatFloat(float64(maxBytes)/(1<<20), 'f', -1, 64)
	return "Ukuran gambar melebihi batas " + mebibytes + " MiB."
}
