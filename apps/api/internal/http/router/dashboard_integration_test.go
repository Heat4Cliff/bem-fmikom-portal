package router

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/domain"
	"github.com/bem-fmikom/bem-fmikom-portal/apps/api/internal/service"
)

func runDashboardIntegration(t *testing.T, ctx context.Context, db *sql.DB, engine http.Handler, adminToken, bphToken, kepalaAToken, kepalaBToken string, divisionA, divisionB int64) {
	t.Helper()
	if response := getWithToken(engine, "/api/v1/admin/dashboard", ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous dashboard status=%d want=401", response.Code)
	}
	type metrics struct {
		Articles struct {
			Total     int64 `json:"total"`
			Published int64 `json:"published"`
			Draft     int64 `json:"draft"`
		} `json:"articles"`
		Registrations struct {
			ActiveForms           int64 `json:"active_forms"`
			ActiveFormSubmissions int64 `json:"active_form_submissions"`
		} `json:"registrations"`
		Aspirations struct {
			Total    int64 `json:"total"`
			Pending  int64 `json:"pending"`
			Read     int64 `json:"read"`
			Resolved int64 `json:"resolved"`
		} `json:"aspirations"`
		Documents struct {
			Total int64 `json:"total"`
		} `json:"documents"`
		Announcements struct {
			Total int64 `json:"total"`
		} `json:"announcements"`
		Scope struct {
			Type     string `json:"type"`
			Division *struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"division"`
		} `json:"scope"`
		ActivePeriod *struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"active_period"`
	}
	read := func(token string) metrics {
		t.Helper()
		response := getWithToken(engine, "/api/v1/admin/dashboard", token)
		if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "private, no-store" {
			t.Fatalf("dashboard status/cache=%d/%q body=%s", response.Code, response.Header().Get("Cache-Control"), response.Body.String())
		}
		if strings.Contains(strings.ToLower(response.Body.String()), "pageview") || strings.Contains(response.Body.String(), "visitor") {
			t.Fatalf("dashboard unexpectedly exposes visitor tracking: %s", response.Body.String())
		}
		var envelope struct {
			Data metrics `json:"data"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode dashboard: %v", err)
		}
		return envelope.Data
	}
	if got := read(adminToken); got.Scope.Type != "organization" {
		t.Fatalf("SUPER_ADMIN dashboard scope=%q", got.Scope.Type)
	}
	if got := read(bphToken); got.Scope.Type != "organization" {
		t.Fatalf("BPH dashboard scope=%q", got.Scope.Type)
	}
	a := read(kepalaAToken)
	b := read(kepalaBToken)
	if a.Scope.Type != "division" || a.Scope.Division == nil || a.Scope.Division.ID != divisionA {
		t.Fatalf("KB A dashboard scope=%+v", a.Scope)
	}
	if b.Scope.Type != "division" || b.Scope.Division == nil || b.Scope.Division.ID != divisionB {
		t.Fatalf("KB B dashboard scope=%+v", b.Scope)
	}
	if response := getWithToken(engine, "/api/v1/admin/dashboard", "invalid-token"); response.Code != http.StatusUnauthorized {
		t.Fatalf("invalid-token dashboard status=%d", response.Code)
	}

	assertCounts := func(role string, got metrics, division *int64) {
		t.Helper()
		filter := ""
		args := []any{}
		if division != nil {
			filter = " AND division_id=?"
			args = append(args, *division)
		}
		var total, published, drafts, aspirationTotal, pending, readCount, resolved, docs int64
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(status='PUBLISHED'),0), COALESCE(SUM(status='DRAFT'),0) FROM articles WHERE deleted_at IS NULL`+filter, args...).Scan(&total, &published, &drafts); err != nil {
			t.Fatal(err)
		}
		if got.Articles.Total != total || got.Articles.Published != published || got.Articles.Draft != drafts {
			t.Fatalf("%s articles got=%+v db=%d/%d/%d", role, got.Articles, total, published, drafts)
		}
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status='PENDING'),0),COALESCE(SUM(status='READ'),0),COALESCE(SUM(status='RESOLVED'),0) FROM aspirations WHERE 1=1`+filter, args...).Scan(&aspirationTotal, &pending, &readCount, &resolved); err != nil {
			t.Fatal(err)
		}
		if got.Aspirations.Total != aspirationTotal || got.Aspirations.Pending != pending || got.Aspirations.Read != readCount || got.Aspirations.Resolved != resolved {
			t.Fatalf("%s aspirations got=%+v db=%d/%d/%d/%d", role, got.Aspirations, aspirationTotal, pending, readCount, resolved)
		}
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents WHERE delete_pending=FALSE`+filter, args...).Scan(&docs); err != nil {
			t.Fatal(err)
		}
		if got.Documents.Total != docs {
			t.Fatalf("%s documents=%d db=%d", role, got.Documents.Total, docs)
		}
		fFilter := ""
		fArgs := []any{}
		if division != nil {
			fFilter = " AND f.division_id=?"
			fArgs = append(fArgs, *division)
		}
		var forms, submissions int64
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_forms f WHERE f.deleted_at IS NULL AND f.status='OPEN' AND UTC_TIMESTAMP(6)>=f.open_at AND UTC_TIMESTAMP(6)<=f.close_at`+fFilter, fArgs...).Scan(&forms); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM registration_submissions rs JOIN registration_forms f ON f.id=rs.registration_form_id WHERE f.deleted_at IS NULL AND f.status='OPEN' AND UTC_TIMESTAMP(6)>=f.open_at AND UTC_TIMESTAMP(6)<=f.close_at`+fFilter, fArgs...).Scan(&submissions); err != nil {
			t.Fatal(err)
		}
		if got.Registrations.ActiveForms != forms || got.Registrations.ActiveFormSubmissions != submissions {
			t.Fatalf("%s registrations got=%+v db=%d/%d", role, got.Registrations, forms, submissions)
		}
	}
	assertCounts("SUPER_ADMIN", read(adminToken), nil)
	assertCounts("BPH", read(bphToken), nil)
	aID, bID := divisionA, divisionB
	assertCounts("KEPALA_BIDANG_A", a, &aID)
	assertCounts("KEPALA_BIDANG_B", b, &bID)

	noDivision := service.NewDashboardService(db)
	if _, err := noDivision.Summary(ctx, domain.User{ID: 999999, Role: domain.RoleKepala, Status: domain.StatusActive}); err != service.ErrDashboardForbidden {
		t.Fatalf("KEPALA_BIDANG without division error=%v want=forbidden", err)
	}
}
