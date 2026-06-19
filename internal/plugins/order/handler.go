// handler.go — order HTTP surface.
package order

import (
	"errors"
	"net/http"

	"github.com/apaichon/harvest-monti/internal/infra/db"
	"github.com/apaichon/harvest-monti/internal/router"
)

// Handler exposes service operations via HTTP.
type Handler struct{ svc *Service }

// NewHandler wires the service.
func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// place — POST /api/v1/orders
func (h *Handler) place(c *router.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	var req PlaceRequest
	if err := c.Bind(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": "invalid body"})
	}
	o, err := h.svc.PlaceFromCart(c.Context(), tenantID, req)
	if errors.Is(err, ErrTotalsMismatch) {
		return c.Status(http.StatusUnprocessableEntity).JSON(map[string]string{"error": err.Error()})
	}
	if errors.Is(err, db.ErrNotFound) {
		return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "cart not found"})
	}
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	return c.Status(http.StatusCreated).JSON(o)
}

// poll — GET /api/v1/public/orders/{order_number}
func (h *Handler) poll(c *router.Ctx) error {
	num := c.Params("order_number")
	o, err := h.svc.GetByNumber(c.Context(), num)
	if errors.Is(err, db.ErrNotFound) {
		return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "not found"})
	}
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(o)
}

// advance — POST /api/v1/operator/orders/{order_id}/advance
func (h *Handler) advance(c *router.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(string)
	var req AdvanceRequest
	if err := c.Bind(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": "invalid body"})
	}
	o, err := h.svc.Advance(c.Context(), tenantID, c.Params("order_id"), req.NextStatus)
	if errors.Is(err, ErrIllegalTransition) {
		return c.Status(http.StatusConflict).JSON(map[string]string{"error": err.Error()})
	}
	if errors.Is(err, db.ErrNotFound) {
		return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "order not found"})
	}
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]string{"error": err.Error()})
	}
	return c.JSON(o)
}
