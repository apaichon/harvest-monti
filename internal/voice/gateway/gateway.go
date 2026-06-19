package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/dispatch"
	"github.com/apaichon/harvest-monti/internal/voice/provider"
	"github.com/apaichon/harvest-monti/internal/voice/switcher"
)

// Endpoint is the canonical WS path per DES-0009 §2.
const Endpoint = "/api/v1/voice/sessions"

// SessionEventSink receives NATS-bound events.
type SessionEventSink interface {
	SessionStarted(ctx context.Context, ev SessionStartedEvent)
	SessionEnded(ctx context.Context, ev SessionEndedEvent)
	LanguageChanged(ctx context.Context, ev LanguageChangedEvent)
}

// SessionStartedEvent — NATS subject monti.voice.session.started.
type SessionStartedEvent struct {
	SessionID string
	TenantID  string
	OutletID  string
	TableCode string
	StartedAt time.Time
}

// SessionEndedEvent — NATS subject monti.voice.session.ended.
type SessionEndedEvent struct {
	SessionID         string
	DurationSeconds   int64
	ProviderUsed      string
	FunctionCallCount int
	Switched          bool
	SwitchReason      string
}

// LanguageChangedEvent — NATS subject monti.session.language_changed.
type LanguageChangedEvent struct {
	SessionID string
	BCP47     string
}

// nopSink is the default sink when none is wired.
type nopSink struct{}

func (nopSink) SessionStarted(context.Context, SessionStartedEvent)   {}
func (nopSink) SessionEnded(context.Context, SessionEndedEvent)       {}
func (nopSink) LanguageChanged(context.Context, LanguageChangedEvent) {}

// Conn abstracts the WebSocket connection. Production wraps gorilla/websocket
// or nhooyr.io/websocket; tests use a simple bytes pipe.
type Conn interface {
	ReadJSON(v any) error
	WriteJSON(v any) error
	ReadBinary() ([]byte, error)
	WriteBinary(b []byte) error
	Close(code int, reason string) error
}

// AuthClaims holds the JWT-derived session identity per DES-0009 §2.
type AuthClaims struct {
	TenantID  string
	OutletID  string
	TableCode string
	Subject   string
}

// Authenticator validates a bearer token from the WS upgrade and returns the
// session-trusted claims.
type Authenticator interface {
	Authenticate(token string) (AuthClaims, error)
}

// Config wires the gateway.
type Config struct {
	Switcher      *switcher.Switcher
	Dispatcher    *dispatch.Dispatcher
	Authenticator Authenticator
	Sink          SessionEventSink
	Logger        *slog.Logger
	Now           func() time.Time
	NewSessionID  func() string // UUIDv7 in production
	IdleTimeout   time.Duration
	SessionCap    time.Duration
}

// Gateway holds the gateway state.
type Gateway struct {
	cfg Config
}

// New builds a Gateway.
func New(cfg Config) *Gateway {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 90 * time.Second
	}
	if cfg.SessionCap == 0 {
		cfg.SessionCap = 30 * time.Minute
	}
	if cfg.Sink == nil {
		cfg.Sink = nopSink{}
	}
	return &Gateway{cfg: cfg}
}

// Handle drives a single WS session end-to-end. Caller is responsible for
// upgrading HTTP to WS and providing the Conn.
func (g *Gateway) Handle(ctx context.Context, conn Conn, bearer string) error {
	claims, err := g.cfg.Authenticator.Authenticate(bearer)
	if err != nil {
		// Per DES-0009 / TEST-0009 TC-2, close with policy-violation code 1008.
		_ = conn.Close(1008, "unauthorized")
		return err
	}

	sessionID := g.cfg.NewSessionID()
	scfg := provider.SessionConfig{
		TenantID:  claims.TenantID,
		OutletID:  claims.OutletID,
		TableCode: claims.TableCode,
		SessionID: sessionID,
	}

	provSess, err := g.cfg.Switcher.StartSession(ctx, scfg)
	if err != nil {
		_ = writeFrame(conn, GatewayFrameError, map[string]any{
			"code":    "provider_start_failed",
			"message": err.Error(),
		})
		_ = conn.Close(1011, "provider start failed")
		return err
	}
	swSess, _ := provSess.(*switcher.Session)

	startedAt := g.cfg.Now()
	g.cfg.Sink.SessionStarted(ctx, SessionStartedEvent{
		SessionID: sessionID,
		TenantID:  claims.TenantID,
		OutletID:  claims.OutletID,
		TableCode: claims.TableCode,
		StartedAt: startedAt,
	})

	if err := writeFrame(conn, GatewayFrameSessionOpen, map[string]any{
		"session_id": sessionID,
		"provider":   currentProvider(swSess),
	}); err != nil {
		_ = provSess.Close(provider.CloseReason{Code: provider.CloseClientGone, Detail: err.Error()})
		return err
	}

	id := dispatch.Identity{
		TenantID:  claims.TenantID,
		OutletID:  claims.OutletID,
		TableCode: claims.TableCode,
		SessionID: sessionID,
	}

	s := &sessionLoop{
		gateway: g,
		ctx:     ctx,
		conn:    conn,
		prov:    provSess,
		swSess:  swSess,
		id:      id,
	}
	defer func() {
		s.closeOnce.Do(func() {
			_ = provSess.Close(provider.CloseReason{Code: provider.CloseNormal})
			g.cfg.Dispatcher.ReleaseSession(sessionID)
			g.cfg.Sink.SessionEnded(ctx, SessionEndedEvent{
				SessionID:         sessionID,
				DurationSeconds:   int64(g.cfg.Now().Sub(startedAt).Seconds()),
				ProviderUsed:      currentProvider(swSess),
				FunctionCallCount: int(s.fnCount.Load()),
				Switched:          currentProvider(swSess) == "grok",
				SwitchReason:      s.switchReason,
			})
		})
	}()

	return s.run()
}

// writeFrame serializes and writes a single gateway frame.
func writeFrame(c Conn, typ string, payload map[string]any) error {
	return c.WriteJSON(GatewayFrame{Type: typ, Payload: payload})
}

// currentProvider safely calls swSess.CurrentProvider — falls back to
// "gemini" when the switcher isn't in play (e.g. test direct-provider runs).
func currentProvider(s *switcher.Session) string {
	if s == nil {
		return "gemini"
	}
	return s.CurrentProvider()
}

// ErrAuthFailed is returned when bearer token validation fails.
var ErrAuthFailed = errors.New("gateway: unauthorized")

// jsonString helps test inspection by serializing payloads to a deterministic
// JSON string.
func jsonString(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
