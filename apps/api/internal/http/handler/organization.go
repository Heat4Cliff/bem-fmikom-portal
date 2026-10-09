package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

type Organization struct{ service *service.OrganizationService }

func NewOrganization(org *service.OrganizationService) *Organization {
	return &Organization{service: org}
}

func (h *Organization) Register(engine *gin.Engine, auth middleware.CurrentUserLoader) {
	public := engine.Group("/api/v1/public")
	public.GET("/periods/active", h.activePeriod)
	public.GET("/divisions", h.publicDivisions)
	public.GET("/organization", h.publicOrganization)

	admin := engine.Group("/api/v1/admin")
	admin.Use(middleware.RequireAuth(auth))
	admin.GET("/periods", h.periods)
	admin.POST("/periods", middleware.RequireRoles(domain.RoleSuperAdmin), h.createPeriod)
	admin.GET("/periods/:id", h.period)
	admin.PATCH("/periods/:id", middleware.RequireRoles(domain.RoleSuperAdmin), h.updatePeriod)
	admin.POST("/periods/:id/activate", middleware.RequireRoles(domain.RoleSuperAdmin), h.activatePeriod)
	admin.POST("/periods/:id/archive", middleware.RequireRoles(domain.RoleSuperAdmin), h.archivePeriod)

	admin.GET("/divisions", h.divisions)
	admin.POST("/divisions", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH), h.createDivision)
	admin.GET("/divisions/:id", h.division)
	admin.PATCH("/divisions/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH), h.updateDivision)
	admin.DELETE("/divisions/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH), h.deleteDivision)

	admin.GET("/members", h.members)
	admin.POST("/members", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH), h.createMember)
	admin.GET("/members/:id", h.member)
	admin.PATCH("/members/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH), h.updateMember)
	admin.DELETE("/members/:id", middleware.RequireRoles(domain.RoleSuperAdmin, domain.RoleBPH), h.deleteMember)
}

func (h *Organization) activePeriod(c *gin.Context) {
	value, err := h.service.ActivePeriod(c.Request.Context())
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}
func (h *Organization) publicDivisions(c *gin.Context) {
	items, err := h.service.PublicDivisions(c.Request.Context())
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", items))
}
func (h *Organization) publicOrganization(c *gin.Context) {
	value, err := h.service.PublicOrganization(c.Request.Context())
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}

func (h *Organization) periods(c *gin.Context) {
	page, limit := queryPage(c)
	result, err := h.service.ListPeriods(c.Request.Context(), page, limit, c.Query("search"), c.Query("status"))
	if err != nil {
		organizationError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Organization) period(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	value, err := h.service.GetPeriod(c.Request.Context(), id)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}
func (h *Organization) createPeriod(c *gin.Context) {
	var in service.PeriodInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SavePeriod(c.Request.Context(), user, 0, in)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Periode berhasil dibuat", value))
}
func (h *Organization) updatePeriod(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	var in service.PeriodInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SavePeriod(c.Request.Context(), user, id, in)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Periode berhasil diperbarui", value))
}
func (h *Organization) activatePeriod(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.ActivatePeriod(c.Request.Context(), user, id); err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Periode berhasil diaktifkan", gin.H{"id": id, "status": "ACTIVE"}))
}
func (h *Organization) archivePeriod(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.ArchivePeriod(c.Request.Context(), user, id); err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Periode berhasil diarsipkan", gin.H{"id": id, "status": "ARCHIVED"}))
}

func organizationFilters(c *gin.Context) (service.OrganizationFilters, bool) {
	page, limit := queryPage(c)
	f := service.OrganizationFilters{Page: page, Limit: limit, Search: c.Query("search"), Status: c.Query("status"), Position: c.Query("position")}
	for key, target := range map[string]**int64{"period_id": &f.PeriodID, "division_id": &f.DivisionID} {
		if raw := c.Query(key); raw != "" {
			value, err := strconv.ParseInt(raw, 10, 64)
			if err != nil || value < 1 {
				organizationError(c, service.ErrOrganizationInvalid)
				return f, false
			}
			*target = &value
		}
	}
	return f, true
}
func (h *Organization) divisions(c *gin.Context) {
	f, ok := organizationFilters(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if user.Role == domain.RoleKepala {
		f.DivisionID = user.DivisionID
	}
	result, err := h.service.ListDivisions(c.Request.Context(), f, &user)
	if err != nil {
		organizationError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Organization) division(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.GetDivision(c.Request.Context(), id, user)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}
func (h *Organization) createDivision(c *gin.Context) {
	var in service.DivisionInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveDivision(c.Request.Context(), user, 0, in)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Divisi berhasil dibuat", value))
}
func (h *Organization) updateDivision(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	var in service.DivisionInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveDivision(c.Request.Context(), user, id, in)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Divisi berhasil diperbarui", value))
}
func (h *Organization) deleteDivision(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.DeleteDivision(c.Request.Context(), user, id); err != nil {
		organizationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Organization) members(c *gin.Context) {
	f, ok := organizationFilters(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	result, err := h.service.ListMembers(c.Request.Context(), f, user)
	if err != nil {
		organizationError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Organization) member(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.GetMember(c.Request.Context(), id, user)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}
func (h *Organization) createMember(c *gin.Context) {
	var in service.MemberInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveMember(c.Request.Context(), user, 0, in)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Pengurus berhasil ditambahkan", value))
}
func (h *Organization) updateMember(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	var in service.MemberInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveMember(c.Request.Context(), user, id, in)
	if err != nil {
		organizationError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Pengurus berhasil diperbarui", value))
}
func (h *Organization) deleteMember(c *gin.Context) {
	id, ok := organizationID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.DeleteMember(c.Request.Context(), user, id); err != nil {
		organizationError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func organizationID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		organizationError(c, service.ErrOrganizationInvalid)
		return 0, false
	}
	return id, true
}
func organizationError(c *gin.Context, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan server"
	switch {
	case errors.Is(err, service.ErrOrganizationInvalid):
		status, code, message = http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed"
	case errors.Is(err, service.ErrOrganizationForbidden):
		status, code, message = http.StatusForbidden, "FORBIDDEN", "Akses ditolak"
	case errors.Is(err, service.ErrOrganizationNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "Data tidak ditemukan"
	case errors.Is(err, service.ErrOrganizationConflict):
		status, code, message = http.StatusConflict, "CONFLICT", "Data bertentangan dengan resource yang sudah ada"
	}
	c.JSON(status, FailureCode(code, message, nil))
}
