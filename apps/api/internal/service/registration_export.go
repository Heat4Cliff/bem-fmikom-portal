package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/repository"
	"github.com/xuri/excelize/v2"
)

var ErrRegistrationExportLimit = errors.New("registration export row limit exceeded")

func (s *RegistrationService) Export(ctx context.Context, user domain.User, formID int64, filters RegistrationSubmissionFilters) ([]byte, int, error) {
	form, err := s.Get(ctx, user, formID)
	if err != nil {
		return nil, 0, err
	}
	where, args := submissionWhere(formID, filters)
	rows, err := s.db.QueryContext(ctx, `SELECT s.id,s.registration_form_id,f.title,d.name,s.full_name,s.nim,s.email,s.phone,s.submitted_at FROM registration_submissions s JOIN registration_forms f ON f.id=s.registration_form_id LEFT JOIN divisions d ON d.id=f.division_id`+where+` ORDER BY s.submitted_at,s.id LIMIT ?`, append(args, RegistrationMaxExportRows+1)...)
	if err != nil {
		return nil, 0, err
	}
	submissions := []RegistrationSubmission{}
	for rows.Next() {
		var item RegistrationSubmission
		var division sql.NullString
		if err := rows.Scan(&item.ID, &item.FormID, &item.FormTitle, &division, &item.FullName, &item.NIM, &item.Email, &item.Phone, &item.SubmittedAt); err != nil {
			rows.Close()
			return nil, 0, err
		}
		item.Division, item.SubmittedAt = nullText(division), item.SubmittedAt.UTC()
		submissions = append(submissions, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, 0, err
	}
	rows.Close()
	if len(submissions) > RegistrationMaxExportRows {
		return nil, len(submissions), ErrRegistrationExportLimit
	}
	if len(submissions) > 0 {
		ids := make([]any, len(submissions))
		placeholders := make([]string, len(submissions))
		for i, item := range submissions {
			ids[i], placeholders[i] = item.ID, "?"
		}
		answerRows, err := s.db.QueryContext(ctx, `SELECT submission_id,field_key_snapshot,label_snapshot,type_snapshot,value_text,value_json,display_order_snapshot FROM registration_submission_answers WHERE submission_id IN (`+strings.Join(placeholders, ",")+") ORDER BY submission_id,display_order_snapshot,id", ids...)
		if err != nil {
			return nil, 0, err
		}
		byID := make(map[int64]int, len(submissions))
		for i := range submissions {
			byID[submissions[i].ID] = i
			submissions[i].Answers = []RegistrationAnswer{}
		}
		for answerRows.Next() {
			var submissionID int64
			var answer RegistrationAnswer
			var value sql.NullString
			var jsonValue []byte
			if err := answerRows.Scan(&submissionID, &answer.FieldKey, &answer.Label, &answer.Type, &value, &jsonValue, &answer.Order); err != nil {
				answerRows.Close()
				return nil, 0, err
			}
			answer.Text = nullText(value)
			if len(jsonValue) > 0 {
				answer.JSON = append(json.RawMessage(nil), jsonValue...)
			}
			if index, exists := byID[submissionID]; exists {
				submissions[index].Answers = append(submissions[index].Answers, answer)
			}
		}
		if err := answerRows.Err(); err != nil {
			answerRows.Close()
			return nil, 0, err
		}
		answerRows.Close()
	}
	file := excelize.NewFile()
	defer file.Close()
	sheet := "Pendaftar"
	file.SetSheetName("Sheet1", sheet)
	headers := []string{"No.", "Nama", "NIM", "Email", "No HP", "Waktu Pendaftaran"}
	fieldLabels := map[string]string{}
	fieldKeys := []string{}
	for _, submission := range submissions {
		for _, answer := range submission.Answers {
			if answer.FieldKey == "full_name" || answer.FieldKey == "nim" || answer.FieldKey == "email" || answer.FieldKey == "phone" {
				continue
			}
			if _, exists := fieldLabels[answer.FieldKey]; !exists {
				fieldLabels[answer.FieldKey] = answer.Label
				fieldKeys = append(fieldKeys, answer.FieldKey)
			}
		}
	}
	if len(fieldKeys) == 0 {
		fields, err := s.loadFields(ctx, s.db, formID)
		if err != nil {
			return nil, 0, err
		}
		for _, field := range fields {
			if field.FieldKey == "full_name" || field.FieldKey == "nim" || field.FieldKey == "email" || field.FieldKey == "phone" {
				continue
			}
			fieldLabels[field.FieldKey] = field.Label
			fieldKeys = append(fieldKeys, field.FieldKey)
		}
	}
	for _, key := range fieldKeys {
		headers = append(headers, fieldLabels[key])
	}
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		if err := file.SetCellStr(sheet, cell, header); err != nil {
			return nil, 0, err
		}
	}
	for rowIndex, submission := range submissions {
		answerValues := map[string]string{}
		for _, answer := range submission.Answers {
			if answer.FieldKey == "full_name" || answer.FieldKey == "nim" || answer.FieldKey == "email" || answer.FieldKey == "phone" {
				continue
			}
			value := ""
			if answer.Text != nil {
				value = *answer.Text
			} else if len(answer.JSON) > 0 {
				var choices []string
				if err := json.Unmarshal(answer.JSON, &choices); err != nil {
					return nil, 0, err
				}
				value = strings.Join(choices, "; ")
			}
			answerValues[answer.FieldKey] = value
		}
		values := []string{fmt.Sprint(rowIndex + 1), submission.FullName, submission.NIM, submission.Email, submission.Phone, submission.SubmittedAt.Format("2006-01-02 15:04:05 UTC")}
		for _, key := range fieldKeys {
			values = append(values, answerValues[key])
		}
		for column, value := range values {
			cell, _ := excelize.CoordinatesToCellName(column+1, rowIndex+2)
			// SetCellStr writes applicant-controlled content as a string cell, never a formula.
			if err := file.SetCellStr(sheet, cell, value); err != nil {
				return nil, 0, err
			}
		}
	}
	if err := file.SetColWidth(sheet, "A", "A", 8); err != nil {
		return nil, 0, err
	}
	for i := 2; i <= len(headers); i++ {
		column, _ := excelize.ColumnNumberToName(i)
		if err := file.SetColWidth(sheet, column, column, 22); err != nil {
			return nil, 0, err
		}
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, 0, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()
	if err := s.log.Write(ctx, tx, repository.ActivityEvent{ActorUserID: &user.ID, Action: "EXPORT", EntityType: "REGISTRATION_SUBMISSIONS", EntityID: fmt.Sprint(form.ID), Details: map[string]any{"form_id": form.ID, "division_id": nullableID(form.DivisionID), "submission_count": len(submissions)}}); err != nil {
		return nil, 0, err
	}
	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}
	return bytes.Clone(buffer.Bytes()), len(submissions), nil
}
