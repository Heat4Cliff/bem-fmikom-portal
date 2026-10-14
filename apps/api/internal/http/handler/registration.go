package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/storage"
	"github.com/gin-gonic/gin"
)

type Registrations struct {
	service  *service.RegistrationService
	files    storage.DocumentStorage
	imageMax int64
}

func NewRegistrations(registrations *service.RegistrationService, files storage.DocumentStorage, imageMax int64) *Registrations {
	return &Registrations{service: registrations, files: files, imageMax: imageMax}
}

func (h *Registrations) Register(engine *gin.Engine, auth middleware.CurrentUserLoader, origins []string, limiter *middleware.RegistrationLimiter) {
	public := engine.Group("/api/v1/public")
	public.GET("/registrations", h.publicList)
	public.GET("/registrations/:slug", h.publicDetail)
	public.POST("/registrations/:id/submissions", middleware.RejectUntrustedBrowserOrigin(origins), limiter.Middleware(), h.submit)

	admin := engine.Group("/api/v1/admin")
	admin.Use(middleware.RequireAuth(auth))
	admin.GET("/registrations", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.list)
	admin.POST("/registrations", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.create)
	admin.GET("/registrations/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.detail)
	admin.PATCH("/registrations/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.update)
	admin.DELETE("/registrations/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.delete)
	admin.POST("/registrations/:id/open", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.open)
	admin.POST("/registrations/:id/close", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.close)
	admin.PUT("/registrations/:id/fields", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.saveFields)
	admin.GET("/registrations/:id/submissions", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.submissions)
	admin.GET("/registrations/:id/submissions/:submissionId", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.submissionDetail)
	admin.GET("/registrations/:id/submissions/:submissionId/files/:fileId", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.submissionFile)
	admin.GET("/registrations/:id/export", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.export)
}

func (h *Registrations) publicList(c *gin.Context) {
	page, limit := queryPage(c)
	filters, ok := registrationFilters(c, page, limit)
	if !ok {
		registrationError(c, service.ErrRegistrationInvalid)
		return
	}
	result, err := h.service.PublicList(c.Request.Context(), filters, time.Now().UTC())
	if err != nil {
		registrationError(c, err)
		return
	}
	listSuccess(c, "Data pendaftaran berhasil diambil", result)
}

func (h *Registrations) publicDetail(c *gin.Context) {
	item, err := h.service.PublicGet(c.Request.Context(), c.Param("slug"), time.Now().UTC())
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data pendaftaran berhasil diambil", item))
}

func (h *Registrations) submit(c *gin.Context) {
	id, ok := registrationID(c, "id")
	if !ok {
		return
	}
	var input service.RegistrationSubmitInput
	images := map[string]service.RegistrationImageUpload{}
	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(strings.ToLower(contentType), "multipart/form-data") {
		if h.parseSubmissionMultipart(c, &input, images) != nil {
			return
		}
	} else if isJSONRequest(c) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.RegistrationMaxRequestBytes)
		if err := decodeStrictJSON(c, &input); err != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(err, &maxBytesError) {
				c.JSON(http.StatusRequestEntityTooLarge, FailureCode("REQUEST_TOO_LARGE", "Ukuran request melebihi batas", nil))
				return
			}
			c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Request body tidak valid", nil))
			return
		}
	} else {
		c.JSON(http.StatusUnsupportedMediaType, FailureCode("UNSUPPORTED_MEDIA_TYPE", "Gunakan JSON atau multipart/form-data", nil))
		return
	}
	receipt, err := h.service.SubmitWithImages(c.Request.Context(), id, input, images, time.Now().UTC())
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Pendaftaran berhasil dikirim", receipt))
}

