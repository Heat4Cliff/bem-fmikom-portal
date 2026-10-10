package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
)

var (
	ErrRegistrationNotFound  = errors.New("registration not found")
	ErrRegistrationForbidden = errors.New("registration access forbidden")
	ErrRegistrationConflict  = errors.New("registration conflict")
	ErrRegistrationInvalid   = errors.New("registration validation failed")
	ErrRegistrationDuplicate = errors.New("duplicate registration submission")
	ErrRegistrationLocked    = errors.New("registration schema is locked")
	ErrRegistrationHasData   = errors.New("registration has submissions")
)

const (
	RegistrationMaxFields       = 50
	RegistrationMaxOptions      = 100
	RegistrationMaxText         = 500
	RegistrationMaxTextArea     = 5000
	RegistrationMaxName         = 160
	RegistrationMaxNIM          = 32
	RegistrationMaxEmail        = 254
	RegistrationMaxPhone        = 32
	RegistrationMaxExportRows   = 1000
	RegistrationMaxRequestBytes = 64 << 10
)

var (
	registrationFieldKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*$`)
	registrationSlugPattern     = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

type RegistrationService struct {
	db  *sql.DB
	log *repository.ActivityLogs
}

func NewRegistrationService(db *sql.DB) *RegistrationService {
	return &RegistrationService{db: db, log: repository.NewActivityLogs(db)}
}

type RegistrationForm struct {
	ID              int64     `json:"id"`
	PeriodID        int64     `json:"period_id"`
	PeriodName      string    `json:"period_name,omitempty"`
	DivisionID      *int64    `json:"division_id"`
	DivisionName    *string   `json:"division_name"`
	Title           string    `json:"title"`
	Slug            string    `json:"slug"`
	Description     *string   `json:"description"`
	OpenAt          time.Time `json:"open_at"`
	CloseAt         time.Time `json:"close_at"`
	Status          string    `json:"status"`
	PublicStatus    string    `json:"public_status,omitempty"`
	SubmissionCount int       `json:"submission_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type RegistrationOption struct {
	ID           int64  `json:"id,omitempty"`
	Value        string `json:"value"`
	Label        string `json:"label"`
	DisplayOrder int    `json:"display_order"`
}

type RegistrationField struct {
	ID           int64                `json:"id,omitempty"`
	FieldKey     string               `json:"field_key"`
	Label        string               `json:"label"`
	FieldType    string               `json:"field_type"`
	Placeholder  *string              `json:"placeholder"`
	Required     bool                 `json:"required"`
	DisplayOrder int                  `json:"display_order"`
	Options      []RegistrationOption `json:"options"`
}

type PublicRegistrationField struct {
	FieldKey    string               `json:"field_key"`
	Label       string               `json:"label"`
	FieldType   string               `json:"field_type"`
	Placeholder *string              `json:"placeholder"`
	Required    bool                 `json:"required"`
	Options     []RegistrationOption `json:"options"`
}

type PublicRegistration struct {
	ID           int64                     `json:"id"`
	Title        string                    `json:"title"`
	Slug         string                    `json:"slug"`
	Description  *string                   `json:"description"`
	PeriodName   string                    `json:"period_name"`
	DivisionName *string                   `json:"division_name"`
	OpenAt       time.Time                 `json:"open_at"`
	CloseAt      time.Time                 `json:"close_at"`
	Status       string                    `json:"status"`
	Fields       []PublicRegistrationField `json:"fields"`
}

type PublicRegistrationPreview struct {
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Description  *string   `json:"description"`
	PeriodName   string    `json:"period_name"`
	DivisionName *string   `json:"division_name"`
	OpenAt       time.Time `json:"open_at"`
	CloseAt      time.Time `json:"close_at"`
	PublicStatus string    `json:"public_status"`
}

type RegistrationOptionInput struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type RegistrationFieldInput struct {
	FieldKey    string                    `json:"field_key"`
	Label       string                    `json:"label"`
	FieldType   string                    `json:"field_type"`
	Placeholder *string                   `json:"placeholder"`
	Required    bool                      `json:"required"`
	Options     []RegistrationOptionInput `json:"options"`
}

