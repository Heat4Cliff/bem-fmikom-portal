package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/security"
	"github.com/go-sql-driver/mysql"
)

var (
	ErrSystemNotFound  = errors.New("system resource not found")
	ErrSystemForbidden = errors.New("system action forbidden")
	ErrSystemConflict  = errors.New("system action conflict")
	ErrSystemInvalid   = errors.New("system input invalid")
)

type LandingSettings struct {
	HeroHeadline          string     `json:"hero_headline"`
	HeroSubheadline       string     `json:"hero_subheadline"`
	HeroPrimaryCTALabel   string     `json:"hero_primary_cta_label"`
	HeroPrimaryCTAURL     string     `json:"hero_primary_cta_url"`
	HeroSecondaryCTALabel *string    `json:"hero_secondary_cta_label"`
	HeroSecondaryCTAURL   *string    `json:"hero_secondary_cta_url"`
	HeroImageURL          *string    `json:"hero_image_url"`
	AboutTitle            string     `json:"about_title"`
	AboutDescription      string     `json:"about_description"`
	Vision                string     `json:"vision"`
	Missions              []string   `json:"missions"`
	ContactEmail          *string    `json:"contact_email"`
	ContactPhone          *string    `json:"contact_phone"`
	ContactAddress        *string    `json:"contact_address"`
	LocationURL           *string    `json:"location_url"`
	InstagramURL          *string    `json:"instagram_url"`
	YouTubeURL            *string    `json:"youtube_url"`
	TikTokURL             *string    `json:"tiktok_url"`
	OtherSocialURL        *string    `json:"other_social_url"`
	UpdatedAt             *time.Time `json:"updated_at,omitempty"`
}

type UserRecord struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Email       string           `json:"email"`
	Role        string           `json:"role"`
	DivisionID  *int64           `json:"division_id"`
	Division    *domain.Division `json:"division"`
	Status      string           `json:"status"`
	LastLoginAt *time.Time       `json:"last_login_at"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type UserInput struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	DivisionID *int64 `json:"division_id"`
	Status     string `json:"status"`
	Password   string `json:"password,omitempty"`
}

type UserFilters struct {
	Page       int
	Limit      int
	Search     string
	Role       string
	Status     string
	DivisionID *int64
}

type ActivityRecord struct {
	ID         int64          `json:"id"`
	Timestamp  time.Time      `json:"timestamp"`
	Actor      *ActivityActor `json:"actor"`
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type"`
	EntityID   *string        `json:"entity_id"`
	Details    any            `json:"details"`
}
type ActivityActor struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
type ActivityFilters struct {
	Page, Limit                                               int
	Search, ActorUserID, Action, EntityType, DateFrom, DateTo string
}

type SystemManagementService struct {
	db         *sql.DB
	log        *repository.ActivityLogs
	bcryptCost int
}

func NewSystemManagementService(db *sql.DB, bcryptCost int) *SystemManagementService {
	return &SystemManagementService{db: db, log: repository.NewActivityLogs(db), bcryptCost: bcryptCost}
}

