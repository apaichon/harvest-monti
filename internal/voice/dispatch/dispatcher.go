// Package dispatch implements the function-call dispatcher described in
// DES-0009 §4 and ADR-0005 D4. Responsibilities:
//
//   - Resolve a FunctionCall name to one of the six canonical handlers.
//   - Inject trusted session identity (tenant_id, outlet_id, table_code,
//     session_id) and REJECT any model-supplied identity arguments.
//   - Serialize calls within a single session (one in-flight at a time).
//   - Enforce 3 s soft / 10 s hard per-call timeouts.
//   - Return a normalized FunctionResult envelope.
package dispatch

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/dispatch/ports"
	"github.com/apaichon/harvest-monti/internal/voice/provider"
)

// Canonical function names (ADR-0005 D4 / DES-0009 §4).
const (
	FnMenuSearch     = "menu_search"
	FnCartAdd        = "cart_add"
	FnCartRemove     = "cart_remove"
	FnOrderSubmit    = "order_submit"
	FnLanguageSwitch = "language_switch"
	FnCallStaff      = "call_staff"
)

// reservedIdentityKeys are the argument names the model is NEVER allowed to
// supply. The dispatcher strips them with a WARN log line and overwrites with
// the trusted session-derived values.
var reservedIdentityKeys = []string{
	"tenant_id", "tenantId",
	"outlet_id", "outletId",
	"table_code", "tableCode",
	"session_id", "sessionId",
}

// Timeouts per DES-0009 §4 final paragraph.
const (
	DefaultSoftTimeout = 3 * time.Second
	DefaultHardTimeout = 10 * time.Second
)

// Identity is the trusted session-derived context the dispatcher injects.
type Identity struct {
	TenantID  string
	OutletID  string
	TableCode string
	SessionID string
}

// Config wires the dispatcher.
type Config struct {
	Menu     ports.MenuPort
	Cart     ports.CartPort
	Order    ports.OrderPort
	Staff    ports.StaffPort
	Language ports.LanguagePort

	SoftTimeout time.Duration
	HardTimeout time.Duration

	Logger *slog.Logger
	Now    func() time.Time
}

// Dispatcher routes FunctionCall → handler with strict identity enforcement.
type Dispatcher struct {
	cfg     Config
	mu      sync.Mutex
	queues  map[string]*sessionQueue // keyed by session_id
}

// sessionQueue serializes calls within one session.
type sessionQueue struct {
	mu sync.Mutex
}

// New builds a Dispatcher.
func New(cfg Config) *Dispatcher {
	if cfg.SoftTimeout == 0 {
		cfg.SoftTimeout = DefaultSoftTimeout
	}
	if cfg.HardTimeout == 0 {
		cfg.HardTimeout = DefaultHardTimeout
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Dispatcher{
		cfg:    cfg,
		queues: make(map[string]*sessionQueue),
	}
}

// Dispatch resolves and invokes a function call. It blocks until the call
// completes or hits the hard timeout. Per-session FIFO is guaranteed.
func (d *Dispatcher) Dispatch(ctx context.Context, id Identity, call provider.FunctionCall) provider.FunctionResult {
	if id.SessionID == "" {
		return fail(call.CallID, "missing_session_id", "dispatcher requires a session identity")
	}

	// Strip model-supplied identity keys before doing anything else.
	args := stripReservedIdentity(call.Arguments, call.Name, id, d.cfg.Logger)

	// Acquire per-session queue lock.
	q := d.queueFor(id.SessionID)
	q.mu.Lock()
	defer q.mu.Unlock()

	// Hard timeout context.
	hardCtx, cancel := context.WithTimeout(ctx, d.cfg.HardTimeout)
	defer cancel()

	// Run in goroutine so we can fire the soft-timeout signal independently.
	resultCh := make(chan provider.FunctionResult, 1)
	go func() {
		resultCh <- d.invoke(hardCtx, id, call.CallID, call.Name, args)
	}()

	softTimer := time.NewTimer(d.cfg.SoftTimeout)
	defer softTimer.Stop()
	softFired := false

	for {
		select {
		case res := <-resultCh:
			return res
		case <-softTimer.C:
			if !softFired {
				softFired = true
				d.cfg.Logger.Warn("dispatcher.soft_timeout",
					"call_id", call.CallID,
					"name", call.Name,
					"session_id", id.SessionID,
					"soft_timeout_ms", d.cfg.SoftTimeout.Milliseconds(),
				)
			}
		case <-hardCtx.Done():
			d.cfg.Logger.Error("dispatcher.hard_timeout",
				"call_id", call.CallID,
				"name", call.Name,
				"session_id", id.SessionID,
			)
			return fail(call.CallID, "timeout",
				fmt.Sprintf("hard timeout after %s", d.cfg.HardTimeout))
		}
	}
}

func (d *Dispatcher) queueFor(sessionID string) *sessionQueue {
	d.mu.Lock()
	defer d.mu.Unlock()
	q, ok := d.queues[sessionID]
	if !ok {
		q = &sessionQueue{}
		d.queues[sessionID] = q
	}
	return q
}

// ReleaseSession drops the per-session queue. Call on session close.
func (d *Dispatcher) ReleaseSession(sessionID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.queues, sessionID)
}