type RegistrationCreateInput struct {
	PeriodID    int64     `json:"period_id"`
	DivisionID  *int64    `json:"division_id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	OpenAt      time.Time `json:"open_at"`
	CloseAt     time.Time `json:"close_at"`
}

type RegistrationUpdateInput struct {
	PeriodID    *int64     `json:"period_id"`
	DivisionID  **int64    `json:"division_id"`
	Title       *string    `json:"title"`
	Slug        *string    `json:"slug"`
	Description **string   `json:"description"`
	OpenAt      *time.Time `json:"open_at"`
	CloseAt     *time.Time `json:"close_at"`
}

type RegistrationFilters struct {
	Page, Limit          int
	Search, Status       string
	DivisionID, PeriodID *int64
	DateFrom, DateTo     *time.Time
}

type RegistrationSubmissionFilters struct {
	Page, Limit      int
	Search           string
	DateFrom, DateTo *time.Time
}

func (s *RegistrationService) Create(ctx context.Context, user domain.User, in RegistrationCreateInput) (RegistrationForm, error) {
	in.Title = strings.TrimSpace(in.Title)
	if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleBPH && user.Role != domain.RoleKepala {
		return RegistrationForm{}, ErrRegistrationForbidden
	}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil || *user.DivisionID < 1 || (in.DivisionID != nil && *in.DivisionID != *user.DivisionID) {
			return RegistrationForm{}, ErrRegistrationForbidden
		}
		in.DivisionID = user.DivisionID
	}
	if in.PeriodID < 1 || in.Title == "" || runeLen(in.Title) > 200 || !validRegistrationDates(in.OpenAt, in.CloseAt) {
		return RegistrationForm{}, ErrRegistrationInvalid
	}
	if err := s.validatePeriodDivision(ctx, in.PeriodID, in.DivisionID); err != nil {
		return RegistrationForm{}, err
	}
	if strings.TrimSpace(in.Slug) == "" {
		in.Slug = slugifyRegistration(in.Title)
	} else {
		in.Slug = strings.TrimSpace(strings.ToLower(in.Slug))
	}
	if !registrationSlugPattern.MatchString(in.Slug) {
		return RegistrationForm{}, ErrRegistrationInvalid
	}
	if in.Description != nil {
		value := strings.TrimSpace(*in.Description)
		in.Description = &value
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RegistrationForm{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `INSERT INTO registration_forms (period_id,division_id,created_by,title,slug,description,open_at,close_at,status) VALUES (?,?,?,?,?,?,?,?,'DRAFT')`, in.PeriodID, in.DivisionID, user.ID, in.Title, in.Slug, in.Description, in.OpenAt.UTC(), in.CloseAt.UTC())
	if err != nil {
		if isRegistrationUniqueConflict(err) {
			return RegistrationForm{}, ErrRegistrationConflict
		}
		return RegistrationForm{}, err
	}
	formID, err := result.LastInsertId()
	if err != nil {
		return RegistrationForm{}, err
	}
	if err := insertRegistrationFields(ctx, tx, formID, defaultRegistrationCoreFields()); err != nil {
		return RegistrationForm{}, err
	}
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "CREATE", EntityType: "REGISTRATION_FORM", EntityID: fmt.Sprint(formID), Details: map[string]any{"form_id": formID, "division_id": nullableID(in.DivisionID), "field_count": 4}}); err != nil {
		return RegistrationForm{}, err
	}
	if err := tx.Commit(); err != nil {
		return RegistrationForm{}, err
	}
	return s.Get(ctx, user, formID)
}

func (s *RegistrationService) List(ctx context.Context, user domain.User, filters RegistrationFilters) (Result[RegistrationForm], error) {
	if err := validateRegistrationFilters(filters); err != nil {
		return Result[RegistrationForm]{}, err
	}
	if !registrationRoleAllowed(user.Role) {
		return Result[RegistrationForm]{}, ErrRegistrationForbidden
	}
	where, args, err := registrationAdminWhere(user, filters)
	if err != nil {
		return Result[RegistrationForm]{}, err
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_forms f JOIN periods p ON p.id=f.period_id LEFT JOIN divisions d ON d.id=f.division_id`+where, args...).Scan(&total); err != nil {
		return Result[RegistrationForm]{}, err
	}
	queryArgs := append(append([]any{}, args...), filters.Limit, (filters.Page-1)*filters.Limit)
	rows, err := s.db.QueryContext(ctx, `SELECT f.id,f.period_id,p.name,f.division_id,d.name,f.title,f.slug,f.description,f.open_at,f.close_at,f.status,f.created_at,f.updated_at,(SELECT COUNT(*) FROM registration_submissions s WHERE s.registration_form_id=f.id) FROM registration_forms f JOIN periods p ON p.id=f.period_id LEFT JOIN divisions d ON d.id=f.division_id`+where+` ORDER BY f.created_at DESC,f.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return Result[RegistrationForm]{}, err
	}
	defer rows.Close()
	items := []RegistrationForm{}
	for rows.Next() {
		item, err := scanRegistrationForm(rows)
		if err != nil {
			return Result[RegistrationForm]{}, err
		}
		items = append(items, item)
	}
	return Result[RegistrationForm]{Items: items, Page: pageMeta(filters.Page, filters.Limit, total)}, rows.Err()
}

func (s *RegistrationService) PublicList(ctx context.Context, filters RegistrationFilters, now time.Time) (Result[PublicRegistrationPreview], error) {
	if err := validateRegistrationFilters(filters); err != nil {
		return Result[PublicRegistrationPreview]{}, err
	}
	where, args := publicRegistrationWhere(filters, now.UTC())
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_forms f JOIN periods p ON p.id=f.period_id LEFT JOIN divisions d ON d.id=f.division_id`+where, args...).Scan(&total); err != nil {
		return Result[PublicRegistrationPreview]{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT f.title,f.slug,f.description,p.name,d.name,f.open_at,f.close_at,
	CASE WHEN f.status='CLOSED' OR ? > f.close_at THEN 'CLOSED' WHEN ? < f.open_at THEN 'UPCOMING' ELSE 'OPEN' END
	FROM registration_forms f JOIN periods p ON p.id=f.period_id LEFT JOIN divisions d ON d.id=f.division_id`+where+` ORDER BY CASE WHEN f.status='OPEN' AND ? >= f.open_at AND ? <= f.close_at THEN 0 WHEN f.status='OPEN' AND ? < f.open_at THEN 1 ELSE 2 END,
	CASE WHEN f.status='CLOSED' OR ? > f.close_at THEN f.close_at ELSE NULL END DESC,f.created_at DESC,f.id DESC LIMIT ? OFFSET ?`, func() []any {
		queryArgs := []any{now.UTC(), now.UTC()}
		queryArgs = append(queryArgs, args...)
		queryArgs = append(queryArgs, now.UTC(), now.UTC(), now.UTC(), now.UTC(), filters.Limit, (filters.Page-1)*filters.Limit)
		return queryArgs
	}()...)
	if err != nil {
		return Result[PublicRegistrationPreview]{}, err
	}
	defer rows.Close()
	items := []PublicRegistrationPreview{}
	for rows.Next() {
		var item PublicRegistrationPreview
		var divisionName sql.NullString
		if err := rows.Scan(&item.Title, &item.Slug, &item.Description, &item.PeriodName, &divisionName, &item.OpenAt, &item.CloseAt, &item.PublicStatus); err != nil {
			return Result[PublicRegistrationPreview]{}, err
		}
		item.DivisionName = nullText(divisionName)
		item.OpenAt, item.CloseAt = item.OpenAt.UTC(), item.CloseAt.UTC()
		items = append(items, item)
	}
	return Result[PublicRegistrationPreview]{Items: items, Page: pageMeta(filters.Page, filters.Limit, total)}, rows.Err()
}

func (s *RegistrationService) PublicGet(ctx context.Context, slug string, now time.Time) (PublicRegistration, error) {
	var result PublicRegistration
	var divisionName sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT f.id,f.title,f.slug,f.description,p.name,d.name,f.open_at,f.close_at,
	f.status
	FROM registration_forms f JOIN periods p ON p.id=f.period_id LEFT JOIN divisions d ON d.id=f.division_id
	WHERE f.slug=? AND f.deleted_at IS NULL AND f.status<>'DRAFT'`, strings.TrimSpace(slug)).Scan(&result.ID, &result.Title, &result.Slug, &result.Description, &result.PeriodName, &divisionName, &result.OpenAt, &result.CloseAt, &result.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return PublicRegistration{}, ErrRegistrationNotFound
	}
	if err != nil {
		return PublicRegistration{}, err
	}
	result.DivisionName = nullText(divisionName)
	result.OpenAt, result.CloseAt = result.OpenAt.UTC(), result.CloseAt.UTC()
	if !now.IsZero() {
		result.Status = effectiveRegistrationStatus(result.Status, result.OpenAt, result.CloseAt, now.UTC())
	}
	fields, err := s.loadFields(ctx, s.db, result.ID)
	if err != nil {
		return PublicRegistration{}, err
	}
	result.Fields = make([]PublicRegistrationField, len(fields))
	for i, field := range fields {
		options := make([]RegistrationOption, len(field.Options))
		for j, option := range field.Options {
			options[j] = RegistrationOption{Value: option.Value, Label: option.Label, DisplayOrder: option.DisplayOrder}
		}
		result.Fields[i] = PublicRegistrationField{FieldKey: field.FieldKey, Label: field.Label, FieldType: field.FieldType, Placeholder: field.Placeholder, Required: field.Required, Options: options}
	}
	return result, nil
}

