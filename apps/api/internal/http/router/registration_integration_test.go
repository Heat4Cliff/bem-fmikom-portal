package router

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
	"github.com/xuri/excelize/v2"
)

func runRegistrationIntegration(t *testing.T, ctx context.Context, db *sql.DB, engine http.Handler, logger *bytes.Buffer, superToken, bphToken, kepalaAToken, kepalaBToken string, divisionA, divisionB int64) {
	t.Helper()
	now := time.Now().UTC()
	var periodID int64
	if err := db.QueryRowContext(ctx, `SELECT period_id FROM divisions WHERE id=?`, divisionA).Scan(&periodID); err != nil {
		t.Fatal(err)
	}
	fields := []map[string]any{
		{"field_key": "full_name", "label": "Nama Lengkap", "field_type": "TEXT", "required": true, "options": []any{}},
		{"field_key": "nim", "label": "NIM", "field_type": "TEXT", "required": true, "options": []any{}},
		{"field_key": "email", "label": "Email", "field_type": "EMAIL", "required": true, "options": []any{}},
		{"field_key": "phone", "label": "No HP", "field_type": "TEXT", "required": true, "options": []any{}},
		{"field_key": "motivation", "label": "Motivasi", "field_type": "TEXTAREA", "required": true, "options": []any{}},
		{"field_key": "score", "label": "Nilai", "field_type": "NUMBER", "required": true, "options": []any{}},
		{"field_key": "unit", "label": "Unit", "field_type": "SELECT", "required": true, "options": []any{map[string]string{"value": "media", "label": "Media"}, map[string]string{"value": "acara", "label": "Acara"}}},
		{"field_key": "shift", "label": "Shift", "field_type": "RADIO", "required": true, "options": []any{map[string]string{"value": "pagi", "label": "Pagi"}, map[string]string{"value": "sore", "label": "Sore"}}},
		{"field_key": "interest", "label": "Minat", "field_type": "CHECKBOX", "required": true, "options": []any{map[string]string{"value": "web", "label": "Web"}, map[string]string{"value": "data", "label": "Data"}}},
		{"field_key": "available_date", "label": "Tanggal", "field_type": "DATE", "required": true, "options": []any{}},
	}
	formA := createRegistrationFixture(t, engine, superToken, periodID, divisionA, "Phase7 Dynamic Form A", now)
	formB := createRegistrationFixture(t, engine, superToken, periodID, divisionB, "Phase7 Dynamic Form B", now)
	_ = createRegistrationFixture(t, engine, superToken, periodID, divisionA, "Phase7 Public Draft", now)
	if res := contentRequest(engine, http.MethodPut, fmt.Sprintf("/api/v1/admin/registrations/%d/fields", formA), mustJSON(t, fields), superToken); res.Code != http.StatusOK {
		t.Fatalf("save dynamic schema: %d %s", res.Code, res.Body.String())
	}
	if res := contentRequest(engine, http.MethodPut, fmt.Sprintf("/api/v1/admin/registrations/%d/fields", formB), mustJSON(t, coreRegistrationFields()), superToken); res.Code != http.StatusOK {
		t.Fatalf("save cross-division schema: %d %s", res.Code, res.Body.String())
	}
	for _, id := range []int64{formA, formB} {
		res := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/registrations/%d/open", id), `{}`, superToken)
		if res.Code != http.StatusOK {
			t.Fatalf("open form %d: %d %s", id, res.Code, res.Body.String())
		}
	}
	// This intentionally remains DRAFT and must never appear in public results.
	publicList := getWithToken(engine, "/api/v1/public/registrations?status=OPEN&page=1&limit=20", "")
	if publicList.Code != http.StatusOK || strings.Contains(publicList.Body.String(), "Phase7 Public Draft") {
		t.Fatalf("public registration listing exposed draft: %d %s", publicList.Code, publicList.Body.String())
	}
	detail := getWithToken(engine, fmt.Sprintf("/api/v1/public/registrations/%d", formA), "")
	if detail.Code != http.StatusNotFound {
		t.Fatalf("numeric public slug should not expose arbitrary registration, got %d", detail.Code)
	}
	var slug string
	if err := db.QueryRowContext(ctx, `SELECT slug FROM registration_forms WHERE id=?`, formA).Scan(&slug); err != nil {
		t.Fatal(err)
	}
	detail = getWithToken(engine, "/api/v1/public/registrations/"+slug, "")
	if detail.Code != http.StatusOK || strings.Contains(detail.Body.String(), "created_by") || strings.Contains(detail.Body.String(), "submission_count") {
		t.Fatalf("public detail is not safe metadata: %d %s", detail.Code, detail.Body.String())
	}
	var detailPayload struct {
		Data struct {
			Status string `json:"status"`
			ID     int64  `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(detail.Body.Bytes(), &detailPayload); err != nil || detailPayload.Data.Status != "OPEN" || detailPayload.Data.ID != formA {
		t.Fatalf("public detail state/id incorrect: %v %+v", err, detailPayload)
	}

	makeAnswers := func(nim, email, motivation string) map[string]any {
		return map[string]any{"full_name": "Applicant Formula", "nim": nim, "email": email, "phone": "+62 812-3456", "motivation": motivation, "score": "4.25", "unit": "media", "shift": "pagi", "interest": []string{"web", "data"}, "available_date": "2026-11-01"}
	}
	answers := makeAnswers(" 26A001 ", "Shared@Example.test", "=HYPERLINK(\"https://attacker.invalid\",\"click\")")
	body := mustJSON(t, map[string]any{"answers": answers})
	submitted := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), body, "")
	if submitted.Code != http.StatusCreated {
		t.Fatalf("valid dynamic submission: %d %s", submitted.Code, submitted.Body.String())
	}
	oversizedBody := `{"answers":{"full_name":"` + strings.Repeat("x", service.RegistrationMaxRequestBytes+1) + `"}}`
	oversized := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), oversizedBody, "")
	if oversized.Code != http.StatusRequestEntityTooLarge || !strings.Contains(oversized.Body.String(), "REQUEST_TOO_LARGE") {
		t.Fatalf("oversized public submission did not return 413: %d %s", oversized.Code, oversized.Body.String())
	}
	for _, private := range []string{"Applicant Formula", "26A001", "Shared@Example.test", "=HYPERLINK"} {
		if strings.Contains(submitted.Body.String(), private) {
			t.Fatalf("public receipt disclosed applicant value %q: %s", private, submitted.Body.String())
		}
	}
	var receipt struct {
		Data struct {
			SubmissionID int64 `json:"submission_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(submitted.Body.Bytes(), &receipt); err != nil || receipt.Data.SubmissionID < 1 {
		t.Fatalf("invalid safe receipt: %v %s", err, submitted.Body.String())
	}
	var storedNIM, normalized, storedEmail string
	if err := db.QueryRowContext(ctx, `SELECT nim,normalized_nim,email FROM registration_submissions WHERE id=?`, receipt.Data.SubmissionID).Scan(&storedNIM, &normalized, &storedEmail); err != nil || storedNIM != "26A001" || normalized != "26A001" || storedEmail != "shared@example.test" {
		t.Fatalf("submission identity normalization mismatch: %q %q %q err=%v", storedNIM, normalized, storedEmail, err)
	}
	var coreAnswerValue sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT value_text FROM registration_submission_answers WHERE submission_id=? AND field_key_snapshot='full_name'`, receipt.Data.SubmissionID).Scan(&coreAnswerValue); err != nil || coreAnswerValue.Valid {
		t.Fatalf("core identity was duplicated in answer snapshot: %v err=%v", coreAnswerValue, err)
	}
	duplicate := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), mustJSON(t, map[string]any{"answers": makeAnswers("26a001", "different@example.test", "duplicate")}), "")
	if duplicate.Code != http.StatusConflict || !strings.Contains(duplicate.Body.String(), "DUPLICATE_SUBMISSION") || strings.Contains(duplicate.Body.String(), "Applicant Formula") {
		t.Fatalf("duplicate normalized NIM was not safely rejected: %d %s", duplicate.Code, duplicate.Body.String())
	}
	// NIM uniqueness is scoped to one form; reuse the same email is allowed.
	otherForm := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formB), mustJSON(t, map[string]any{"answers": map[string]string{"full_name": "Second Applicant", "nim": "26A001", "email": "shared@example.test", "phone": "0812345678"}}), "")
	if otherForm.Code != http.StatusCreated {
		t.Fatalf("NIM/email policy across a different form: %d %s", otherForm.Code, otherForm.Body.String())
	}

	// Every dynamic field type is server validated; unknown fields and forged choice values fail.
	invalid := makeAnswers("26A002", "another@example.test", "valid")
	invalid["unexpected"] = "tamper"
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), mustJSON(t, map[string]any{"answers": invalid}), ""); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unknown submitted field accepted: %d %s", response.Code, response.Body.String())
	}
	invalid = makeAnswers("26A003", "another@example.test", "valid")
	invalid["unit"] = "admin"
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), mustJSON(t, map[string]any{"answers": invalid}), ""); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("forged option value accepted: %d %s", response.Code, response.Body.String())
	}
	invalid = makeAnswers("26A004", "another@example.test", "valid")
	delete(invalid, "motivation")
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), mustJSON(t, map[string]any{"answers": invalid}), ""); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing required dynamic field accepted: %d %s", response.Code, response.Body.String())
	}
	invalid = makeAnswers("26A005", "another@example.test", "valid")
	invalid["score"] = 4.25
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), mustJSON(t, map[string]any{"answers": invalid}), ""); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("wrong JSON type for a dynamic field accepted: %d %s", response.Code, response.Body.String())
	}
	invalid = makeAnswers("26A006", "another@example.test", strings.Repeat("x", service.RegistrationMaxTextArea+1))
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), mustJSON(t, map[string]any{"answers": invalid}), ""); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("overlong TEXTAREA answer accepted: %d %s", response.Code, response.Body.String())
	}
	duplicateAnswerKey := `{"answers":{"full_name":"Duplicate Key","nim":"26DUPKEY","email":"dupkey@example.test","phone":"0812345678","motivation":"first","motivation":"second","score":"4.25","unit":"media","shift":"pagi","interest":["web"],"available_date":"2026-11-01"}}`
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), duplicateAnswerKey, ""); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("duplicate answer key accepted: %d %s", response.Code, response.Body.String())
	}

	// Same-form duplicate race must be resolved by DB uniqueness, not an in-memory check.
	raceAnswers := mustJSON(t, map[string]any{"answers": makeAnswers("26RACE", "race@example.test", "concurrent")})
	var group sync.WaitGroup
	statuses := make(chan int, 2)
	for i := 0; i < 2; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), raceAnswers, "")
			statuses <- response.Code
		}()
	}
	group.Wait()
	close(statuses)
	created, conflicts := 0, 0
	for status := range statuses {
		if status == http.StatusCreated {
			created++
		}
		if status == http.StatusConflict {
			conflicts++
		}
	}
	if created != 1 || conflicts != 1 {
		t.Fatalf("concurrent duplicate submission outcomes: created=%d conflicts=%d", created, conflicts)
	}

	fields[4]["label"] = "Motivasi Histori Baru"
	if response := contentRequest(engine, http.MethodPut, fmt.Sprintf("/api/v1/admin/registrations/%d/fields", formA), mustJSON(t, fields), superToken); response.Code != http.StatusOK {
		t.Fatalf("safe label edit after submission denied: %d %s", response.Code, response.Body.String())
	}
	changed := append([]map[string]any(nil), fields...)
	changed = append(changed, map[string]any{"field_key": "new_field", "label": "New", "field_type": "TEXT", "required": false, "options": []any{}})
	if response := contentRequest(engine, http.MethodPut, fmt.Sprintf("/api/v1/admin/registrations/%d/fields", formA), mustJSON(t, changed), superToken); response.Code != http.StatusConflict {
		t.Fatalf("structural schema change after submission accepted: %d %s", response.Code, response.Body.String())
	}
	var snapshotLabel string
	if err := db.QueryRowContext(ctx, `SELECT label_snapshot FROM registration_submission_answers WHERE submission_id=? AND field_key_snapshot='motivation'`, receipt.Data.SubmissionID).Scan(&snapshotLabel); err != nil || snapshotLabel != "Motivasi" {
		t.Fatalf("answer label snapshot changed with builder label: %q err=%v", snapshotLabel, err)
	}

	// Division scope is enforced for the form and every submission operation.
	for _, tc := range []struct {
		token  string
		form   int64
		status int
		label  string
	}{
		{kepalaAToken, formB, http.StatusNotFound, "KB cross-division form"},
		{kepalaBToken, formA, http.StatusNotFound, "KB reverse cross-division form"},
	} {
		if response := getWithToken(engine, fmt.Sprintf("/api/v1/admin/registrations/%d", tc.form), tc.token); response.Code != tc.status {
			t.Fatalf("%s: %d %s", tc.label, response.Code, response.Body.String())
		}
	}
	if response := getWithToken(engine, fmt.Sprintf("/api/v1/admin/registrations/%d", formB), bphToken); response.Code != http.StatusOK {
		t.Fatalf("BPH cross-division access denied: %d %s", response.Code, response.Body.String())
	}
	if response := getWithToken(engine, fmt.Sprintf("/api/v1/admin/registrations/%d", formB), superToken); response.Code != http.StatusOK {
		t.Fatalf("SUPER_ADMIN cross-division access denied: %d %s", response.Code, response.Body.String())
	}
	for _, tc := range []struct {
		method string
		path   string
		body   string
		label  string
	}{
		{http.MethodPatch, fmt.Sprintf("/api/v1/admin/registrations/%d", formB), `{"title":"Tampered cross-division"}`, "KB cross-division metadata update"},
		{http.MethodPut, fmt.Sprintf("/api/v1/admin/registrations/%d/fields", formB), mustJSON(t, coreRegistrationFields()), "KB cross-division schema update"},
		{http.MethodPost, fmt.Sprintf("/api/v1/admin/registrations/%d/open", formB), `{}`, "KB cross-division open"},
		{http.MethodPost, fmt.Sprintf("/api/v1/admin/registrations/%d/close", formB), `{}`, "KB cross-division close"},
	} {
		if response := contentRequest(engine, tc.method, tc.path, tc.body, kepalaAToken); response.Code != http.StatusNotFound {
			t.Fatalf("%s was not denied: %d %s", tc.label, response.Code, response.Body.String())
		}
	}
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/registrations", registrationCreateJSON(t, periodID, divisionB, "KB Wrong Division", now), kepalaAToken); response.Code != http.StatusForbidden {
		t.Fatalf("KEPALA_BIDANG created form in other division: %d %s", response.Code, response.Body.String())
	}
	for _, path := range []string{
		fmt.Sprintf("/api/v1/admin/registrations/%d/submissions", formB),
		fmt.Sprintf("/api/v1/admin/registrations/%d/submissions/1", formB),
		fmt.Sprintf("/api/v1/admin/registrations/%d/export", formB),
	} {
		if response := getWithToken(engine, path, kepalaAToken); response.Code != http.StatusNotFound {
			t.Fatalf("cross-division nested resource %s exposed: %d %s", path, response.Code, response.Body.String())
		}
	}
	var divisionBefore sql.NullInt64
	if err := db.QueryRowContext(ctx, `SELECT division_id FROM users WHERE email='kepala@example.org'`).Scan(&divisionBefore); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET division_id=NULL WHERE email='kepala@example.org'`); err != nil {
		t.Fatal(err)
	}
	if response := getWithToken(engine, "/api/v1/admin/registrations", kepalaAToken); response.Code != http.StatusForbidden && response.Code != http.StatusUnauthorized {
		t.Fatalf("divisionless KEPALA_BIDANG should be denied: %d %s", response.Code, response.Body.String())
	}
	if _, err := db.ExecContext(ctx, `UPDATE users SET division_id=? WHERE email='kepala@example.org'`, divisionBefore.Int64); err != nil {
		t.Fatal(err)
	}

	// Admin list/detail are paginated and private; export writes formula-looking cells as strings.
	adminList := getWithToken(engine, fmt.Sprintf("/api/v1/admin/registrations/%d/submissions?page=1&limit=20", formA), bphToken)
	if adminList.Code != http.StatusOK || !strings.Contains(adminList.Body.String(), "Applicant Formula") {
		t.Fatalf("BPH submission list failed: %d %s", adminList.Code, adminList.Body.String())
	}
	detailResponse := getWithToken(engine, fmt.Sprintf("/api/v1/admin/registrations/%d/submissions/%d", formA, receipt.Data.SubmissionID), superToken)
	if detailResponse.Code != http.StatusOK || !strings.Contains(detailResponse.Body.String(), `"label":"Motivasi"`) || !strings.Contains(detailResponse.Body.String(), "Applicant Formula") {
		t.Fatalf("submission detail/snapshot failed: %d %s", detailResponse.Code, detailResponse.Body.String())
	}

	request := httptestRequest(t, fmt.Sprintf("/api/v1/admin/registrations/%d/export", formA), bphToken)
	exportResponse := httptestServe(engine, request)
	if exportResponse.Code != http.StatusOK || exportResponse.Header().Get("Content-Type") != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" || exportResponse.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("XLSX export failed: %d %s", exportResponse.Code, exportResponse.Body.String())
	}
	book, err := excelize.OpenReader(bytes.NewReader(exportResponse.Body.Bytes()))
	if err != nil {
		t.Fatalf("open exported xlsx: %v", err)
	}
	defer book.Close()
	rows, err := book.GetRows("Pendaftar")
	if err != nil || len(rows) != 3 || len(rows[0]) != 12 {
		t.Fatalf("XLSX row/column count mismatch: rows=%d err=%v", len(rows), err)
	}
	for column, want := range []string{"No.", "Nama", "NIM", "Email", "No HP", "Waktu Pendaftaran", "Motivasi", "Nilai", "Unit", "Shift", "Minat", "Tanggal"} {
		cell, _ := excelize.CoordinatesToCellName(column+1, 1)
		if got, err := book.GetCellValue("Pendaftar", cell); err != nil || got != want {
			t.Fatalf("XLSX header %s mismatch: got=%q want=%q err=%v", cell, got, want, err)
		}
	}
	if got, err := book.GetCellValue("Pendaftar", "G2"); err != nil || got != `=HYPERLINK("https://attacker.invalid","click")` {
		t.Fatalf("formula-like applicant answer missing as literal string: %q err=%v", got, err)
	}
	if got, err := book.GetCellValue("Pendaftar", "K2"); err != nil || got != "web; data" {
		t.Fatalf("checkbox values were not serialized: %q err=%v", got, err)
	}
	if got, err := book.GetCellValue("Pendaftar", "L2"); err != nil || got != "2026-11-01" {
		t.Fatalf("date answer was not preserved: %q err=%v", got, err)
	}
	if formula, err := book.GetCellFormula("Pendaftar", "G2"); err != nil || formula != "" {
		t.Fatalf("formula injection produced executable formula %q err=%v", formula, err)
	}
	var exportEvents int
	var exportDetails string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),MAX(CAST(details_json AS CHAR)) FROM activity_logs WHERE action='EXPORT' AND entity_type='REGISTRATION_SUBMISSIONS' AND entity_id=?`, fmt.Sprint(formA)).Scan(&exportEvents, &exportDetails); err != nil || exportEvents != 1 || strings.Contains(exportDetails, "Applicant Formula") || strings.Contains(exportDetails, "shared@example.test") {
		t.Fatalf("export audit event leaked PII: count=%d details=%s err=%v", exportEvents, exportDetails, err)
	}
	if strings.Contains(logger.String(), "Applicant Formula") || strings.Contains(logger.String(), "26A001") || strings.Contains(logger.String(), "Shared@Example.test") {
		t.Fatal("registration PII appeared in structured request logs")
	}

	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/registrations/%d/close", formA), `{}`, bphToken); response.Code != http.StatusOK {
		t.Fatalf("BPH close cross-division registration failed: %d %s", response.Code, response.Body.String())
	}
	closedSubmit := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), mustJSON(t, map[string]any{"answers": makeAnswers("26CLOSED", "closed@example.test", "late")}), "")
	if closedSubmit.Code != http.StatusConflict || !strings.Contains(closedSubmit.Body.String(), "REGISTRATION_CLOSED") {
		t.Fatalf("closed registration accepted submission: %d %s", closedSubmit.Code, closedSubmit.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/registrations/%d/open", formA), `{}`, superToken); response.Code != http.StatusConflict {
		t.Fatalf("closed form reopened: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/registrations/%d", formA), "", superToken); response.Code != http.StatusConflict {
		t.Fatalf("form with submissions was deleted: %d %s", response.Code, response.Body.String())
	}

	// Repeated unique attempts are bounded independently from login and aspirations.
	limitedCfg := config.Config{AppEnv: "test", FrontendOrigins: []string{"http://localhost:3000"}, JWTSecret: bytes.Repeat([]byte("r"), 32), AccessTokenTTL: time.Minute * 15, RefreshTokenTTL: time.Hour, Cookie: config.Cookie{Name: "bem_refresh", SameSite: "lax", Path: "/api/v1/auth"}, BcryptCost: 4, LoginLimitCount: 100, LoginLimitWindow: time.Minute, AspirationLimitCount: 100, AspirationLimitWindow: time.Minute, RegistrationLimitCount: 1, RegistrationLimitWindow: time.Minute * 30}
	limitedEngine, err := NewWithDependencies(slog.New(slog.NewJSONHandler(io.Discard, nil)), limitedCfg, db)
	if err != nil {
		t.Fatal(err)
	}
	firstRate := contentRequest(limitedEngine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), `{}`, "")
	secondRate := contentRequest(limitedEngine, http.MethodPost, fmt.Sprintf("/api/v1/public/registrations/%d/submissions", formA), `{}`, "")
	if firstRate.Code != http.StatusUnprocessableEntity || secondRate.Code != http.StatusTooManyRequests || secondRate.Header().Get("Retry-After") == "" {
		t.Fatalf("registration submission limiter failed: first=%d second=%d retry=%s", firstRate.Code, secondRate.Code, secondRate.Header().Get("Retry-After"))
	}
}

func createRegistrationFixture(t *testing.T, engine http.Handler, token string, periodID, divisionID int64, title string, now time.Time) int64 {
	t.Helper()
	response := contentRequest(engine, http.MethodPost, "/api/v1/admin/registrations", registrationCreateJSON(t, periodID, divisionID, title, now), token)
	if response.Code != http.StatusCreated {
		t.Fatalf("create registration %q: %d %s", title, response.Code, response.Body.String())
	}
	return responseID(t, response.Body.String())
}

func registrationCreateJSON(t *testing.T, periodID, divisionID int64, title string, now time.Time) string {
	t.Helper()
	return mustJSON(t, map[string]any{"period_id": periodID, "division_id": divisionID, "title": title, "open_at": now.Add(-time.Hour).Format(time.RFC3339Nano), "close_at": now.Add(48 * time.Hour).Format(time.RFC3339Nano)})
}

func coreRegistrationFields() []map[string]any {
	return []map[string]any{
		{"field_key": "full_name", "label": "Nama Lengkap", "field_type": "TEXT", "required": true, "options": []any{}},
		{"field_key": "nim", "label": "NIM", "field_type": "TEXT", "required": true, "options": []any{}},
		{"field_key": "email", "label": "Email", "field_type": "EMAIL", "required": true, "options": []any{}},
		{"field_key": "phone", "label": "No HP", "field_type": "TEXT", "required": true, "options": []any{}},
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func httptestRequest(t *testing.T, path, token string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "http://localhost"+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func httptestServe(engine http.Handler, request *http.Request) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	return response
}
