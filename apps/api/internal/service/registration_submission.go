package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
)

var registrationPhonePattern = regexp.MustCompile(`^[+0-9() .-]+$`)

type RegistrationSubmitInput struct {
	Answers map[string]json.RawMessage `json:"answers"`
}

func (in *RegistrationSubmitInput) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var body struct {
		Answers json.RawMessage `json:"answers"`
	}
	if err := decoder.Decode(&body); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("registration request has trailing JSON")
	}
	if len(body.Answers) == 0 || bytes.Equal(bytes.TrimSpace(body.Answers), []byte("null")) {
		in.Answers = nil
		return nil
	}
	answersDecoder := json.NewDecoder(bytes.NewReader(body.Answers))
	token, err := answersDecoder.Token()
	if err != nil || token != json.Delim('{') {
		return fmt.Errorf("answers must be an object")
	}
	in.Answers = make(map[string]json.RawMessage)
	for answersDecoder.More() {
		keyToken, err := answersDecoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("answer key must be a string")
		}
		if _, exists := in.Answers[key]; exists {
			return fmt.Errorf("duplicate answer key")
		}
		var value json.RawMessage
		if err := answersDecoder.Decode(&value); err != nil {
			return err
		}
		in.Answers[key] = value
	}
	if _, err := answersDecoder.Token(); err != nil {
		return err
	}
	if err := answersDecoder.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("answers contain trailing JSON")
	}
	return nil
}

type RegistrationReceipt struct {
	SubmissionID int64     `json:"submission_id"`
	SubmittedAt  time.Time `json:"submitted_at"`
	Title        string    `json:"registration_title"`
}

type RegistrationAnswer struct {
	FieldKey string          `json:"field_key"`
	Label    string          `json:"label"`
	Type     string          `json:"type"`
	Text     *string         `json:"value_text"`
	JSON     json.RawMessage `json:"value_json,omitempty"`
	Order    int             `json:"display_order"`
}

type RegistrationSubmission struct {
	ID          int64                `json:"id"`
	FormID      int64                `json:"registration_form_id"`
	FormTitle   string               `json:"registration_title"`
	Division    *string              `json:"division_name"`
	FullName    string               `json:"full_name"`
	NIM         string               `json:"nim"`
	Email       string               `json:"email"`
	Phone       string               `json:"phone"`
	SubmittedAt time.Time            `json:"submitted_at"`
	Answers     []RegistrationAnswer `json:"answers,omitempty"`
}

func (s *RegistrationService) Submit(ctx context.Context, formID int64, in RegistrationSubmitInput, now time.Time) (RegistrationReceipt, error) {
	if formID < 1 || len(in.Answers) < 4 || len(in.Answers) > RegistrationMaxFields {
		return RegistrationReceipt{}, ErrRegistrationInvalid
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RegistrationReceipt{}, err
	}
	defer tx.Rollback()
	form, err := lockRegistrationForm(ctx, tx, formID)
	if err != nil {
		return RegistrationReceipt{}, err
	}
	if form.Status != "OPEN" || now.Before(form.OpenAt) || now.After(form.CloseAt) {
		return RegistrationReceipt{}, ErrRegistrationClosed
	}
	fields, err := loadRegistrationFields(ctx, tx, formID)
	if err != nil {
		return RegistrationReceipt{}, err
	}
	if err := validateRegistrationFields(toRegistrationFieldInputs(fields)); err != nil {
		return RegistrationReceipt{}, err
	}
	values := make(map[string]any, len(fields))
	for _, field := range fields {
		raw, ok := in.Answers[field.FieldKey]
		if !ok {
			if field.Required {
				return RegistrationReceipt{}, ErrRegistrationInvalid
			}
			values[field.FieldKey] = nil
			continue
		}
		value, err := validateRegistrationAnswer(field, raw)
		if err != nil {
			return RegistrationReceipt{}, ErrRegistrationInvalid
		}
		values[field.FieldKey] = value
	}
	for key := range in.Answers {
		if _, ok := values[key]; !ok {
			return RegistrationReceipt{}, ErrRegistrationInvalid
		}
	}
	fullName := strings.TrimSpace(values["full_name"].(string))
	nim := strings.TrimSpace(values["nim"].(string))
	email := strings.ToLower(strings.TrimSpace(values["email"].(string)))
	phone := strings.TrimSpace(values["phone"].(string))
	if fullName == "" || runeLen(fullName) > RegistrationMaxName || nim == "" || runeLen(nim) > RegistrationMaxNIM || email == "" || runeLen(email) > RegistrationMaxEmail || phone == "" || runeLen(phone) > RegistrationMaxPhone || !registrationPhonePattern.MatchString(phone) {
		return RegistrationReceipt{}, ErrRegistrationInvalid
	}
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return RegistrationReceipt{}, ErrRegistrationInvalid
	}
	normalizedNIM := strings.ToUpper(nim)
	var duplicate int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_submissions WHERE registration_form_id=? AND normalized_nim=?`, formID, normalizedNIM).Scan(&duplicate); err != nil {
		return RegistrationReceipt{}, err
	}
	if duplicate > 0 {
		return RegistrationReceipt{}, ErrRegistrationDuplicate
	}
	submittedAt := now.UTC()
	result, err := tx.ExecContext(ctx, `INSERT INTO registration_submissions (registration_form_id,full_name,nim,normalized_nim,email,phone,submitted_at) VALUES (?,?,?,?,?,?,?)`, formID, fullName, nim, normalizedNIM, email, phone, submittedAt)
	if err != nil {
		if isRegistrationUniqueConflict(err) {
			return RegistrationReceipt{}, ErrRegistrationDuplicate
		}
		return RegistrationReceipt{}, err
	}
	submissionID, err := result.LastInsertId()
	if err != nil {
		return RegistrationReceipt{}, err
	}
	for _, field := range fields {
		answer := values[field.FieldKey]
		var textValue, jsonValue any
		if field.FieldKey == "full_name" || field.FieldKey == "nim" || field.FieldKey == "email" || field.FieldKey == "phone" || answer == nil {
			// Core identity values live once in registration_submissions.
		} else if field.FieldType == "CHECKBOX" {
			encoded, marshalErr := json.Marshal(answer)
			if marshalErr != nil {
				return RegistrationReceipt{}, ErrRegistrationInvalid
			}
			jsonValue = encoded
		} else {
			textValue = answer
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO registration_submission_answers (submission_id,field_id,field_key_snapshot,label_snapshot,type_snapshot,value_text,value_json,display_order_snapshot) VALUES (?,?,?,?,?,?,?,?)`, submissionID, field.ID, field.FieldKey, field.Label, field.FieldType, textValue, jsonValue, field.DisplayOrder); err != nil {
			return RegistrationReceipt{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		if isRegistrationUniqueConflict(err) {
			return RegistrationReceipt{}, ErrRegistrationDuplicate
		}
		return RegistrationReceipt{}, err
	}
	return RegistrationReceipt{SubmissionID: submissionID, SubmittedAt: submittedAt, Title: form.Title}, nil
}