func (h *Registrations) parseSubmissionMultipart(c *gin.Context, input *service.RegistrationSubmitInput, images map[string]service.RegistrationImageUpload) error {
	maxRequest := int64(service.RegistrationMaxImageFields)*h.imageMax + service.RegistrationMaxRequestBytes + (128 << 10)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequest)
	reader, err := c.Request.MultipartReader()
	if err != nil {
		registrationMultipartImageError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Format pengiriman gambar tidak valid.")
		return err
	}
	answersSeen := false
	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			var maxErr *http.MaxBytesError
			if errors.As(nextErr, &maxErr) {
				registrationMultipartImageError(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "Ukuran pengiriman gambar melebihi batas.")
			} else {
				c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Multipart tidak valid", nil))
			}
			return nextErr
		}
		name := part.FormName()
		if name == "answers" && part.FileName() == "" {
			if answersSeen {
				part.Close()
				c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Jawaban duplikat", nil))
				return errors.New("duplicate answers part")
			}
			answersSeen = true
			data, readErr := io.ReadAll(io.LimitReader(part, service.RegistrationMaxRequestBytes+1))
			part.Close()
			// Multipart's `answers` part is the answer object itself. Wrap it in
			// the same envelope used by JSON-only requests so the strict custom
			// decoder still rejects duplicate keys and trailing JSON.
			wrappedAnswers := make([]byte, 0, len(data)+12)
			wrappedAnswers = append(wrappedAnswers, `{"answers":`...)
			wrappedAnswers = append(wrappedAnswers, data...)
			wrappedAnswers = append(wrappedAnswers, '}')
			if readErr != nil || int64(len(data)) > service.RegistrationMaxRequestBytes || json.Unmarshal(wrappedAnswers, input) != nil {
				c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Bagian jawaban tidak valid", nil))
				return errors.New("invalid multipart answers")
			}
			continue
		}
		if !strings.HasPrefix(name, "image:") || part.FileName() == "" {
			part.Close()
			registrationMultipartImageError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Lampiran gambar tidak dikenali.")
			return errors.New("unknown multipart part")
		}
		fieldKey := strings.TrimPrefix(name, "image:")
		if fieldKey == "" || len(fieldKey) > 64 || len(images) >= service.RegistrationMaxImageFields {
			part.Close()
			registrationMultipartImageError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Jumlah lampiran gambar tidak valid.")
			return errors.New("invalid image field")
		}
		if _, duplicate := images[fieldKey]; duplicate {
			part.Close()
			registrationMultipartImageError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Lampiran gambar dikirim lebih dari sekali.")
			return errors.New("duplicate image part")
		}
		data, readErr := io.ReadAll(io.LimitReader(part, h.imageMax+1))
		filename := part.FileName()
		contentMIME, _, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
		part.Close()
		if readErr != nil || int64(len(data)) > h.imageMax {
			registrationMultipartImageError(c, http.StatusRequestEntityTooLarge, "REQUEST_TOO_LARGE", "Ukuran gambar maksimal 5 MiB.")
			return errors.New("image too large")
		}
		images[fieldKey] = service.RegistrationImageUpload{Filename: filename, MIMEType: contentMIME, Data: data}
	}
	if !answersSeen {
		c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Bagian answers wajib ada", nil))
		return errors.New("answers part missing")
	}
	return nil
}

func registrationMultipartImageError(c *gin.Context, status int, code, message string) {
	c.JSON(status, FailureCode(code, message, map[string][]string{"image": {message}}))
}

func (h *Registrations) list(c *gin.Context) {
	page, limit := queryPage(c)
	filters, ok := registrationFilters(c, page, limit)
	if !ok {
		registrationError(c, service.ErrRegistrationInvalid)
		return
	}
	user, _ := middleware.CurrentUser(c)
	result, err := h.service.List(c.Request.Context(), user, filters)
	if err != nil {
		registrationError(c, err)
		return
	}
	listSuccess(c, "Data pendaftaran berhasil diambil", result)
}

func (h *Registrations) create(c *gin.Context) {
	var input service.RegistrationCreateInput
	if !bindContent(c, &input) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	item, err := h.service.Create(c.Request.Context(), user, input)
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Formulir pendaftaran dibuat", item))
}

func (h *Registrations) detail(c *gin.Context) {
	id, ok := registrationID(c, "id")
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	item, err := h.service.Get(c.Request.Context(), user, id)
	if err != nil {
		registrationError(c, err)
		return
	}
	fields, err := h.service.Fields(c.Request.Context(), user, id)
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data formulir berhasil diambil", gin.H{"form": item, "fields": fields}))
}

func (h *Registrations) update(c *gin.Context) {
	id, ok := registrationID(c, "id")
	if !ok {
		return
	}
	var input service.RegistrationUpdateInput
	if !bindContent(c, &input) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	item, err := h.service.Update(c.Request.Context(), user, id, input)
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Metadata pendaftaran diperbarui", item))
}

func (h *Registrations) delete(c *gin.Context) {
	id, ok := registrationID(c, "id")
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.Delete(c.Request.Context(), user, id); err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Formulir pendaftaran dihapus", nil))
}

func (h *Registrations) open(c *gin.Context)  { h.changeStatus(c, true) }
func (h *Registrations) close(c *gin.Context) { h.changeStatus(c, false) }

func (h *Registrations) changeStatus(c *gin.Context, open bool) {
	id, ok := registrationID(c, "id")
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	var item service.RegistrationForm
	var err error
	if open {
		item, err = h.service.Open(c.Request.Context(), user, id)
	} else {
		item, err = h.service.Close(c.Request.Context(), user, id)
	}
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Status pendaftaran diperbarui", item))
}

func (h *Registrations) saveFields(c *gin.Context) {
	id, ok := registrationID(c, "id")
	if !ok {
		return
	}
	var input []service.RegistrationFieldInput
	if !bindContent(c, &input) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	fields, err := h.service.SaveFields(c.Request.Context(), user, id, input)
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Struktur formulir diperbarui", fields))
}

func (h *Registrations) submissions(c *gin.Context) {
	formID, ok := registrationID(c, "id")
	if !ok {
		return
	}
	page, limit := queryPage(c)
	filters, valid := registrationSubmissionFilters(c, page, limit)
	if !valid {
		registrationError(c, service.ErrRegistrationInvalid)
		return
	}
	user, _ := middleware.CurrentUser(c)
	result, err := h.service.ListSubmissions(c.Request.Context(), user, formID, filters)
	if err != nil {
		registrationError(c, err)
		return
	}
	listSuccess(c, "Data pendaftar berhasil diambil", result)
}

