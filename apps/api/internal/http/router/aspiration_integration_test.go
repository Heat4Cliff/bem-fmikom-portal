package router

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/config"
)

func runAspirationIntegration(t *testing.T, ctx context.Context, db *sql.DB, engine http.Handler, cfg config.Config, logger *slog.Logger, logs *bytes.Buffer, superToken, bphToken, kepalaAToken, kepalaBToken string) {
	t.Helper()

	periodID := createPeriod(t, engine, superToken, `{"name":"Phase6 Active Period","start_date":"2031-01-01","end_date":"2031-12-31"}`)
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/periods/%d/activate", periodID), `{}`, superToken); response.Code != http.StatusOK {
		t.Fatalf("activate aspiration test period=%d %s", response.Code, response.Body.String())
	}
	divisionA := createDivision(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"name":"Phase6 Divisi A","status":"ACTIVE"}`, periodID))
	divisionB := createDivision(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"name":"Phase6 Divisi B","status":"ACTIVE"}`, periodID))
	createDivision(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"name":"Fakultas","status":"ACTIVE","show_in_public_structure":false}`, periodID))
	inactiveDivision := createDivision(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"name":"Phase6 Divisi Inactive","status":"INACTIVE"}`, periodID))
	archivedPeriod := createPeriod(t, engine, superToken, `{"name":"Phase6 Archived Period","start_date":"2032-01-01","end_date":"2032-12-31"}`)
	archivedDivision := createDivision(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"name":"Phase6 Divisi Archived","status":"ACTIVE"}`, archivedPeriod))
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/divisions", "", ""); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Fakultas") || strings.Contains(response.Body.String(), "Phase6 Divisi Inactive") || strings.Contains(response.Body.String(), "Phase6 Divisi Archived") {
		t.Fatalf("public division options included inactive or archived division: %d %s", response.Code, response.Body.String())
	}

	var oldA, oldB int64
	if err := db.QueryRowContext(ctx, `SELECT division_id FROM users WHERE email='kepala@example.org'`).Scan(&oldA); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT division_id FROM users WHERE email='kepala-b@example.org'`).Scan(&oldB); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.ExecContext(context.Background(), `UPDATE users SET division_id=CASE email WHEN 'kepala@example.org' THEN ? WHEN 'kepala-b@example.org' THEN ? END WHERE email IN ('kepala@example.org','kepala-b@example.org')`, oldA, oldB); err != nil {
			t.Errorf("restore integration user divisions: %v", err)
		}
	}()
	if _, err := db.ExecContext(ctx, `UPDATE users SET division_id=CASE email WHEN 'kepala@example.org' THEN ? WHEN 'kepala-b@example.org' THEN ? END WHERE email IN ('kepala@example.org','kepala-b@example.org')`, divisionA, divisionB); err != nil {
		t.Fatal(err)
	}

	identifiedBody := fmt.Sprintf(`{"division_id":%d,"name":"ASPIRATION-PII-NAME","nim":"ASPIRATION-PII-NIM","email":"aspiration-pii@example.test","subject":"Identified Phase6 issue","message":"ASPIRATION-PII-MESSAGE <script>alert(1)</script>","is_anonymous":false}`, divisionA)
	identified := contentRequest(engine, http.MethodPost, "/api/v1/public/aspirations", identifiedBody, "")
	if identified.Code != http.StatusCreated {
		t.Fatalf("identified aspiration submission=%d %s", identified.Code, identified.Body.String())
	}
	identifiedID := responseID(t, identified.Body.String())
	for _, private := range []string{"ASPIRATION-PII-NAME", "ASPIRATION-PII-NIM", "aspiration-pii@example.test", "ASPIRATION-PII-MESSAGE", "<script>"} {
		if strings.Contains(identified.Body.String(), private) {
			t.Fatalf("public receipt leaked submitted content %q: %s", private, identified.Body.String())
		}
	}
	if !strings.Contains(identified.Body.String(), `"status":"PENDING"`) || !strings.Contains(identified.Body.String(), `"submitted_at":`) {
		t.Fatalf("safe confirmation receipt missing status/time: %s", identified.Body.String())
	}
	var storedName, storedNIM, storedEmail string
	if err := db.QueryRowContext(ctx, `SELECT name,nim,email FROM aspirations WHERE id=?`, identifiedID).Scan(&storedName, &storedNIM, &storedEmail); err != nil || storedName != "ASPIRATION-PII-NAME" || storedNIM != "ASPIRATION-PII-NIM" || storedEmail != "aspiration-pii@example.test" {
		t.Fatalf("identified aspiration identity was not stored as submitted: name=%q nim=%q email=%q err=%v", storedName, storedNIM, storedEmail, err)
	}

	anonymousBody := fmt.Sprintf(`{"division_id":%d,"name":"ASPIRATION-ANON-NAME","nim":"ASPIRATION-ANON-NIM","email":"anon-secret@example.test","subject":"Anonymous Phase6 issue","message":"ASPIRATION-ANON-MESSAGE","is_anonymous":true}`, divisionB)
	anonymous := contentRequest(engine, http.MethodPost, "/api/v1/public/aspirations", anonymousBody, "")
	if anonymous.Code != http.StatusCreated {
		t.Fatalf("anonymous aspiration submission=%d %s", anonymous.Code, anonymous.Body.String())
	}
	anonymousID := responseID(t, anonymous.Body.String())
	for _, private := range []string{"ASPIRATION-ANON-NAME", "ASPIRATION-ANON-NIM", "anon-secret@example.test", "ASPIRATION-ANON-MESSAGE"} {
		if strings.Contains(anonymous.Body.String(), private) {
			t.Fatalf("anonymous receipt leaked submitted identity/content %q: %s", private, anonymous.Body.String())
		}
	}
	var name, nim, email sql.NullString
	var isAnonymous bool
	if err := db.QueryRowContext(ctx, `SELECT name,nim,email,is_anonymous FROM aspirations WHERE id=?`, anonymousID).Scan(&name, &nim, &email, &isAnonymous); err != nil || !isAnonymous || name.Valid || nim.Valid || email.Valid {
		t.Fatalf("anonymous identity was not stripped: name=%v nim=%v email=%v anonymous=%v err=%v", name, nim, email, isAnonymous, err)
	}
	var storedMessage string
	if err := db.QueryRowContext(ctx, `SELECT message FROM aspirations WHERE id=?`, identifiedID).Scan(&storedMessage); err != nil || !strings.Contains(storedMessage, "<script>") {
		t.Fatalf("plain-text message was transformed unexpectedly: %q err=%v", storedMessage, err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO aspirations(division_id,name,subject,message,is_anonymous,status,submitted_at) VALUES(?, 'spoofed identity', 'Constraint test', 'Body', TRUE, 'PENDING', UTC_TIMESTAMP(6))`, divisionA); err == nil {
		t.Fatal("database accepted identity fields on an anonymous aspiration")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO aspirations(division_id,subject,message,is_anonymous,status,submitted_at) VALUES(999999, 'Constraint test', 'Body', TRUE, 'PENDING', UTC_TIMESTAMP(6))`); err == nil {
		t.Fatal("database accepted a missing aspiration division foreign key")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO aspirations(division_id,subject,message,is_anonymous,status,submitted_at) VALUES(?, 'Constraint test', 'Body', TRUE, 'REJECTED', UTC_TIMESTAMP(6))`, divisionA); err == nil {
		t.Fatal("database accepted a status outside the fixed aspiration workflow")
	}

	invalidBodies := []struct {
		name string
		body string
	}{
		{"missing subject", fmt.Sprintf(`{"division_id":%d,"message":"missing subject","is_anonymous":true}`, divisionA)},
		{"invalid email", fmt.Sprintf(`{"division_id":%d,"email":"not-an-email","subject":"Subject","message":"Body","is_anonymous":false}`, divisionA)},
		{"message too long", fmt.Sprintf(`{"division_id":%d,"subject":"Subject","message":%q,"is_anonymous":true}`, divisionA, strings.Repeat("x", 5001))},
		{"unknown division", `{"division_id":999999,"subject":"Subject","message":"Body","is_anonymous":true}`},
		{"inactive division", fmt.Sprintf(`{"division_id":%d,"subject":"Subject","message":"Body","is_anonymous":true}`, inactiveDivision)},
		{"archived period division", fmt.Sprintf(`{"division_id":%d,"subject":"Subject","message":"Body","is_anonymous":true}`, archivedDivision)},
	}
	for _, test := range invalidBodies {
		t.Run(test.name, func(t *testing.T) {
			response := contentRequest(engine, http.MethodPost, "/api/v1/public/aspirations", test.body, "")
			if response.Code != http.StatusUnprocessableEntity {
				t.Fatalf("invalid request accepted: %d %s", response.Code, response.Body.String())
			}
		})
	}
	if response := postJSONWithOrigin(engine, "/api/v1/public/aspirations", identifiedBody, "https://attacker.example"); response.Code != http.StatusForbidden {
		t.Fatalf("untrusted browser Origin should be rejected, got %d %s", response.Code, response.Body.String())
	}

	listA := contentRequest(engine, http.MethodGet, "/api/v1/admin/aspirations?limit=100", "", kepalaAToken)
	if listA.Code != http.StatusOK || !strings.Contains(listA.Body.String(), fmt.Sprintf(`"id":%d`, identifiedID)) || strings.Contains(listA.Body.String(), fmt.Sprintf(`"id":%d`, anonymousID)) {
		t.Fatalf("KB_A list did not enforce division scope: %d %s", listA.Code, listA.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/aspirations/%d", anonymousID), "", kepalaAToken); response.Code != http.StatusNotFound {
		t.Fatalf("KB_A cross-division detail should be concealed: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/aspirations/%d/status", anonymousID), `{"status":"READ"}`, kepalaAToken); response.Code != http.StatusForbidden {
		t.Fatalf("KB_A cross-division status mutation should be forbidden: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/aspirations/%d/status", identifiedID), `{"status":"READ"}`, kepalaBToken); response.Code != http.StatusForbidden {
		t.Fatalf("KB_B cross-division status mutation should be forbidden: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/admin/aspirations?limit=100", "", bphToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), fmt.Sprintf(`"id":%d`, identifiedID)) || !strings.Contains(response.Body.String(), fmt.Sprintf(`"id":%d`, anonymousID)) {
		t.Fatalf("BPH could not read cross-division aspirations: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/aspirations/%d", anonymousID), "", superToken); response.Code != http.StatusOK {
		t.Fatalf("SUPER_ADMIN could not read cross-division aspiration: %d %s", response.Code, response.Body.String())
	}

	read := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/aspirations/%d/status", identifiedID), `{"status":"READ"}`, kepalaAToken)
	if read.Code != http.StatusOK || !strings.Contains(read.Body.String(), `"status":"READ"`) {
		t.Fatalf("PENDING to READ failed: %d %s", read.Code, read.Body.String())
	}
	var readAt time.Time
	if err := db.QueryRowContext(ctx, `SELECT read_at FROM aspirations WHERE id=?`, identifiedID).Scan(&readAt); err != nil || readAt.IsZero() || readAt.Location() != time.UTC {
		t.Fatalf("read_at missing or not UTC: %v err=%v", readAt, err)
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/aspirations/%d/status", anonymousID), `{"status":"RESOLVED"}`, bphToken); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("direct PENDING to RESOLVED should be rejected by sequential workflow: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/aspirations/%d/status", identifiedID), `{"status":"RESOLVED"}`, bphToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"status":"RESOLVED"`) {
		t.Fatalf("READ to RESOLVED failed: %d %s", response.Code, response.Body.String())
	}
	var resolvedAt time.Time
	if err := db.QueryRowContext(ctx, `SELECT resolved_at FROM aspirations WHERE id=?`, identifiedID).Scan(&resolvedAt); err != nil || resolvedAt.IsZero() || resolvedAt.Location() != time.UTC || !resolvedAt.After(readAt) {
		t.Fatalf("resolved_at missing, non-UTC, or before read_at: read=%v resolved=%v err=%v", readAt, resolvedAt, err)
	}
	if response := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/aspirations/%d/status", identifiedID), `{"status":"READ"}`, superToken); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("resolved aspiration was reopened: %d %s", response.Code, response.Body.String())
	}

	limitedConfig := cfg
	limitedConfig.AspirationLimitCount = 1
	limitedEngine, err := NewWithDependencies(logger, limitedConfig, db)
	if err != nil {
		t.Fatalf("create aspiration-limited router: %v", err)
	}
	limitedBody := fmt.Sprintf(`{"division_id":%d,"subject":"Limited test","message":"Rate check","is_anonymous":true}`, divisionA)
	first := contentRequest(limitedEngine, http.MethodPost, "/api/v1/public/aspirations", limitedBody, "")
	second := contentRequest(limitedEngine, http.MethodPost, "/api/v1/public/aspirations", limitedBody, "")
	if first.Code != http.StatusCreated || second.Code != http.StatusTooManyRequests || second.Header().Get("Retry-After") == "" {
		t.Fatalf("aspiration IP rate limit not enforced: first=%d second=%d headers=%v", first.Code, second.Code, second.Header())
	}

	var details string
	if err := db.QueryRowContext(ctx, `SELECT details_json FROM activity_logs WHERE entity_type='ASPIRATION' AND entity_id=? AND action='UPDATE' ORDER BY id DESC LIMIT 1`, fmt.Sprint(identifiedID)).Scan(&details); err != nil {
		t.Fatalf("aspiration status activity missing: %v", err)
	}
	for _, marker := range []string{"ASPIRATION-PII-NAME", "ASPIRATION-PII-NIM", "aspiration-pii@example.test", "ASPIRATION-PII-MESSAGE", "<script>", "ASPIRATION-ANON-NAME", "ASPIRATION-ANON-NIM", "anon-secret@example.test", "ASPIRATION-ANON-MESSAGE"} {
		if strings.Contains(details, marker) || strings.Contains(logs.String(), marker) {
			t.Fatalf("aspiration PII leaked to activity details or structured request logs: %s", marker)
		}
	}
	if !strings.Contains(details, `"old_status":"READ"`) || !strings.Contains(details, `"new_status":"RESOLVED"`) || !strings.Contains(details, `"division_id":`) {
		t.Fatalf("activity details lack safe status metadata: %s", details)
	}
}