func (s *RegistrationService) ListSubmissions(ctx context.Context, user domain.User, formID int64, filters RegistrationSubmissionFilters) (Result[RegistrationSubmission], error) {
	if filters.Page < 1 || filters.Limit < 1 || filters.Limit > 100 || (filters.DateFrom != nil && filters.DateTo != nil && filters.DateTo.Before(*filters.DateFrom)) {
		return Result[RegistrationSubmission]{}, ErrRegistrationInvalid
	}
	form, err := s.Get(ctx, user, formID)
	if err != nil {
		return Result[RegistrationSubmission]{}, err
	}
	where, args := submissionWhere(formID, filters)
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_submissions s`+where, args...).Scan(&total); err != nil {
		return Result[RegistrationSubmission]{}, err
	}
	queryArgs := append(append([]any{}, args...), filters.Limit, (filters.Page-1)*filters.Limit)
	rows, err := s.db.QueryContext(ctx, `SELECT s.id,s.registration_form_id,?,d.name,s.full_name,s.nim,s.email,s.phone,s.submitted_at FROM registration_submissions s JOIN registration_forms f ON f.id=s.registration_form_id LEFT JOIN divisions d ON d.id=f.division_id`+where+` ORDER BY s.submitted_at DESC,s.id DESC LIMIT ? OFFSET ?`, append([]any{form.Title}, queryArgs...)...)
	if err != nil {
		return Result[RegistrationSubmission]{}, err
	}
	defer rows.Close()
	items := []RegistrationSubmission{}
	for rows.Next() {
		var item RegistrationSubmission
		var division sql.NullString
		if err := rows.Scan(&item.ID, &item.FormID, &item.FormTitle, &division, &item.FullName, &item.NIM, &item.Email, &item.Phone, &item.SubmittedAt); err != nil {
			return Result[RegistrationSubmission]{}, err
		}
		item.Division, item.SubmittedAt = nullText(division), item.SubmittedAt.UTC()
		items = append(items, item)
	}
	return Result[RegistrationSubmission]{Items: items, Page: pageMeta(filters.Page, filters.Limit, total)}, rows.Err()
}

func (s *RegistrationService) GetSubmission(ctx context.Context, user domain.User, formID, submissionID int64) (RegistrationSubmission, error) {
	form, err := s.Get(ctx, user, formID)
	if err != nil {
		return RegistrationSubmission{}, err
	}
	var item RegistrationSubmission
	var division sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT s.id,s.registration_form_id,f.title,d.name,s.full_name,s.nim,s.email,s.phone,s.submitted_at FROM registration_submissions s JOIN registration_forms f ON f.id=s.registration_form_id LEFT JOIN divisions d ON d.id=f.division_id WHERE s.id=? AND s.registration_form_id=?`, submissionID, formID).Scan(&item.ID, &item.FormID, &item.FormTitle, &division, &item.FullName, &item.NIM, &item.Email, &item.Phone, &item.SubmittedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return RegistrationSubmission{}, ErrRegistrationNotFound
	}
	if err != nil {
		return RegistrationSubmission{}, err
	}
	_ = form // Get enforces the database-authoritative division scope.
	item.Division, item.SubmittedAt = nullText(division), item.SubmittedAt.UTC()
	rows, err := s.db.QueryContext(ctx, `SELECT field_key_snapshot,label_snapshot,type_snapshot,value_text,value_json,display_order_snapshot FROM registration_submission_answers WHERE submission_id=? ORDER BY display_order_snapshot,id`, submissionID)
	if err != nil {
		return RegistrationSubmission{}, err
	}
	defer rows.Close()
	item.Answers = []RegistrationAnswer{}
	for rows.Next() {
		var answer RegistrationAnswer
		var text sql.NullString
		var jsonValue []byte
		if err := rows.Scan(&answer.FieldKey, &answer.Label, &answer.Type, &text, &jsonValue, &answer.Order); err != nil {
			return RegistrationSubmission{}, err
		}
		answer.Text = nullText(text)
		if len(jsonValue) > 0 {
			answer.JSON = append(json.RawMessage(nil), jsonValue...)
		}
		item.Answers = append(item.Answers, answer)
	}
	return item, rows.Err()
}

