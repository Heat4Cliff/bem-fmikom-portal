package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/security"
)

var (
	ErrContentNotFound  = errors.New("content not found")
	ErrContentForbidden = errors.New("content access forbidden")
	ErrContentConflict  = errors.New("content conflict")
	ErrContentInvalid   = errors.New("content validation failed")
)

type ContentService struct {
	db  *sql.DB
	log *repository.ActivityLogs
}

func NewContentService(db *sql.DB) *ContentService {
	return &ContentService{db: db, log: repository.NewActivityLogs(db)}
}

type Page struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
type Result[T any] struct {
	Items []T
	Page  Page
}

type ArticleCategory struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type Article struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Slug         string            `json:"slug"`
	Excerpt      *string           `json:"excerpt,omitempty"`
	ContentHTML  string            `json:"content_html"`
	ThumbnailURL *string           `json:"thumbnail_url,omitempty"`
	DivisionID   *int64            `json:"division_id"`
	DivisionName *string           `json:"division_name,omitempty"`
	AuthorID     int64             `json:"author_id"`
	Status       string            `json:"status"`
	PublishedAt  *time.Time        `json:"published_at"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Categories   []ArticleCategory `json:"categories"`
}
type GalleryCategory struct {
	ID           int64         `json:"id"`
	Title        string        `json:"title"`
	Slug         string        `json:"slug"`
	Description  *string       `json:"description,omitempty"`
	PeriodID     *int64        `json:"period_id"`
	DivisionID   *int64        `json:"division_id"`
	DivisionName *string       `json:"division_name,omitempty"`
	EventDate    *string       `json:"event_date,omitempty"`
	ItemCount    int           `json:"item_count,omitempty"`
	Items        []GalleryItem `json:"items,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
type GalleryItem struct {
	ID              int64     `json:"id"`
	CategoryID      int64     `json:"category_id"`
	Title           string    `json:"title"`
	Description     *string   `json:"description,omitempty"`
	GooglePhotosURL string    `json:"google_photos_url"`
	PreviewImageURL *string   `json:"preview_image_url,omitempty"`
	SortOrder       int       `json:"sort_order"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type Announcement struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	ContentHTML string     `json:"content_html"`
	ImageURL    *string    `json:"image_url,omitempty"`
	StartAt     time.Time  `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	IsActive    bool       `json:"is_active"`
	Status      string     `json:"status"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CategoryInput struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
}
type ArticleInput struct {
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	Excerpt      *string `json:"excerpt"`
	ContentHTML  string  `json:"content_html"`
	ThumbnailURL *string `json:"thumbnail_url"`
	DivisionID   *int64  `json:"division_id"`
	CategoryIDs  []int64 `json:"category_ids"`
}
type GalleryCategoryInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Description *string `json:"description"`
	PeriodID    *int64  `json:"period_id"`
	DivisionID  *int64  `json:"division_id"`
	EventDate   *string `json:"event_date"`
}
type GalleryItemInput struct {
	CategoryID      int64   `json:"category_id"`
	Title           string  `json:"title"`
	Description     *string `json:"description"`
	GooglePhotosURL string  `json:"google_photos_url"`
	PreviewImageURL *string `json:"preview_image_url"`
	SortOrder       int     `json:"sort_order"`
}
type AnnouncementInput struct {
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	ContentHTML string     `json:"content_html"`
	ImageURL    *string    `json:"image_url"`
	StartAt     time.Time  `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var internalPublicMediaPath = regexp.MustCompile(`^/api/v1/public/media/([1-9][0-9]*)/?$`)

type ContentValidationError struct {
	Fields map[string][]string
}

func (e *ContentValidationError) Error() string { return ErrContentInvalid.Error() }
func (e *ContentValidationError) Unwrap() error { return ErrContentInvalid }

func invalidContentField(field, message string) error {
	return &ContentValidationError{Fields: map[string][]string{field: {message}}}
}

func normalizeSlug(value, title string) (string, error) {
	if strings.TrimSpace(value) == "" {
		value = slugify(title)
	}
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) > 240 || !slugPattern.MatchString(value) {
		return "", ErrContentInvalid
	}
	return value, nil
}

func slugify(value string) string {
	var out strings.Builder
	space := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && out.Len() > 0 {
				out.WriteByte('-')
			}
			out.WriteRune(r)
			space = false
		} else {
			space = true
		}
	}
	return strings.Trim(out.String(), "-")
}

func normalizeRichText(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", ErrContentInvalid
	}
	clean, err := security.SanitizeRichText(value)
	if err != nil {
		return "", ErrContentInvalid
	}
	if strings.TrimSpace(clean) == "" {
		return "", ErrContentInvalid
	}
	return clean, nil
}

func validateOptionalURL(value *string, required bool) error {
	if value == nil || strings.TrimSpace(*value) == "" {
		if required {
			return ErrContentInvalid
		}
		return nil
	}
	u, err := url.ParseRequestURI(strings.TrimSpace(*value))
	if err != nil || strings.ToLower(u.Scheme) != "https" || u.Host == "" || u.User != nil {
		return ErrContentInvalid
	}
	return nil
}

func validateHTTPS(value string) error {
	u, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil || strings.ToLower(u.Scheme) != "https" || u.Host == "" || u.User != nil {
		return ErrContentInvalid
	}
	return nil
}

