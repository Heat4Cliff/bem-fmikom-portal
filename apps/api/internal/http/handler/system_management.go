package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

type SystemManagement struct {
	service   *service.SystemManagementService
	dashboard *service.DashboardService
}

func NewSystemManagement(s *service.SystemManagementService) *SystemManagement {
	return &SystemManagement{service: s}
}

func NewSystemManagementWithDashboard(s *service.SystemManagementService, dashboard *service.DashboardService) *SystemManagement {
	return &SystemManagement{service: s, dashboard: dashboard}
}

func (h *SystemManagement) Register(engine *gin.Engine, auth middleware.CurrentUserLoader) {
	public := engine.Group("/api/v1/public")
	public.GET("/settings", h.publicSettings)
	admin := engine.Group("/api/v1/admin")
	admin.Use(middleware.RequireAuth(auth))
	admin.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		c.Next()
	})
	admin.GET("/dashboard", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH, domain.RoleKepala), h.getDashboard)
	admin.GET("/settings", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH), h.getSettings)
	admin.PATCH("/settings", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH), h.patchSettings)
	admin.GET("/users", middleware.RequireRoles(domain.RoleSuperAdmin), h.listUsers)
	admin.POST("/users", middleware.RequireRoles(domain.RoleSuperAdmin), h.createUser)
	admin.GET("/users/:id", middleware.RequireRoles(domain.RoleSuperAdmin), h.getUser)
	admin.PATCH("/users/:id", middleware.RequireRoles(domain.RoleSuperAdmin), h.updateUser)
	admin.POST("/users/:id/deactivate", middleware.RequireRoles(domain.RoleSuperAdmin), h.deactivateUser)
	admin.POST("/users/:id/reactivate", middleware.RequireRoles(domain.RoleSuperAdmin), h.reactivateUser)
	admin.POST("/users/:id/reset-password", middleware.RequireRoles(domain.RoleSuperAdmin), h.resetPassword)
	admin.GET("/activity-logs", middleware.RequireRoles(domain.RoleSuperAdmin), h.listActivity)
	admin.GET("/activity-logs/:id", middleware.RequireRoles(domain.RoleSuperAdmin), h.getActivity)
}

