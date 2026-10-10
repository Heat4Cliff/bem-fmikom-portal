package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
)

var (
	ErrAspirationNotFound  = errors.New("aspiration not found")
	ErrAspirationForbidden = errors.New("aspiration access forbidden")
	ErrAspirationInvalid   = errors.New("aspiration validation failed")
)

const (
	AspirationSubjectMax = 200
	AspirationMessageMax = 5000
	AspirationNameMax    = 160
	AspirationNIMMax     = 32
	AspirationEmailMax   = 254
)

type AspirationService struct {
	db  *sql.DB
	log *repository.ActivityLogs
}

func NewAspirationService(db *sql.DB) *AspirationService {
	return &AspirationService{db: db, log: repository.NewActivityLogs(db)}
}

type AspirationInput struct {
	DivisionID  int64  `json:"division_id"`
	Name        string `json:"name"`
	NIM         string `json:"nim"`
	Email       string `json:"email"`
	Subject     string `json:"subject"`
	Message     string `json:"message"`
	IsAnonymous bool   `json:"is_anonymous"`
}

type AspirationReceipt struct {
	ID          int64     `json:"id"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
}

type Aspiration struct {
	ID          int64      `json:"id"`
	DivisionID  int64      `json:"division_id"`
	Division    string     `json:"division_name"`
	Name        *string    `json:"name"`
	NIM         *string    `json:"nim"`
	Email       *string    `json:"email"`
	Subject     string     `json:"subject"`
	Message     string     `json:"message"`
	IsAnonymous bool       `json:"is_anonymous"`
	Status      string     `json:"status"`
	SubmittedAt time.Time  `json:"submitted_at"`
	ReadAt      *time.Time `json:"read_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AspirationFilters struct {
	Page       int
	Limit      int
	Search     string
	Status     string
	DivisionID *int64
	DateFrom   *time.Time
	DateTo     *time.Time
}