// normalizeImageReference accepts optional legacy HTTPS images and only those
// internal media URLs that resolve to a persisted asset for the requested use.
// Internal absolute URLs are normalized to the stable API-relative path.
func (s *ContentService) normalizeImageReference(ctx context.Context, value *string, purpose MediaPurpose, field string) (*string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	raw := strings.TrimSpace(*value)
	var parsed *url.URL
	path := ""
	if strings.HasPrefix(strings.ToLower(raw), "media:") {
		id := raw[len("media:"):]
		if id == "" || strings.Trim(id, "0123456789") != "" || strings.HasPrefix(id, "0") {
			return nil, invalidContentField(field, "Format gambar tidak valid.")
		}
		path = "/api/v1/public/media/" + id
	} else {
		var err error
		parsed, err = url.ParseRequestURI(raw)
		if err != nil {
			return nil, invalidContentField(field, "Format gambar tidak valid.")
		}
		path = parsed.Path
		if parsed.IsAbs() {
			if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil {
				return nil, invalidContentField(field, "Format gambar tidak valid.")
			}
		} else if !strings.HasPrefix(raw, "/") {
			return nil, invalidContentField(field, "Format gambar tidak valid.")
		}
	}
	if match := internalPublicMediaPath.FindStringSubmatch(path); match != nil {
		if parsed != nil && (parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "") {
			return nil, invalidContentField(field, "Format gambar tidak valid.")
		}
		id, parseErr := strconv.ParseInt(match[1], 10, 64)
		if parseErr != nil || id < 1 {
			return nil, invalidContentField(field, "Format gambar tidak valid.")
		}
		var storedPurpose string
		if err := s.db.QueryRowContext(ctx, `SELECT purpose FROM media_assets WHERE id=?`, id).Scan(&storedPurpose); err != nil || storedPurpose != string(purpose) {
			return nil, invalidContentField(field, "Media gambar tidak ditemukan atau tidak sesuai untuk field ini.")
		}
		normalized := "/api/v1/public/media/" + strconv.FormatInt(id, 10)
		return &normalized, nil
	}
	if parsed == nil || !parsed.IsAbs() || !strings.EqualFold(parsed.Scheme, "https") || parsed.Hostname() == "" || parsed.User != nil {
		return nil, invalidContentField(field, "Format gambar tidak valid.")
	}
	return &raw, nil
}

func validateTitle(title string) error {
	if n := len(strings.TrimSpace(title)); n == 0 || n > 240 {
		return ErrContentInvalid
	}
	return nil
}
func optionalString(value *string, max int) error {
	if value != nil && len(strings.TrimSpace(*value)) > max {
		return ErrContentInvalid
	}
	return nil
}

func cloneContentID(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func contentPage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
func pageMeta(page, limit, total int) Page {
	pages := 0
	if total > 0 {
		pages = (total + limit - 1) / limit
	}
	return Page{Page: page, Limit: limit, Total: total, TotalPages: pages}
}

func (s *ContentService) canManageSharedCategories(user domain.User) bool {
	return user.Role == domain.RoleSuperAdmin || user.Role == domain.RoleBPH
}
func (s *ContentService) divisionScope(user domain.User) (string, []any) {
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil {
			return " AND 1=0", nil
		}
		return " AND division_id = ?", []any{*user.DivisionID}
	}
	return "", nil
}

