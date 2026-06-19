// ws.go — order status WebSocket endpoint.
//
// DES-0007 §7 designates WS as the primary transport with 10-second
// polling fallback. The handler is structured so the underlying transport
// is swappable: in production, main() installs a real WS upgrader; in
// tests we install an in-memory upgrader that records every frame.
package order

import (
	"net/http"

	"github.com/apaichon/harvest-monti/internal/router"
)

// stream — GET /api/v1/public/orders/{order_number}/stream
//
// Opens a WS connection bound to the supplied order_number and writes one
// frame per status change. Falls back to a single snapshot frame + close
// when no upgrader is installed (used by curl-based smoke tests).
func (h *Handler) stream(c *router.Ctx) error {
	num := c.Params("order_number")
	o, err := h.svc.GetByNumber(c.Context(), num)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "not found"})
	}

	conn, err := c.UpgradeWS()
	if err != nil {
		// Graceful fallback: serve one snapshot and tell the client to poll
		// per DES-0007 §7. The body shape mirrors what a frame would carry.
		c.Status(http.StatusOK)
		return c.JSON(map[string]any{
			"order":          o,
			"transport":      "snapshot-fallback",
			"poll_seconds":   10,
			"reason":         err.Error(),
		})
	}
	defer conn.Close()

	// Initial snapshot frame.
	if err := conn.WriteJSON(map[string]any{
		"type":  "snapshot",
		"order": o,
	}); err != nil {
		return nil
	}

	ch, unsub := h.svc.Subscribe(num)
	defer unsub()

	ctx := c.Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, ok := <-ch:
			if !ok {
				return nil
			}
			if err := conn.WriteJSON(map[string]any{
				"type":  "status",
				"order": ev,
			}); err != nil {
				return nil
			}
			if ev.Status == StatusCompleted || ev.Status == StatusCancelled {
				return nil
			}
		}
	}
}