func (h *Registrations) submissionDetail(c *gin.Context) {
	formID, ok := registrationID(c, "id")
	if !ok {
		return
	}
	submissionID, ok := registrationID(c, "submissionId")
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	item, err := h.service.GetSubmission(c.Request.Context(), user, formID, submissionID)
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Detail pendaftar berhasil diambil", item))
}

func (h *Registrations) submissionFile(c *gin.Context) {
	formID, ok := registrationID(c, "id")
	if !ok {
		return
	}
	submissionID, ok := registrationID(c, "submissionId")
	if !ok {
		return
	}
	fileID, ok := registrationID(c, "fileId")
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	reader, file, err := h.service.OpenSubmissionFile(c.Request.Context(), user, formID, submissionID, fileID)
	if err != nil {
		registrationError(c, err)
		return
	}
	defer reader.Close()
	info, err := reader.Stat()
	if err != nil {
		registrationError(c, err)
		return
	}
	c.Header("Content-Type", file.MIMEType)
	c.Header("Content-Disposition", `inline; filename*=UTF-8''`+urlQueryEscape(file.OriginalName))
	c.Header("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, reader)
}

func (h *Registrations) export(c *gin.Context) {
	formID, ok := registrationID(c, "id")
	if !ok {
		return
	}
	filters, valid := registrationSubmissionFilters(c, 1, service.RegistrationMaxExportRows)
	if !valid {
		registrationError(c, service.ErrRegistrationInvalid)
		return
	}
	user, _ := middleware.CurrentUser(c)
	data, count, err := h.service.Export(c.Request.Context(), user, formID, filters)
	if err != nil {
		registrationError(c, err)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="registration-%d.xlsx"`, formID))
	c.Header("X-Export-Row-Count", strconv.Itoa(count))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

func registrationFilters(c *gin.Context, page, limit int) (service.RegistrationFilters, bool) {
	filters := service.RegistrationFilters{Page: page, Limit: limit, Search: c.Query("search"), Status: strings.ToUpper(strings.TrimSpace(c.Query("status")))}
	if raw := c.Query("division_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id < 1 {
			return filters, false
		}
		filters.DivisionID = &id
	}
	if raw := c.Query("period_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id < 1 {
			return filters, false
		}
		filters.PeriodID = &id
	}
	var ok bool
	filters.DateFrom, filters.DateTo, ok = aspirationDateFilters(c)
	return filters, ok
}

func registrationSubmissionFilters(c *gin.Context, page, limit int) (service.RegistrationSubmissionFilters, bool) {
	filters := service.RegistrationSubmissionFilters{Page: page, Limit: limit, Search: c.Query("search")}
	var ok bool
	filters.DateFrom, filters.DateTo, ok = aspirationDateFilters(c)
	return filters, ok
}

func registrationID(c *gin.Context, parameter string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(parameter), 10, 64)
	if err != nil || id < 1 {
		registrationError(c, service.ErrRegistrationInvalid)
		return 0, false
	}
	return id, true
}

func registrationError(c *gin.Context, err error) {
	var fieldErrors *service.RegistrationSubmissionFieldErrors
	if errors.As(err, &fieldErrors) {
		fields := make(map[string][]string, len(fieldErrors.Fields))
		for key, message := range fieldErrors.Fields {
			fields[key] = []string{message}
		}
		c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Periksa data formulir.", fields))
		return
	}
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan server"
	switch {
	case errors.Is(err, service.ErrRegistrationInvalid):
		status, code, message = http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed"
	case errors.Is(err, service.ErrRegistrationForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", "Akses ditolak"
	case errors.Is(err, service.ErrRegistrationNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan"
	case errors.Is(err, service.ErrRegistrationConflict), errors.Is(err, service.ErrRegistrationLocked), errors.Is(err, service.ErrRegistrationHasData):
		status, code, message = http.StatusConflict, "CONFLICT", "Perubahan tidak dapat dilakukan"
	case errors.Is(err, service.ErrRegistrationDuplicate):
		status, code, message = http.StatusConflict, "DUPLICATE_SUBMISSION", "Anda sudah pernah mengirim pendaftaran untuk formulir ini"
	case errors.Is(err, service.ErrRegistrationClosed):
		status, code, message = http.StatusConflict, "REGISTRATION_CLOSED", "Pendaftaran sudah tidak menerima kiriman"
	case errors.Is(err, service.ErrRegistrationExportLimit):
		status, code, message = http.StatusRequestEntityTooLarge, "EXPORT_LIMIT_EXCEEDED", "Ekspor melebihi batas 1.000 baris"
	case errors.Is(err, service.ErrRegistrationImageTooLarge):
		status, code, message = http.StatusRequestEntityTooLarge, "REGISTRATION_IMAGE_TOO_LARGE", "Ukuran gambar melebihi batas 5 MiB"
	}
	c.JSON(status, FailureCode(code, message, nil))
}
