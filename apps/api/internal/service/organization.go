package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
)

var (
	ErrOrganizationNotFound  = errors.New("organization resource not found")
	ErrOrganizationForbidden = errors.New("organization resource forbidden")
	ErrOrganizationConflict  = errors.New("organization resource conflict")
	ErrOrganizationInvalid   = errors.New("organization data invalid")
)

var organizationPositions = map[string]int{
	"KETUA": 1, "WAKIL_KETUA": 2, "SEKRETARIS": 3, "BENDAHARA": 4, "KEPALA_BIDANG": 5, "ANGGOTA": 6,
}

type OrganizationService struct {
	db  *sql.DB
	log *repository.ActivityLogs
}

func NewOrganizationService(db *sql.DB) *OrganizationService {
	return &OrganizationService{db: db, log: repository.NewActivityLogs(db)}
}

type Period struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Status        string `json:"status"`
	DivisionCount int    `json:"division_count,omitempty"`
	MemberCount   int    `json:"member_count,omitempty"`
}

type Division struct {
	ID           int64          `json:"id"`
	PeriodID     int64          `json:"period_id"`
	PeriodName   string         `json:"period_name,omitempty"`
	Name         string         `json:"name"`
	Description  *string        `json:"description"`
	Status       string         `json:"status"`
	DisplayOrder int            `json:"display_order"`
	Head         *MemberSummary `json:"head,omitempty"`
	MemberCount  int            `json:"member_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type MemberSummary struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Position string  `json:"position"`
	NIM      *string `json:"nim"`
	PhotoURL *string `json:"photo_url"`
}

type OrganizationMember struct {
	ID           int64     `json:"id"`
	PeriodID     int64     `json:"period_id"`
	PeriodName   string    `json:"period_name,omitempty"`
	DivisionID   *int64    `json:"division_id"`
	DivisionName *string   `json:"division_name"`
	Name         string    `json:"name"`
	NIM          *string   `json:"nim"`
	Position     string    `json:"position"`
	PhotoURL     *string   `json:"photo_url"`
	DisplayOrder int       `json:"display_order"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Organization struct {
	Period     *PublicPeriod                `json:"period"`
	Leadership []PublicOrganizationMember   `json:"leadership"`
	Divisions  []PublicOrganizationDivision `json:"divisions"`
}

type PublicPeriod struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type PublicOrganizationMember struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	NIM      *string `json:"nim"`
	Position string  `json:"position"`
	PhotoURL *string `json:"photo_url"`
}

type PublicOrganizationDivision struct {
	ID      int64                      `json:"id"`
	Name    string                     `json:"name"`
	Head    *PublicOrganizationMember  `json:"head"`
	Members []PublicOrganizationMember `json:"members"`
}

