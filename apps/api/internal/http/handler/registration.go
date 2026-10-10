package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

type Registrations struct{ service *service.RegistrationService }

func NewRegistrations(registrations *service.RegistrationService) *Registrations {
	return &Registrations{service: registrations}
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
	if !isJSONRequest(c) {
		c.JSON(http.StatusUnsupportedMediaType, FailureCode("UNSUPPORTED_MEDIA_TYPE", "Gunakan format JSON", nil))
		return
	}
	id, ok := registrationID(c, "id")
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.RegistrationMaxRequestBytes)
	var input service.RegistrationSubmitInput
	if err := decodeStrictJSON(c, &input); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			c.JSON(http.StatusRequestEntityTooLarge, FailureCode("REQUEST_TOO_LARGE", "Ukuran request melebihi batas", nil))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Request body tidak valid", nil))
		return
	}
	receipt, err := h.service.Submit(c.Request.Context(), id, input, time.Now().UTC())
	if err != nil {
		registrationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Pendaftaran berhasil dikirim", receipt))
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
	}
	c.JSON(status, FailureCode(code, message, nil))
}