func (s *AspirationService) Submit(ctx context.Context, in AspirationInput) (AspirationReceipt, error) {
	in.Subject = strings.TrimSpace(in.Subject)
	in.Message = strings.TrimSpace(in.Message)
	in.Name = strings.TrimSpace(in.Name)
	in.NIM = strings.TrimSpace(in.NIM)
	in.Email = strings.TrimSpace(in.Email)
	if in.DivisionID < 1 || runeLen(in.Subject) == 0 || runeLen(in.Subject) > AspirationSubjectMax || runeLen(in.Message) == 0 || runeLen(in.Message) > AspirationMessageMax {
		return AspirationReceipt{}, ErrAspirationInvalid
	}
	var name, nim, email any
	if !in.IsAnonymous {
		if runeLen(in.Name) > AspirationNameMax || runeLen(in.NIM) > AspirationNIMMax || runeLen(in.Email) > AspirationEmailMax {
			return AspirationReceipt{}, ErrAspirationInvalid
		}
		if in.Email != "" {
			parsed, err := mail.ParseAddress(in.Email)
			if err != nil || parsed.Address != in.Email {
				return AspirationReceipt{}, ErrAspirationInvalid
			}
			email = in.Email
		}
		if in.Name != "" {
			name = in.Name
		}
		if in.NIM != "" {
			nim = in.NIM
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AspirationReceipt{}, err
	}
	defer tx.Rollback()
	var divisionID int64
	err = tx.QueryRowContext(ctx, `SELECT d.id FROM divisions d JOIN periods p ON p.id=d.period_id WHERE d.id=? AND d.status='ACTIVE' AND p.status='ACTIVE' FOR UPDATE`, in.DivisionID).Scan(&divisionID)
	if errors.Is(err, sql.ErrNoRows) {
		return AspirationReceipt{}, ErrAspirationInvalid
	}
	if err != nil {
		return AspirationReceipt{}, err
	}
	submittedAt := time.Now().UTC()
	result, err := tx.ExecContext(ctx, `INSERT INTO aspirations(division_id,name,nim,email,subject,message,is_anonymous,status,submitted_at) VALUES(?,?,?,?,?,?,?,'PENDING',?)`, divisionID, name, nim, email, in.Subject, in.Message, in.IsAnonymous, submittedAt)
	if err != nil {
		return AspirationReceipt{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return AspirationReceipt{}, err
	}
	if err := tx.Commit(); err != nil {
		return AspirationReceipt{}, err
	}
	return AspirationReceipt{ID: id, Status: "PENDING", SubmittedAt: submittedAt}, nil
}

func (s *AspirationService) List(ctx context.Context, user domain.User, filters AspirationFilters) (Result[Aspiration], error) {
	if filters.Page < 1 || filters.Limit < 1 || filters.Limit > 100 || (filters.Status != "" && !validAspirationStatus(filters.Status)) || (filters.DateFrom != nil && filters.DateTo != nil && filters.DateTo.Before(*filters.DateFrom)) {
		return Result[Aspiration]{}, ErrAspirationInvalid
	}
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH && user.Role != domain.RoleKepala {
		return Result[Aspiration]{}, ErrAspirationForbidden
	}
	where, args := aspirationWhere(user, filters)
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM aspirations a`+where, args...).Scan(&total); err != nil {
		return Result[Aspiration]{}, err
	}
	queryArgs := append(append([]any{}, args...), filters.Limit, (filters.Page-1)*filters.Limit)
	rows, err := s.db.QueryContext(ctx, `SELECT a.id,a.division_id,d.name,a.name,a.nim,a.email,a.subject,a.message,a.is_anonymous,a.status,a.submitted_at,a.read_at,a.resolved_at,a.created_at,a.updated_at FROM aspirations a JOIN divisions d ON d.id=a.division_id`+where+` ORDER BY a.submitted_at DESC,a.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return Result[Aspiration]{}, err
	}
	defer rows.Close()
	items := []Aspiration{}
	for rows.Next() {
		item, err := scanAspiration(rows)
		if err != nil {
			return Result[Aspiration]{}, err
		}
		items = append(items, item)
	}
	return Result[Aspiration]{Items: items, Page: pageMeta(filters.Page, filters.Limit, total)}, rows.Err()
}

func (s *AspirationService) Get(ctx context.Context, user domain.User, id int64) (Aspiration, error) {
	if id < 1 {
		return Aspiration{}, ErrAspirationInvalid
	}
	query := `SELECT a.id,a.division_id,d.name,a.name,a.nim,a.email,a.subject,a.message,a.is_anonymous,a.status,a.submitted_at,a.read_at,a.resolved_at,a.created_at,a.updated_at FROM aspirations a JOIN divisions d ON d.id=a.division_id WHERE a.id=?`
	args := []any{id}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil || *user.DivisionID < 1 {
			return Aspiration{}, ErrAspirationForbidden
		}
		query += ` AND a.division_id=?`
		args = append(args, *user.DivisionID)
	} else if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH {
		return Aspiration{}, ErrAspirationForbidden
	}
	item, err := scanAspiration(s.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return Aspiration{}, ErrAspirationNotFound
	}
	return item, err
}

