package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/http/middleware"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/gin-gonic/gin"
)

type Content struct{ service *service.ContentService }

func NewContent(content *service.ContentService) *Content { return &Content{service: content} }

func (h *Content) Register(engine *gin.Engine, auth middleware.CurrentUserLoader) {
	public := engine.Group("/api/v1/public")
	public.GET("/articles", h.publicArticles)
	public.GET("/articles/:slug", h.publicArticle)
	public.GET("/gallery-categories", h.publicGalleryCategories)
	public.GET("/gallery-categories/:slug", h.publicGalleryCategory)
	public.GET("/announcements", h.publicAnnouncements)
	public.GET("/announcements/active", h.activeAnnouncements)
	public.GET("/announcements/:slug", h.publicAnnouncement)

	admin := engine.Group("/api/v1/admin")
	admin.Use(middleware.RequireAuth(auth))
	admin.GET("/article-categories", h.articleCategories)
	admin.POST("/article-categories", h.createArticleCategory)
	admin.PATCH("/article-categories/:id", h.updateArticleCategory)
	admin.DELETE("/article-categories/:id", h.deleteArticleCategory)
	admin.GET("/articles", h.adminArticles)
	admin.POST("/articles", h.createArticle)
	admin.GET("/articles/:id", h.adminArticle)
	admin.PATCH("/articles/:id", h.updateArticle)
	admin.DELETE("/articles/:id", h.deleteArticle)
	admin.POST("/articles/:id/publish", h.publishArticle)
	admin.POST("/articles/:id/unpublish", h.unpublishArticle)
	admin.GET("/gallery-categories", h.adminGalleryCategories)
	admin.POST("/gallery-categories", h.createGalleryCategory)
	admin.PATCH("/gallery-categories/:id", h.updateGalleryCategory)
	admin.DELETE("/gallery-categories/:id", h.deleteGalleryCategory)
	admin.GET("/gallery-items", h.adminGalleryItems)
	admin.POST("/gallery-items", h.createGalleryItem)
	admin.PATCH("/gallery-items/:id", h.updateGalleryItem)
	admin.DELETE("/gallery-items/:id", h.deleteGalleryItem)
	admin.GET("/announcements", h.adminAnnouncements)
	admin.POST("/announcements", h.createAnnouncement)
	admin.GET("/announcements/:id", h.adminAnnouncement)
	admin.PATCH("/announcements/:id", h.updateAnnouncement)
	admin.DELETE("/announcements/:id", h.deleteAnnouncement)
	admin.POST("/announcements/:id/activate", h.activateAnnouncement)
	admin.POST("/announcements/:id/deactivate", h.deactivateAnnouncement)
}

