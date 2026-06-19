// routes.go — mounts the cart HTTP endpoints from TASK-0009.
package cart

import "github.com/apaichon/harvest-monti/internal/router"

// Register mounts the four cart endpoints under the caller-provided group
// (typically /api/v1).
func Register(r *router.Router, h *Handler) {
	r.Post("/carts", h.open)
	r.Post("/carts/{cart_id}/items", h.addItem)
	r.Delete("/carts/{cart_id}/items/{line_id}", h.removeLine)
	r.Post("/carts/{cart_id}/promo", h.promo)
}
