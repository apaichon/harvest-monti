// Package middleware holds the auth + tenant chain used by every non-public
// MONTI route. Public routes under /api/v1/public skip auth per REQ-0012.
//
// At runtime the JWT layer is the harvest-core JWKS client. For the
// initial scaffold we ship a small stub that accepts a tenant-id header
// for local testing; production wiring (replacing the stub) is a one-file
// swap in cmd/monti-api/main.go.
package middleware

import (
	"net/http"

	"github.com/apaichon/harvest-monti/internal/router"
)

// TenantHeader is the HTTP header name carrying the tenant id. Same as
// harvest-core convention for cross-repo consistency.
const TenantHeader = "X-Tenant-ID"

// AuthorizationHeader carries the bearer JWT.
const AuthorizationHeader = "Authorization"

// Tenant extracts X-Tenant-ID and stores it in Ctx.Locals("tenant_id").
// Returns 400 if missing. Mount this on every authenticated route group.
func Tenant() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Ctx) error {
			t := c.Request().Header.Get(TenantHeader)
			if t == "" {
				return c.Status(http.StatusBadRequest).JSON(map[string]string{
					"error": "X-Tenant-ID required",
				})
			}
			c.Locals("tenant_id", t)
			return next(c)
		}
	}
}

// JWTAuth is a stub that records the bearer header and treats any non-empty
// Authorization as authenticated. The production wiring replaces this with
// the harvest-core JWKS verifier — same Middleware signature, no plugin
// code changes.
func JWTAuth() router.Middleware {
	return func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Ctx) error {
			auth := c.Request().Header.Get(AuthorizationHeader)
			if auth == "" {
				return c.Status(http.StatusUnauthorized).JSON(map[string]string{
					"error": "missing bearer token",
				})
			}
			c.Locals("user_id", "stub-user")
			return next(c)
		}
	}
}