func (h *SystemManagement) getDashboard(c *gin.Context) {
	actor, ok := middleware.CurrentUser(c)
	if !ok || h.dashboard == nil {
		systemError(c, service.ErrDashboardForbidden)
		return
	}
	value, err := h.dashboard.Summary(c.Request.Context(), actor)
	if err != nil {
		if errors.Is(err, service.ErrDashboardForbidden) {
			c.JSON(http.StatusForbidden, FailureCode("FORBIDDEN", "Akses ditolak", nil))
			return
		}
		systemError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Ringkasan dashboard berhasil diambil", value))
}

func (h *SystemManagement) publicSettings(c *gin.Context) {
	v, e := h.service.GetLandingSettings(c.Request.Context())
	if e != nil {
		systemError(c, e)
		return
	}
	c.Header("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	c.JSON(http.StatusOK, Success("Data berhasil diambil", v))
}
func (h *SystemManagement) getSettings(c *gin.Context) {
	v, e := h.service.GetLandingSettings(c.Request.Context())
	if e != nil {
		systemError(c, e)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.JSON(http.StatusOK, Success("Pengaturan berhasil diambil", v))
}
func (h *SystemManagement) patchSettings(c *gin.Context) {
	var v service.LandingSettings
	if !bindContent(c, &v) {
		return
	}
	actor, _ := middleware.CurrentUser(c)
	saved, e := h.service.SaveLandingSettings(c.Request.Context(), actor, v)
	if e != nil {
		systemError(c, e)
		return
	}
	c.Header("Cache-Control", "private, no-store")
	c.JSON(http.StatusOK, Success("Perubahan Landing Page berhasil disimpan.", saved))
}

func (h *SystemManagement) listUsers(c *gin.Context) {
	f := service.UserFilters{Page: queryPageValue(c, "page", 1), Limit: queryPageValue(c, "limit", 20), Search: c.Query("search"), Role: strings.ToUpper(c.Query("role")), Status: strings.ToUpper(c.Query("status"))}
	if raw := c.Query("division_id"); raw != "" {
		id, e := strconv.ParseInt(raw, 10, 64)
		if e != nil || id < 1 {
			systemError(c, service.ErrSystemInvalid)
			return
		}
		f.DivisionID = &id
	}
	result, e := h.service.ListUsers(c.Request.Context(), f)
	if e != nil {
		systemError(c, e)
		return
	}
	listSuccess(c, "Daftar pengguna berhasil diambil", result)
}
func (h *SystemManagement) getUser(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	v, e := h.service.GetUser(c.Request.Context(), id)
	if e != nil {
		systemError(c, e)
		return
	}
	c.JSON(http.StatusOK, Success("Pengguna berhasil diambil", v))
}
func (h *SystemManagement) createUser(c *gin.Context) {
	var in service.UserInput
	if !bindContent(c, &in) {
		return
	}
	actor, _ := middleware.CurrentUser(c)
	v, e := h.service.CreateUser(c.Request.Context(), actor, in)
	if e != nil {
		systemError(c, e)
		return
	}
	c.JSON(http.StatusCreated, Success("Pengguna berhasil dibuat", v))
}
func (h *SystemManagement) updateUser(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var in service.UserInput
	if !bindContent(c, &in) {
		return
	}
	actor, _ := middleware.CurrentUser(c)
	v, e := h.service.UpdateUser(c.Request.Context(), actor, id, in)
	if e != nil {
		systemError(c, e)
		return
	}
	c.JSON(http.StatusOK, Success("Pengguna berhasil diperbarui", v))
}
func (h *SystemManagement) deactivateUser(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	actor, _ := middleware.CurrentUser(c)
	if e := h.service.DeactivateUser(c.Request.Context(), actor, id); e != nil {
		systemError(c, e)
		return
	}
	c.JSON(http.StatusOK, Success("Pengguna dinonaktifkan", gin.H{"id": id, "status": "INACTIVE"}))
}
func (h *SystemManagement) reactivateUser(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	actor, _ := middleware.CurrentUser(c)
	if e := h.service.ReactivateUser(c.Request.Context(), actor, id); e != nil {
		systemError(c, e)
		return
	}
	c.JSON(http.StatusOK, Success("Pengguna diaktifkan kembali", gin.H{"id": id, "status": "ACTIVE"}))
}
func (h *SystemManagement) resetPassword(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var in struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if !bindContent(c, &in) {
		return
	}
	if in.Password == "" || in.Password != in.ConfirmPassword {
		systemError(c, service.ErrSystemInvalid)
		return
	}
	actor, _ := middleware.CurrentUser(c)
	if e := h.service.ResetPassword(c.Request.Context(), actor, id, in.Password); e != nil {
		systemError(c, e)
		return
	}
	c.JSON(http.StatusOK, Success("Password berhasil direset; semua sesi refresh lama dicabut", gin.H{"id": id, "refresh_sessions_revoked": true}))
}

func (h *SystemManagement) listActivity(c *gin.Context) {
	f := service.ActivityFilters{Page: queryPageValue(c, "page", 1), Limit: queryPageValue(c, "limit", 20), Search: c.Query("search"), ActorUserID: c.Query("actor_user_id"), Action: strings.ToUpper(c.Query("action")), EntityType: strings.ToUpper(c.Query("entity_type")), DateFrom: c.Query("date_from"), DateTo: c.Query("date_to")}
	if f.ActorUserID != "" {
		id, e := strconv.ParseInt(f.ActorUserID, 10, 64)
		if e != nil || id < 1 {
			systemError(c, service.ErrSystemInvalid)
			return
		}
	}
	for _, d := range []string{f.DateFrom, f.DateTo} {
		if d != "" {
			if _, e := time.Parse("2006-01-02", d); e != nil {
				systemError(c, service.ErrSystemInvalid)
				return
			}
		}
	}
	if f.DateFrom != "" && f.DateTo != "" && f.DateFrom > f.DateTo {
		systemError(c, service.ErrSystemInvalid)
		return
	}
	result, e := h.service.ListActivity(c.Request.Context(), f)
	if e != nil {
		systemError(c, e)
		return
	}
	listSuccess(c, "Activity log berhasil diambil", result)
}
func (h *SystemManagement) getActivity(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	v, e := h.service.GetActivity(c.Request.Context(), id)
	if e != nil {
		systemError(c, e)
		return
	}
	c.JSON(http.StatusOK, Success("Detail activity log berhasil diambil", v))
}

func systemID(c *gin.Context) (int64, bool) {
	id, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil || id < 1 {
		systemError(c, service.ErrSystemInvalid)
		return 0, false
	}
	return id, true
}
func queryPageValue(c *gin.Context, key string, fallback int) int {
	raw := c.Query(key)
	if raw == "" {
		return fallback
	}
	n, e := strconv.Atoi(raw)
	if e != nil || n < 1 {
		return fallback
	}
	return n
}
func systemError(c *gin.Context, e error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan server"
	switch {
	case errors.Is(e, service.ErrSystemInvalid):
		status, code, message = http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Data tidak valid"
	case errors.Is(e, service.ErrSystemForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", "Akses ditolak"
	case errors.Is(e, service.ErrSystemNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan"
	case errors.Is(e, service.ErrSystemConflict):
		status, code, message = http.StatusConflict, "CONFLICT", "Operasi bertentangan dengan kebijakan sistem"
	}
	c.JSON(status, FailureCode(code, message, nil))
}
