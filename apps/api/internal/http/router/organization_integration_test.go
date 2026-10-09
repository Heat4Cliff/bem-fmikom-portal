package router

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

func runOrganizationIntegration(t *testing.T, ctx context.Context, db *sql.DB, engine http.Handler, superToken, bphToken, kepalaToken string) {
	t.Helper()

	active := contentRequest(engine, http.MethodGet, "/api/v1/public/periods/active", "", "")
	if active.Code != http.StatusOK || !strings.Contains(active.Body.String(), `"name":"2025/2026"`) {
		t.Fatalf("public active period = %d %s", active.Code, active.Body.String())
	}
	if !strings.Contains(active.Body.String(), `"start_date":"2025-01-01"`) || strings.Contains(active.Body.String(), "created_at") {
		t.Fatalf("public period exposes wrong shape: %s", active.Body.String())
	}

	periodA := createPeriod(t, engine, superToken, `{"name":"Phase5 A","start_date":"2027-01-01","end_date":"2027-12-31"}`)
	periodB := createPeriod(t, engine, superToken, `{"name":"Phase5 B","start_date":"2028-01-01","end_date":"2028-12-31"}`)
	wrongDate := contentRequest(engine, http.MethodPost, "/api/v1/admin/periods", `{"name":"Bad range","start_date":"2028-12-31","end_date":"2028-01-01"}`, superToken)
	if wrongDate.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid period date range status=%d body=%s", wrongDate.Code, wrongDate.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/periods", `{"name":"No","start_date":"2029-01-01","end_date":"2029-12-31"}`, bphToken); response.Code != http.StatusForbidden {
		t.Fatalf("BPH period create status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/periods/%d/activate", periodA), `{}`, bphToken); response.Code != http.StatusForbidden {
		t.Fatalf("BPH period activation status=%d", response.Code)
	}

	divA := createDivision(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"name":"Phase5 Divisi A","status":"ACTIVE","display_order":1}`, periodA))
	divB := createDivision(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"name":"Phase5 Divisi B","status":"ACTIVE","display_order":0}`, periodA))
	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/divisions/%d", divA), "", kepalaToken); response.Code != http.StatusNotFound {
		t.Fatalf("KB cross-division read status=%d", response.Code)
	}

	createMember(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"name":"Phase5 Ketua","position":"KETUA","status":"ACTIVE","display_order":0}`, periodA))
	headA := createMember(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"division_id":%d,"name":"Phase5 Kepala A","position":"KEPALA_BIDANG","status":"ACTIVE","display_order":0}`, periodA, divA))
	createMember(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"division_id":%d,"name":"Phase5 Anggota A","position":"ANGGOTA","status":"ACTIVE","display_order":1}`, periodA, divA))
	createMember(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"division_id":%d,"name":"Phase5 Kepala B","position":"KEPALA_BIDANG","status":"ACTIVE","display_order":0}`, periodA, divB))
	inactiveMember := createMember(t, engine, bphToken, fmt.Sprintf(`{"period_id":%d,"division_id":%d,"name":"Phase5 Tidak Publik","position":"ANGGOTA","status":"INACTIVE"}`, periodA, divA))
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/members", fmt.Sprintf(`{"period_id":%d,"name":"Kepala Tanpa Divisi","position":"KEPALA_BIDANG"}`, periodA), bphToken); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("head without division status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/members", fmt.Sprintf(`{"period_id":%d,"division_id":%d,"name":"Mismatch","position":"ANGGOTA"}`, periodB, divA), bphToken); response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("member/period mismatch status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/members", fmt.Sprintf(`{"period_id":%d,"division_id":%d,"name":"Duplikat Kepala","position":"KEPALA_BIDANG","status":"ACTIVE"}`, periodA, divA), bphToken); response.Code != http.StatusConflict {
		t.Fatalf("duplicate active head status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/members/%d", headA), "", kepalaToken); response.Code != http.StatusNotFound {
		t.Fatalf("KB cross-member detail status=%d", response.Code)
	}
	if response := contentRequest(engine, http.MethodPost, "/api/v1/admin/members", fmt.Sprintf(`{"period_id":%d,"division_id":%d,"name":"KB Spoof","position":"ANGGOTA"}`, periodA, divB), kepalaToken); response.Code != http.StatusForbidden {
		t.Fatalf("KB member mutation status=%d", response.Code)
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/members/%d", headA), "", kepalaToken); response.Code != http.StatusForbidden {
		t.Fatalf("KB member delete status=%d", response.Code)
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/divisions/%d", divA), "", bphToken); response.Code != http.StatusConflict {
		t.Fatalf("division dependency delete status=%d body=%s", response.Code, response.Body.String())
	}

	archivedID := int64(0)
	if err := db.QueryRowContext(ctx, `SELECT id FROM periods WHERE name='Archived'`).Scan(&archivedID); err != nil {
		t.Fatal(err)
	}
	var ownDivision int64
	if err := db.QueryRowContext(ctx, `SELECT division_id FROM users WHERE role='KEPALA_BIDANG' AND email='kepala@example.org'`).Scan(&ownDivision); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO organization_members(period_id,division_id,name,position,status) VALUES(?,?,'Phase5 Scoped KB','ANGGOTA','ACTIVE')`, archivedID, ownDivision); err != nil {
		t.Fatal(err)
	}
	kbMembers := contentRequest(engine, http.MethodGet, "/api/v1/admin/members?limit=100", "", kepalaToken)
	if kbMembers.Code != http.StatusOK || !strings.Contains(kbMembers.Body.String(), "Phase5 Scoped KB") || strings.Contains(kbMembers.Body.String(), "Phase5 Kepala A") {
		t.Fatalf("KB member list scope = %d %s", kbMembers.Code, kbMembers.Body.String())
	}
	kbDivisions := contentRequest(engine, http.MethodGet, "/api/v1/admin/divisions?limit=100", "", kepalaToken)
	if kbDivisions.Code != http.StatusOK || !strings.Contains(kbDivisions.Body.String(), fmt.Sprint(ownDivision)) {
		t.Fatalf("KB division read = %d %s", kbDivisions.Code, kbDivisions.Body.String())
	}

	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/periods/%d/activate", periodA), `{}`, superToken); response.Code != http.StatusOK {
		t.Fatalf("activate A status=%d body=%s", response.Code, response.Body.String())
	}
	var statuses struct{ A, B string }
	if err := db.QueryRowContext(ctx, `SELECT (SELECT status FROM periods WHERE id=?),(SELECT status FROM periods WHERE id=?)`, periodA, periodB).Scan(&statuses.A, &statuses.B); err != nil {
		t.Fatal(err)
	}
	if statuses.A != "ACTIVE" || statuses.B != "ARCHIVED" {
		t.Fatalf("period activation states A=%s B=%s", statuses.A, statuses.B)
	}
	var activeCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM periods WHERE status='ACTIVE'`).Scan(&activeCount); err != nil || activeCount != 1 {
		t.Fatalf("active period count=%d err=%v", activeCount, err)
	}
	publicOrg := contentRequest(engine, http.MethodGet, "/api/v1/public/organization", "", "")
	if publicOrg.Code != http.StatusOK || !strings.Contains(publicOrg.Body.String(), "Phase5 Ketua") || !strings.Contains(publicOrg.Body.String(), "Phase5 Kepala A") || strings.Contains(publicOrg.Body.String(), "Phase5 Tidak Publik") || strings.Contains(publicOrg.Body.String(), "created_at") || strings.Contains(publicOrg.Body.String(), "display_order") || strings.Contains(publicOrg.Body.String(), "INACTIVE") {
		t.Fatalf("public organization visibility/shape = %d %s", publicOrg.Code, publicOrg.Body.String())
	}
	publicDivisions := contentRequest(engine, http.MethodGet, "/api/v1/public/divisions", "", "")
	if publicDivisions.Code != http.StatusOK || strings.Index(publicDivisions.Body.String(), "Phase5 Divisi B") > strings.Index(publicDivisions.Body.String(), "Phase5 Divisi A") {
		t.Fatalf("public division order = %d %s", publicDivisions.Code, publicDivisions.Body.String())
	}
	if response := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/members/%d", inactiveMember), "", bphToken); response.Code != http.StatusNoContent {
		t.Fatalf("member soft-delete status=%d", response.Code)
	}
	var storedStatus string
	if err := db.QueryRowContext(ctx, `SELECT status FROM organization_members WHERE id=?`, inactiveMember).Scan(&storedStatus); err != nil || storedStatus != "INACTIVE" {
		t.Fatalf("member delete did not preserve as inactive: status=%s err=%v", storedStatus, err)
	}

	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/periods/%d/activate", periodB), `{}`, superToken); response.Code != http.StatusOK {
		t.Fatalf("activate B status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/organization", "", ""); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Phase5 B") || strings.Contains(response.Body.String(), "Phase5 Ketua") {
		t.Fatalf("public organization did not switch period: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, fmt.Sprintf("/api/v1/admin/members?period_id=%d&limit=100", periodA), "", superToken); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Phase5 Ketua") {
		t.Fatalf("historical members unavailable: %d %s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/periods/%d/archive", periodB), `{}`, superToken); response.Code != http.StatusOK {
		t.Fatalf("archive B status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/periods/active", "", ""); response.Code != http.StatusNotFound {
		t.Fatalf("no-active-period response status=%d body=%s", response.Code, response.Body.String())
	}
	if response := contentRequest(engine, http.MethodGet, "/api/v1/public/organization", "", ""); response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"period":null`) {
		t.Fatalf("public empty organization state=%d %s", response.Code, response.Body.String())
	}

	periodC := createPeriod(t, engine, superToken, `{"name":"Phase5 C","start_date":"2029-01-01","end_date":"2029-12-31"}`)
	periodD := createPeriod(t, engine, superToken, `{"name":"Phase5 D","start_date":"2030-01-01","end_date":"2030-12-31"}`)
	var wg sync.WaitGroup
	results := make(chan int, 2)
	for _, id := range []int64{periodC, periodD} {
		wg.Add(1)
		go func(periodID int64) {
			defer wg.Done()
			results <- contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/periods/%d/activate", periodID), `{}`, superToken).Code
		}(id)
	}
	wg.Wait()
	close(results)
	for code := range results {
		if code != http.StatusOK && code != http.StatusConflict {
			t.Fatalf("concurrent activation returned unexpected status %d", code)
		}
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM periods WHERE status='ACTIVE'`).Scan(&activeCount); err != nil || activeCount != 1 {
		t.Fatalf("concurrent activation violated active period invariant: count=%d err=%v", activeCount, err)
	}

	for _, entity := range []string{"PERIOD", "DIVISION", "ORGANIZATION_MEMBER"} {
		var count int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activity_logs WHERE entity_type=? AND action IN ('CREATE','UPDATE','DELETE','ACTIVATE','ARCHIVE')`, entity).Scan(&count); err != nil || count == 0 {
			t.Fatalf("missing organization activity rows for %s: count=%d err=%v", entity, count, err)
		}
	}
}

func createPeriod(t *testing.T, engine http.Handler, token, body string) int64 {
	t.Helper()
	response := contentRequest(engine, http.MethodPost, "/api/v1/admin/periods", body, token)
	if response.Code != http.StatusCreated {
		t.Fatalf("create period status=%d body=%s", response.Code, response.Body.String())
	}
	return responseID(t, response.Body.String())
}
func createDivision(t *testing.T, engine http.Handler, token, body string) int64 {
	t.Helper()
	response := contentRequest(engine, http.MethodPost, "/api/v1/admin/divisions", body, token)
	if response.Code != http.StatusCreated {
		t.Fatalf("create division status=%d body=%s", response.Code, response.Body.String())
	}
	return responseID(t, response.Body.String())
}
func createMember(t *testing.T, engine http.Handler, token, body string) int64 {
	t.Helper()
	response := contentRequest(engine, http.MethodPost, "/api/v1/admin/members", body, token)
	if response.Code != http.StatusCreated {
		t.Fatalf("create member status=%d body=%s", response.Code, response.Body.String())
	}
	return responseID(t, response.Body.String())
}
func responseID(t *testing.T, body string) int64 {
	t.Helper()
	var value struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &value); err != nil {
		t.Fatalf("parse response id: %v (%s)", err, body)
	}
	if value.Data.ID < 1 {
		t.Fatalf("response id missing: %s", body)
	}
	return value.Data.ID
}