func (s *RegistrationService) Get(ctx context.Context, user domain.User, id int64) (RegistrationForm, error) {
	if id < 1 || !registrationRoleAllowed(user.Role) {
		return RegistrationForm{}, ErrRegistrationForbidden
	}
	item, err := scanRegistrationForm(s.db.QueryRowContext(ctx, `SELECT f.id,f.period_id,p.name,f.division_id,d.name,f.title,f.slug,f.description,f.open_at,f.close_at,f.status,f.created_at,f.updated_at,(SELECT COUNT(*) FROM registration_submissions s WHERE s.registration_form_id=f.id)
	FROM registration_forms f JOIN periods p ON p.id=f.period_id LEFT JOIN divisions d ON d.id=f.division_id WHERE f.id=? AND f.deleted_at IS NULL`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return RegistrationForm{}, ErrRegistrationNotFound
	}
	if err != nil {
		return RegistrationForm{}, err
	}
	if !registrationCanAccess(user, item.DivisionID) {
		return RegistrationForm{}, ErrRegistrationNotFound
	}
	return item, nil
}

func (s *RegistrationService) Fields(ctx context.Context, user domain.User, id int64) ([]RegistrationField, error) {
	if _, err := s.Get(ctx, user, id); err != nil {
		return nil, err
	}
	return s.loadFields(ctx, s.db, id)
}

func (s *RegistrationService) Update(ctx context.Context, user domain.User, id int64, in RegistrationUpdateInput) (RegistrationForm, error) {
	if id < 1 || !registrationRoleAllowed(user.Role) {
		return RegistrationForm{}, ErrRegistrationForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RegistrationForm{}, err
	}
	defer tx.Rollback()
	form, err := lockRegistrationForm(ctx, tx, id)
	if err != nil {
		return RegistrationForm{}, err
	}
	if !registrationCanAccess(user, form.DivisionID) {
		return RegistrationForm{}, ErrRegistrationNotFound
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_submissions WHERE registration_form_id=?`, id).Scan(&count); err != nil {
		return RegistrationForm{}, err
	}
	if count > 0 && (in.PeriodID != nil || in.DivisionID != nil || in.Slug != nil || in.OpenAt != nil) {
		return RegistrationForm{}, ErrRegistrationConflict
	}
	if in.Title != nil {
		form.Title = strings.TrimSpace(*in.Title)
	}
	if in.Slug != nil {
		form.Slug = strings.ToLower(strings.TrimSpace(*in.Slug))
	}
	if in.Description != nil {
		form.Description = *in.Description
		if form.Description != nil {
			value := strings.TrimSpace(**in.Description)
			form.Description = &value
		}
	}
	if in.PeriodID != nil {
		form.PeriodID = *in.PeriodID
	}
	if in.DivisionID != nil {
		form.DivisionID = *in.DivisionID
	}
	if in.OpenAt != nil {
		form.OpenAt = in.OpenAt.UTC()
	}
	if in.CloseAt != nil {
		form.CloseAt = in.CloseAt.UTC()
	}
	if form.Title == "" || runeLen(form.Title) > 200 || !registrationSlugPattern.MatchString(form.Slug) || !validRegistrationDates(form.OpenAt, form.CloseAt) {
		return RegistrationForm{}, ErrRegistrationInvalid
	}
	if user.Role == domain.RoleKepala && (form.DivisionID == nil || user.DivisionID == nil || *form.DivisionID != *user.DivisionID) {
		return RegistrationForm{}, ErrRegistrationForbidden
	}
	if err := s.validatePeriodDivisionTx(ctx, tx, form.PeriodID, form.DivisionID); err != nil {
		return RegistrationForm{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE registration_forms SET period_id=?,division_id=?,title=?,slug=?,description=?,open_at=?,close_at=? WHERE id=?`, form.PeriodID, form.DivisionID, form.Title, form.Slug, form.Description, form.OpenAt, form.CloseAt, id); err != nil {
		if isRegistrationUniqueConflict(err) {
			return RegistrationForm{}, ErrRegistrationConflict
		}
		return RegistrationForm{}, err
	}
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "UPDATE", EntityType: "REGISTRATION_FORM", EntityID: fmt.Sprint(id), Details: map[string]any{"form_id": id, "division_id": nullableID(form.DivisionID)}}); err != nil {
		return RegistrationForm{}, err
	}
	if err := tx.Commit(); err != nil {
		return RegistrationForm{}, err
	}
	return s.Get(ctx, user, id)
}