func (s *AspirationService) UpdateStatus(ctx context.Context, user domain.User, id int64, newStatus string) (Aspiration, error) {
	if id < 1 || !validAspirationStatus(newStatus) {
		return Aspiration{}, ErrAspirationInvalid
	}
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH && user.Role != domain.RoleKepala {
		return Aspiration{}, ErrAspirationForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Aspiration{}, err
	}
	defer tx.Rollback()
	var divisionID int64
	var oldStatus string
	err = tx.QueryRowContext(ctx, `SELECT division_id,status FROM aspirations WHERE id=? FOR UPDATE`, id).Scan(&divisionID, &oldStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return Aspiration{}, ErrAspirationNotFound
	}
	if err != nil {
		return Aspiration{}, err
	}
	if user.Role == domain.RoleKepala && (user.DivisionID == nil || *user.DivisionID != divisionID) {
		return Aspiration{}, ErrAspirationForbidden
	}
	if !allowedAspirationTransition(oldStatus, newStatus) {
		return Aspiration{}, ErrAspirationInvalid
	}
	now := time.Now().UTC()
	var readAt, resolvedAt any
	if newStatus == "READ" {
		readAt = now
	}
	if newStatus == "RESOLVED" {
		resolvedAt = now
	}
	if _, err := tx.ExecContext(ctx, `UPDATE aspirations SET status=?,read_at=COALESCE(read_at,?),resolved_at=COALESCE(resolved_at,?) WHERE id=?`, newStatus, readAt, resolvedAt, id); err != nil {
		return Aspiration{}, err
	}
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{
		ActorUserID: &user.ID,
		Action:      "UPDATE",
		EntityType:  "ASPIRATION",
		EntityID:    fmt.Sprint(id),
		Details:     map[string]any{"aspiration_id": id, "old_status": oldStatus, "new_status": newStatus, "division_id": divisionID},
	}); err != nil {
		return Aspiration{}, err
	}
	if err := tx.Commit(); err != nil {
		return Aspiration{}, err
	}
	return s.Get(ctx, user, id)
}

func aspirationWhere(user domain.User, filters AspirationFilters) (string, []any) {
	parts := []string{"1=1"}
	args := []any{}
	search := strings.ToLower(strings.TrimSpace(filters.Search))
	if search != "" {
		parts = append(parts, `(LOCATE(?,LOWER(a.subject))>0 OR LOCATE(?,LOWER(COALESCE(a.name,'')))>0 OR LOCATE(?,LOWER(COALESCE(a.nim,'')))>0 OR LOCATE(?,LOWER(COALESCE(a.email,'')))>0)`)
		args = append(args, search, search, search, search)
	}
	if filters.Status != "" {
		parts = append(parts, `a.status=?`)
		args = append(args, filters.Status)
	}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil || *user.DivisionID < 1 {
			parts = append(parts, `1=0`)
		} else {
			parts = append(parts, `a.division_id=?`)
			args = append(args, *user.DivisionID)
		}
	} else if filters.DivisionID != nil {
		parts = append(parts, `a.division_id=?`)
		args = append(args, *filters.DivisionID)
	}
	if filters.DateFrom != nil {
		parts = append(parts, `a.submitted_at>=?`)
		args = append(args, filters.DateFrom.UTC())
	}
	if filters.DateTo != nil {
		parts = append(parts, `a.submitted_at<?`)
		args = append(args, filters.DateTo.AddDate(0, 0, 1).UTC())
	}
	return ` WHERE ` + strings.Join(parts, ` AND `), args
}

type aspirationScanner interface{ Scan(...any) error }

func scanAspiration(row aspirationScanner) (Aspiration, error) {
	var item Aspiration
	var name, nim, email sql.NullString
	var readAt, resolvedAt sql.NullTime
	err := row.Scan(&item.ID, &item.DivisionID, &item.Division, &name, &nim, &email, &item.Subject, &item.Message, &item.IsAnonymous, &item.Status, &item.SubmittedAt, &readAt, &resolvedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Aspiration{}, err
	}
	item.Name, item.NIM, item.Email = nullableAspirationString(name), nullableAspirationString(nim), nullableAspirationString(email)
	item.ReadAt, item.ResolvedAt = nullableAspirationTime(readAt), nullableAspirationTime(resolvedAt)
	item.SubmittedAt = item.SubmittedAt.UTC()
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	if item.ReadAt != nil {
		value := item.ReadAt.UTC()
		item.ReadAt = &value
	}
	if item.ResolvedAt != nil {
		value := item.ResolvedAt.UTC()
		item.ResolvedAt = &value
	}
	return item, nil
}

func nullableAspirationString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableAspirationTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func runeLen(value string) int { return utf8.RuneCountInString(value) }

func validAspirationStatus(status string) bool {
	return status == "PENDING" || status == "READ" || status == "RESOLVED"
}

func allowedAspirationTransition(oldStatus, newStatus string) bool {
	return (oldStatus == "PENDING" && newStatus == "READ") || (oldStatus == "READ" && newStatus == "RESOLVED")
}