func (h *Content) publicArticles(c *gin.Context) {
	page, limit := queryPage(c)
	result, err := h.service.ListArticles(c.Request.Context(), nil, true, page, limit, c.Query("search"), c.Query("category"), "", nil)
	if err != nil {
		contentError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Content) publicArticle(c *gin.Context) {
	value, err := h.service.GetArticle(c.Request.Context(), nil, c.Param("slug"), true)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}
func (h *Content) publicGalleryCategories(c *gin.Context) {
	page, limit := queryPage(c)
	result, err := h.service.ListGalleryCategories(c.Request.Context(), nil, page, limit, c.Query("search"))
	if err != nil {
		contentError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Content) publicGalleryCategory(c *gin.Context) {
	value, err := h.service.GetGalleryCategory(c.Request.Context(), nil, c.Param("slug"))
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}
func (h *Content) publicAnnouncements(c *gin.Context) {
	page, limit := queryPage(c)
	from, to, valid := announcementDateFilters(c)
	if !valid {
		contentError(c, service.ErrContentInvalid)
		return
	}
	result, err := h.service.ListAnnouncements(c.Request.Context(), true, false, page, limit, c.Query("search"), nil, from, to)
	if err != nil {
		contentError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Content) activeAnnouncements(c *gin.Context) {
	page, limit := queryPage(c)
	result, err := h.service.ListAnnouncements(c.Request.Context(), true, true, page, limit, "", nil, nil, nil)
	if err != nil {
		contentError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Content) publicAnnouncement(c *gin.Context) {
	value, err := h.service.GetAnnouncement(c.Request.Context(), c.Param("slug"), true)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}

func (h *Content) articleCategories(c *gin.Context) {
	items, err := h.service.ListArticleCategories(c.Request.Context())
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", items))
}
func (h *Content) createArticleCategory(c *gin.Context) {
	var in service.CategoryInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	v, e := h.service.SaveArticleCategory(c.Request.Context(), user, 0, in)
	if e != nil {
		contentError(c, e)
		return
	}
	c.JSON(http.StatusCreated, Success("Kategori berhasil dibuat", v))
}
func (h *Content) updateArticleCategory(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	var in service.CategoryInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	v, e := h.service.SaveArticleCategory(c.Request.Context(), user, id, in)
	if e != nil {
		contentError(c, e)
		return
	}
	c.JSON(http.StatusOK, Success("Kategori berhasil diperbarui", v))
}
func (h *Content) deleteArticleCategory(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if e := h.service.DeleteArticleCategory(c.Request.Context(), user, id); e != nil {
		contentError(c, e)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Content) adminArticles(c *gin.Context) {
	page, limit := queryPage(c)
	user, _ := middleware.CurrentUser(c)
	var divisionID *int64
	if raw := c.Query("division_id"); raw != "" {
		parsed, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || parsed < 1 {
			contentError(c, service.ErrContentInvalid)
			return
		}
		divisionID = &parsed
	}
	result, err := h.service.ListArticles(c.Request.Context(), &user, false, page, limit, c.Query("search"), c.Query("category"), c.Query("status"), divisionID)
	if err != nil {
		contentError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Content) createArticle(c *gin.Context) {
	var in service.ArticleInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveArticle(c.Request.Context(), user, 0, in)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Berita berhasil dibuat", value))
}
func (h *Content) adminArticle(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.GetArticle(c.Request.Context(), &user, fmt.Sprint(id), false)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}
func (h *Content) updateArticle(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	var in service.ArticleInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveArticle(c.Request.Context(), user, id, in)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Berita berhasil diperbarui", value))
}
func (h *Content) deleteArticle(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.DeleteArticle(c.Request.Context(), user, id); err != nil {
		contentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Content) publishArticle(c *gin.Context)   { h.changeArticle(c, true) }
func (h *Content) unpublishArticle(c *gin.Context) { h.changeArticle(c, false) }
func (h *Content) changeArticle(c *gin.Context, publish bool) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.ChangeArticlePublication(c.Request.Context(), user, id, publish)
	if err != nil {
		contentError(c, err)
		return
	}
	message := "Berita dikembalikan ke draft"
	if publish {
		message = "Berita berhasil dipublikasikan"
	}
	c.JSON(http.StatusOK, Success(message, value))
}

func (h *Content) adminGalleryCategories(c *gin.Context) {
	page, limit := queryPage(c)
	user, _ := middleware.CurrentUser(c)
	result, err := h.service.ListGalleryCategories(c.Request.Context(), &user, page, limit, c.Query("search"))
	if err != nil {
		contentError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Content) createGalleryCategory(c *gin.Context) {
	var in service.GalleryCategoryInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveGalleryCategory(c.Request.Context(), user, 0, in)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Album berhasil dibuat", value))
}
func (h *Content) updateGalleryCategory(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	var in service.GalleryCategoryInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveGalleryCategory(c.Request.Context(), user, id, in)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Album berhasil diperbarui", value))
}
func (h *Content) deleteGalleryCategory(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.DeleteGalleryCategory(c.Request.Context(), user, id); err != nil {
		contentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Content) adminGalleryItems(c *gin.Context) {
	page, limit := queryPage(c)
	user, _ := middleware.CurrentUser(c)
	categoryID := int64(0)
	if v := c.Query("category_id"); v != "" {
		parsed, e := strconv.ParseInt(v, 10, 64)
		if e != nil || parsed < 1 {
			contentError(c, service.ErrContentInvalid)
			return
		}
		categoryID = parsed
	}
	result, err := h.service.ListGalleryItems(c.Request.Context(), user, page, limit, categoryID)
	if err != nil {
		contentError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Content) createGalleryItem(c *gin.Context) {
	var in service.GalleryItemInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveGalleryItem(c.Request.Context(), user, 0, in)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Item galeri berhasil dibuat", value))
}
func (h *Content) updateGalleryItem(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	var in service.GalleryItemInput
	if !bindContent(c, &in) {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveGalleryItem(c.Request.Context(), user, id, in)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Item galeri berhasil diperbarui", value))
}
func (h *Content) deleteGalleryItem(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.DeleteGalleryItem(c.Request.Context(), user, id); err != nil {
		contentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Content) adminAnnouncements(c *gin.Context) {
	page, limit := queryPage(c)
	from, to, valid := announcementDateFilters(c)
	if !valid {
		contentError(c, service.ErrContentInvalid)
		return
	}
	var active *bool
	if raw := c.Query("active"); raw != "" {
		parsed, parseErr := strconv.ParseBool(raw)
		if parseErr != nil {
			contentError(c, service.ErrContentInvalid)
			return
		}
		active = &parsed
	}
	result, err := h.service.ListAnnouncements(c.Request.Context(), false, false, page, limit, c.Query("search"), active, from, to)
	if err != nil {
		contentError(c, err)
		return
	}
	listSuccess(c, "Data berhasil diambil", result)
}
func (h *Content) createAnnouncement(c *gin.Context) {
	in, ok := bindAnnouncement(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveAnnouncement(c.Request.Context(), user, 0, in)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, Success("Pengumuman berhasil dibuat", value))
}
func (h *Content) adminAnnouncement(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	value, err := h.service.GetAnnouncement(c.Request.Context(), fmt.Sprint(id), false)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Data berhasil diambil", value))
}
func (h *Content) updateAnnouncement(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	in, ok := bindAnnouncement(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.SaveAnnouncement(c.Request.Context(), user, id, in)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Pengumuman berhasil diperbarui", value))
}

func bindAnnouncement(c *gin.Context) (service.AnnouncementInput, bool) {
	var raw struct {
		Title       string          `json:"title"`
		Slug        string          `json:"slug"`
		ContentHTML string          `json:"content_html"`
		ImageURL    *string         `json:"image_url"`
		StartAt     json.RawMessage `json:"start_at"`
		EndAt       json.RawMessage `json:"end_at"`
	}
	if !bindContent(c, &raw) {
		return service.AnnouncementInput{}, false
	}
	parse := func(name string, value json.RawMessage, optional bool) (*time.Time, bool) {
		if len(value) == 0 || string(value) == "null" {
			if optional {
				return nil, true
			}
			c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Validation failed", map[string][]string{name: {"Tanggal harus valid."}}))
			return nil, false
		}
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Validation failed", map[string][]string{name: {"Tanggal harus menggunakan format ISO 8601."}}))
			return nil, false
		}
		parsed, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Validation failed", map[string][]string{name: {"Tanggal harus menggunakan format ISO 8601."}}))
			return nil, false
		}
		return &parsed, true
	}
	start, valid := parse("start_at", raw.StartAt, false)
	if !valid {
		return service.AnnouncementInput{}, false
	}
	end, valid := parse("end_at", raw.EndAt, true)
	if !valid {
		return service.AnnouncementInput{}, false
	}
	return service.AnnouncementInput{Title: raw.Title, Slug: raw.Slug, ContentHTML: raw.ContentHTML, ImageURL: raw.ImageURL, StartAt: *start, EndAt: end}, true
}

func (h *Content) deleteAnnouncement(c *gin.Context) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	if err := h.service.DeleteAnnouncement(c.Request.Context(), user, id); err != nil {
		contentError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Content) activateAnnouncement(c *gin.Context)   { h.changeAnnouncement(c, true) }
func (h *Content) deactivateAnnouncement(c *gin.Context) { h.changeAnnouncement(c, false) }
func (h *Content) changeAnnouncement(c *gin.Context, active bool) {
	id, ok := contentID(c)
	if !ok {
		return
	}
	user, _ := middleware.CurrentUser(c)
	value, err := h.service.ChangeAnnouncementActive(c.Request.Context(), user, id, active)
	if err != nil {
		contentError(c, err)
		return
	}
	c.JSON(http.StatusOK, Success("Status pengumuman diperbarui", value))
}

func queryPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	return page, limit
}

func announcementDateFilters(c *gin.Context) (*time.Time, *time.Time, bool) {
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
	return from, to, okFrom && okTo
}
func contentID(c *gin.Context) (int64, bool) {
	value, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || value < 1 {
		contentError(c, service.ErrContentInvalid)
		return 0, false
	}
	return value, true
}
func bindContent(c *gin.Context, dst any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		c.JSON(http.StatusUnprocessableEntity, Failure("Validation failed", map[string][]string{"body": {"Request body tidak valid"}}))
		return false
	}
	return true
}
func listSuccess[T any](c *gin.Context, message string, result service.Result[T]) {
	c.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: result.Items, Meta: &Meta{Page: result.Page.Page, Limit: result.Page.Limit, Total: result.Page.Total, TotalPages: result.Page.TotalPages}})
}
func contentError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "INTERNAL_ERROR"
	message := "Terjadi kesalahan server"
	var validation *service.ContentValidationError
	if errors.As(err, &validation) {
		c.JSON(http.StatusUnprocessableEntity, FailureCode("VALIDATION_ERROR", "Validation failed", validation.Fields))
		return
	}
	switch {
	case errors.Is(err, service.ErrContentInvalid):
		status = http.StatusUnprocessableEntity
		code = "VALIDATION_ERROR"
		message = "Validation failed"
	case errors.Is(err, service.ErrContentForbidden):
		status = http.StatusForbidden
		code = "FORBIDDEN"
		message = "Akses ditolak"
	case errors.Is(err, service.ErrContentNotFound):
		status = http.StatusNotFound
		code = "NOT_FOUND"
		message = "Data tidak ditemukan"
	case errors.Is(err, service.ErrContentConflict):
		status = http.StatusConflict
		code = "CONFLICT"
		message = "Data bertentangan dengan resource yang sudah ada"
	}
	c.JSON(status, FailureCode(code, message, map[string][]string(nil)))
}