type PublicDivision struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type PeriodInput struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type DivisionInput struct {
	PeriodID     int64   `json:"period_id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Status       string  `json:"status"`
	DisplayOrder int     `json:"display_order"`
}

type MemberInput struct {
	PeriodID     int64   `json:"period_id"`
	DivisionID   *int64  `json:"division_id"`
	Name         string  `json:"name"`
	NIM          *string `json:"nim"`
	Position     string  `json:"position"`
	PhotoURL     *string `json:"photo_url"`
	DisplayOrder int     `json:"display_order"`
	Status       string  `json:"status"`
}

type OrganizationFilters struct {
	Page, Limit              int
	Search, Status, Position string
	PeriodID, DivisionID     *int64
}

func (s *OrganizationService) ActivePeriod(ctx context.Context) (*Period, error) {
	p, err := scanPeriod(s.db.QueryRowContext(ctx, `SELECT id,name,DATE_FORMAT(start_date,'%Y-%m-%d'),DATE_FORMAT(end_date,'%Y-%m-%d'),status FROM periods WHERE status='ACTIVE'`))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrOrganizationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func scanPeriod(row *sql.Row) (Period, error) {
	var p Period
	err := row.Scan(&p.ID, &p.Name, &p.StartDate, &p.EndDate, &p.Status)
	return p, err
}

func (s *OrganizationService) PublicDivisions(ctx context.Context) ([]PublicDivision, error) {
	p, err := s.ActivePeriod(ctx)
	if errors.Is(err, ErrOrganizationNotFound) {
		return []PublicDivision{}, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,description FROM divisions WHERE period_id=? AND status='ACTIVE' ORDER BY display_order,name,id`, p.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []PublicDivision{}
	for rows.Next() {
		var d PublicDivision
		if err := rows.Scan(&d.ID, &d.Name, &d.Description); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (s *OrganizationService) PublicOrganization(ctx context.Context) (Organization, error) {
	result := Organization{Leadership: []PublicOrganizationMember{}, Divisions: []PublicOrganizationDivision{}}
	p, err := s.ActivePeriod(ctx)
	if errors.Is(err, ErrOrganizationNotFound) {
		return result, nil
	}
	if err != nil {
		return result, err
	}
	result.Period = &PublicPeriod{ID: p.ID, Name: p.Name, StartDate: p.StartDate, EndDate: p.EndDate}
	rows, err := s.db.QueryContext(ctx, `SELECT id,name,nim,position,photo_url FROM organization_members WHERE period_id=? AND division_id IS NULL AND status='ACTIVE' AND position IN ('KETUA','WAKIL_KETUA','SEKRETARIS','BENDAHARA') ORDER BY FIELD(position,'KETUA','WAKIL_KETUA','SEKRETARIS','BENDAHARA'),display_order,name,id`, p.ID)
	if err != nil {
		return result, err
	}
	result.Leadership, err = scanPublicMembers(rows)
	rows.Close()
	if err != nil {
		return result, err
	}
	rows, err = s.db.QueryContext(ctx, `SELECT d.id,d.name,
		h.id,h.name,h.nim,h.photo_url,
		m.id,m.name,m.nim,m.photo_url
		FROM divisions d
		LEFT JOIN organization_members h ON h.division_id=d.id AND h.period_id=d.period_id AND h.position='KEPALA_BIDANG' AND h.status='ACTIVE'
		LEFT JOIN organization_members m ON m.division_id=d.id AND m.period_id=d.period_id AND m.status='ACTIVE' AND m.position='ANGGOTA'
		WHERE d.period_id=? AND d.status='ACTIVE'
		ORDER BY d.display_order,d.name,d.id,COALESCE(m.display_order,-1),m.name,m.id`, p.ID)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	byID := map[int64]int{}
	for rows.Next() {
		var d PublicOrganizationDivision
		var headID, mID sql.NullInt64
		var headName, headNIM, headPhoto, mName, mNIM, mPhoto sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &headID, &headName, &headNIM, &headPhoto, &mID, &mName, &mNIM, &mPhoto); err != nil {
			return result, err
		}
		idx, ok := byID[d.ID]
		if !ok {
			idx = len(result.Divisions)
			byID[d.ID] = idx
			result.Divisions = append(result.Divisions, PublicOrganizationDivision{ID: d.ID, Name: d.Name, Members: []PublicOrganizationMember{}})
		}
		division := &result.Divisions[idx]
		if headID.Valid && division.Head == nil {
			division.Head = &PublicOrganizationMember{ID: headID.Int64, Name: headName.String, NIM: ptrString(headNIM), Position: "KEPALA_BIDANG", PhotoURL: ptrString(headPhoto)}
		}
		if mID.Valid && (len(division.Members) == 0 || division.Members[len(division.Members)-1].ID != mID.Int64) {
			division.Members = append(division.Members, PublicOrganizationMember{ID: mID.Int64, Name: mName.String, NIM: ptrString(mNIM), Position: "ANGGOTA", PhotoURL: ptrString(mPhoto)})
		}
	}
	return result, rows.Err()
}

func ptrString(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}
func ptrInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func scanPublicMembers(rows *sql.Rows) ([]PublicOrganizationMember, error) {
	items := []PublicOrganizationMember{}
	for rows.Next() {
		var m PublicOrganizationMember
		var nim, photo sql.NullString
		if err := rows.Scan(&m.ID, &m.Name, &nim, &m.Position, &photo); err != nil {
			return nil, err
		}
		m.NIM, m.PhotoURL = ptrString(nim), ptrString(photo)
		items = append(items, m)
	}
	return items, rows.Err()
}

func (s *OrganizationService) ListPeriods(ctx context.Context, page, limit int, search, status string) (Result[Period], error) {
	if page < 1 || limit < 1 || limit > 100 || (status != "" && status != "ACTIVE" && status != "ARCHIVED") {
		return Result[Period]{}, ErrOrganizationInvalid
	}
	where := " WHERE (?='' OR name LIKE ?) AND (?='' OR status=?)"
	like := "%" + strings.TrimSpace(search) + "%"
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM periods"+where, search, like, status, status).Scan(&total); err != nil {
		return Result[Period]{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id,p.name,DATE_FORMAT(p.start_date,'%Y-%m-%d'),DATE_FORMAT(p.end_date,'%Y-%m-%d'),p.status,
	(SELECT COUNT(*) FROM divisions d WHERE d.period_id=p.id),(SELECT COUNT(*) FROM organization_members m WHERE m.period_id=p.id AND m.status='ACTIVE')
	FROM periods p`+where+` ORDER BY p.start_date DESC,p.id DESC LIMIT ? OFFSET ?`, search, like, status, status, limit, (page-1)*limit)
	if err != nil {
		return Result[Period]{}, err
	}
	defer rows.Close()
	items := []Period{}
	for rows.Next() {
		var p Period
		if err := rows.Scan(&p.ID, &p.Name, &p.StartDate, &p.EndDate, &p.Status, &p.DivisionCount, &p.MemberCount); err != nil {
			return Result[Period]{}, err
		}
		items = append(items, p)
	}
	return Result[Period]{Items: items, Page: Page{Page: page, Limit: limit, Total: total, TotalPages: (total + limit - 1) / limit}}, rows.Err()
}

func (s *OrganizationService) GetPeriod(ctx context.Context, id int64) (Period, error) {
	var p Period
	err := s.db.QueryRowContext(ctx, `SELECT p.id,p.name,DATE_FORMAT(p.start_date,'%Y-%m-%d'),DATE_FORMAT(p.end_date,'%Y-%m-%d'),p.status,(SELECT COUNT(*) FROM divisions d WHERE d.period_id=p.id),(SELECT COUNT(*) FROM organization_members m WHERE m.period_id=p.id AND m.status='ACTIVE') FROM periods p WHERE p.id=?`, id).Scan(&p.ID, &p.Name, &p.StartDate, &p.EndDate, &p.Status, &p.DivisionCount, &p.MemberCount)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrOrganizationNotFound
	}
	return p, err
}

func (s *OrganizationService) SavePeriod(ctx context.Context, user domain.User, id int64, in PeriodInput) (Period, error) {
	if user.Role != domain.RoleSuperAdmin {
		return Period{}, ErrOrganizationForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	start, e1 := time.Parse("2006-01-02", in.StartDate)
	end, e2 := time.Parse("2006-01-02", in.EndDate)
	if in.Name == "" || len(in.Name) > 120 || e1 != nil || e2 != nil || end.Before(start) {
		return Period{}, ErrOrganizationInvalid
	}
	created := id == 0
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Period{}, err
	}
	defer tx.Rollback()
	if created {
		res, e := tx.ExecContext(ctx, `INSERT INTO periods(name,start_date,end_date,status) VALUES(?,?,?,'ARCHIVED')`, in.Name, start, end)
		if e != nil {
			return Period{}, mapOrganizationDBError(e)
		}
		id, _ = res.LastInsertId()
	} else {
		var exists int
		if e := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM periods WHERE id=? FOR UPDATE`, id).Scan(&exists); e != nil {
			return Period{}, e
		}
		if exists == 0 {
			return Period{}, ErrOrganizationNotFound
		}
		if _, e := tx.ExecContext(ctx, `UPDATE periods SET name=?,start_date=?,end_date=? WHERE id=?`, in.Name, start, end, id); e != nil {
			return Period{}, mapOrganizationDBError(e)
		}
	}
	if err := s.writeActivity(ctx, tx, user, actionForCreated(created, "UPDATE"), "PERIOD", id, map[string]any{"name": in.Name, "start_date": in.StartDate, "end_date": in.EndDate}); err != nil {
		return Period{}, err
	}
	if err := tx.Commit(); err != nil {
		return Period{}, err
	}
	return s.GetPeriod(ctx, id)
}

