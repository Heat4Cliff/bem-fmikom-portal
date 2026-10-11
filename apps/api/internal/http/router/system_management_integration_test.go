package router

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func runSystemManagementIntegration(t *testing.T, ctx context.Context, db *sql.DB, engine http.Handler, adminToken, bphToken, kepalaToken string, divisionA, divisionB int64) {
	t.Helper()
	public := getWithToken(engine, "/api/v1/public/settings", "")
	if public.Code != http.StatusOK || !strings.Contains(public.Body.String(), `"hero_headline"`) || strings.Contains(public.Body.String(), `"updated_by"`) {
		t.Fatalf("public landing settings contract failed: %d %s", public.Code, public.Body.String())
	}
	if public.Header().Get("Cache-Control") != "public, max-age=60, stale-while-revalidate=300" {
		t.Fatalf("unexpected public settings cache: %s", public.Header().Get("Cache-Control"))
	}
	if got := getWithToken(engine, "/api/v1/admin/settings", ""); got.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated settings request = %d, want 401", got.Code)
	}
	if got := getWithToken(engine, "/api/v1/admin/settings", kepalaToken); got.Code != http.StatusForbidden {
		t.Fatalf("KEPALA_BIDANG settings request = %d, want 403", got.Code)
	}
	if got := getWithToken(engine, "/api/v1/admin/settings", bphToken); got.Code != http.StatusOK {
		t.Fatalf("BPH could not read settings: %d %s", got.Code, got.Body.String())
	}
	settings := map[string]any{
		"hero_headline": "Headline Uji Phase 9", "hero_subheadline": "Subheadline yang diperbarui untuk integrasi.",
		"hero_primary_cta_label": "Lihat berita", "hero_primary_cta_url": "/berita", "hero_secondary_cta_label": nil, "hero_secondary_cta_url": nil,
		"hero_image_url": "https://example.test/hero.jpg", "about_title": "Tentang Uji", "about_description": "Deskripsi konten biasa.",
		"vision": "Visi uji", "missions": []string{"Misi satu", "Misi dua"}, "contact_email": "bem@example.test", "contact_phone": nil,
		"contact_address": nil, "location_url": nil, "instagram_url": "https://instagram.com/example", "youtube_url": nil,
		"tiktok_url": nil, "other_social_url": nil,
	}
	encoded, _ := json.Marshal(settings)
	if got := contentRequest(engine, http.MethodPatch, "/api/v1/admin/settings", string(encoded), bphToken); got.Code != http.StatusOK || !strings.Contains(got.Body.String(), "Perubahan Landing Page berhasil disimpan.") {
		t.Fatalf("BPH settings update failed: %d %s", got.Code, got.Body.String())
	}
	if got := getWithToken(engine, "/api/v1/public/settings", ""); got.Code != http.StatusOK || !strings.Contains(got.Body.String(), "Headline Uji Phase 9") {
		t.Fatalf("public settings did not reflect saved value: %d %s", got.Code, got.Body.String())
	}
	unsafe := map[string]any{}
	for k, v := range settings {
		unsafe[k] = v
	}
	unsafe["hero_image_url"] = "javascript:alert(1)"
	unsafeBody, _ := json.Marshal(unsafe)
	if got := contentRequest(engine, http.MethodPatch, "/api/v1/admin/settings", string(unsafeBody), adminToken); got.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unsafe landing URL accepted: %d %s", got.Code, got.Body.String())
	}
	unsafe["hero_image_url"] = "https://example.test/hero.jpg"
	unsafe["missions"] = []string{""}
	unsafeBody, _ = json.Marshal(unsafe)
	if got := contentRequest(engine, http.MethodPatch, "/api/v1/admin/settings", string(unsafeBody), adminToken); got.Code != http.StatusUnprocessableEntity {
		t.Fatalf("empty mission accepted: %d %s", got.Code, got.Body.String())
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM landing_settings`).Scan(new(int)); err != nil {
		t.Fatalf("read settings singleton: %v", err)
	}
	var singleton int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM landing_settings`).Scan(&singleton); err != nil || singleton != 1 {
		t.Fatalf("landing settings is not a singleton: count=%d err=%v", singleton, err)
	}
	var changedFields string
	if err := db.QueryRowContext(ctx, `SELECT CAST(details_json AS CHAR) FROM activity_logs WHERE entity_type='LANDING_PAGE_SETTINGS' ORDER BY id DESC LIMIT 1`).Scan(&changedFields); err != nil || strings.Contains(changedFields, "Subheadline yang diperbarui") || !strings.Contains(changedFields, "hero_headline") {
		t.Fatalf("settings audit details contain payload or omit changed fields: %s err=%v", changedFields, err)
	}

	create := func(name, email, role string, division *int64) int64 {
		payload := map[string]any{"name": name, "email": email, "role": role, "status": "ACTIVE", "password": "temporary phase9 password"}
		payload["division_id"] = division
		body, _ := json.Marshal(payload)
		response := contentRequest(engine, http.MethodPost, "/api/v1/admin/users", string(body), adminToken)
		if response.Code != http.StatusCreated || strings.Contains(response.Body.String(), "password_hash") {
			t.Fatalf("create %s failed or leaked password data: %d %s", role, response.Code, response.Body.String())
		}
		var decoded struct {
			Data struct {
				ID int64 `json:"id"`
			} `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil || decoded.Data.ID < 1 {
			t.Fatalf("parse created user: %v %s", err, response.Body.String())
		}
		return decoded.Data.ID
	}
	secondAdmin := create("Super Admin B", "super-b@example.test", "SUPER_ADMIN", nil)
	thirdAdmin := create("Super Admin C", "super-c@example.test", "SUPER_ADMIN", nil)
	bphID := create("BPH Phase 9", "phase9-bph@example.test", "BPH", nil)
	kepalaID := create("KB Phase 9", "phase9-kb@example.test", "KEPALA_BIDANG", &divisionA)
	invalidDivisionBody := fmt.Sprintf(`{"name":"Invalid KB","email":"invalid-kb@example.test","role":"KEPALA_BIDANG","division_id":%d,"status":"ACTIVE","password":"temporary phase9 password"}`, divisionA+100000)
	if got := contentRequest(engine, http.MethodPost, "/api/v1/admin/users", invalidDivisionBody, adminToken); got.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid KB division accepted: %d %s", got.Code, got.Body.String())
	}
	if got := contentRequest(engine, http.MethodPost, "/api/v1/admin/users", `{"name":"Bad BPH","email":"bad-bph@example.test","role":"BPH","division_id":1,"status":"ACTIVE","password":"temporary phase9 password"}`, adminToken); got.Code != http.StatusUnprocessableEntity {
		t.Fatalf("BPH division assignment accepted: %d %s", got.Code, got.Body.String())
	}
	if got := contentRequest(engine, http.MethodPost, "/api/v1/admin/users", `{"name":"No Division KB","email":"no-division@example.test","role":"KEPALA_BIDANG","status":"ACTIVE","password":"temporary phase9 password"}`, adminToken); got.Code != http.StatusUnprocessableEntity {
		t.Fatalf("divisionless KB accepted: %d %s", got.Code, got.Body.String())
	}
	for _, path := range []string{"/api/v1/admin/users", "/api/v1/admin/users/1", "/api/v1/admin/activity-logs", "/api/v1/admin/activity-logs/1"} {
		for _, token := range []string{bphToken, kepalaToken} {
			if got := getWithToken(engine, path, token); got.Code != http.StatusForbidden {
				t.Fatalf("non-SUPER_ADMIN read %s = %d, want 403", path, got.Code)
			}
		}
	}
	list := getWithToken(engine, "/api/v1/admin/users?search=phase9&role=KEPALA_BIDANG&division_id="+fmt.Sprint(divisionA)+"&status=ACTIVE&page=1&limit=5", adminToken)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "phase9-kb@example.test") || strings.Contains(list.Body.String(), "password_hash") {
		t.Fatalf("user filters/safe DTO failed: %d %s", list.Code, list.Body.String())
	}
	if got := getWithToken(engine, "/api/v1/admin/users/"+fmt.Sprint(kepalaID), adminToken); got.Code != http.StatusOK || strings.Contains(got.Body.String(), "password_hash") {
		t.Fatalf("user detail failed: %d %s", got.Code, got.Body.String())
	}
	roleChange := contentRequest(engine, http.MethodPatch, "/api/v1/admin/users/"+fmt.Sprint(kepalaID), `{"name":"KB Phase 9","email":"phase9-kb@example.test","role":"BPH","division_id":null,"status":"ACTIVE"}`, adminToken)
	if roleChange.Code != http.StatusOK || !strings.Contains(roleChange.Body.String(), `"division_id":null`) {
		t.Fatalf("KB role change did not clear division transactionally: %d %s", roleChange.Code, roleChange.Body.String())
	}
	roleChange = contentRequest(engine, http.MethodPatch, "/api/v1/admin/users/"+fmt.Sprint(kepalaID), fmt.Sprintf(`{"name":"KB Phase 9","email":"phase9-kb@example.test","role":"KEPALA_BIDANG","division_id":%d,"status":"ACTIVE"}`, divisionB), adminToken)
	if roleChange.Code != http.StatusOK || !strings.Contains(roleChange.Body.String(), fmt.Sprintf(`"division_id":%d`, divisionB)) {
		t.Fatalf("BPH to KB role change failed: %d %s", roleChange.Code, roleChange.Body.String())
	}
	if got := contentRequest(engine, http.MethodPatch, "/api/v1/admin/users/"+fmt.Sprint(1), `{"name":"Changed","email":"admin@example.org","role":"BPH","division_id":null,"status":"ACTIVE"}`, adminToken); got.Code != http.StatusConflict {
		t.Fatalf("self-demotion was not rejected: %d %s", got.Code, got.Body.String())
	}
	if got := contentRequest(engine, http.MethodPost, "/api/v1/admin/users/1/deactivate", `{}`, adminToken); got.Code != http.StatusConflict {
		t.Fatalf("self-deactivation was not rejected: %d %s", got.Code, got.Body.String())
	}

	// Concurrent role/status changes serialize through the super-admin row locks.
	var concurrent sync.WaitGroup
	responses := make(chan int, 2)
	concurrent.Add(2)
	go func() {
		defer concurrent.Done()
		responses <- contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/deactivate", secondAdmin), `{}`, adminToken).Code
	}()
	go func() {
		defer concurrent.Done()
		responses <- contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/deactivate", thirdAdmin), `{}`, adminToken).Code
	}()
	concurrent.Wait()
	close(responses)
	for code := range responses {
		if code != http.StatusOK && code != http.StatusConflict {
			t.Fatalf("concurrent admin mutation returned %d", code)
		}
	}
	var activeAdmins int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE role='SUPER_ADMIN' AND status='ACTIVE'`).Scan(&activeAdmins); err != nil || activeAdmins < 1 {
		t.Fatalf("concurrent admin mutation removed all active superadmins: count=%d err=%v", activeAdmins, err)
	}
	for _, id := range []int64{secondAdmin, thirdAdmin} {
		if err := serviceIgnoreError(contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/reactivate", id), `{}`, adminToken)); err != nil {
			t.Fatal(err)
		}
	}
	if got := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/deactivate", secondAdmin), `{}`, adminToken); got.Code != http.StatusOK {
		t.Fatalf("safe superadmin deactivation failed: %d %s", got.Code, got.Body.String())
	}
	if got := contentRequest(engine, http.MethodPost, "/api/v1/admin/users/1/deactivate", `{}`, adminToken); got.Code != http.StatusConflict {
		t.Fatalf("last active superadmin was not protected: %d %s", got.Code, got.Body.String())
	}
	if got := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/reactivate", secondAdmin), `{}`, adminToken); got.Code != http.StatusOK {
		t.Fatalf("superadmin reactivation failed: %d %s", got.Code, got.Body.String())
	}

	// Login then reset: both the old password and refresh family must stop working.
	userLogin := postJSON(engine, "/api/v1/auth/login", `{"email":"phase9-bph@example.test","password":"temporary phase9 password"}`, nil)
	if userLogin.Code != http.StatusOK {
		t.Fatalf("new user login failed: %d %s", userLogin.Code, userLogin.Body.String())
	}
	oldCookie := responseCookie(t, userLogin, "bem_refresh")
	resetBody := `{"password":"a replacement phase9 password","confirm_password":"a replacement phase9 password"}`
	if got := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/reset-password", bphID), resetBody, adminToken); got.Code != http.StatusOK || strings.Contains(got.Body.String(), "replacement phase9 password") {
		t.Fatalf("password reset response unsafe/failed: %d %s", got.Code, got.Body.String())
	}
	if got := postWithCookie(engine, "/api/v1/auth/refresh", oldCookie, nil); got.Code != http.StatusUnauthorized {
		t.Fatalf("old refresh session still worked after password reset: %d %s", got.Code, got.Body.String())
	}
	if got := postJSON(engine, "/api/v1/auth/login", `{"email":"phase9-bph@example.test","password":"temporary phase9 password"}`, nil); got.Code != http.StatusUnauthorized {
		t.Fatalf("old password still worked after reset: %d", got.Code)
	}
	if got := postJSON(engine, "/api/v1/auth/login", `{"email":"phase9-bph@example.test","password":"a replacement phase9 password"}`, nil); got.Code != http.StatusOK {
		t.Fatalf("new password failed after reset: %d %s", got.Code, got.Body.String())
	}
	var passwordHash string
	if err := db.QueryRowContext(ctx, "SELECT password_hash FROM users WHERE id=?", bphID).Scan(&passwordHash); err != nil || strings.Contains(passwordHash, "replacement phase9 password") {
		t.Fatalf("plaintext password persisted: %s err=%v", passwordHash, err)
	}
	if got := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/deactivate", kepalaID), `{}`, adminToken); got.Code != http.StatusOK {
		t.Fatalf("deactivate user failed: %d %s", got.Code, got.Body.String())
	}
	if got := postJSON(engine, "/api/v1/auth/login", `{"email":"phase9-kb@example.test","password":"temporary phase9 password"}`, nil); got.Code != http.StatusUnauthorized {
		t.Fatalf("inactive user authenticated: %d", got.Code)
	}
	if got := contentRequest(engine, http.MethodPost, fmt.Sprintf("/api/v1/admin/users/%d/reactivate", kepalaID), `{}`, adminToken); got.Code != http.StatusOK {
		t.Fatalf("reactivate user failed: %d %s", got.Code, got.Body.String())
	}

	// Redaction is recursive and display-only; stored historical event remains untouched.
	result, err := db.ExecContext(ctx, `INSERT INTO activity_logs(actor_user_id,action,entity_type,entity_id,details_json,ip_address,user_agent) VALUES(?,?,?,?,?,?,?)`, 1, "UPDATE", "REDACTION_TEST", "1", `{"visible":"safe-marker","password":"password-marker","access_token":"access-marker","nested":{"refresh_token":"refresh-marker","authorization":"auth-marker","secret":"secret-marker","storage_path":"path-marker"}}`, "192.0.2.8", "test-agent")
	if err != nil {
		t.Fatalf("insert redaction fixture: %v", err)
	}
	logID, _ := result.LastInsertId()
	activityList := getWithToken(engine, "/api/v1/admin/activity-logs?actor_user_id=1&action=UPDATE&entity_type=REDACTION_TEST&page=1&limit=5", adminToken)
	if activityList.Code != http.StatusOK || !strings.Contains(activityList.Body.String(), "safe-marker") || !strings.Contains(activityList.Body.String(), "[REDACTED]") || strings.Contains(activityList.Body.String(), "password-marker") || strings.Contains(activityList.Body.String(), "192.0.2.8") || strings.Contains(activityList.Body.String(), "test-agent") {
		t.Fatalf("activity list redaction/filter failed: %d %s", activityList.Code, activityList.Body.String())
	}
	activityDetail := getWithToken(engine, fmt.Sprintf("/api/v1/admin/activity-logs/%d", logID), adminToken)
	if activityDetail.Code != http.StatusOK || strings.Contains(activityDetail.Body.String(), "refresh-marker") || strings.Contains(activityDetail.Body.String(), "auth-marker") || strings.Contains(activityDetail.Body.String(), "path-marker") {
		t.Fatalf("activity detail leaked nested secrets: %d %s", activityDetail.Code, activityDetail.Body.String())
	}
	if got := contentRequest(engine, http.MethodPatch, fmt.Sprintf("/api/v1/admin/activity-logs/%d", logID), `{}`, adminToken); got.Code == http.StatusOK {
		t.Fatalf("activity log unexpectedly mutable: %s", got.Body.String())
	}
	if got := contentRequest(engine, http.MethodDelete, fmt.Sprintf("/api/v1/admin/activity-logs/%d", logID), ``, adminToken); got.Code == http.StatusOK {
		t.Fatalf("activity log unexpectedly deletable: %s", got.Body.String())
	}
	var stored string
	if err := db.QueryRowContext(ctx, "SELECT CAST(details_json AS CHAR) FROM activity_logs WHERE id=?", logID).Scan(&stored); err != nil || !strings.Contains(stored, "password-marker") {
		t.Fatalf("redaction modified historical row: %s err=%v", stored, err)
	}
}

func serviceIgnoreError(response *httptest.ResponseRecorder) error {
	if response.Code != http.StatusOK {
		return fmt.Errorf("reactivate response %d: %s", response.Code, response.Body.String())
	}
	return nil
}
