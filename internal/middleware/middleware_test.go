package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apaichon/harvest-monti/internal/router"
)

func TestTenantMiddlewareRejectsMissingHeader(t *testing.T) {
	r := router.New()
	g := r.Group("/api/v1")
	g.Use(Tenant())
	g.Get("/ping", func(c *router.Ctx) error { return c.JSON(map[string]string{"pong": "1"}) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 without tenant header, got %d", rec.Code)
	}
}

func TestTenantMiddlewareInjectsLocal(t *testing.T) {
	r := router.New()
	g := r.Group("/api/v1")
	g.Use(Tenant())
	var seen string
	g.Get("/ping", func(c *router.Ctx) error {
		seen, _ = c.Locals("tenant_id").(string)
		return c.JSON(map[string]string{"pong": "1"})
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.Header.Set(TenantHeader, "tenant-monti-demo")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
	if seen != "tenant-monti-demo" {
		t.Fatalf("tenant local not propagated, got %q", seen)
	}
}

func TestJWTAuthRejectsMissingHeader(t *testing.T) {
	r := router.New()
	g := r.Group("/api/v1")
	g.Use(JWTAuth())
	g.Get("/secure", func(c *router.Ctx) error { return c.JSON(map[string]bool{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/secure", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}