func submissionWhere(formID int64, filters RegistrationSubmissionFilters) (string, []any) {
	parts := []string{"s.registration_form_id=?"}
	args := []any{formID}
	if search := strings.TrimSpace(filters.Search); search != "" {
		parts = append(parts, `(LOCATE(?,LOWER(s.full_name))>0 OR LOCATE(?,LOWER(s.nim))>0 OR LOCATE(?,LOWER(s.email))>0)`)
		needle := strings.ToLower(search)
		args = append(args, needle, needle, needle)
	}
	if filters.DateFrom != nil {
		parts = append(parts, "s.submitted_at>=?")
		args = append(args, filters.DateFrom.UTC())
	}
	if filters.DateTo != nil {
		parts = append(parts, "s.submitted_at<?")
		args = append(args, filters.DateTo.AddDate(0, 0, 1).UTC())
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func toRegistrationFieldInputs(fields []RegistrationField) []RegistrationFieldInput {
	inputs := make([]RegistrationFieldInput, len(fields))
	for i, field := range fields {
		inputs[i] = RegistrationFieldInput{FieldKey: field.FieldKey, Label: field.Label, FieldType: field.FieldType, Placeholder: field.Placeholder, Required: field.Required}
		for _, option := range field.Options {
			inputs[i].Options = append(inputs[i].Options, RegistrationOptionInput{Value: option.Value, Label: option.Label})
		}
	}
	return inputs
}

func validateRegistrationAnswer(field RegistrationField, raw json.RawMessage) (any, error) {
	if strings.TrimSpace(string(raw)) == "null" {
		if field.Required {
			return nil, ErrRegistrationInvalid
		}
		return nil, nil
	}
	var value any
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if value == nil || value == "" {
		if field.Required {
			return nil, ErrRegistrationInvalid
		}
		return nil, nil
	}
	allowed := map[string]bool{}
	for _, option := range field.Options {
		allowed[option.Value] = true
	}
	switch field.FieldType {
	case "TEXT", "TEXTAREA", "EMAIL", "NUMBER", "SELECT", "RADIO", "DATE":
		text, ok := value.(string)
		if !ok {
			return nil, ErrRegistrationInvalid
		}
		text = strings.TrimSpace(text)
		max := RegistrationMaxText
		if field.FieldType == "TEXTAREA" {
			max = RegistrationMaxTextArea
		}
		if text == "" && !field.Required {
			return nil, nil
		}
		if text == "" || runeLen(text) > max {
			return nil, ErrRegistrationInvalid
		}
		switch field.FieldType {
		case "EMAIL":
			parsed, err := mail.ParseAddress(strings.ToLower(text))
			if err != nil || parsed.Address != strings.ToLower(text) {
				return nil, ErrRegistrationInvalid
			}
			text = strings.ToLower(text)
		case "NUMBER":
			if _, err := strconv.ParseFloat(text, 64); err != nil {
				return nil, ErrRegistrationInvalid
			}
		case "DATE":
			if _, err := time.Parse("2006-01-02", text); err != nil {
				return nil, ErrRegistrationInvalid
			}
		case "SELECT", "RADIO":
			if !allowed[text] {
				return nil, ErrRegistrationInvalid
			}
		}
		return text, nil
	case "CHECKBOX":
		values, ok := value.([]any)
		if !ok || (field.Required && len(values) == 0) || len(values) > len(allowed) {
			return nil, ErrRegistrationInvalid
		}
		seen := map[string]bool{}
		result := make([]string, 0, len(values))
		for _, value := range values {
			text, ok := value.(string)
			if !ok || !allowed[text] || seen[text] {
				return nil, ErrRegistrationInvalid
			}
			seen[text] = true
			result = append(result, text)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown field type")
	}
}

var ErrRegistrationClosed = errors.New("registration is not accepting submissions")

var _ = repository.ActivityEvent{}
