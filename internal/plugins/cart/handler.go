// handler.go — HTTP surface for cart operations.
package cart

import (
	"errors"
	"net/http"

	"github.com/apaichon/harvest-monti/internal/infra/db"
	"github.com/apaichon/harvest-monti/internal/router"
)

// Handler binds the cart service to HTTP.
type Handler struct{ svc *Service }

// NewHandler wires the Service.
func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// open — POST /api/v1/carts
func (h *Handler) open(c *router.Ctx) error {
	var req OpenCartRequest
	if err := c.Bind(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": "invalid body"})
	}
	if t, _ := c.Locals("tenant_id").(string); t != "" {
		req.TenantID = t
	}
	out, err := h.svc.Open(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(out)
}

// addItem — POST /api/v1/carts/{cart_id}/items
func (h *Handler) addItem(c *router.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}
	cartID := c.Params("cart_id")
	var req AddItemRequest
	if err := c.Bind(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": "invalid body"})
	}
	out, err := h.svc.AddItem(c.Context(), tenantID, cartID, req)
	if errors.Is(err, db.ErrNotFound) {
		return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "cart not found"})
	}
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(out)
}

// removeLine — DELETE /api/v1/carts/{cart_id}/items/{line_id}
func (h *Handler) removeLine(c *router.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}
	out, err := h.svc.RemoveLine(c.Context(), tenantID, c.Params("cart_id"), c.Params("line_id"))
	if errors.Is(err, db.ErrNotFound) {
		return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "not found"})
	}
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(out)
}

// promo — POST /api/v1/carts/{cart_id}/promo
func (h *Handler) promo(c *router.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	if tenantID == "" {
		tenantID = c.Query("tenant_id")
	}
	var req PromoRequest
	if err := c.Bind(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": "invalid body"})
	}
	out, err := h.svc.ApplyPromo(c.Context(), tenantID, c.Params("cart_id"), req.Code)
	if errors.Is(err, ErrInvalidPromo) {
		return c.Status(http.StatusUnprocessableEntity).JSON(map[string]string{"error": "invalid promo"})
	}
	if errors.Is(err, db.ErrNotFound) {
		return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "cart not found"})
	}
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(out)
}