func (s *OrganizationService) ActivatePeriod(ctx context.Context, user domain.User, id int64) error {
	if user.Role != domain.RoleSuperAdmin {
		return ErrOrganizationForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Lock the shared period rows in a stable order; the generated unique active slot is the final invariant.
	rows, err := tx.QueryContext(ctx, `SELECT id FROM periods ORDER BY id FOR UPDATE`)
	if err != nil {
		return err
	}
	ids := []int64{}
	for rows.Next() {
		var v int64
		if e := rows.Scan(&v); e != nil {
			rows.Close()
			return e
		}
		ids = append(ids, v)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	found := false
	for _, v := range ids {
		if v == id {
			found = true
			break
		}
	}
	if !found {
		return ErrOrganizationNotFound
	}
	var wasActive bool
	if err := tx.QueryRowContext(ctx, `SELECT status='ACTIVE' FROM periods WHERE id=?`, id).Scan(&wasActive); err != nil {
		return err
	}
	if wasActive {
		return tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, `UPDATE periods SET status='ARCHIVED' WHERE status='ACTIVE'`); err != nil {
		return mapOrganizationDBError(err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE periods SET status='ACTIVE' WHERE id=?`, id); err != nil {
		return mapOrganizationDBError(err)
	}
	if err := s.writeActivity(ctx, tx, user, "ACTIVATE", "PERIOD", id, map[string]any{"status": "ACTIVE"}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OrganizationService) ArchivePeriod(ctx context.Context, user domain.User, id int64) error {
	if user.Role != domain.RoleSuperAdmin {
		return ErrOrganizationForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var name, status string
	if err := tx.QueryRowContext(ctx, `SELECT name,status FROM periods WHERE id=? FOR UPDATE`, id).Scan(&name, &status); errors.Is(err, sql.ErrNoRows) {
		return ErrOrganizationNotFound
	} else if err != nil {
		return err
	}
	if status == "ARCHIVED" {
		return tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, `UPDATE periods SET status='ARCHIVED' WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.writeActivity(ctx, tx, user, "ARCHIVE", "PERIOD", id, map[string]any{"name": name, "status": "ARCHIVED"}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OrganizationService) ListDivisions(ctx context.Context, filters OrganizationFilters, user *domain.User) (Result[Division], error) {
	return s.listDivisions(ctx, filters, user)
}
func (s *OrganizationService) listDivisions(ctx context.Context, f OrganizationFilters, user *domain.User) (Result[Division], error) {
	if f.Page < 1 || f.Limit < 1 || f.Limit > 100 || (f.Status != "" && f.Status != "ACTIVE" && f.Status != "INACTIVE") {
		return Result[Division]{}, ErrOrganizationInvalid
	}
	clauses := []string{"(?='' OR d.name LIKE ?)", "(?='' OR d.status=?)"}
	args := []any{f.Search, "%" + strings.TrimSpace(f.Search) + "%", f.Status, f.Status}
	if f.PeriodID != nil {
		clauses = append(clauses, "d.period_id=?")
		args = append(args, *f.PeriodID)
	}
	if f.DivisionID != nil {
		clauses = append(clauses, "d.id=?")
		args = append(args, *f.DivisionID)
	}
	where := " WHERE " + strings.Join(clauses, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM divisions d`+where, args...).Scan(&total); err != nil {
		return Result[Division]{}, err
	}
	query := `SELECT d.id,d.period_id,p.name,d.name,d.description,d.status,d.display_order,d.created_at,d.updated_at,
	(SELECT COUNT(*) FROM organization_members m WHERE m.division_id=d.id AND m.period_id=d.period_id AND m.status='ACTIVE'),
	(SELECT id FROM organization_members h WHERE h.division_id=d.id AND h.period_id=d.period_id AND h.position='KEPALA_BIDANG' AND h.status='ACTIVE' LIMIT 1),
	(SELECT name FROM organization_members h WHERE h.division_id=d.id AND h.period_id=d.period_id AND h.position='KEPALA_BIDANG' AND h.status='ACTIVE' LIMIT 1)
	FROM divisions d JOIN periods p ON p.id=d.period_id` + where + ` ORDER BY d.period_id DESC,d.display_order,d.name,d.id LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), f.Limit, (f.Page-1)*f.Limit)
	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return Result[Division]{}, err
	}
	defer rows.Close()
	items := []Division{}
	for rows.Next() {
		var d Division
		var headID sql.NullInt64
		var headName sql.NullString
		if err := rows.Scan(&d.ID, &d.PeriodID, &d.PeriodName, &d.Name, &d.Description, &d.Status, &d.DisplayOrder, &d.CreatedAt, &d.UpdatedAt, &d.MemberCount, &headID, &headName); err != nil {
			return Result[Division]{}, err
		}
		if headID.Valid {
			d.Head = &MemberSummary{ID: headID.Int64, Name: headName.String, Position: "KEPALA_BIDANG"}
		}
		items = append(items, d)
	}
	return Result[Division]{Items: items, Page: Page{Page: f.Page, Limit: f.Limit, Total: total, TotalPages: (total + f.Limit - 1) / f.Limit}}, rows.Err()
}

func (s *OrganizationService) GetDivision(ctx context.Context, id int64, user domain.User) (Division, error) {
	f := OrganizationFilters{Page: 1, Limit: 1, DivisionID: &id}
	if user.Role == domain.RoleKepala {
		f.DivisionID = user.DivisionID
	}
	r, e := s.ListDivisions(ctx, f, &user)
	if e != nil {
		return Division{}, e
	}
	if user.Role == domain.RoleKepala && (user.DivisionID == nil || *user.DivisionID != id) {
		return Division{}, ErrOrganizationNotFound
	}
	if len(r.Items) == 0 {
		return Division{}, ErrOrganizationNotFound
	}
	return r.Items[0], nil
}

func (s *OrganizationService) SaveDivision(ctx context.Context, user domain.User, id int64, in DivisionInput) (Division, error) {
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH {
		return Division{}, ErrOrganizationForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	if in.Status == "" {
		in.Status = "ACTIVE"
	}
	if in.PeriodID < 1 || in.Name == "" || len(in.Name) > 120 || (in.Status != "ACTIVE" && in.Status != "INACTIVE") || in.DisplayOrder < 0 {
		return Division{}, ErrOrganizationInvalid
	}
	var periodStatus string
	if err := s.db.QueryRowContext(ctx, `SELECT status FROM periods WHERE id=?`, in.PeriodID).Scan(&periodStatus); errors.Is(err, sql.ErrNoRows) {
		return Division{}, ErrOrganizationInvalid
	} else if err != nil {
		return Division{}, err
	}
	if user.Role == domain.RoleBPH && id > 0 {
		var currentPeriod int64
		if err := s.db.QueryRowContext(ctx, `SELECT period_id FROM divisions WHERE id=?`, id).Scan(&currentPeriod); errors.Is(err, sql.ErrNoRows) {
			return Division{}, ErrOrganizationNotFound
		} else if err != nil {
			return Division{}, err
		}
		if currentPeriod != in.PeriodID {
			return Division{}, ErrOrganizationConflict
		}
	}
	created := id == 0
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Division{}, err
	}
	defer tx.Rollback()
	if created {
		res, e := tx.ExecContext(ctx, `INSERT INTO divisions(period_id,name,description,status,display_order) VALUES(?,?,?,?,?)`, in.PeriodID, in.Name, in.Description, in.Status, in.DisplayOrder)
		if e != nil {
			return Division{}, mapOrganizationDBError(e)
		}
		id, _ = res.LastInsertId()
	} else {
		var exists int
		if e := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM divisions WHERE id=? FOR UPDATE`, id).Scan(&exists); e != nil {
			return Division{}, e
		}
		if exists == 0 {
			return Division{}, ErrOrganizationNotFound
		}
		if _, e := tx.ExecContext(ctx, `UPDATE divisions SET period_id=?,name=?,description=?,status=?,display_order=? WHERE id=?`, in.PeriodID, in.Name, in.Description, in.Status, in.DisplayOrder, id); e != nil {
			return Division{}, mapOrganizationDBError(e)
		}
	}
	if err := s.writeActivity(ctx, tx, user, actionForCreated(created, "UPDATE"), "DIVISION", id, map[string]any{"name": in.Name, "status": in.Status}); err != nil {
		return Division{}, err
	}
	if err := tx.Commit(); err != nil {
		return Division{}, err
	}
	return s.GetDivision(ctx, id, user)
}

func (s *OrganizationService) DeleteDivision(ctx context.Context, user domain.User, id int64) error {
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH {
		return ErrOrganizationForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var name string
	if err := tx.QueryRowContext(ctx, `SELECT name FROM divisions WHERE id=? FOR UPDATE`, id).Scan(&name); errors.Is(err, sql.ErrNoRows) {
		return ErrOrganizationNotFound
	} else if err != nil {
		return err
	}
	var refs int
	if err := tx.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM organization_members WHERE division_id=?)+(SELECT COUNT(*) FROM users WHERE division_id=?)+(SELECT COUNT(*) FROM articles WHERE division_id=?)+(SELECT COUNT(*) FROM gallery_categories WHERE division_id=?)`, id, id, id, id).Scan(&refs); err != nil {
		return err
	}
	if refs > 0 {
		return ErrOrganizationConflict
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM divisions WHERE id=?`, id); err != nil {
		return mapOrganizationDBError(err)
	}
	if err := s.writeActivity(ctx, tx, user, "DELETE", "DIVISION", id, map[string]any{"name": name}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OrganizationService) ListMembers(ctx context.Context, f OrganizationFilters, user domain.User) (Result[OrganizationMember], error) {
	if f.Page < 1 || f.Limit < 1 || f.Limit > 100 || (f.Status != "" && f.Status != "ACTIVE" && f.Status != "INACTIVE") || (f.Position != "" && organizationPositions[f.Position] == 0) {
		return Result[OrganizationMember]{}, ErrOrganizationInvalid
	}
	clauses := []string{"(?='' OR m.name LIKE ? OR m.nim LIKE ?)", "(?='' OR m.status=?)", "(?='' OR m.position=?)"}
	args := []any{f.Search, "%" + strings.TrimSpace(f.Search) + "%", "%" + strings.TrimSpace(f.Search) + "%", f.Status, f.Status, f.Position, f.Position}
	if f.PeriodID != nil {
		clauses = append(clauses, "m.period_id=?")
		args = append(args, *f.PeriodID)
	}
	if f.DivisionID != nil {
		clauses = append(clauses, "m.division_id=?")
		args = append(args, *f.DivisionID)
	}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil {
			return Result[OrganizationMember]{}, ErrOrganizationForbidden
		}
		clauses = append(clauses, "m.division_id=?")
		args = append(args, *user.DivisionID)
	}
	where := " WHERE " + strings.Join(clauses, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM organization_members m`+where, args...).Scan(&total); err != nil {
		return Result[OrganizationMember]{}, err
	}
	query := `SELECT m.id,m.period_id,p.name,m.division_id,d.name,m.name,m.nim,m.position,m.photo_url,m.display_order,m.status,m.created_at,m.updated_at FROM organization_members m JOIN periods p ON p.id=m.period_id LEFT JOIN divisions d ON d.id=m.division_id` + where + ` ORDER BY m.period_id DESC,FIELD(m.position,'KETUA','WAKIL_KETUA','SEKRETARIS','BENDAHARA','KEPALA_BIDANG','ANGGOTA'),m.display_order,m.name,m.id LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), f.Limit, (f.Page-1)*f.Limit)
	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return Result[OrganizationMember]{}, err
	}
	defer rows.Close()
	items, err := scanAdminMembers(rows)
	return Result[OrganizationMember]{Items: items, Page: Page{Page: f.Page, Limit: f.Limit, Total: total, TotalPages: (total + f.Limit - 1) / f.Limit}}, err
}

func scanAdminMembers(rows *sql.Rows) ([]OrganizationMember, error) {
	items := []OrganizationMember{}
	for rows.Next() {
		var m OrganizationMember
		var d, nim, photo sql.NullString
		var did sql.NullInt64
		if err := rows.Scan(&m.ID, &m.PeriodID, &m.PeriodName, &did, &d, &m.Name, &nim, &m.Position, &photo, &m.DisplayOrder, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		m.DivisionID, m.DivisionName, m.NIM, m.PhotoURL = ptrInt64(did), ptrString(d), ptrString(nim), ptrString(photo)
		items = append(items, m)
	}
	return items, rows.Err()
}

func (s *OrganizationService) GetMember(ctx context.Context, id int64, user domain.User) (OrganizationMember, error) {
	query := `SELECT m.id,m.period_id,p.name,m.division_id,d.name,m.name,m.nim,m.position,m.photo_url,m.display_order,m.status,m.created_at,m.updated_at FROM organization_members m JOIN periods p ON p.id=m.period_id LEFT JOIN divisions d ON d.id=m.division_id WHERE m.id=?`
	args := []any{id}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil {
			return OrganizationMember{}, ErrOrganizationForbidden
		}
		query += ` AND m.division_id=?`
		args = append(args, *user.DivisionID)
	}
	var m OrganizationMember
	var divisionID sql.NullInt64
	var divisionName, nim, photo sql.NullString
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&m.ID, &m.PeriodID, &m.PeriodName, &divisionID, &divisionName, &m.Name, &nim, &m.Position, &photo, &m.DisplayOrder, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return m, ErrOrganizationNotFound
	}
	m.DivisionID, m.DivisionName, m.NIM, m.PhotoURL = ptrInt64(divisionID), ptrString(divisionName), ptrString(nim), ptrString(photo)
	return m, err
}

func validateMemberInput(in *MemberInput) error {
	in.Name = strings.TrimSpace(in.Name)
	in.Position = strings.ToUpper(strings.TrimSpace(in.Position))
	in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	if in.Status == "" {
		in.Status = "ACTIVE"
	}
	if in.Name == "" || len(in.Name) > 160 || in.PeriodID < 1 || organizationPositions[in.Position] == 0 || in.DisplayOrder < 0 || (in.Status != "ACTIVE" && in.Status != "INACTIVE") {
		return ErrOrganizationInvalid
	}
	if in.Position == "KEPALA_BIDANG" && in.DivisionID == nil {
		return ErrOrganizationInvalid
	}
	if in.Position == "ANGGOTA" && in.DivisionID == nil {
		return ErrOrganizationInvalid
	}
	if in.DivisionID != nil && *in.DivisionID < 1 {
		return ErrOrganizationInvalid
	}
	if in.NIM != nil {
		v := strings.TrimSpace(*in.NIM)
		if v == "" {
			in.NIM = nil
		} else if len(v) > 32 {
			return ErrOrganizationInvalid
		} else {
			in.NIM = &v
		}
	}
	if in.PhotoURL != nil {
		v := strings.TrimSpace(*in.PhotoURL)
		if v == "" {
			in.PhotoURL = nil
		} else {
			u, e := url.ParseRequestURI(v)
			if e != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
				return ErrOrganizationInvalid
			}
			in.PhotoURL = &v
		}
	}
	return nil
}

func (s *OrganizationService) SaveMember(ctx context.Context, user domain.User, id int64, in MemberInput) (OrganizationMember, error) {
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH {
		return OrganizationMember{}, ErrOrganizationForbidden
	}
	if err := validateMemberInput(&in); err != nil {
		return OrganizationMember{}, err
	}
	if in.DivisionID != nil {
		var period int64
		if err := s.db.QueryRowContext(ctx, `SELECT period_id FROM divisions WHERE id=?`, *in.DivisionID).Scan(&period); errors.Is(err, sql.ErrNoRows) {
			return OrganizationMember{}, ErrOrganizationInvalid
		} else if err != nil {
			return OrganizationMember{}, err
		}
		if period != in.PeriodID {
			return OrganizationMember{}, ErrOrganizationInvalid
		}
	}
	created := id == 0
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return OrganizationMember{}, err
	}
	defer tx.Rollback()
	if created {
		res, e := tx.ExecContext(ctx, `INSERT INTO organization_members(period_id,division_id,name,nim,position,photo_url,display_order,status) VALUES(?,?,?,?,?,?,?,?)`, in.PeriodID, in.DivisionID, in.Name, in.NIM, in.Position, in.PhotoURL, in.DisplayOrder, in.Status)
		if e != nil {
			return OrganizationMember{}, mapOrganizationDBError(e)
		}
		id, _ = res.LastInsertId()
	} else {
		var exists int
		if e := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM organization_members WHERE id=? FOR UPDATE`, id).Scan(&exists); e != nil {
			return OrganizationMember{}, e
		}
		if exists == 0 {
			return OrganizationMember{}, ErrOrganizationNotFound
		}
		if _, e := tx.ExecContext(ctx, `UPDATE organization_members SET period_id=?,division_id=?,name=?,nim=?,position=?,photo_url=?,display_order=?,status=? WHERE id=?`, in.PeriodID, in.DivisionID, in.Name, in.NIM, in.Position, in.PhotoURL, in.DisplayOrder, in.Status, id); e != nil {
			return OrganizationMember{}, mapOrganizationDBError(e)
		}
	}
	if err := s.writeActivity(ctx, tx, user, actionForCreated(created, "UPDATE"), "ORGANIZATION_MEMBER", id, map[string]any{"name": in.Name, "position": in.Position, "status": in.Status}); err != nil {
		return OrganizationMember{}, err
	}
	if err := tx.Commit(); err != nil {
		return OrganizationMember{}, err
	}
	return s.GetMember(ctx, id, user)
}

func (s *OrganizationService) DeleteMember(ctx context.Context, user domain.User, id int64) error {
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH {
		return ErrOrganizationForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var name, position, status string
	if err := tx.QueryRowContext(ctx, `SELECT name,position,status FROM organization_members WHERE id=? FOR UPDATE`, id).Scan(&name, &position, &status); errors.Is(err, sql.ErrNoRows) {
		return ErrOrganizationNotFound
	} else if err != nil {
		return err
	}
	if status == "INACTIVE" {
		return tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, `UPDATE organization_members SET status='INACTIVE' WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.writeActivity(ctx, tx, user, "DELETE", "ORGANIZATION_MEMBER", id, map[string]any{"name": name, "position": position, "status": "INACTIVE"}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *OrganizationService) writeActivity(ctx context.Context, tx *sql.Tx, user domain.User, action, entity string, id int64, details map[string]any) error {
	return s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: action, EntityType: entity, EntityID: fmt.Sprint(id), Details: details})
}

func mapOrganizationDBError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "duplicate") || strings.Contains(message, "foreign key") || strings.Contains(message, "constraint") || strings.Contains(message, "deadlock") {
		return fmt.Errorf("%w: %v", ErrOrganizationConflict, err)
	}
	return err
}
