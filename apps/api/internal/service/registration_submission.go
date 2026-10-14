package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/storage"
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
	FieldKey   string                  `json:"field_key"`
	Label      string                  `json:"label"`
	Type       string                  `json:"type"`
	Text       *string                 `json:"value_text"`
	JSON       json.RawMessage         `json:"value_json,omitempty"`
	Order      int                     `json:"display_order"`
	Attachment *RegistrationAttachment `json:"attachment,omitempty"`
}

type RegistrationImageUpload struct {
	Filename string
	MIMEType string
	Data     []byte
}

// RegistrationSubmissionFieldErrors carries safe, actionable validation
// feedback keyed by the stable registration field key. It deliberately omits
// database and storage internals from public responses.
type RegistrationSubmissionFieldErrors struct {
	Fields map[string]string
}

func (e *RegistrationSubmissionFieldErrors) Error() string {
	return "registration submission validation failed"
}
func (e *RegistrationSubmissionFieldErrors) Unwrap() error { return ErrRegistrationInvalid }

func registrationSubmissionFieldError(key, message string) error {
	return &RegistrationSubmissionFieldErrors{Fields: map[string]string{key: message}}
}

type RegistrationAttachment struct {
	ID           int64  `json:"id"`
	OriginalName string `json:"original_filename"`
	MIMEType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	DownloadURL  string `json:"download_url"`
}

type privateRegistrationFile struct {
	StorageKey   string
	OriginalName string
	MIMEType     string
	SizeBytes    int64
	Width        int
	Height       int
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
	return s.SubmitWithImages(ctx, formID, in, nil, now)
}