func (d *Dispatcher) invoke(ctx context.Context, id Identity, callID, name string, args map[string]any) provider.FunctionResult {
	switch name {
	case FnMenuSearch:
		return d.menuSearch(ctx, id, callID, args)
	case FnCartAdd:
		return d.cartAdd(ctx, id, callID, args)
	case FnCartRemove:
		return d.cartRemove(ctx, id, callID, args)
	case FnOrderSubmit:
		return d.orderSubmit(ctx, id, callID, args)
	case FnLanguageSwitch:
		return d.languageSwitch(ctx, id, callID, args)
	case FnCallStaff:
		return d.callStaff(ctx, id, callID, args)
	default:
		return fail(callID, "unknown_function", "no handler registered for "+name)
	}
}

func (d *Dispatcher) menuSearch(ctx context.Context, id Identity, callID string, args map[string]any) provider.FunctionResult {
	if d.cfg.Menu == nil {
		return fail(callID, "not_implemented", "menu port not wired")
	}
	q := ports.MenuSearchQuery{
		TenantID:  id.TenantID,
		OutletID:  id.OutletID,
		TableCode: id.TableCode,
		SessionID: id.SessionID,
		Query:     stringArg(args, "query"),
		Allergens: stringSliceArg(args, "allergens"),
		Category:  stringArg(args, "category"),
	}
	items, err := d.cfg.Menu.Search(ctx, q)
	if err != nil {
		return fail(callID, "menu_search_failed", err.Error())
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{
			"id":          it.ID,
			"name":        it.Name,
			"description": it.Description,
			"price_cents": it.PriceCents,
			"currency":    it.Currency,
			"allergens":   it.Allergens,
		})
	}
	return ok(callID, map[string]any{"items": out, "count": len(items)})
}

func (d *Dispatcher) cartAdd(ctx context.Context, id Identity, callID string, args map[string]any) provider.FunctionResult {
	if d.cfg.Cart == nil {
		return fail(callID, "not_implemented", "cart port not wired")
	}
	req := ports.CartAddRequest{
		TenantID:  id.TenantID,
		OutletID:  id.OutletID,
		TableCode: id.TableCode,
		SessionID: id.SessionID,
		ItemID:    stringArg(args, "item_id"),
		Qty:       intArg(args, "qty", 1),
		Modifiers: stringMapArg(args, "modifiers"),
	}
	if req.ItemID == "" {
		return fail(callID, "bad_arguments", "item_id is required")
	}
	state, err := d.cfg.Cart.Add(ctx, req)
	if err != nil {
		return fail(callID, "cart_add_failed", err.Error())
	}
	return ok(callID, cartStateMap(state))
}

func (d *Dispatcher) cartRemove(ctx context.Context, id Identity, callID string, args map[string]any) provider.FunctionResult {
	if d.cfg.Cart == nil {
		return fail(callID, "not_implemented", "cart port not wired")
	}
	req := ports.CartRemoveRequest{
		TenantID:  id.TenantID,
		OutletID:  id.OutletID,
		TableCode: id.TableCode,
		SessionID: id.SessionID,
		LineID:    stringArg(args, "line_id"),
	}
	if req.LineID == "" {
		return fail(callID, "bad_arguments", "line_id is required")
	}
	state, err := d.cfg.Cart.Remove(ctx, req)
	if err != nil {
		return fail(callID, "cart_remove_failed", err.Error())
	}
	return ok(callID, cartStateMap(state))
}

func (d *Dispatcher) orderSubmit(ctx context.Context, id Identity, callID string, args map[string]any) provider.FunctionResult {
	if d.cfg.Order == nil {
		return fail(callID, "not_implemented", "order port not wired")
	}
	req := ports.OrderSubmitRequest{
		TenantID:      id.TenantID,
		OutletID:      id.OutletID,
		TableCode:     id.TableCode,
		SessionID:     id.SessionID,
		PaymentMethod: stringArg(args, "payment_method"),
	}
	ticket, err := d.cfg.Order.Submit(ctx, req)
	if err != nil {
		return fail(callID, "order_submit_failed", err.Error())
	}
	return ok(callID, map[string]any{
		"order_id":       ticket.OrderID,
		"session_id":     ticket.SessionID,
		"payment_status": ticket.PaymentStatus,
		"total_cents":    ticket.TotalCents,
		"currency":       ticket.Currency,
	})
}

