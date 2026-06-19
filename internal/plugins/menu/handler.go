// handler.go — Fiber-shaped HTTP handlers for the menu plugin.
package menu

import (
	"errors"
	"net/http"
	"strings"

	"github.com/apaichon/harvest-monti/internal/infra/db"
	"github.com/apaichon/harvest-monti/internal/router"
)

// Handler holds the bound service.
type Handler struct {
	svc *Service
}

// NewHandler wires the Service.
func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// list — GET /api/v1/menu
func (h *Handler) list(c *router.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		// Public route fallback — read tenant from query for the kiosk's
		// pre-auth menu fetch (REQ-0012 / DES-0007 §7 recommendation flow).
		tenantID = c.Query("tenant_id")
	}
	f := ItemFilter{}
	if ex := c.Query("exclude_allergens"); ex != "" {
		f.ExcludeAllergens = strings.Split(ex, ",")
	}
	out, err := h.svc.List(c.Context(), tenantID, f)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(out)
}

// getItem — GET /api/v1/menu/items/{item_id}
func (h *Handler) getItem(c *router.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}
	itemID := c.Params("item_id")
	it, err := h.svc.GetItem(c.Context(), tenantID, itemID)
	if errors.Is(err, db.ErrNotFound) {
		// Per TC-10, return 404 not 403, even for cross-tenant probes.
		return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "item not found"})
	}
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(it)
}

// recommendations — GET /api/v1/recommendations
func (h *Handler) recommendations(c *router.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}
	items, err := h.svc.Recommendations(c.Context(), tenantID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(map[string]any{"items": items})
}