func (s *RegistrationService) Delete(ctx context.Context, user domain.User, id int64) error {
	if id < 1 || !registrationRoleAllowed(user.Role) {
		return ErrRegistrationForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	form, err := lockRegistrationForm(ctx, tx, id)
	if err != nil {
		return err
	}
	if !registrationCanAccess(user, form.DivisionID) {
		return ErrRegistrationNotFound
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_submissions WHERE registration_form_id=?`, id).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrRegistrationHasData
	}
	if _, err := tx.ExecContext(ctx, `UPDATE registration_forms SET deleted_at=UTC_TIMESTAMP(6) WHERE id=?`, id); err != nil {
		return err
	}
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "DELETE", EntityType: "REGISTRATION_FORM", EntityID: fmt.Sprint(id), Details: map[string]any{"form_id": id, "division_id": nullableID(form.DivisionID)}}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *RegistrationService) Open(ctx context.Context, user domain.User, id int64) (RegistrationForm, error) {
	return s.changeStatus(ctx, user, id, "OPEN")
}

func (s *RegistrationService) Close(ctx context.Context, user domain.User, id int64) (RegistrationForm, error) {
	return s.changeStatus(ctx, user, id, "CLOSED")
}

func (s *RegistrationService) changeStatus(ctx context.Context, user domain.User, id int64, status string) (RegistrationForm, error) {
	if id < 1 || !registrationRoleAllowed(user.Role) {
		return RegistrationForm{}, ErrRegistrationForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RegistrationForm{}, err
	}
	defer tx.Rollback()
	form, err := lockRegistrationForm(ctx, tx, id)
	if err != nil {
		return RegistrationForm{}, err
	}
	if !registrationCanAccess(user, form.DivisionID) {
		return RegistrationForm{}, ErrRegistrationNotFound
	}
	if (status == "OPEN" && form.Status != "DRAFT") || (status == "CLOSED" && form.Status != "OPEN") {
		return RegistrationForm{}, ErrRegistrationConflict
	}
	if status == "OPEN" {
		if !validRegistrationDates(form.OpenAt, form.CloseAt) {
			return RegistrationForm{}, ErrRegistrationInvalid
		}
		if err := validateFormSchema(ctx, tx, id); err != nil {
			return RegistrationForm{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE registration_forms SET status=? WHERE id=?`, status, id); err != nil {
		return RegistrationForm{}, err
	}
	action := "ACTIVATE"
	if status == "CLOSED" {
		action = "DEACTIVATE"
	}
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: action, EntityType: "REGISTRATION_FORM", EntityID: fmt.Sprint(id), Details: map[string]any{"form_id": id, "division_id": nullableID(form.DivisionID)}}); err != nil {
		return RegistrationForm{}, err
	}
	if err := tx.Commit(); err != nil {
		return RegistrationForm{}, err
	}
	return s.Get(ctx, user, id)
}

