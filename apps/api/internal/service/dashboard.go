package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
)

var ErrDashboardForbidden = errors.New("dashboard access forbidden")

type DashboardSummary struct {
	GeneratedAt   time.Time             `json:"generated_at"`
	Scope         DashboardScope        `json:"scope"`
	ActivePeriod  *DashboardPeriod      `json:"active_period"`
	Articles      DashboardArticleStats `json:"articles"`
	Registrations DashboardRegStats     `json:"registrations"`
	Aspirations   DashboardAspStats     `json:"aspirations"`
	Documents     DashboardCount        `json:"documents"`
	Announcements DashboardCount        `json:"announcements"`
}

type DashboardScope struct {
	Type     string           `json:"type"`
	Division *domain.Division `json:"division"`
}
type DashboardPeriod struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}
type DashboardCount struct {
	Total int64 `json:"total"`
}
type DashboardArticleStats struct {
	Total     int64 `json:"total"`
	Published int64 `json:"published"`
	Draft     int64 `json:"draft"`
}
type DashboardRegStats struct {
	ActiveForms           int64 `json:"active_forms"`
	ActiveFormSubmissions int64 `json:"active_form_submissions"`
}
type DashboardAspStats struct {
	Total    int64 `json:"total"`
	Pending  int64 `json:"pending"`
	Read     int64 `json:"read"`
	Resolved int64 `json:"resolved"`
}

type DashboardService struct{ db *sql.DB }

func NewDashboardService(db *sql.DB) *DashboardService { return &DashboardService{db: db} }

// Summary returns bounded aggregate counts. Role-derived scope is never accepted from request input.
func (s *DashboardService) Summary(ctx context.Context, user domain.User) (DashboardSummary, error) {
	var result DashboardSummary
	if user.Status != domain.StatusActive {
		return result, ErrDashboardForbidden
	}
	whereArticles, whereForms, whereAspirations, whereDocuments := "", "", "", ""
	switch user.Role {
	case domain.RoleSuperAdmin, domain.RoleBPH:
		if user.DivisionID != nil {
			return result, ErrDashboardForbidden
		}
		result.Scope = DashboardScope{Type: "organization"}
	case domain.RoleKepala:
		if user.DivisionID == nil || *user.DivisionID < 1 {
			return result, ErrDashboardForbidden
		}
		var division domain.Division
		if err := s.db.QueryRowContext(ctx, `SELECT id,name FROM divisions WHERE id=?`, *user.DivisionID).Scan(&division.ID, &division.Name); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return result, ErrDashboardForbidden
			}
			return result, fmt.Errorf("load dashboard division: %w", err)
		}
		whereArticles, whereForms, whereAspirations, whereDocuments = " AND division_id=?", " AND f.division_id=?", " AND division_id=?", " AND division_id=?"
		result.Scope = DashboardScope{Type: "division", Division: &division}
	default:
		return result, ErrDashboardForbidden
	}

	// Scalar subqueries avoid join multiplication while keeping this to one aggregate round trip.
	query := `SELECT UTC_TIMESTAMP(6),
	 (SELECT id FROM periods WHERE status='ACTIVE' LIMIT 1),
	 (SELECT name FROM periods WHERE status='ACTIVE' LIMIT 1),
	 (SELECT start_date FROM periods WHERE status='ACTIVE' LIMIT 1),
	 (SELECT end_date FROM periods WHERE status='ACTIVE' LIMIT 1),
	 (SELECT COUNT(*) FROM articles WHERE deleted_at IS NULL` + whereArticles + `),
	 (SELECT COUNT(*) FROM articles WHERE deleted_at IS NULL AND status='PUBLISHED'` + whereArticles + `),
	 (SELECT COUNT(*) FROM articles WHERE deleted_at IS NULL AND status='DRAFT'` + whereArticles + `),
	 (SELECT COUNT(*) FROM registration_forms f WHERE f.deleted_at IS NULL AND f.status='OPEN' AND UTC_TIMESTAMP(6)>=f.open_at AND UTC_TIMESTAMP(6)<=f.close_at` + whereForms + `),
	 (SELECT COUNT(*) FROM registration_submissions rs INNER JOIN registration_forms f ON f.id=rs.registration_form_id WHERE f.deleted_at IS NULL AND f.status='OPEN' AND UTC_TIMESTAMP(6)>=f.open_at AND UTC_TIMESTAMP(6)<=f.close_at` + whereForms + `),
	 (SELECT COUNT(*) FROM aspirations WHERE 1=1` + whereAspirations + `),
	 (SELECT COUNT(*) FROM aspirations WHERE status='PENDING'` + whereAspirations + `),
	 (SELECT COUNT(*) FROM aspirations WHERE status='READ'` + whereAspirations + `),
	 (SELECT COUNT(*) FROM aspirations WHERE status='RESOLVED'` + whereAspirations + `),
	 (SELECT COUNT(*) FROM documents WHERE delete_pending=FALSE` + whereDocuments + `),
	 (SELECT COUNT(*) FROM announcements WHERE is_active=TRUE AND deleted_at IS NULL AND start_at<=UTC_TIMESTAMP(6) AND (end_at IS NULL OR end_at>=UTC_TIMESTAMP(6)))`
	// Every scoped subquery receives the division ID; the submission aggregate scopes on f.division_id.
	queryArgs := make([]any, 0, 10)
	if user.Role == domain.RoleKepala {
		for range 10 {
			queryArgs = append(queryArgs, *user.DivisionID)
		}
	}
	var generated time.Time
	var periodID, totalArticles, published, drafts, activeForms, submissions, aspirationTotal, pending, read, resolved, documents, announcements sql.NullInt64
	var periodName sql.NullString
	var startDate, endDate sql.NullTime
	err := s.db.QueryRowContext(ctx, query, queryArgs...).Scan(&generated, &periodID, &periodName, &startDate, &endDate,
		&totalArticles, &published, &drafts, &activeForms, &submissions, &aspirationTotal, &pending, &read, &resolved, &documents, &announcements)
	if err != nil {
		return result, fmt.Errorf("aggregate dashboard summary: %w", err)
	}
	result.GeneratedAt = generated.UTC()
	result.Articles = DashboardArticleStats{Total: totalArticles.Int64, Published: published.Int64, Draft: drafts.Int64}
	result.Registrations = DashboardRegStats{ActiveForms: activeForms.Int64, ActiveFormSubmissions: submissions.Int64}
	result.Aspirations = DashboardAspStats{Total: aspirationTotal.Int64, Pending: pending.Int64, Read: read.Int64, Resolved: resolved.Int64}
	result.Documents = DashboardCount{Total: documents.Int64}
	result.Announcements = DashboardCount{Total: announcements.Int64}
	if periodID.Valid {
		result.ActivePeriod = &DashboardPeriod{ID: periodID.Int64, Name: periodName.String, StartDate: startDate.Time, EndDate: endDate.Time}
	}
	return result, nil
}
