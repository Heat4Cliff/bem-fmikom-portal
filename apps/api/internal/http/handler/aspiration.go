package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

type Aspirations struct{ service *service.AspirationService }

type publicAspirationRequest struct {
	DivisionID  int64  `json:"division_id"`
	Name        string `json:"name"`
	NIM         string `json:"nim"`
	Email       string `json:"email"`
	Subject     string `json:"subject"`
	Message     string `json:"message"`
	IsAnonymous *bool  `json:"is_anonymous"`
}

func NewAspirations(aspirations *service.AspirationService) *Aspirations {
	return &Aspirations{service: aspirations}
}

func (h *Aspirations) Register(engine *gin.Engine, auth middleware.CurrentUserLoader, origins []string, limiter *middleware.AspirationLimiter) {
	public := engine.Group("/api/v1/public")
	public.Use(middleware.RejectUntrustedBrowserOrigin(origins))
	public.POST("/aspirations", limiter.Middleware(), h.submit)

	admin := engine.Group("/api/v1/admin")
	admin.Use(middleware.RequireAuth(auth))
	admin.GET("/aspirations", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.list)
	admin.GET("/aspirations/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.detail)
	admin.PATCH("/aspirations/:id/status", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.updateStatus)
}

func (h *Aspirations) submit(c *gin.Context) {
	if !isJSONRequest(c) {
		c.JSON(http.StatusUnsupportedMediaType, FailureCode("UNSUPPORTED_MEDIA_TYPE", "Gunakan format JSON", nil))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	var body publicAspirationRequest
	if err := decodeStrictJSON(c, &body); err != nil || body.IsAnonymous == nil {
		c.JSON(http.StatusUnprocessableEntity, Failure("Validation failed", map[string][]string{"body": {"Request body tidak valid"}}))
		return
	}
	input := service.AspirationInput{DivisionID: body.DivisionID, Name: body.Name, NIM: body.NIM, Email: body.Email, Subject: body.Subject, Message: body.Message, IsAnonymous: *body.IsAnonymous}
	receipt, err := h.service.Submit(c.Request.Context(), input)
	if err != nil {
		aspirationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Aspirasi berhasil dikirim", receipt))
}

func (h *Aspirations) list(c *gin.Context) {
	page, limit := queryPage(c)
	filters := service.AspirationFilters{Page: page, Limit: limit, Search: c.Query("search"), Status: strings.ToUpper(strings.TrimSpace(c.Query("status")))}
	if raw := c.Query("division_id"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value < 1 {
			aspirationError(c, service.ErrAspirationInvalid)
			return
		}
		filters.DivisionID = &value
	}
	var valid bool
	filters.DateFrom, filters.DateTo, valid = aspirationDateFilters(c)
	if !valid {
		aspirationError(c, service.ErrAspirationInvalid)
		return
	}
	user, _ := middleware.CurrentUser(c)
	result, err := h.service.List(c.Request.Context(), user, filters)
	if err != nil {
		aspirationError(c, err)
		return
	}
	listSuccess(c, "Data aspirasi berhasil diambil", result)
}

func (h *Aspirations) detail(c *gin.Context) {
	id, ok := aspirationID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	item, err := h.service.Get(c.Request.Context(), user, id)
	if err != nil {
		aspirationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data aspirasi berhasil diambil", item))
}

func (h *Aspirations) updateStatus(c *gin.Context) {
	id, ok := aspirationID(c)
	if !ok {
		return
	}
	if !isJSONRequest(c) {
		c.JSON(http.StatusUnsupportedMediaType, FailureCode("UNSUPPORTED_MEDIA_TYPE", "Gunakan format JSON", nil))
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)
	var body struct {
		Status string `json:"status"`
	}
	if err := decodeStrictJSON(c, &body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, Failure("Validation failed", map[string][]string{"body": {"Request body tidak valid"}}))
		return
	}
	user, _ := middleware.CurrentUser(c)
	item, err := h.service.UpdateStatus(c.Request.Context(), user, id, strings.ToUpper(strings.TrimSpace(body.Status)))
	if err != nil {
		aspirationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Status aspirasi diperbarui", item))
}

func aspirationID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		aspirationError(c, service.ErrAspirationInvalid)
		return 0, false
	}
	return id, true
}

func aspirationDateFilters(c *gin.Context) (*time.Time, *time.Time, bool) {
	parse := func(raw string) (*time.Time, bool) {
		if raw == "" {
			return nil, true
		}
		value, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return nil, false
		}
		return &value, true
	}
	from, okFrom := parse(c.Query("date_from"))
	to, okTo := parse(c.Query("date_to"))
	if from != nil && to != nil && to.Before(*from) {
		return nil, nil, false
	}
	return from, to, okFrom && okTo
}

func isJSONRequest(c *gin.Context) bool {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	return err == nil && mediaType == "application/json"
}

func decodeStrictJSON(c *gin.Context, target any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err == nil {
		return errors.New("multiple JSON values")
	} else if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func aspirationError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan server"
	switch {
	case errors.Is(err, service.ErrAspirationInvalid):
		status, code, message = http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed"
	case errors.Is(err, service.ErrAspirationForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", "Akses ditolak"
	case errors.Is(err, service.ErrAspirationNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan"
	}
	c.JSON(status, FailureCode(code, message, nil))
}