func (d *Dispatcher) languageSwitch(ctx context.Context, id Identity, callID string, args map[string]any) provider.FunctionResult {
	if d.cfg.Language == nil {
		return fail(callID, "not_implemented", "language port not wired")
	}
	bcp := stringArg(args, "bcp47")
	if bcp == "" {
		return fail(callID, "bad_arguments", "bcp47 is required")
	}
	state, err := d.cfg.Language.Switch(ctx, ports.LanguageSwitchRequest{
		TenantID:  id.TenantID,
		OutletID:  id.OutletID,
		TableCode: id.TableCode,
		SessionID: id.SessionID,
		BCP47:     bcp,
	})
	if err != nil {
		return fail(callID, "language_switch_failed", err.Error())
	}
	return ok(callID, map[string]any{
		"session_id": state.SessionID,
		"bcp47":      state.BCP47,
	})
}

func (d *Dispatcher) callStaff(ctx context.Context, id Identity, callID string, args map[string]any) provider.FunctionResult {
	if d.cfg.Staff == nil {
		return fail(callID, "not_implemented", "staff port not wired")
	}
	ping, err := d.cfg.Staff.Call(ctx, ports.StaffCallRequest{
		TenantID:  id.TenantID,
		OutletID:  id.OutletID,
		TableCode: id.TableCode,
		SessionID: id.SessionID,
		Reason:    stringArg(args, "reason"),
	})
	if err != nil {
		return fail(callID, "call_staff_failed", err.Error())
	}
	return ok(callID, map[string]any{
		"ping_id":     ping.PingID,
		"notified_at": ping.NotifiedAt,
	})
}

// stripReservedIdentity removes the reserved identity keys from the model's
// argument map and logs a WARN if any were present. The trusted identity is
// not copied INTO the map (the dispatcher passes it via the typed Identity
// struct); the map only carries call-specific business arguments.
func stripReservedIdentity(args map[string]any, callName string, id Identity, log *slog.Logger) map[string]any {
	if args == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(args))
	for k, v := range args {
		if isReserved(k) {
			log.Warn("dispatcher.tenant_override",
				"function", callName,
				"stripped_key", k,
				"model_supplied", fmt.Sprintf("%v", v),
				"enforced_tenant", id.TenantID,
				"enforced_outlet", id.OutletID,
				"enforced_session", id.SessionID,
			)
			continue
		}
		out[k] = v
	}
	return out
}

func isReserved(k string) bool {
	for _, r := range reservedIdentityKeys {
		if k == r {
			return true
		}
	}
	return false
}

// ---------- arg helpers ----------

func stringArg(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func intArg(args map[string]any, key string, fallback int) int {
	if args == nil {
		return fallback
	}
	switch v := args[key].(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	}
	return fallback
}

func stringSliceArg(args map[string]any, key string) []string {
	if args == nil {
		return nil
	}
	raw, ok := args[key].([]any)
	if ok {
		out := make([]string, 0, len(raw))
		for _, r := range raw {
			if s, ok := r.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	if direct, ok := args[key].([]string); ok {
		return direct
	}
	return nil
}

func stringMapArg(args map[string]any, key string) map[string]string {
	if args == nil {
		return nil
	}
	if raw, ok := args[key].(map[string]any); ok {
		out := make(map[string]string, len(raw))
		for k, v := range raw {
			if s, ok := v.(string); ok {
				out[k] = s
			}
		}
		return out
	}
	if direct, ok := args[key].(map[string]string); ok {
		return direct
	}
	return nil
}

func cartStateMap(s ports.CartState) map[string]any {
	lines := make([]map[string]any, 0, len(s.Lines))
	for _, l := range s.Lines {
		lines = append(lines, map[string]any{
			"line_id":    l.LineID,
			"item_id":    l.ItemID,
			"name":       l.Name,
			"qty":        l.Qty,
			"modifiers":  l.Modifiers,
			"unit_cents": l.UnitCents,
		})
	}
	return map[string]any{
		"session_id":     s.SessionID,
		"lines":          lines,
		"subtotal_cents": s.Subtotal,
		"currency":       s.Currency,
	}
}

func ok(callID string, payload map[string]any) provider.FunctionResult {
	return provider.FunctionResult{
		CallID:  callID,
		Success: true,
		Payload: payload,
	}
}

func fail(callID, code, msg string) provider.FunctionResult {
	return provider.FunctionResult{
		CallID:  callID,
		Success: false,
		Payload: map[string]any{"error_code": code},
		Error:   msg,
	}
}

// ErrUnknownFunction is exported for callers that want to detect the
// unknown-name path without string-matching.
var ErrUnknownFunction = errors.New("dispatch: unknown function name")