func (s *RegistrationService) SaveFields(ctx context.Context, user domain.User, id int64, fields []RegistrationFieldInput) ([]RegistrationField, error) {
	if id < 1 || !registrationRoleAllowed(user.Role) {
		return nil, ErrRegistrationForbidden
	}
	if err := validateRegistrationFields(fields); err != nil {
		return nil, err
	}
	normalizeRegistrationFields(fields)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	form, err := lockRegistrationForm(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if !registrationCanAccess(user, form.DivisionID) {
		return nil, ErrRegistrationNotFound
	}
	var submissions int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_submissions WHERE registration_form_id=?`, id).Scan(&submissions); err != nil {
		return nil, err
	}
	oldFields, err := loadRegistrationFields(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if submissions > 0 {
		if !sameRegistrationStructure(oldFields, fields) {
			return nil, ErrRegistrationLocked
		}
		for i := range fields {
			if _, err := tx.ExecContext(ctx, `UPDATE registration_fields SET label=?,placeholder=? WHERE registration_form_id=? AND field_key=?`, strings.TrimSpace(fields[i].Label), fields[i].Placeholder, id, fields[i].FieldKey); err != nil {
				return nil, err
			}
		}
	} else {
		if form.Status != "DRAFT" {
			return nil, ErrRegistrationLocked
		}
		if _, err := tx.ExecContext(ctx, `DELETE o FROM registration_field_options o JOIN registration_fields f ON f.id=o.field_id WHERE f.registration_form_id=?`, id); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM registration_fields WHERE registration_form_id=?`, id); err != nil {
			return nil, err
		}
		if err := insertRegistrationFields(ctx, tx, id, fields); err != nil {
			return nil, err
		}
	}
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "UPDATE", EntityType: "REGISTRATION_FORM", EntityID: fmt.Sprint(id), Details: map[string]any{"form_id": id, "division_id": nullableID(form.DivisionID), "field_count": len(fields)}}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s.loadFields(ctx, s.db, id)
}

func (s *RegistrationService) validatePeriodDivision(ctx context.Context, periodID int64, divisionID *int64) error {
	return s.validatePeriodDivisionTx(ctx, s.db, periodID, divisionID)
}

func (s *RegistrationService) validatePeriodDivisionTx(ctx context.Context, q registrationQueryer, periodID int64, divisionID *int64) error {
	if periodID < 1 {
		return ErrRegistrationInvalid
	}
	var count int
	query := `SELECT COUNT(*) FROM periods WHERE id=?`
	args := []any{periodID}
	if divisionID != nil {
		query = `SELECT COUNT(*) FROM divisions WHERE period_id=? AND id=?`
		args = []any{periodID, *divisionID}
	}
	if err := q.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return err
	}
	if count != 1 {
		return ErrRegistrationInvalid
	}
	return nil
}

type registrationQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *RegistrationService) loadFields(ctx context.Context, q registrationQueryer, formID int64) ([]RegistrationField, error) {
	return loadRegistrationFields(ctx, q, formID)
}