func (s *RegistrationService) SubmitWithImages(ctx context.Context, formID int64, in RegistrationSubmitInput, images map[string]RegistrationImageUpload, now time.Time) (RegistrationReceipt, error) {
	if formID < 1 || len(in.Answers) > RegistrationMaxFields {
		return RegistrationReceipt{}, ErrRegistrationInvalid
	}
	for _, key := range []string{"full_name", "nim", "email", "phone"} {
		if _, exists := in.Answers[key]; !exists {
			label := map[string]string{"full_name": "Nama lengkap", "nim": "NIM", "email": "Email", "phone": "Nomor HP"}[key]
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, label+" wajib diisi.")
		}
	}
	if len(images) > RegistrationMaxImageFields || (len(images) > 0 && (s.files == nil || s.imageMax < 1)) {
		return RegistrationReceipt{}, registrationSubmissionFieldError("image", "Lampiran gambar tidak dapat diproses.")
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
	imageFields := make(map[string]RegistrationField)
	for _, field := range fields {
		if field.FieldType == "IMAGE" {
			imageFields[field.FieldKey] = field
		}
	}
	if len(images) > len(imageFields) {
		return RegistrationReceipt{}, registrationSubmissionFieldError("image", "Jumlah lampiran tidak sesuai dengan kolom gambar formulir.")
	}
	prepared := make(map[string]preparedRegistrationFile, len(images))
	cleanupKeys := []string{}
	committed := false
	defer func() {
		if !committed && s.files != nil {
			for _, key := range cleanupKeys {
				_ = s.files.Delete(context.Background(), key)
			}
		}
	}()
	for key, field := range imageFields {
		upload, exists := images[key]
		if !exists {
			if field.Required {
				return RegistrationReceipt{}, registrationSubmissionFieldError(key, field.Label+" wajib diunggah.")
			}
			continue
		}
		name := safeMediaFilename(upload.Filename)
		ext := strings.ToLower(filepath.Ext(name))
		processed, processErr := processMediaImage(upload.Data, ext, upload.MIMEType, s.imageMax)
		if errors.Is(processErr, ErrMediaTooLarge) {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "Ukuran gambar maksimal 5 MiB.")
		}
		if name == "" {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "Nama file gambar tidak valid.")
		}
		if errors.Is(processErr, ErrMediaUnsupported) {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "Format gambar tidak didukung atau tidak cocok dengan isi file.")
		}
		if errors.Is(processErr, ErrMediaAnimationUnsupported) {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "GIF animasi tidak didukung.")
		}
		if errors.Is(processErr, ErrMediaDimensions) {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "Dimensi gambar melebihi batas yang diizinkan.")
		}
		if processErr != nil {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "File bukan gambar yang valid atau rusak.")
		}
		id, randomErr := randomRegistrationFileID()
		if randomErr != nil {
			return RegistrationReceipt{}, randomErr
		}
		stage := ".staging/" + id + ".part"
		final := "registration/" + id + processed.storedExt
		cleanupKeys = append(cleanupKeys, stage, final)
		prepared[key] = preparedRegistrationFile{field: field, filename: name, originalMIME: acceptedImageTypes[ext], processed: processed, stageKey: stage, storageKey: final}
	}
	for key := range images {
		if _, ok := imageFields[key]; !ok {
			return RegistrationReceipt{}, registrationSubmissionFieldError("image", "Lampiran gambar tidak cocok dengan kolom formulir.")
		}
		if _, ok := in.Answers[key]; ok {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "Lampiran gambar dikirim dengan format jawaban yang keliru.")
		}
	}
	values := make(map[string]any, len(fields))
	for _, field := range fields {
		if field.FieldType == "IMAGE" {
			values[field.FieldKey] = nil
			continue
		}
		raw, ok := in.Answers[field.FieldKey]
		if !ok {
			if field.Required {
				return RegistrationReceipt{}, registrationSubmissionFieldError(field.FieldKey, field.Label+" wajib diisi.")
			}
			values[field.FieldKey] = nil
			continue
		}
		value, err := validateRegistrationAnswer(field, raw)
		if err != nil {
			message := field.Label + " tidak valid."
			if field.FieldType == "EMAIL" {
				message = "Format email tidak valid."
			}
			return RegistrationReceipt{}, registrationSubmissionFieldError(field.FieldKey, message)
		}
		values[field.FieldKey] = value
	}
	for key := range in.Answers {
		if field, isImage := imageFields[key]; isImage && field.FieldType == "IMAGE" {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "Unggah gambar pada kolom lampiran, bukan sebagai jawaban teks.")
		}
		if _, ok := values[key]; !ok {
			return RegistrationReceipt{}, registrationSubmissionFieldError(key, "Kolom formulir tidak dikenali.")
		}
	}
	fullName := strings.TrimSpace(values["full_name"].(string))
	nim := strings.TrimSpace(values["nim"].(string))
	email := strings.ToLower(strings.TrimSpace(values["email"].(string)))
	phone := strings.TrimSpace(values["phone"].(string))
	if fullName == "" || runeLen(fullName) > RegistrationMaxName {
		return RegistrationReceipt{}, registrationSubmissionFieldError("full_name", "Nama lengkap wajib diisi dan tidak boleh melebihi batas karakter.")
	}
	if nim == "" || runeLen(nim) > RegistrationMaxNIM {
		return RegistrationReceipt{}, registrationSubmissionFieldError("nim", "NIM wajib diisi dan tidak boleh melebihi batas karakter.")
	}
	if email == "" || runeLen(email) > RegistrationMaxEmail {
		return RegistrationReceipt{}, registrationSubmissionFieldError("email", "Email wajib diisi dan tidak boleh melebihi batas karakter.")
	}
	if phone == "" || runeLen(phone) > RegistrationMaxPhone || !registrationPhonePattern.MatchString(phone) {
		return RegistrationReceipt{}, registrationSubmissionFieldError("phone", "Nomor HP wajib diisi dan hanya boleh berisi angka, spasi, +, tanda kurung, titik, atau tanda hubung.")
	}
	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return RegistrationReceipt{}, registrationSubmissionFieldError("email", "Format email tidak valid.")
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
	for fieldKey, file := range prepared {
		object, err := s.files.Save(ctx, file.stageKey, bytes.NewReader(file.processed.storedBytes), s.imageMax)
		if err != nil {
			return RegistrationReceipt{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO registration_submission_files (submission_id,field_id,field_key_snapshot,original_filename,storage_key,original_mime_type,stored_mime_type,size_bytes,width,height,sha256) VALUES (?,?,?,?,?,?,?,?,?,?,?)`, submissionID, file.field.ID, fieldKey, file.filename, file.storageKey, file.originalMIME, file.processed.storedMIME, object.SizeBytes, file.processed.width, file.processed.height, object.SHA256); err != nil {
			return RegistrationReceipt{}, err
		}
		if err := s.files.Promote(ctx, file.stageKey, file.storageKey); err != nil {
			return RegistrationReceipt{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		if isRegistrationUniqueConflict(err) {
			return RegistrationReceipt{}, ErrRegistrationDuplicate
		}
		return RegistrationReceipt{}, err
	}
	committed = true
	return RegistrationReceipt{SubmissionID: submissionID, SubmittedAt: submittedAt, Title: form.Title}, nil
}

type preparedRegistrationFile struct {
	field        RegistrationField
	filename     string
	originalMIME string
	processed    processedImage
	stageKey     string
	storageKey   string
}

func randomRegistrationFileID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("generate private attachment key")
	}
	return hex.EncodeToString(value[:]), nil
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
		if answer.Type == "IMAGE" {
			var file RegistrationAttachment
			fileErr := s.db.QueryRowContext(ctx, `SELECT id,original_filename,stored_mime_type,size_bytes,width,height FROM registration_submission_files WHERE submission_id=? AND field_key_snapshot=?`, submissionID, answer.FieldKey).Scan(&file.ID, &file.OriginalName, &file.MIMEType, &file.SizeBytes, &file.Width, &file.Height)
			if fileErr == nil {
				file.DownloadURL = fmt.Sprintf("/admin/registrations/%d/submissions/%d/files/%d", formID, submissionID, file.ID)
				answer.Attachment = &file
			} else if !errors.Is(fileErr, sql.ErrNoRows) {
				return RegistrationSubmission{}, fileErr
			}
		}
		item.Answers = append(item.Answers, answer)
	}
	return item, rows.Err()
}

func (s *RegistrationService) OpenSubmissionFile(ctx context.Context, user domain.User, formID, submissionID, fileID int64) (storage.DocumentReader, privateRegistrationFile, error) {
	if _, err := s.Get(ctx, user, formID); err != nil {
		return nil, privateRegistrationFile{}, err
	}
	var file privateRegistrationFile
	err := s.db.QueryRowContext(ctx, `SELECT f.storage_key,f.original_filename,f.stored_mime_type,f.size_bytes,f.width,f.height FROM registration_submission_files f JOIN registration_submissions s ON s.id=f.submission_id WHERE s.registration_form_id=? AND s.id=? AND f.id=?`, formID, submissionID, fileID).Scan(&file.StorageKey, &file.OriginalName, &file.MIMEType, &file.SizeBytes, &file.Width, &file.Height)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, privateRegistrationFile{}, ErrRegistrationNotFound
	}
	if err != nil {
		return nil, privateRegistrationFile{}, err
	}
	reader, err := s.files.Open(ctx, file.StorageKey)
	if err != nil {
		return nil, privateRegistrationFile{}, err
	}
	return reader, file, nil
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