func (s *SystemManagementService) GetLandingSettings(ctx context.Context) (LandingSettings, error) {
	var value LandingSettings
	var missions []byte
	var updated sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT hero_headline,hero_subheadline,hero_primary_cta_label,hero_primary_cta_url,hero_secondary_cta_label,hero_secondary_cta_url,hero_image_url,about_title,about_description,vision,missions_json,contact_email,contact_phone,contact_address,location_url,instagram_url,youtube_url,tiktok_url,other_social_url,updated_at FROM landing_settings WHERE id=1`).Scan(
		&value.HeroHeadline, &value.HeroSubheadline, &value.HeroPrimaryCTALabel, &value.HeroPrimaryCTAURL, &value.HeroSecondaryCTALabel, &value.HeroSecondaryCTAURL, &value.HeroImageURL, &value.AboutTitle, &value.AboutDescription, &value.Vision, &missions, &value.ContactEmail, &value.ContactPhone, &value.ContactAddress, &value.LocationURL, &value.InstagramURL, &value.YouTubeURL, &value.TikTokURL, &value.OtherSocialURL, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return LandingSettings{}, ErrSystemNotFound
	}
	if err != nil {
		return LandingSettings{}, fmt.Errorf("read landing settings")
	}
	if err := json.Unmarshal(missions, &value.Missions); err != nil {
		return LandingSettings{}, fmt.Errorf("decode landing missions")
	}
	if value.Missions == nil {
		value.Missions = []string{}
	}
	if updated.Valid {
		t := updated.Time.UTC()
		value.UpdatedAt = &t
	}
	return value, nil
}

func (s *SystemManagementService) SaveLandingSettings(ctx context.Context, actor domain.User, in LandingSettings) (LandingSettings, error) {
	if !RoleAllowed(actor.Role, domain.RoleSuperAdmin, domain.RoleBPH) {
		return LandingSettings{}, ErrSystemForbidden
	}
	if err := validateLandingSettings(in); err != nil {
		return LandingSettings{}, err
	}
	old, err := s.GetLandingSettings(ctx)
	if err != nil {
		return LandingSettings{}, err
	}
	missions, _ := json.Marshal(in.Missions)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return LandingSettings{}, fmt.Errorf("begin settings update")
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `UPDATE landing_settings SET hero_headline=?,hero_subheadline=?,hero_primary_cta_label=?,hero_primary_cta_url=?,hero_secondary_cta_label=?,hero_secondary_cta_url=?,hero_image_url=?,about_title=?,about_description=?,vision=?,missions_json=?,contact_email=?,contact_phone=?,contact_address=?,location_url=?,instagram_url=?,youtube_url=?,tiktok_url=?,other_social_url=?,updated_by=? WHERE id=1`, in.HeroHeadline, in.HeroSubheadline, in.HeroPrimaryCTALabel, in.HeroPrimaryCTAURL, in.HeroSecondaryCTALabel, in.HeroSecondaryCTAURL, in.HeroImageURL, in.AboutTitle, in.AboutDescription, in.Vision, missions, in.ContactEmail, in.ContactPhone, in.ContactAddress, in.LocationURL, in.InstagramURL, in.YouTubeURL, in.TikTokURL, in.OtherSocialURL, actor.ID)
	if err != nil {
		return LandingSettings{}, fmt.Errorf("update landing settings")
	}
	changed := changedLandingFields(old, in)
	if len(changed) > 0 {
		actorID := actor.ID
		if err = s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &actorID, Action: "UPDATE", EntityType: "LANDING_PAGE_SETTINGS", EntityID: "1", Details: map[string]any{"changed_fields": changed}}); err != nil {
			return LandingSettings{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return LandingSettings{}, fmt.Errorf("commit landing settings")
	}
	return s.GetLandingSettings(ctx)
}

func validateLandingSettings(v LandingSettings) error {
	bounded := func(x string, max int) bool { return strings.TrimSpace(x) != "" && utf8.RuneCountInString(x) <= max }
	if !bounded(v.HeroHeadline, 180) || !bounded(v.HeroSubheadline, 500) || !bounded(v.HeroPrimaryCTALabel, 80) || !validActionURL(v.HeroPrimaryCTAURL) || !bounded(v.AboutTitle, 180) || utf8.RuneCountInString(v.AboutDescription) > 5000 || utf8.RuneCountInString(v.Vision) > 3000 || v.Missions == nil || len(v.Missions) > 10 {
		return ErrSystemInvalid
	}
	if (v.HeroSecondaryCTALabel == nil) != (v.HeroSecondaryCTAURL == nil) {
		return ErrSystemInvalid
	}
	if v.HeroSecondaryCTALabel != nil && (!bounded(*v.HeroSecondaryCTALabel, 80) || !validActionURL(*v.HeroSecondaryCTAURL)) {
		return ErrSystemInvalid
	}
	for _, mission := range v.Missions {
		if strings.TrimSpace(mission) == "" || utf8.RuneCountInString(mission) > 300 {
			return ErrSystemInvalid
		}
	}
	if !validExternalURL(v.HeroImageURL) || !validExternalURL(v.LocationURL) || !validExternalURL(v.InstagramURL) || !validExternalURL(v.YouTubeURL) || !validExternalURL(v.TikTokURL) || !validExternalURL(v.OtherSocialURL) {
		return ErrSystemInvalid
	}
	if v.ContactEmail != nil {
		e := strings.TrimSpace(*v.ContactEmail)
		parsed, err := mail.ParseAddress(e)
		if err != nil || parsed.Address != e || len(e) > 254 {
			return ErrSystemInvalid
		}
	}
	if v.ContactPhone != nil && utf8.RuneCountInString(*v.ContactPhone) > 64 {
		return ErrSystemInvalid
	}
	if v.ContactAddress != nil && utf8.RuneCountInString(*v.ContactAddress) > 500 {
		return ErrSystemInvalid
	}
	return nil
}
func validActionURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 500 || strings.HasPrefix(raw, "//") {
		return false
	}
	if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "#") {
		return true
	}
	return validURLString(raw)
}
func validExternalURL(raw *string) bool {
	if raw == nil {
		return true
	}
	if *raw == "" || len(*raw) > 2048 {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(*raw))
	return err == nil && u.Host != "" && u.User == nil && strings.EqualFold(u.Scheme, "https")
}
func validURLString(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Host != "" && u.User == nil && strings.EqualFold(u.Scheme, "https")
}
func changedLandingFields(a, b LandingSettings) []string {
	out := []string{}
	checks := []struct {
		name    string
		changed bool
	}{{"hero_headline", a.HeroHeadline != b.HeroHeadline}, {"hero_subheadline", a.HeroSubheadline != b.HeroSubheadline}, {"hero_primary_cta_label", a.HeroPrimaryCTALabel != b.HeroPrimaryCTALabel}, {"hero_primary_cta_url", a.HeroPrimaryCTAURL != b.HeroPrimaryCTAURL}, {"hero_secondary_cta_label", !sameString(a.HeroSecondaryCTALabel, b.HeroSecondaryCTALabel)}, {"hero_secondary_cta_url", !sameString(a.HeroSecondaryCTAURL, b.HeroSecondaryCTAURL)}, {"hero_image_url", !sameString(a.HeroImageURL, b.HeroImageURL)}, {"about_title", a.AboutTitle != b.AboutTitle}, {"about_description", a.AboutDescription != b.AboutDescription}, {"vision", a.Vision != b.Vision}, {"missions", !sameMissions(a.Missions, b.Missions)}, {"contact_email", !sameString(a.ContactEmail, b.ContactEmail)}, {"contact_phone", !sameString(a.ContactPhone, b.ContactPhone)}, {"contact_address", !sameString(a.ContactAddress, b.ContactAddress)}, {"location_url", !sameString(a.LocationURL, b.LocationURL)}, {"instagram_url", !sameString(a.InstagramURL, b.InstagramURL)}, {"youtube_url", !sameString(a.YouTubeURL, b.YouTubeURL)}, {"tiktok_url", !sameString(a.TikTokURL, b.TikTokURL)}, {"other_social_url", !sameString(a.OtherSocialURL, b.OtherSocialURL)}}
	for _, c := range checks {
		if c.changed {
			out = append(out, c.name)
		}
	}
	return out
}
func sameString(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
func sameMissions(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (s *SystemManagementService) ListUsers(ctx context.Context, f UserFilters) (Result[UserRecord], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if strings.TrimSpace(f.Search) != "" {
		where = append(where, "(u.name LIKE ? OR u.email LIKE ?)")
		q := "%" + strings.TrimSpace(f.Search) + "%"
		args = append(args, q, q)
	}
	if f.Role != "" {
		if f.Role != "SUPER_ADMIN" && f.Role != "BPH" && f.Role != "KEPALA_BIDANG" {
			return Result[UserRecord]{}, ErrSystemInvalid
		}
		where = append(where, "u.role=?")
		args = append(args, f.Role)
	}
	if f.Status != "" {
		if f.Status != "ACTIVE" && f.Status != "INACTIVE" {
			return Result[UserRecord]{}, ErrSystemInvalid
		}
		where = append(where, "u.status=?")
		args = append(args, f.Status)
	}
	if f.DivisionID != nil {
		where = append(where, "u.division_id=?")
		args = append(args, *f.DivisionID)
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users u WHERE "+clause, args...).Scan(&total); err != nil {
		return Result[UserRecord]{}, fmt.Errorf("count users")
	}
	listArgs := append(append([]any{}, args...), f.Limit, (f.Page-1)*f.Limit)
	rows, err := s.db.QueryContext(ctx, `SELECT u.id,u.name,u.email,u.role,u.division_id,d.name,u.status,u.last_login_at,u.created_at,u.updated_at FROM users u LEFT JOIN divisions d ON d.id=u.division_id WHERE `+clause+` ORDER BY u.created_at DESC,u.id DESC LIMIT ? OFFSET ?`, listArgs...)
	if err != nil {
		return Result[UserRecord]{}, fmt.Errorf("list users")
	}
	defer rows.Close()
	items := []UserRecord{}
	for rows.Next() {
		v, e := scanUserRecord(rows)
		if e != nil {
			return Result[UserRecord]{}, e
		}
		items = append(items, v)
	}
	if err = rows.Err(); err != nil {
		return Result[UserRecord]{}, fmt.Errorf("iterate users")
	}
	return Result[UserRecord]{Items: items, Page: Page{Page: f.Page, Limit: f.Limit, Total: total, TotalPages: (total + f.Limit - 1) / f.Limit}}, nil
}
func (s *SystemManagementService) GetUser(ctx context.Context, id int64) (UserRecord, error) {
	row := s.db.QueryRowContext(ctx, `SELECT u.id,u.name,u.email,u.role,u.division_id,d.name,u.status,u.last_login_at,u.created_at,u.updated_at FROM users u LEFT JOIN divisions d ON d.id=u.division_id WHERE u.id=?`, id)
	v, e := scanUserRecord(row)
	if errors.Is(e, sql.ErrNoRows) {
		return UserRecord{}, ErrSystemNotFound
	}
	return v, e
}

type rowScanner interface{ Scan(...any) error }

func scanUserRecord(row rowScanner) (UserRecord, error) {
	var v UserRecord
	var did sql.NullInt64
	var dname sql.NullString
	var last sql.NullTime
	if e := row.Scan(&v.ID, &v.Name, &v.Email, &v.Role, &did, &dname, &v.Status, &last, &v.CreatedAt, &v.UpdatedAt); e != nil {
		return v, e
	}
	if did.Valid {
		v.DivisionID = &did.Int64
		v.Division = &domain.Division{ID: did.Int64, Name: dname.String}
	}
	if last.Valid {
		t := last.Time.UTC()
		v.LastLoginAt = &t
	}
	v.CreatedAt = v.CreatedAt.UTC()
	v.UpdatedAt = v.UpdatedAt.UTC()
	return v, nil
}
func (s *SystemManagementService) CreateUser(ctx context.Context, actor domain.User, in UserInput) (UserRecord, error) {
	if !RoleAllowed(actor.Role, domain.RoleSuperAdmin) {
		return UserRecord{}, ErrSystemForbidden
	}
	name, email, role, status, division, e := normalizeUserInput(in)
	if e != nil {
		return UserRecord{}, e
	}
	hash, e := security.HashPassword(in.Password, s.bcryptCost)
	if e != nil {
		return UserRecord{}, ErrSystemInvalid
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return UserRecord{}, fmt.Errorf("begin user create")
	}
	defer tx.Rollback()
	if e = validateDivisionTx(ctx, tx, role, division); e != nil {
		return UserRecord{}, e
	}
	result, e := tx.ExecContext(ctx, `INSERT INTO users(name,email,password_hash,role,division_id,status) VALUES(?,?,?,?,?,?)`, name, email, hash, role, division, status)
	if e != nil {
		var me *mysql.MySQLError
		if errors.As(e, &me) && me.Number == 1062 {
			return UserRecord{}, ErrSystemConflict
		}
		return UserRecord{}, fmt.Errorf("create user")
	}
	id, e := result.LastInsertId()
	if e != nil {
		return UserRecord{}, fmt.Errorf("read user id")
	}
	actorID := actor.ID
	if e = s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &actorID, Action: "CREATE", EntityType: "USER", EntityID: fmt.Sprint(id), Details: map[string]any{"user_id": id, "role": role, "status": status, "division_id": division}}); e != nil {
		return UserRecord{}, e
	}
	if e = tx.Commit(); e != nil {
		return UserRecord{}, fmt.Errorf("commit user create")
	}
	return s.GetUser(ctx, id)
}
func (s *SystemManagementService) UpdateUser(ctx context.Context, actor domain.User, id int64, in UserInput) (UserRecord, error) {
	if !RoleAllowed(actor.Role, domain.RoleSuperAdmin) {
		return UserRecord{}, ErrSystemForbidden
	}
	if id == actor.ID && (strings.TrimSpace(in.Role) != "SUPER_ADMIN" || strings.TrimSpace(in.Status) != "ACTIVE") {
		return UserRecord{}, ErrSystemConflict
	}
	name, email, role, status, division, e := normalizeUserInput(in)
	if e != nil {
		return UserRecord{}, e
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return UserRecord{}, fmt.Errorf("begin user update")
	}
	defer tx.Rollback()
	if e = lockSuperAdmins(ctx, tx); e != nil {
		return UserRecord{}, e
	}
	before, e := lockUser(ctx, tx, id)
	if errors.Is(e, sql.ErrNoRows) {
		return UserRecord{}, ErrSystemNotFound
	}
	if e != nil {
		return UserRecord{}, e
	}
	if e = validateDivisionTx(ctx, tx, role, division); e != nil {
		return UserRecord{}, e
	}
	if before.Role == "SUPER_ADMIN" && before.Status == "ACTIVE" && (role != "SUPER_ADMIN" || status != "ACTIVE") {
		var active int
		if e = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role='SUPER_ADMIN' AND status='ACTIVE'`).Scan(&active); e != nil {
			return UserRecord{}, e
		}
		if active <= 1 {
			return UserRecord{}, ErrSystemConflict
		}
	}
	_, e = tx.ExecContext(ctx, `UPDATE users SET name=?,email=?,role=?,division_id=?,status=? WHERE id=?`, name, email, role, division, status, id)
	if e != nil {
		var me *mysql.MySQLError
		if errors.As(e, &me) && me.Number == 1062 {
			return UserRecord{}, ErrSystemConflict
		}
		return UserRecord{}, fmt.Errorf("update user")
	}
	if status == "INACTIVE" && before.Status != "INACTIVE" {
		if _, e = tx.ExecContext(ctx, `UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,UTC_TIMESTAMP(6)) WHERE user_id=?`, id); e != nil {
			return UserRecord{}, fmt.Errorf("revoke user sessions")
		}
	}
	details := map[string]any{"user_id": id}
	changed := []string{}
	if before.Name != name {
		changed = append(changed, "name")
	}
	if before.Email != email {
		changed = append(changed, "email")
	}
	if before.Role != role {
		changed = append(changed, "role")
		details["role_change"] = map[string]string{"from": before.Role, "to": role}
	}
	if !sameNullableInt64(before.DivisionID, division) {
		changed = append(changed, "division_id")
		details["division_change"] = true
	}
	if before.Status != status {
		changed = append(changed, "status")
		details["status_change"] = map[string]string{"from": before.Status, "to": status}
	}
	if len(changed) > 0 {
		details["changed_fields"] = changed
		actorID := actor.ID
		action := "UPDATE"
		if before.Status != status {
			if status == "INACTIVE" {
				action = "DEACTIVATE"
			} else {
				action = "ACTIVATE"
			}
		}
		if e = s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &actorID, Action: action, EntityType: "USER", EntityID: fmt.Sprint(id), Details: details}); e != nil {
			return UserRecord{}, e
		}
	}
	if e = tx.Commit(); e != nil {
		return UserRecord{}, fmt.Errorf("commit user update")
	}
	return s.GetUser(ctx, id)
}
func (s *SystemManagementService) DeactivateUser(ctx context.Context, actor domain.User, id int64) error {
	return s.setUserStatus(ctx, actor, id, "INACTIVE")
}
func (s *SystemManagementService) ReactivateUser(ctx context.Context, actor domain.User, id int64) error {
	return s.setUserStatus(ctx, actor, id, "ACTIVE")
}