func (s *ContentService) ListArticleCategories(ctx context.Context) ([]ArticleCategory, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,slug,description,created_at,updated_at FROM article_categories WHERE deleted_at IS NULL ORDER BY name,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ArticleCategory{}
	for rows.Next() {
		var item ArticleCategory
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *ContentService) SaveArticleCategory(ctx context.Context, user domain.User, id int64, input CategoryInput) (ArticleCategory, error) {
	if !s.canManageSharedCategories(user) {
		return ArticleCategory{}, ErrContentForbidden
	}
	input.Name = strings.TrimSpace(input.Name)
	if len(input.Name) == 0 || len(input.Name) > 120 || optionalString(input.Description, 2048) != nil {
		return ArticleCategory{}, ErrContentInvalid
	}
	slug, err := normalizeSlug(input.Slug, input.Name)
	if err != nil {
		return ArticleCategory{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ArticleCategory{}, err
	}
	defer tx.Rollback()
	created := id == 0
	var item ArticleCategory
	if id == 0 {
		res, e := tx.ExecContext(ctx, `INSERT INTO article_categories(name,slug,description) VALUES(?,?,?)`, input.Name, slug, input.Description)
		if e != nil {
			return item, sqlContentError(e)
		}
		item.ID, _ = res.LastInsertId()
		item.Name = input.Name
		item.Slug = slug
		item.Description = input.Description
	} else {
		res, e := tx.ExecContext(ctx, `UPDATE article_categories SET name=?,slug=?,description=? WHERE id=? AND deleted_at IS NULL`, input.Name, slug, input.Description, id)
		if e != nil {
			return item, sqlContentError(e)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return item, ErrContentNotFound
		}
		item.ID = id
		item.Name = input.Name
		item.Slug = slug
		item.Description = input.Description
	}
	if err := s.activity(ctx, tx, user, actionForCreated(created, "UPDATE"), "ARTICLE_CATEGORY", item.ID, map[string]any{"name": item.Name}); err != nil {
		return item, err
	}
	if err := tx.Commit(); err != nil {
		return item, err
	}
	return s.articleCategory(ctx, item.ID)
}

func actionForCreated(created bool, update string) string {
	if created {
		return "CREATE"
	}
	return update
}

func (s *ContentService) articleCategory(ctx context.Context, id int64) (ArticleCategory, error) {
	var v ArticleCategory
	err := s.db.QueryRowContext(ctx, `SELECT id,name,slug,description,created_at,updated_at FROM article_categories WHERE id=? AND deleted_at IS NULL`, id).Scan(&v.ID, &v.Name, &v.Slug, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	return v, contentNotFound(err)
}

func (s *ContentService) DeleteArticleCategory(ctx context.Context, user domain.User, id int64) error {
	if !s.canManageSharedCategories(user) {
		return ErrContentForbidden
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var count int
	if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM article_category_map m JOIN articles a ON a.id=m.article_id AND a.deleted_at IS NULL WHERE m.category_id=?`, id).Scan(&count); e != nil {
		return e
	}
	if count > 0 {
		return ErrContentConflict
	}
	res, e := tx.ExecContext(ctx, `UPDATE article_categories SET deleted_at=UTC_TIMESTAMP(6) WHERE id=? AND deleted_at IS NULL`, id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrContentNotFound
	}
	if e = s.activity(ctx, tx, user, "DELETE", "ARTICLE_CATEGORY", id, nil); e != nil {
		return e
	}
	return tx.Commit()
}

func (s *ContentService) ListArticles(ctx context.Context, user *domain.User, publicOnly bool, page, limit int, search, category, status string, divisionID *int64) (Result[Article], error) {
	page, limit = contentPage(page, limit)
	where := `a.deleted_at IS NULL`
	args := []any{}
	if publicOnly {
		where += ` AND a.status='PUBLISHED' AND a.published_at <= UTC_TIMESTAMP(6)`
	} else if user != nil && user.Role == domain.RoleKepala {
		if user.DivisionID == nil {
			return Result[Article]{}, ErrContentForbidden
		}
		where += ` AND a.division_id=?`
		args = append(args, *user.DivisionID)
	} else if !publicOnly && divisionID != nil {
		where += ` AND a.division_id=?`
		args = append(args, *divisionID)
	}
	if search != "" {
		where += ` AND (a.title LIKE ? OR a.excerpt LIKE ?)`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	if status != "" && !publicOnly {
		if status != "DRAFT" && status != "PUBLISHED" {
			return Result[Article]{}, ErrContentInvalid
		}
		where += ` AND a.status=?`
		args = append(args, status)
	}
	if category != "" {
		where += ` AND EXISTS (SELECT 1 FROM article_category_map m JOIN article_categories c ON c.id=m.category_id WHERE m.article_id=a.id AND c.slug=? AND c.deleted_at IS NULL)`
		args = append(args, category)
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles a WHERE `+where, args...).Scan(&total); err != nil {
		return Result[Article]{}, err
	}
	queryArgs := append(append([]any{}, args...), limit, (page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,a.title,a.slug,a.excerpt,a.content_html,a.thumbnail_url,a.division_id,d.name,a.author_id,a.status,a.published_at,a.created_at,a.updated_at FROM articles a LEFT JOIN divisions d ON d.id=a.division_id WHERE `+where+` ORDER BY COALESCE(a.published_at,a.created_at) DESC,a.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return Result[Article]{}, err
	}
	defer rows.Close()
	items := []Article{}
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Slug, &a.Excerpt, &a.ContentHTML, &a.ThumbnailURL, &a.DivisionID, &a.DivisionName, &a.AuthorID, &a.Status, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return Result[Article]{}, err
		}
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return Result[Article]{}, err
	}
	for i := range items {
		items[i].Categories, err = s.articleCategories(ctx, items[i].ID)
		if err != nil {
			return Result[Article]{}, err
		}
	}
	return Result[Article]{Items: items, Page: pageMeta(page, limit, total)}, nil
}

func (s *ContentService) articleCategories(ctx context.Context, id int64) ([]ArticleCategory, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT c.id,c.name,c.slug,c.description,c.created_at,c.updated_at FROM article_categories c JOIN article_category_map m ON m.category_id=c.id WHERE m.article_id=? AND c.deleted_at IS NULL ORDER BY c.name`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []ArticleCategory{}
	for rows.Next() {
		var c ArticleCategory
		if e := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *ContentService) GetArticle(ctx context.Context, user *domain.User, idOrSlug string, publicOnly bool) (Article, error) {
	query := `SELECT a.id,a.title,a.slug,a.excerpt,a.content_html,a.thumbnail_url,a.division_id,d.name,a.author_id,a.status,a.published_at,a.created_at,a.updated_at FROM articles a LEFT JOIN divisions d ON d.id=a.division_id WHERE a.deleted_at IS NULL AND `
	args := []any{idOrSlug}
	if publicOnly {
		query += `a.slug=? AND a.status='PUBLISHED' AND a.published_at<=UTC_TIMESTAMP(6)`
	} else {
		query += `a.id=?`
		if user != nil && user.Role == domain.RoleKepala {
			if user.DivisionID == nil {
				return Article{}, ErrContentForbidden
			}
			query += ` AND a.division_id=?`
			args = append(args, *user.DivisionID)
		}
	}
	var a Article
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&a.ID, &a.Title, &a.Slug, &a.Excerpt, &a.ContentHTML, &a.ThumbnailURL, &a.DivisionID, &a.DivisionName, &a.AuthorID, &a.Status, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return a, contentNotFound(err)
	}
	a.Categories, err = s.articleCategories(ctx, a.ID)
	return a, err
}

func (s *ContentService) SaveArticle(ctx context.Context, user domain.User, id int64, input ArticleInput) (Article, error) {
	if err := validateTitle(input.Title); err != nil {
		return Article{}, err
	}
	if optionalString(input.Excerpt, 500) != nil {
		return Article{}, ErrContentInvalid
	}
	thumbnail, err := s.normalizeImageReference(ctx, input.ThumbnailURL, MediaArticleCover, "thumbnail_url")
	if err != nil {
		return Article{}, err
	}
	input.ThumbnailURL = thumbnail
	body, err := normalizeRichText(input.ContentHTML)
	if err != nil {
		return Article{}, err
	}
	slug, err := normalizeSlug(input.Slug, input.Title)
	if err != nil {
		return Article{}, err
	}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil {
			return Article{}, ErrContentForbidden
		}
		input.DivisionID = cloneContentID(user.DivisionID)
	}
	if input.DivisionID != nil {
		var exists int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM divisions WHERE id=?`, *input.DivisionID).Scan(&exists); err != nil {
			return Article{}, err
		}
		if exists == 0 {
			return Article{}, ErrContentInvalid
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Article{}, err
	}
	defer tx.Rollback()
	created := id == 0
	if id == 0 {
		res, e := tx.ExecContext(ctx, `INSERT INTO articles(title,slug,excerpt,content_html,thumbnail_url,division_id,author_id,status) VALUES(?,?,?,?,?,?,?,'DRAFT')`, strings.TrimSpace(input.Title), slug, input.Excerpt, body, input.ThumbnailURL, input.DivisionID, user.ID)
		if e != nil {
			return Article{}, sqlContentError(e)
		}
		id, _ = res.LastInsertId()
	} else {
		if _, e := s.scopedArticleTx(ctx, tx, user, id); e != nil {
			return Article{}, e
		}
		res, e := tx.ExecContext(ctx, `UPDATE articles SET title=?,slug=?,excerpt=?,content_html=?,thumbnail_url=?,division_id=? WHERE id=? AND deleted_at IS NULL`, strings.TrimSpace(input.Title), slug, input.Excerpt, body, input.ThumbnailURL, input.DivisionID, id)
		if e != nil {
			return Article{}, sqlContentError(e)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			var exists int
			if checkErr := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM articles WHERE id=? AND deleted_at IS NULL`, id).Scan(&exists); checkErr != nil {
				return Article{}, checkErr
			}
			if exists == 0 {
				return Article{}, ErrContentNotFound
			}
		}
	}
	categoryIDs := uniqueIDs(input.CategoryIDs)
	if len(categoryIDs) != len(input.CategoryIDs) {
		return Article{}, ErrContentInvalid
	}
	if len(categoryIDs) > 0 {
		placeholders := strings.TrimRight(strings.Repeat("?,", len(categoryIDs)), ",")
		args := make([]any, 0, len(categoryIDs))
		for _, categoryID := range categoryIDs {
			args = append(args, categoryID)
		}
		var validCategories int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM article_categories WHERE deleted_at IS NULL AND id IN (`+placeholders+`)`, args...).Scan(&validCategories); err != nil {
			return Article{}, err
		}
		if validCategories != len(categoryIDs) {
			return Article{}, ErrContentInvalid
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM article_category_map WHERE article_id=?`, id); err != nil {
		return Article{}, err
	}
	for _, cid := range categoryIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO article_category_map(article_id,category_id) SELECT ?,id FROM article_categories WHERE id=? AND deleted_at IS NULL`, id, cid); err != nil {
			return Article{}, sqlContentError(err)
		}
	}
	if err = s.activity(ctx, tx, user, actionForCreated(created, "UPDATE"), "ARTICLE", id, map[string]any{"title": strings.TrimSpace(input.Title)}); err != nil {
		return Article{}, err
	}
	if err = tx.Commit(); err != nil {
		return Article{}, err
	}
	return s.GetArticle(ctx, &user, fmt.Sprint(id), false)
}

func (s *ContentService) scopedArticleTx(ctx context.Context, tx *sql.Tx, user domain.User, id int64) (int64, error) {
	var div sql.NullInt64
	err := tx.QueryRowContext(ctx, `SELECT division_id FROM articles WHERE id=? AND deleted_at IS NULL`, id).Scan(&div)
	if err != nil {
		return 0, contentNotFound(err)
	}
	if user.Role == domain.RoleKepala && (!div.Valid || user.DivisionID == nil || div.Int64 != *user.DivisionID) {
		return 0, ErrContentForbidden
	}
	return id, nil
}
func uniqueIDs(ids []int64) []int64 {
	seen := map[int64]bool{}
	out := []int64{}
	for _, id := range ids {
		if id > 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func (s *ContentService) ChangeArticlePublication(ctx context.Context, user domain.User, id int64, publish bool) (Article, error) {
	if user.Role == domain.RoleKepala && user.DivisionID == nil {
		return Article{}, ErrContentForbidden
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return Article{}, e
	}
	defer tx.Rollback()
	if _, e = s.scopedArticleTx(ctx, tx, user, id); e != nil {
		return Article{}, e
	}
	var action, status string
	var result sql.Result
	if publish {
		action = "PUBLISH"
		status = "PUBLISHED"
		result, e = tx.ExecContext(ctx, `UPDATE articles SET status='PUBLISHED',published_at=COALESCE(published_at,UTC_TIMESTAMP(6)) WHERE id=? AND deleted_at IS NULL AND status='DRAFT'`, id)
	} else {
		action = "UNPUBLISH"
		status = "DRAFT"
		result, e = tx.ExecContext(ctx, `UPDATE articles SET status='DRAFT',published_at=NULL WHERE id=? AND deleted_at IS NULL AND status='PUBLISHED'`, id)
	}
	if e != nil {
		return Article{}, e
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return Article{}, ErrContentConflict
	}
	if e = s.activity(ctx, tx, user, action, "ARTICLE", id, map[string]any{"status": status}); e != nil {
		return Article{}, e
	}
	if e = tx.Commit(); e != nil {
		return Article{}, e
	}
	return s.GetArticle(ctx, &user, fmt.Sprint(id), false)
}

func (s *ContentService) DeleteArticle(ctx context.Context, user domain.User, id int64) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = s.scopedArticleTx(ctx, tx, user, id); e != nil {
		return e
	}
	res, e := tx.ExecContext(ctx, `UPDATE articles SET deleted_at=UTC_TIMESTAMP(6) WHERE id=? AND deleted_at IS NULL`, id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrContentNotFound
	}
	if e = s.activity(ctx, tx, user, "DELETE", "ARTICLE", id, nil); e != nil {
		return e
	}
	return tx.Commit()
}

func (s *ContentService) ListGalleryCategories(ctx context.Context, user *domain.User, page, limit int, search string) (Result[GalleryCategory], error) {
	page, limit = contentPage(page, limit)
	where := `c.deleted_at IS NULL`
	args := []any{}
	if user != nil && user.Role == domain.RoleKepala {
		if user.DivisionID == nil {
			return Result[GalleryCategory]{}, ErrContentForbidden
		}
		where += ` AND c.division_id=?`
		args = append(args, *user.DivisionID)
	}
	if search != "" {
		where += ` AND c.title LIKE ?`
		args = append(args, "%"+search+"%")
	}
	var total int
	if e := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM gallery_categories c WHERE `+where, args...).Scan(&total); e != nil {
		return Result[GalleryCategory]{}, e
	}
	qargs := append(append([]any{}, args...), limit, (page-1)*limit)
	rows, e := s.db.QueryContext(ctx, `SELECT c.id,c.title,c.slug,c.description,c.period_id,c.division_id,d.name,c.event_date,c.created_at,c.updated_at,(SELECT COUNT(*) FROM gallery_items i WHERE i.category_id=c.id AND i.deleted_at IS NULL) FROM gallery_categories c LEFT JOIN divisions d ON d.id=c.division_id WHERE `+where+` ORDER BY COALESCE(c.event_date,'1000-01-01') DESC,c.id DESC LIMIT ? OFFSET ?`, qargs...)
	if e != nil {
		return Result[GalleryCategory]{}, e
	}
	defer rows.Close()
	items := []GalleryCategory{}
	for rows.Next() {
		var v GalleryCategory
		var date sql.NullTime
		if e := rows.Scan(&v.ID, &v.Title, &v.Slug, &v.Description, &v.PeriodID, &v.DivisionID, &v.DivisionName, &date, &v.CreatedAt, &v.UpdatedAt, &v.ItemCount); e != nil {
			return Result[GalleryCategory]{}, e
		}
		if date.Valid {
			d := date.Time.Format("2006-01-02")
			v.EventDate = &d
		}
		items = append(items, v)
	}
	if err := rows.Err(); err != nil {
		return Result[GalleryCategory]{}, err
	}
	if err := rows.Close(); err != nil {
		return Result[GalleryCategory]{}, err
	}
	if user == nil {
		for i := range items {
			items[i].Items, e = s.galleryItems(ctx, items[i].ID)
			if e != nil {
				return Result[GalleryCategory]{}, e
			}
		}
	}
	return Result[GalleryCategory]{Items: items, Page: pageMeta(page, limit, total)}, nil
}

func (s *ContentService) GetGalleryCategory(ctx context.Context, user *domain.User, slug string) (GalleryCategory, error) {
	var v GalleryCategory
	var date sql.NullTime
	e := s.db.QueryRowContext(ctx, `SELECT c.id,c.title,c.slug,c.description,c.period_id,c.division_id,d.name,c.event_date,c.created_at,c.updated_at FROM gallery_categories c LEFT JOIN divisions d ON d.id=c.division_id WHERE c.slug=? AND c.deleted_at IS NULL`, slug).Scan(&v.ID, &v.Title, &v.Slug, &v.Description, &v.PeriodID, &v.DivisionID, &v.DivisionName, &date, &v.CreatedAt, &v.UpdatedAt)
	if e != nil {
		return v, contentNotFound(e)
	}
	if user != nil && user.Role == domain.RoleKepala && (user.DivisionID == nil || v.DivisionID == nil || *user.DivisionID != *v.DivisionID) {
		return v, ErrContentForbidden
	}
	if date.Valid {
		d := date.Time.Format("2006-01-02")
		v.EventDate = &d
	}
	v.Items, e = s.galleryItems(ctx, v.ID)
	return v, e
}
func (s *ContentService) galleryItems(ctx context.Context, categoryID int64) ([]GalleryItem, error) {
	rows, e := s.db.QueryContext(ctx, `SELECT id,category_id,title,description,google_photos_url,preview_image_url,sort_order,created_at,updated_at FROM gallery_items WHERE category_id=? AND deleted_at IS NULL ORDER BY sort_order,id`, categoryID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []GalleryItem{}
	for rows.Next() {
		var v GalleryItem
		if e := rows.Scan(&v.ID, &v.CategoryID, &v.Title, &v.Description, &v.GooglePhotosURL, &v.PreviewImageURL, &v.SortOrder, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *ContentService) SaveGalleryCategory(ctx context.Context, user domain.User, id int64, input GalleryCategoryInput) (GalleryCategory, error) {
	if user.Role == domain.RoleKepala {
		return GalleryCategory{}, ErrContentForbidden
	}
	if validateTitle(input.Title) != nil {
		return GalleryCategory{}, invalidContentField("title", "Judul kategori wajib diisi.")
	}
	if optionalString(input.Description, 2048) != nil {
		return GalleryCategory{}, invalidContentField("description", "Deskripsi terlalu panjang.")
	}
	slug, e := normalizeSlug(input.Slug, input.Title)
	if e != nil {
		return GalleryCategory{}, invalidContentField("slug", "Slug kategori tidak valid.")
	}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil {
			return GalleryCategory{}, ErrContentForbidden
		}
		input.DivisionID = cloneContentID(user.DivisionID)
	}
	if input.EventDate != nil {
		if _, e := time.Parse("2006-01-02", *input.EventDate); e != nil {
			return GalleryCategory{}, invalidContentField("event_date", "Tanggal kegiatan tidak valid.")
		}
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return GalleryCategory{}, e
	}
	defer tx.Rollback()
	created := id == 0
	if id == 0 {
		res, err := tx.ExecContext(ctx, `INSERT INTO gallery_categories(title,slug,description,period_id,division_id,event_date) VALUES(?,?,?,?,?,?)`, strings.TrimSpace(input.Title), slug, input.Description, input.PeriodID, input.DivisionID, input.EventDate)
		if err != nil {
			return GalleryCategory{}, sqlContentError(err)
		}
		id, _ = res.LastInsertId()
	} else {
		if e = s.scopedGalleryCategoryTx(ctx, tx, user, id); e != nil {
			return GalleryCategory{}, e
		}
		_, e = tx.ExecContext(ctx, `UPDATE gallery_categories SET title=?,slug=?,description=?,period_id=?,division_id=?,event_date=? WHERE id=? AND deleted_at IS NULL`, strings.TrimSpace(input.Title), slug, input.Description, input.PeriodID, input.DivisionID, input.EventDate, id)
		if e != nil {
			return GalleryCategory{}, sqlContentError(e)
		}
	}
	if e = s.activity(ctx, tx, user, actionForCreated(created, "UPDATE"), "GALLERY_CATEGORY", id, map[string]any{"title": strings.TrimSpace(input.Title)}); e != nil {
		return GalleryCategory{}, e
	}
	if e = tx.Commit(); e != nil {
		return GalleryCategory{}, e
	}
	return s.GetGalleryCategory(ctx, &user, slug)
}
func (s *ContentService) scopedGalleryCategoryTx(ctx context.Context, tx *sql.Tx, user domain.User, id int64) error {
	var div sql.NullInt64
	e := tx.QueryRowContext(ctx, `SELECT division_id FROM gallery_categories WHERE id=? AND deleted_at IS NULL`, id).Scan(&div)
	if e != nil {
		return contentNotFound(e)
	}
	if user.Role == domain.RoleKepala && (user.DivisionID == nil || !div.Valid || div.Int64 != *user.DivisionID) {
		return ErrContentForbidden
	}
	return nil
}
func (s *ContentService) DeleteGalleryCategory(ctx context.Context, user domain.User, id int64) error {
	if user.Role == domain.RoleKepala {
		return ErrContentForbidden
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if e = s.scopedGalleryCategoryTx(ctx, tx, user, id); e != nil {
		return e
	}
	var count int
	if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM gallery_items WHERE category_id=? AND deleted_at IS NULL`, id).Scan(&count); e != nil {
		return e
	}
	if count > 0 {
		return ErrContentConflict
	}
	res, e := tx.ExecContext(ctx, `UPDATE gallery_categories SET deleted_at=UTC_TIMESTAMP(6) WHERE id=? AND deleted_at IS NULL`, id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrContentNotFound
	}
	if e = s.activity(ctx, tx, user, "DELETE", "GALLERY_CATEGORY", id, nil); e != nil {
		return e
	}
	return tx.Commit()
}

func (s *ContentService) ListGalleryItems(ctx context.Context, user domain.User, page, limit int, categoryID int64) (Result[GalleryItem], error) {
	page, limit = contentPage(page, limit)
	where := `i.deleted_at IS NULL`
	args := []any{}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil {
			return Result[GalleryItem]{}, ErrContentForbidden
		}
		where += ` AND c.division_id=?`
		args = append(args, *user.DivisionID)
	}
	if categoryID > 0 {
		where += ` AND i.category_id=?`
		args = append(args, categoryID)
	}
	var total int
	e := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM gallery_items i JOIN gallery_categories c ON c.id=i.category_id AND c.deleted_at IS NULL WHERE `+where, args...).Scan(&total)
	if e != nil {
		return Result[GalleryItem]{}, e
	}
	qargs := append(append([]any{}, args...), limit, (page-1)*limit)
	rows, e := s.db.QueryContext(ctx, `SELECT i.id,i.category_id,i.title,i.description,i.google_photos_url,i.preview_image_url,i.sort_order,i.created_at,i.updated_at FROM gallery_items i JOIN gallery_categories c ON c.id=i.category_id AND c.deleted_at IS NULL WHERE `+where+` ORDER BY c.event_date DESC,i.sort_order,i.id LIMIT ? OFFSET ?`, qargs...)
	if e != nil {
		return Result[GalleryItem]{}, e
	}
	defer rows.Close()
	out := []GalleryItem{}
	for rows.Next() {
		var v GalleryItem
		if e := rows.Scan(&v.ID, &v.CategoryID, &v.Title, &v.Description, &v.GooglePhotosURL, &v.PreviewImageURL, &v.SortOrder, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return Result[GalleryItem]{}, e
		}
		out = append(out, v)
	}
	return Result[GalleryItem]{Items: out, Page: pageMeta(page, limit, total)}, rows.Err()
}

func (s *ContentService) SaveGalleryItem(ctx context.Context, user domain.User, id int64, input GalleryItemInput) (GalleryItem, error) {
	if validateTitle(input.Title) != nil {
		return GalleryItem{}, invalidContentField("title", "Judul item wajib diisi.")
	}
	if input.CategoryID < 1 {
		return GalleryItem{}, invalidContentField("category_id", "Pilih kategori galeri.")
	}
	if optionalString(input.Description, 2048) != nil {
		return GalleryItem{}, invalidContentField("description", "Deskripsi terlalu panjang.")
	}
	if validateHTTPS(input.GooglePhotosURL) != nil {
		return GalleryItem{}, invalidContentField("google_photos_url", "Tautan album harus berupa URL HTTPS yang valid.")
	}
	preview, err := s.normalizeImageReference(ctx, input.PreviewImageURL, MediaGalleryPreview, "preview_image_url")
	if err != nil {
		return GalleryItem{}, err
	}
	input.PreviewImageURL = preview
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return GalleryItem{}, e
	}
	defer tx.Rollback()
	created := id == 0
	if id > 0 {
		var oldCategory int64
		if e = tx.QueryRowContext(ctx, `SELECT category_id FROM gallery_items WHERE id=? AND deleted_at IS NULL`, id).Scan(&oldCategory); e != nil {
			return GalleryItem{}, contentNotFound(e)
		}
		if e = s.scopedGalleryCategoryTx(ctx, tx, user, oldCategory); e != nil {
			return GalleryItem{}, e
		}
	}
	if e = s.scopedGalleryCategoryTx(ctx, tx, user, input.CategoryID); e != nil {
		return GalleryItem{}, e
	}
	if id == 0 {
		res, err := tx.ExecContext(ctx, `INSERT INTO gallery_items(category_id,title,description,google_photos_url,preview_image_url,sort_order) VALUES(?,?,?,?,?,?)`, input.CategoryID, strings.TrimSpace(input.Title), input.Description, input.GooglePhotosURL, input.PreviewImageURL, input.SortOrder)
		if err != nil {
			return GalleryItem{}, err
		}
		id, _ = res.LastInsertId()
	} else {
		_, e = tx.ExecContext(ctx, `UPDATE gallery_items SET category_id=?,title=?,description=?,google_photos_url=?,preview_image_url=?,sort_order=? WHERE id=? AND deleted_at IS NULL`, input.CategoryID, strings.TrimSpace(input.Title), input.Description, input.GooglePhotosURL, input.PreviewImageURL, input.SortOrder, id)
		if e != nil {
			return GalleryItem{}, e
		}
	}
	if e = s.activity(ctx, tx, user, actionForCreated(created, "UPDATE"), "GALLERY_ITEM", id, map[string]any{"title": strings.TrimSpace(input.Title)}); e != nil {
		return GalleryItem{}, e
	}
	if e = tx.Commit(); e != nil {
		return GalleryItem{}, e
	}
	var v GalleryItem
	e = s.db.QueryRowContext(ctx, `SELECT id,category_id,title,description,google_photos_url,preview_image_url,sort_order,created_at,updated_at FROM gallery_items WHERE id=?`, id).Scan(&v.ID, &v.CategoryID, &v.Title, &v.Description, &v.GooglePhotosURL, &v.PreviewImageURL, &v.SortOrder, &v.CreatedAt, &v.UpdatedAt)
	return v, e
}
func (s *ContentService) DeleteGalleryItem(ctx context.Context, user domain.User, id int64) error {
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var categoryID int64
	if e = tx.QueryRowContext(ctx, `SELECT category_id FROM gallery_items WHERE id=? AND deleted_at IS NULL`, id).Scan(&categoryID); e != nil {
		return contentNotFound(e)
	}
	if e = s.scopedGalleryCategoryTx(ctx, tx, user, categoryID); e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `UPDATE gallery_items SET deleted_at=UTC_TIMESTAMP(6) WHERE id=?`, id)
	if e != nil {
		return e
	}
	if e = s.activity(ctx, tx, user, "DELETE", "GALLERY_ITEM", id, nil); e != nil {
		return e
	}
	return tx.Commit()
}

func (s *ContentService) ListAnnouncements(ctx context.Context, publicOnly bool, activeOnly bool, page, limit int, search string, activeFilter *bool, dateFrom, dateTo *time.Time) (Result[Announcement], error) {
	page, limit = contentPage(page, limit)
	where := `deleted_at IS NULL`
	args := []any{}
	if publicOnly {
		where += ` AND is_active=TRUE`
		if activeOnly {
			where += ` AND start_at<=UTC_TIMESTAMP(6) AND (end_at IS NULL OR end_at>=UTC_TIMESTAMP(6))`
		}
	} else if activeFilter != nil {
		where += ` AND is_active=?`
		args = append(args, *activeFilter)
	}
	if dateFrom != nil {
		where += ` AND start_at>=?`
		args = append(args, dateFrom.UTC())
	}
	if dateTo != nil {
		where += ` AND start_at<?`
		args = append(args, dateTo.UTC().Add(24*time.Hour))
	}
	if search != "" {
		where += ` AND title LIKE ?`
		args = append(args, "%"+search+"%")
	}
	var total int
	e := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM announcements WHERE `+where, args...).Scan(&total)
	if e != nil {
		return Result[Announcement]{}, e
	}
	qargs := append(append([]any{}, args...), limit, (page-1)*limit)
	rows, e := s.db.QueryContext(ctx, `SELECT id,title,slug,content_html,image_url,start_at,end_at,is_active,created_by,created_at,updated_at FROM announcements WHERE `+where+` ORDER BY start_at DESC,id DESC LIMIT ? OFFSET ?`, qargs...)
	if e != nil {
		return Result[Announcement]{}, e
	}
	defer rows.Close()
	items := []Announcement{}
	now := time.Now().UTC()
	for rows.Next() {
		var a Announcement
		if e := rows.Scan(&a.ID, &a.Title, &a.Slug, &a.ContentHTML, &a.ImageURL, &a.StartAt, &a.EndAt, &a.IsActive, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); e != nil {
			return Result[Announcement]{}, e
		}
		a.Status = announcementStatus(a, now)
		items = append(items, a)
	}
	return Result[Announcement]{Items: items, Page: pageMeta(page, limit, total)}, rows.Err()
}
func announcementStatus(a Announcement, now time.Time) string {
	if !a.IsActive {
		return "INACTIVE"
	}
	if now.Before(a.StartAt) {
		return "SCHEDULED"
	}
	if a.EndAt != nil && now.After(*a.EndAt) {
		return "EXPIRED"
	}
	return "ACTIVE"
}
func (s *ContentService) GetAnnouncement(ctx context.Context, idOrSlug string, public bool) (Announcement, error) {
	query := `SELECT id,title,slug,content_html,image_url,start_at,end_at,is_active,created_by,created_at,updated_at FROM announcements WHERE deleted_at IS NULL AND `
	if public {
		query += `slug=? AND is_active=TRUE`
	} else {
		query += `id=?`
	}
	var a Announcement
	e := s.db.QueryRowContext(ctx, query, idOrSlug).Scan(&a.ID, &a.Title, &a.Slug, &a.ContentHTML, &a.ImageURL, &a.StartAt, &a.EndAt, &a.IsActive, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if e != nil {
		return a, contentNotFound(e)
	}
	a.Status = announcementStatus(a, time.Now().UTC())
	return a, nil
}
func (s *ContentService) SaveAnnouncement(ctx context.Context, user domain.User, id int64, input AnnouncementInput) (Announcement, error) {
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH {
		return Announcement{}, ErrContentForbidden
	}
	if validateTitle(input.Title) != nil {
		return Announcement{}, invalidContentField("title", "Judul wajib diisi.")
	}
	body, e := normalizeRichText(input.ContentHTML)
	if e != nil {
		return Announcement{}, invalidContentField("content_html", "Isi pengumuman wajib diisi.")
	}
	slug, e := normalizeSlug(input.Slug, input.Title)
	if e != nil {
		return Announcement{}, invalidContentField("slug", "Slug pengumuman tidak valid.")
	}
	image, imageErr := s.normalizeImageReference(ctx, input.ImageURL, MediaAnnouncementImage, "image_url")
	if imageErr != nil {
		return Announcement{}, imageErr
	}
	input.ImageURL = image
	if input.StartAt.IsZero() {
		return Announcement{}, invalidContentField("start_at", "Tanggal mulai wajib diisi dan harus valid.")
	}
	if input.EndAt != nil && input.EndAt.Before(input.StartAt) {
		return Announcement{}, invalidContentField("end_at", "Tanggal selesai harus setelah tanggal mulai.")
	}
	input.StartAt = input.StartAt.UTC()
	if input.EndAt != nil {
		end := input.EndAt.UTC()
		input.EndAt = &end
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return Announcement{}, e
	}
	defer tx.Rollback()
	created := id == 0
	if id == 0 {
		res, err := tx.ExecContext(ctx, `INSERT INTO announcements(title,slug,content_html,image_url,start_at,end_at,is_active,created_by) VALUES(?,?,?,?,?,?,FALSE,?)`, strings.TrimSpace(input.Title), slug, body, input.ImageURL, input.StartAt, input.EndAt, user.ID)
		if err != nil {
			return Announcement{}, sqlContentError(err)
		}
		id, _ = res.LastInsertId()
	} else {
		if _, e = s.announcementTx(ctx, tx, id); e != nil {
			return Announcement{}, e
		}
		_, e = tx.ExecContext(ctx, `UPDATE announcements SET title=?,slug=?,content_html=?,image_url=?,start_at=?,end_at=? WHERE id=? AND deleted_at IS NULL`, strings.TrimSpace(input.Title), slug, body, input.ImageURL, input.StartAt, input.EndAt, id)
		if e != nil {
			return Announcement{}, sqlContentError(e)
		}
	}
	if e = s.activity(ctx, tx, user, actionForCreated(created, "UPDATE"), "ANNOUNCEMENT", id, map[string]any{"title": strings.TrimSpace(input.Title)}); e != nil {
		return Announcement{}, e
	}
	if e = tx.Commit(); e != nil {
		return Announcement{}, e
	}
	return s.GetAnnouncement(ctx, fmt.Sprint(id), false)
}
func (s *ContentService) announcementTx(ctx context.Context, tx *sql.Tx, id int64) (Announcement, error) {
	var a Announcement
	e := tx.QueryRowContext(ctx, `SELECT id,title,slug,content_html,image_url,start_at,end_at,is_active,created_by,created_at,updated_at FROM announcements WHERE id=? AND deleted_at IS NULL`, id).Scan(&a.ID, &a.Title, &a.Slug, &a.ContentHTML, &a.ImageURL, &a.StartAt, &a.EndAt, &a.IsActive, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	if e != nil {
		return a, contentNotFound(e)
	}
	return a, nil
}
func (s *ContentService) ChangeAnnouncementActive(ctx context.Context, user domain.User, id int64, active bool) (Announcement, error) {
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH {
		return Announcement{}, ErrContentForbidden
	}
	action := "DEACTIVATE"
	if active {
		action = "ACTIVATE"
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return Announcement{}, e
	}
	defer tx.Rollback()
	if _, e = s.announcementTx(ctx, tx, id); e != nil {
		return Announcement{}, e
	}
	res, e := tx.ExecContext(ctx, `UPDATE announcements SET is_active=? WHERE id=? AND deleted_at IS NULL AND is_active<>?`, active, id, active)
	if e != nil {
		return Announcement{}, e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return Announcement{}, ErrContentConflict
	}
	if e = s.activity(ctx, tx, user, action, "ANNOUNCEMENT", id, map[string]any{"status": map[bool]string{true: "ACTIVE", false: "INACTIVE"}[active]}); e != nil {
		return Announcement{}, e
	}
	if e = tx.Commit(); e != nil {
		return Announcement{}, e
	}
	return s.GetAnnouncement(ctx, fmt.Sprint(id), false)
}
func (s *ContentService) DeleteAnnouncement(ctx context.Context, user domain.User, id int64) error {
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH {
		return ErrContentForbidden
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = s.announcementTx(ctx, tx, id); e != nil {
		return e
	}
	_, e = tx.ExecContext(ctx, `UPDATE announcements SET deleted_at=UTC_TIMESTAMP(6) WHERE id=?`, id)
	if e != nil {
		return e
	}
	if e = s.activity(ctx, tx, user, "DELETE", "ANNOUNCEMENT", id, nil); e != nil {
		return e
	}
	return tx.Commit()
}

func (s *ContentService) activity(ctx context.Context, tx *sql.Tx, user domain.User, action, entity string, id int64, details map[string]any) error {
	return s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: action, EntityType: entity, EntityID: fmt.Sprint(id), Details: details})
}
func contentNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrContentNotFound
	}
	return err
}
func sqlContentError(err error) error {
	var mysqlErr interface{ Error() string }
	if errors.As(err, &mysqlErr) {
		message := strings.ToLower(mysqlErr.Error())
		if strings.Contains(message, "duplicate entry") {
			return ErrContentConflict
		}
	}
	return err
}