func loadRegistrationFields(ctx context.Context, q registrationQueryer, formID int64) ([]RegistrationField, error) {
	rows, err := q.QueryContext(ctx, `SELECT id,field_key,label,field_type,placeholder,required,display_order FROM registration_fields WHERE registration_form_id=? ORDER BY display_order,id`, formID)
	if err != nil {
		return nil, err
	}
	fields := []RegistrationField{}
	byID := map[int64]int{}
	for rows.Next() {
		var field RegistrationField
		if err := rows.Scan(&field.ID, &field.FieldKey, &field.Label, &field.FieldType, &field.Placeholder, &field.Required, &field.DisplayOrder); err != nil {
			rows.Close()
			return nil, err
		}
		field.Options = []RegistrationOption{}
		byID[field.ID] = len(fields)
		fields = append(fields, field)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if len(fields) == 0 {
		return fields, nil
	}
	optionRows, err := q.QueryContext(ctx, `SELECT o.id,o.field_id,o.value,o.label,o.display_order FROM registration_field_options o JOIN registration_fields f ON f.id=o.field_id WHERE f.registration_form_id=? ORDER BY o.field_id,o.display_order,o.id`, formID)
	if err != nil {
		return nil, err
	}
	defer optionRows.Close()
	for optionRows.Next() {
		var option RegistrationOption
		var fieldID int64
		if err := optionRows.Scan(&option.ID, &fieldID, &option.Value, &option.Label, &option.DisplayOrder); err != nil {
			return nil, err
		}
		if index, ok := byID[fieldID]; ok {
			fields[index].Options = append(fields[index].Options, option)
		}
	}
	return fields, optionRows.Err()
}

func validateFormSchema(ctx context.Context, q registrationQueryer, formID int64) error {
	fields, err := loadRegistrationFields(ctx, q, formID)
	if err != nil {
		return err
	}
	inputs := make([]RegistrationFieldInput, len(fields))
	for i, field := range fields {
		inputs[i] = RegistrationFieldInput{FieldKey: field.FieldKey, Label: field.Label, FieldType: field.FieldType, Placeholder: field.Placeholder, Required: field.Required}
		for _, option := range field.Options {
			inputs[i].Options = append(inputs[i].Options, RegistrationOptionInput{Value: option.Value, Label: option.Label})
		}
	}
	return validateRegistrationFields(inputs)
}

func insertRegistrationFields(ctx context.Context, tx *sql.Tx, formID int64, fields []RegistrationFieldInput) error {
	for i, field := range fields {
		result, err := tx.ExecContext(ctx, `INSERT INTO registration_fields (registration_form_id,field_key,label,field_type,placeholder,required,display_order) VALUES (?,?,?,?,?,?,?)`, formID, field.FieldKey, strings.TrimSpace(field.Label), field.FieldType, field.Placeholder, field.Required, i+1)
		if err != nil {
			return err
		}
		fieldID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		for j, option := range field.Options {
			if _, err := tx.ExecContext(ctx, `INSERT INTO registration_field_options (field_id,value,label,display_order) VALUES (?,?,?,?)`, fieldID, strings.TrimSpace(option.Value), strings.TrimSpace(option.Label), j+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func defaultRegistrationCoreFields() []RegistrationFieldInput {
	return []RegistrationFieldInput{
		{FieldKey: "full_name", Label: "Nama Lengkap", FieldType: "TEXT", Required: true},
		{FieldKey: "nim", Label: "NIM", FieldType: "TEXT", Required: true},
		{FieldKey: "email", Label: "Email", FieldType: "EMAIL", Required: true},
		{FieldKey: "phone", Label: "No. HP", FieldType: "TEXT", Required: true},
	}
}

func validateRegistrationFields(fields []RegistrationFieldInput) error {
	if len(fields) < 4 || len(fields) > RegistrationMaxFields {
		return ErrRegistrationInvalid
	}
	seen := map[string]bool{}
	core := map[string]RegistrationFieldInput{}
	for _, field := range fields {
		field.FieldKey = strings.TrimSpace(field.FieldKey)
		label := strings.TrimSpace(field.Label)
		if !registrationFieldKeyPattern.MatchString(field.FieldKey) || seen[field.FieldKey] || label == "" || runeLen(label) > 160 {
			return ErrRegistrationInvalid
		}
		seen[field.FieldKey] = true
		field.FieldType = strings.ToUpper(strings.TrimSpace(field.FieldType))
		switch field.FieldType {
		case "TEXT", "TEXTAREA", "EMAIL", "NUMBER", "SELECT", "RADIO", "CHECKBOX", "DATE":
		default:
			return ErrRegistrationInvalid
		}
		if field.Placeholder != nil && runeLen(strings.TrimSpace(*field.Placeholder)) > 255 {
			return ErrRegistrationInvalid
		}
		choice := field.FieldType == "SELECT" || field.FieldType == "RADIO" || field.FieldType == "CHECKBOX"
		if choice && (len(field.Options) == 0 || len(field.Options) > RegistrationMaxOptions) || !choice && len(field.Options) > 0 {
			return ErrRegistrationInvalid
		}
		optionValues := map[string]bool{}
		for _, option := range field.Options {
			value, optionLabel := strings.TrimSpace(option.Value), strings.TrimSpace(option.Label)
			if value == "" || optionLabel == "" || runeLen(value) > 160 || runeLen(optionLabel) > 160 || optionValues[value] {
				return ErrRegistrationInvalid
			}
			optionValues[value] = true
		}
		if isRegistrationCoreKey(field.FieldKey) {
			core[field.FieldKey] = field
			if !field.Required {
				return ErrRegistrationInvalid
			}
		}
	}
	for key, expectedType := range map[string]string{"full_name": "TEXT", "nim": "TEXT", "email": "EMAIL", "phone": "TEXT"} {
		field, ok := core[key]
		if !ok || field.FieldType != expectedType {
			return ErrRegistrationInvalid
		}
	}
	return nil
}

func normalizeRegistrationFields(fields []RegistrationFieldInput) {
	for i := range fields {
		fields[i].FieldKey = strings.TrimSpace(fields[i].FieldKey)
		fields[i].Label = strings.TrimSpace(fields[i].Label)
		fields[i].FieldType = strings.ToUpper(strings.TrimSpace(fields[i].FieldType))
		if fields[i].Placeholder != nil {
			value := strings.TrimSpace(*fields[i].Placeholder)
			fields[i].Placeholder = &value
		}
		for j := range fields[i].Options {
			fields[i].Options[j].Value = strings.TrimSpace(fields[i].Options[j].Value)
			fields[i].Options[j].Label = strings.TrimSpace(fields[i].Options[j].Label)
		}
	}
}

func sameRegistrationStructure(old []RegistrationField, next []RegistrationFieldInput) bool {
	if len(old) != len(next) {
		return false
	}
	for i, field := range old {
		candidate := next[i]
		if field.FieldKey != candidate.FieldKey || field.FieldType != strings.ToUpper(candidate.FieldType) || field.Required != candidate.Required || field.DisplayOrder != i+1 || len(field.Options) != len(candidate.Options) {
			return false
		}
		for j, option := range field.Options {
			if option.Value != strings.TrimSpace(candidate.Options[j].Value) || option.Label != strings.TrimSpace(candidate.Options[j].Label) || option.DisplayOrder != j+1 {
				return false
			}
		}
	}
	return true
}

func validRegistrationDates(openAt, closeAt time.Time) bool {
	return !openAt.IsZero() && !closeAt.IsZero() && openAt.Before(closeAt)
}

func effectiveRegistrationStatus(stored string, openAt, closeAt, now time.Time) string {
	if stored == "CLOSED" || now.After(closeAt) {
		return "CLOSED"
	}
	if now.Before(openAt) {
		return "UPCOMING"
	}
	return "OPEN"
}

func validateRegistrationFilters(filters RegistrationFilters) error {
	if filters.Page < 1 || filters.Limit < 1 || filters.Limit > 100 || (filters.Status != "" && filters.Status != "DRAFT" && filters.Status != "OPEN" && filters.Status != "CLOSED" && filters.Status != "UPCOMING") || (filters.DateFrom != nil && filters.DateTo != nil && filters.DateTo.Before(*filters.DateFrom)) {
		return ErrRegistrationInvalid
	}
	if filters.DivisionID != nil && *filters.DivisionID < 1 || filters.PeriodID != nil && *filters.PeriodID < 1 {
		return ErrRegistrationInvalid
	}
	return nil
}

func registrationAdminWhere(user domain.User, filters RegistrationFilters) (string, []any, error) {
	parts := []string{"f.deleted_at IS NULL"}
	args := []any{}
	if user.Role == domain.RoleKepala {
		if user.DivisionID == nil || *user.DivisionID < 1 {
			return " WHERE 1=0", nil, ErrRegistrationForbidden
		}
		parts = append(parts, "f.division_id=?")
		args = append(args, *user.DivisionID)
	} else if !registrationRoleAllowed(user.Role) {
		return "", nil, ErrRegistrationForbidden
	} else if filters.DivisionID != nil {
		parts = append(parts, "f.division_id=?")
		args = append(args, *filters.DivisionID)
	}
	if filters.PeriodID != nil {
		parts = append(parts, "f.period_id=?")
		args = append(args, *filters.PeriodID)
	}
	if filters.Status != "" && filters.Status != "UPCOMING" {
		parts = append(parts, "f.status=?")
		args = append(args, filters.Status)
	} else if filters.Status == "UPCOMING" {
		parts = append(parts, "f.status='OPEN' AND UTC_TIMESTAMP(6)<f.open_at")
	}
	if search := strings.TrimSpace(filters.Search); search != "" {
		parts = append(parts, "LOCATE(?,LOWER(f.title))>0")
		args = append(args, strings.ToLower(search))
	}
	if filters.DateFrom != nil {
		parts = append(parts, "f.open_at>=?")
		args = append(args, filters.DateFrom.UTC())
	}
	if filters.DateTo != nil {
		parts = append(parts, "f.open_at<?")
		args = append(args, filters.DateTo.AddDate(0, 0, 1).UTC())
	}
	return " WHERE " + strings.Join(parts, " AND "), args, nil
}

func publicRegistrationWhere(filters RegistrationFilters, now time.Time) (string, []any) {
	parts := []string{"f.deleted_at IS NULL", "f.status<>'DRAFT'"}
	args := []any{}
	if filters.Status != "" {
		switch filters.Status {
		case "UPCOMING":
			parts = append(parts, "f.status='OPEN' AND ? < f.open_at")
			args = append(args, now)
		case "OPEN":
			parts = append(parts, "f.status='OPEN' AND ? >= f.open_at AND ? <= f.close_at")
			args = append(args, now, now)
		case "CLOSED":
			parts = append(parts, "(f.status='CLOSED' OR (f.status='OPEN' AND ? > f.close_at))")
			args = append(args, now)
		}
	}
	if filters.DivisionID != nil {
		parts = append(parts, "f.division_id=?")
		args = append(args, *filters.DivisionID)
	}
	if filters.PeriodID != nil {
		parts = append(parts, "f.period_id=?")
		args = append(args, *filters.PeriodID)
	}
	if search := strings.TrimSpace(filters.Search); search != "" {
		parts = append(parts, "LOCATE(?,LOWER(f.title))>0")
		args = append(args, strings.ToLower(search))
	}
	if filters.DateFrom != nil {
		parts = append(parts, "f.open_at>=?")
		args = append(args, filters.DateFrom.UTC())
	}
	if filters.DateTo != nil {
		parts = append(parts, "f.open_at<?")
		args = append(args, filters.DateTo.AddDate(0, 0, 1).UTC())
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func lockRegistrationForm(ctx context.Context, tx *sql.Tx, id int64) (RegistrationForm, error) {
	item, err := scanRegistrationForm(tx.QueryRowContext(ctx, `SELECT f.id,f.period_id,p.name,f.division_id,d.name,f.title,f.slug,f.description,f.open_at,f.close_at,f.status,f.created_at,f.updated_at,(SELECT COUNT(*) FROM registration_submissions s WHERE s.registration_form_id=f.id)
	FROM registration_forms f JOIN periods p ON p.id=f.period_id LEFT JOIN divisions d ON d.id=f.division_id WHERE f.id=? AND f.deleted_at IS NULL FOR UPDATE`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return RegistrationForm{}, ErrRegistrationNotFound
	}
	return item, err
}

type registrationScanner interface{ Scan(...any) error }

func scanRegistrationForm(row registrationScanner) (RegistrationForm, error) {
	var item RegistrationForm
	var divisionID sql.NullInt64
	var divisionName sql.NullString
	err := row.Scan(&item.ID, &item.PeriodID, &item.PeriodName, &divisionID, &divisionName, &item.Title, &item.Slug, &item.Description, &item.OpenAt, &item.CloseAt, &item.Status, &item.CreatedAt, &item.UpdatedAt, &item.SubmissionCount)
	if err != nil {
		return RegistrationForm{}, err
	}
	item.DivisionID, item.DivisionName = nullID(divisionID), nullText(divisionName)
	item.OpenAt, item.CloseAt = item.OpenAt.UTC(), item.CloseAt.UTC()
	item.CreatedAt, item.UpdatedAt = item.CreatedAt.UTC(), item.UpdatedAt.UTC()
	return item, nil
}

func registrationRoleAllowed(role domain.Role) bool {
	return role == domain.RoleSuperAdmin || role == domain.RoleBPH || role == domain.RoleKepala
}

func registrationCanAccess(user domain.User, divisionID *int64) bool {
	if user.Role == domain.RoleSuperAdmin || user.Role == domain.RoleBPH {
		return true
	}
	return user.Role == domain.RoleKepala && user.DivisionID != nil && divisionID != nil && *user.DivisionID == *divisionID
}

func isRegistrationCoreKey(key string) bool {
	return key == "full_name" || key == "nim" || key == "email" || key == "phone"
}

func nullID(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}

func nullText(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func nullableID(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func slugifyRegistration(value string) string {
	var builder strings.Builder
	pendingDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			if pendingDash && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(r)
			pendingDash = false
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) || unicode.IsPunct(r) {
			pendingDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func isRegistrationUniqueConflict(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}