func (s *SystemManagementService) setUserStatus(ctx context.Context, actor domain.User, id int64, status string) error {
	if actor.Role != domain.RoleSuperAdmin {
		return ErrSystemForbidden
	}
	if status != "ACTIVE" && status != "INACTIVE" {
		return ErrSystemInvalid
	}
	if actor.ID == id && status == "INACTIVE" {
		return ErrSystemConflict
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin user status change")
	}
	defer tx.Rollback()
	if err = lockSuperAdmins(ctx, tx); err != nil {
		return err
	}
	user, err := lockUser(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSystemNotFound
	}
	if err != nil {
		return err
	}
	if user.Status == status {
		return nil
	}
	if user.Role == "SUPER_ADMIN" && user.Status == "ACTIVE" && status == "INACTIVE" {
		var active int
		if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role='SUPER_ADMIN' AND status='ACTIVE'`).Scan(&active); err != nil {
			return err
		}
		if active <= 1 {
			return ErrSystemConflict
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE users SET status=? WHERE id=?`, status, id); err != nil {
		return fmt.Errorf("update user status")
	}
	if status == "INACTIVE" {
		if _, err = tx.ExecContext(ctx, `UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,UTC_TIMESTAMP(6)) WHERE user_id=?`, id); err != nil {
			return fmt.Errorf("revoke user sessions")
		}
	}
	actorID := actor.ID
	action := "ACTIVATE"
	if status == "INACTIVE" {
		action = "DEACTIVATE"
	}
	if err = s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &actorID, Action: action, EntityType: "USER", EntityID: fmt.Sprint(id), Details: map[string]any{"user_id": id, "changed_fields": []string{"status"}, "status_change": map[string]string{"from": user.Status, "to": status}}}); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit user status change")
	}
	return nil
}
func (s *SystemManagementService) ResetPassword(ctx context.Context, actor domain.User, id int64, password string) error {
	if actor.Role != domain.RoleSuperAdmin {
		return ErrSystemForbidden
	}
	hash, e := security.HashPassword(password, s.bcryptCost)
	if e != nil {
		return ErrSystemInvalid
	}
	tx, e := s.db.BeginTx(ctx, nil)
	if e != nil {
		return fmt.Errorf("begin password reset")
	}
	defer tx.Rollback()
	var exists int
	if e = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE id=? FOR UPDATE`, id).Scan(&exists); errors.Is(e, sql.ErrNoRows) {
		return ErrSystemNotFound
	}
	if e != nil {
		return e
	}
	if _, e = tx.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=?`, hash, id); e != nil {
		return fmt.Errorf("reset user password")
	}
	if _, e = tx.ExecContext(ctx, `UPDATE refresh_sessions SET revoked_at=COALESCE(revoked_at,UTC_TIMESTAMP(6)) WHERE user_id=?`, id); e != nil {
		return fmt.Errorf("revoke password reset sessions")
	}
	actorID := actor.ID
	if e = s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &actorID, Action: "UPDATE", EntityType: "USER", EntityID: fmt.Sprint(id), Details: map[string]any{"user_id": id, "operation": "reset_password", "changed_fields": []string{"password"}, "refresh_sessions_revoked": true}}); e != nil {
		return e
	}
	if e = tx.Commit(); e != nil {
		return fmt.Errorf("commit password reset")
	}
	return nil
}

