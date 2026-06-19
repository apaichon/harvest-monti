// public.go — wires the /api/v1/public group with no auth (REQ-0012:
// public routes under /api/v1/public skip JWT).
//
// The QR redeem endpoint lives here so the mobile app can decode a deep
// link before any user is authenticated; the order status WS lives here
// so kiosks at the table can read their own order without a login.
package router

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/apaichon/harvest-monti/internal/infra/qr"
)

// PublicDeps holds the dependencies the public group needs.
type PublicDeps struct {
	QRKey         []byte                // HS256 signing key (from harvest-core)
	JtiBlacklist  JtiStore              // 10-min replay guard
	Now           func() time.Time
}

// QRRedeemHandler installs POST /api/v1/public/qr/redeem on the supplied
// router group. The endpoint accepts {token} and returns the decoded
// {session_id, tenant_id, outlet_id, table_code} envelope.
func QRRedeemHandler(deps PublicDeps) HandlerFunc {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	return func(c *Ctx) error {
		var body struct {
			Token string `json:"token"`
		}
		if err := c.Bind(&body); err != nil || body.Token == "" {
			return c.Status(http.StatusBadRequest).JSON(map[string]string{"error": "token required"})
		}
		claims, err := qr.Verify(body.Token, deps.QRKey, deps.Now())
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(map[string]string{"error": err.Error()})
		}
		if deps.JtiBlacklist != nil && claims.Jti != "" {
			if deps.JtiBlacklist.Seen(claims.Jti) {
				return c.Status(http.StatusConflict).JSON(map[string]string{"error": "token replay"})
			}
			deps.JtiBlacklist.Mark(claims.Jti, 10*time.Minute)
		}
		sessionID := newSessionID()
		return c.JSON(map[string]string{
			"session_id":  sessionID,
			"tenant_id":   claims.TenantID,
			"outlet_id":   claims.OutletID,
			"table_code":  claims.TableCode,
		})
	}
}

// newSessionID returns an opaque kiosk/mobile session id.
func newSessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "sess-fallback"
	}
	return "sess-" + hex.EncodeToString(b[:])
}

// JtiStore is the JWT-id replay blacklist (10-min Redis key in prod;
// in-memory map in tests).
type JtiStore interface {
	Seen(jti string) bool
	Mark(jti string, ttl time.Duration)
}

// MemoryJtiStore is the test-friendly implementation.
type MemoryJtiStore struct {
	mu   sync.Mutex
	seen map[string]time.Time
	now  func() time.Time
}

// NewMemoryJtiStore returns an empty store.
func NewMemoryJtiStore(now func() time.Time) *MemoryJtiStore {
	if now == nil {
		now = time.Now
	}
	return &MemoryJtiStore{seen: map[string]time.Time{}, now: now}
}

// Seen returns true if the jti was marked and the entry has not expired.
func (m *MemoryJtiStore) Seen(jti string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	exp, ok := m.seen[jti]
	if !ok {
		return false
	}
	if m.now().After(exp) {
		delete(m.seen, jti)
		return false
	}
	return true
}

// Mark records the jti with TTL.
func (m *MemoryJtiStore) Mark(jti string, ttl time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seen[jti] = m.now().Add(ttl)
}

// ErrPublicDepsMissing is a convenience for boot wiring.
var ErrPublicDepsMissing = errors.New("router: PublicDeps incomplete")