func normalizeUserInput(in UserInput) (string, string, string, string, *int64, error) {
	name := strings.TrimSpace(in.Name)
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if name == "" || utf8.RuneCountInString(name) > 160 {
		return "", "", "", "", nil, ErrSystemInvalid
	}
	m, e := mail.ParseAddress(email)
	if e != nil || m.Address != email || len(email) > 254 {
		return "", "", "", "", nil, ErrSystemInvalid
	}
	role := strings.ToUpper(strings.TrimSpace(in.Role))
	if role != "SUPER_ADMIN" && role != "BPH" && role != "KEPALA_BIDANG" {
		return "", "", "", "", nil, ErrSystemInvalid
	}
	status := strings.ToUpper(strings.TrimSpace(in.Status))
	if status != "ACTIVE" && status != "INACTIVE" {
		return "", "", "", "", nil, ErrSystemInvalid
	}
	division := in.DivisionID
	if role == "KEPALA_BIDANG" {
		if division == nil || *division < 1 {
			return "", "", "", "", nil, ErrSystemInvalid
		}
	} else if division != nil {
		return "", "", "", "", nil, ErrSystemInvalid
	}
	return name, email, role, status, division, nil
}
func validateDivisionTx(ctx context.Context, tx *sql.Tx, role string, division *int64) error {
	if role != "KEPALA_BIDANG" {
		return nil
	}
	var count int
	e := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM divisions WHERE id=? AND status='ACTIVE'`, *division).Scan(&count)
	if e != nil {
		return e
	}
	if count != 1 {
		return ErrSystemInvalid
	}
	return nil
}

type lockedUser struct {
	Name, Email, Role, Status string
	DivisionID                *int64
}

func lockUser(ctx context.Context, tx *sql.Tx, id int64) (lockedUser, error) {
	var v lockedUser
	var d sql.NullInt64
	e := tx.QueryRowContext(ctx, `SELECT name,email,role,status,division_id FROM users WHERE id=? FOR UPDATE`, id).Scan(&v.Name, &v.Email, &v.Role, &v.Status, &d)
	if d.Valid {
		v.DivisionID = &d.Int64
	}
	return v, e
}
func lockSuperAdmins(ctx context.Context, tx *sql.Tx) error {
	rows, e := tx.QueryContext(ctx, `SELECT id FROM users WHERE role='SUPER_ADMIN' ORDER BY id FOR UPDATE`)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if e = rows.Scan(&id); e != nil {
			return e
		}
	}
	return rows.Err()
}
func sameNullableInt64(a, b *int64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func (s *SystemManagementService) ListActivity(ctx context.Context, f ActivityFilters) (Result[ActivityRecord], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	where := []string{"1=1"}
	args := []any{}
	if f.Search != "" {
		where = append(where, "(COALESCE(u.name,'') LIKE ? OR COALESCE(u.email,'') LIKE ? OR l.entity_type LIKE ? OR COALESCE(l.entity_id,'') LIKE ? OR COALESCE(CAST(l.details_json AS CHAR),'') LIKE ?)")
		q := "%" + strings.TrimSpace(f.Search) + "%"
		for i := 0; i < 5; i++ {
			args = append(args, q)
		}
	}
	if f.ActorUserID != "" {
		where = append(where, "l.actor_user_id=?")
		args = append(args, f.ActorUserID)
	}
	if f.Action != "" {
		where = append(where, "l.action=?")
		args = append(args, strings.ToUpper(f.Action))
	}
	if f.EntityType != "" {
		where = append(where, "l.entity_type=?")
		args = append(args, strings.ToUpper(f.EntityType))
	}
	if f.DateFrom != "" {
		where = append(where, "l.created_at>=?")
		args = append(args, f.DateFrom+" 00:00:00")
	}
	if f.DateTo != "" {
		where = append(where, "l.created_at<?")
		to, e := time.Parse("2006-01-02", f.DateTo)
		if e != nil {
			return Result[ActivityRecord]{}, ErrSystemInvalid
		}
		args = append(args, to.AddDate(0, 0, 1).Format("2006-01-02"))
	}
	clause := strings.Join(where, " AND ")
	var total int
	if e := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM activity_logs l LEFT JOIN users u ON u.id=l.actor_user_id WHERE "+clause, args...).Scan(&total); e != nil {
		return Result[ActivityRecord]{}, fmt.Errorf("count activity logs")
	}
	listArgs := append(append([]any{}, args...), f.Limit, (f.Page-1)*f.Limit)
	rows, e := s.db.QueryContext(ctx, `SELECT l.id,l.created_at,l.actor_user_id,u.name,u.email,u.role,l.action,l.entity_type,l.entity_id,l.details_json FROM activity_logs l LEFT JOIN users u ON u.id=l.actor_user_id WHERE `+clause+` ORDER BY l.created_at DESC,l.id DESC LIMIT ? OFFSET ?`, listArgs...)
	if e != nil {
		return Result[ActivityRecord]{}, fmt.Errorf("list activity logs")
	}
	defer rows.Close()
	items := []ActivityRecord{}
	for rows.Next() {
		v, e := scanActivity(rows)
		if e != nil {
			return Result[ActivityRecord]{}, e
		}
		items = append(items, v)
	}
	if e = rows.Err(); e != nil {
		return Result[ActivityRecord]{}, e
	}
	return Result[ActivityRecord]{Items: items, Page: Page{Page: f.Page, Limit: f.Limit, Total: total, TotalPages: (total + f.Limit - 1) / f.Limit}}, nil
}
func (s *SystemManagementService) GetActivity(ctx context.Context, id int64) (ActivityRecord, error) {
	row := s.db.QueryRowContext(ctx, `SELECT l.id,l.created_at,l.actor_user_id,u.name,u.email,u.role,l.action,l.entity_type,l.entity_id,l.details_json FROM activity_logs l LEFT JOIN users u ON u.id=l.actor_user_id WHERE l.id=?`, id)
	v, e := scanActivity(row)
	if errors.Is(e, sql.ErrNoRows) {
		return ActivityRecord{}, ErrSystemNotFound
	}
	return v, e
}
func scanActivity(row rowScanner) (ActivityRecord, error) {
	var v ActivityRecord
	var aid sql.NullInt64
	var name, email, role, entityID, rawDetails sql.NullString
	if e := row.Scan(&v.ID, &v.Timestamp, &aid, &name, &email, &role, &v.Action, &v.EntityType, &entityID, &rawDetails); e != nil {
		return v, e
	}
	v.Timestamp = v.Timestamp.UTC()
	if aid.Valid {
		v.Actor = &ActivityActor{ID: aid.Int64, Name: name.String, Email: email.String, Role: role.String}
	}
	if entityID.Valid {
		x := entityID.String
		v.EntityID = &x
	}
	if rawDetails.Valid {
		var value any
		if json.Unmarshal([]byte(rawDetails.String), &value) == nil {
			v.Details = redactActivityValue(value)
		} else {
			v.Details = map[string]any{"value": "[REDACTED]"}
		}
	}
	return v, nil
}

var sensitiveDetailKey = regexp.MustCompile(`(?i)(password|token|jwt|secret|cookie|authorization|storage[_-]?(path|key))`)

func redactActivityValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			if sensitiveDetailKey.MatchString(k) {
				out[k] = "[REDACTED]"
			} else {
				out[k] = redactActivityValue(item)
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = redactActivityValue(item)
		}
		return out
	default:
		return value
	}
}
