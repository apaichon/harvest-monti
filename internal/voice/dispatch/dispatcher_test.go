package dispatch_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/apaichon/harvest-monti/internal/voice/dispatch"
	"github.com/apaichon/harvest-monti/internal/voice/dispatch/ports"
	"github.com/apaichon/harvest-monti/internal/voice/provider"
)

// fakeMenu records the last query and returns canned items.
type fakeMenu struct {
	lastQuery ports.MenuSearchQuery
	items     []ports.MenuItemLite
	err       error
}

func (f *fakeMenu) Search(_ context.Context, q ports.MenuSearchQuery) ([]ports.MenuItemLite, error) {
	f.lastQuery = q
	return f.items, f.err
}

type fakeCart struct {
	mu      sync.Mutex
	adds    []ports.CartAddRequest
	state   ports.CartState
	addOnce func()
	addErr  error
}

func (f *fakeCart) Add(_ context.Context, r ports.CartAddRequest) (ports.CartState, error) {
	f.mu.Lock()
	f.adds = append(f.adds, r)
	once := f.addOnce
	f.mu.Unlock()
	if once != nil {
		once()
	}
	if f.addErr != nil {
		return ports.CartState{}, f.addErr
	}
	return f.state, nil
}

func (f *fakeCart) Remove(_ context.Context, r ports.CartRemoveRequest) (ports.CartState, error) {
	return f.state, nil
}

type fakeOrder struct{ ticket ports.OrderTicket }

func (f *fakeOrder) Submit(_ context.Context, r ports.OrderSubmitRequest) (ports.OrderTicket, error) {
	return f.ticket, nil
}

type fakeStaff struct{ ping ports.StaffPing }

func (f *fakeStaff) Call(_ context.Context, r ports.StaffCallRequest) (ports.StaffPing, error) {
	return f.ping, nil
}

type fakeLang struct{ state ports.LanguageState }

func (f *fakeLang) Switch(_ context.Context, r ports.LanguageSwitchRequest) (ports.LanguageState, error) {
	f.state.SessionID = r.SessionID
	f.state.BCP47 = r.BCP47
	return f.state, nil
}

func newDispatcher(t *testing.T, menu *fakeMenu, cart *fakeCart) *dispatch.Dispatcher {
	t.Helper()
	return dispatch.New(dispatch.Config{
		Menu:        menu,
		Cart:        cart,
		Order:       &fakeOrder{ticket: ports.OrderTicket{OrderID: "o-1", PaymentStatus: "stubbed"}},
		Staff:       &fakeStaff{ping: ports.StaffPing{PingID: "p-1", NotifiedAt: 1}},
		Language:    &fakeLang{},
		SoftTimeout: 100 * time.Millisecond,
		HardTimeout: 500 * time.Millisecond,
	})
}

// TC: dispatcher strips model-supplied tenant_id and overwrites with the
// session-trusted identity (DES-0009 §4 step 2, ADR-0005 D4, TEST-0009 TC-9).
func TestDispatcher_RejectsModelSuppliedTenantID(t *testing.T) {
	menu := &fakeMenu{items: []ports.MenuItemLite{{ID: "i1", Name: "Pad Thai"}}}
	cart := &fakeCart{state: ports.CartState{SessionID: "sess-A"}}
	d := newDispatcher(t, menu, cart)

	id := dispatch.Identity{
		TenantID:  "tenant-A",
		OutletID:  "outlet-1",
		TableCode: "A12",
		SessionID: "sess-A",
	}

	call := provider.FunctionCall{
		Name:   dispatch.FnCartAdd,
		CallID: "c1",
		Arguments: map[string]any{
			"item_id":    "i1",
			"qty":        2,
			"tenant_id":  "tenant-B-forged", // model attempt
			"session_id": "sess-EVIL",       // model attempt
			"outletId":   "outlet-EVIL",      // camelCase variant must also be stripped
		},
	}

	res := d.Dispatch(context.Background(), id, call)
	if !res.Success {
		t.Fatalf("expected success, got error %q", res.Error)
	}
	if len(cart.adds) != 1 {
		t.Fatalf("expected exactly one cart.Add call, got %d", len(cart.adds))
	}
	got := cart.adds[0]
	if got.TenantID != "tenant-A" {
		t.Errorf("tenant_id not enforced: want %q got %q", "tenant-A", got.TenantID)
	}
	if got.OutletID != "outlet-1" {
		t.Errorf("outlet_id not enforced: want %q got %q", "outlet-1", got.OutletID)
	}
	if got.SessionID != "sess-A" {
		t.Errorf("session_id not enforced: want %q got %q", "sess-A", got.SessionID)
	}
}

// TC: dispatcher runs calls strictly sequentially for one session (DES-0009 §4
// step 6, TEST-0009 TC-10).
func TestDispatcher_PerSessionSerialization(t *testing.T) {
	menu := &fakeMenu{}
	var inflight atomic.Int32
	var peak atomic.Int32
	cart := &fakeCart{
		state: ports.CartState{SessionID: "s"},
		addOnce: func() {
			n := inflight.Add(1)
			if n > peak.Load() {
				peak.Store(n)
			}
			time.Sleep(20 * time.Millisecond)
			inflight.Add(-1)
		},
	}
	d := newDispatcher(t, menu, cart)
	id := dispatch.Identity{TenantID: "t", OutletID: "o", TableCode: "T1", SessionID: "s"}

	const N = 5
	results := make(chan provider.FunctionResult, N)
	for i := 0; i < N; i++ {
		i := i
		go func() {
			results <- d.Dispatch(context.Background(), id, provider.FunctionCall{
				Name:      dispatch.FnCartAdd,
				CallID:    "c",
				Arguments: map[string]any{"item_id": "i", "qty": i + 1},
			})
		}()
	}
	for i := 0; i < N; i++ {
		r := <-results
		if !r.Success {
			t.Fatalf("dispatch failure: %s", r.Error)
		}
	}
	if peak.Load() != 1 {
		t.Errorf("expected peak concurrency 1 (serialized), got %d", peak.Load())
	}
	if len(cart.adds) != N {
		t.Errorf("expected %d add calls, got %d", N, len(cart.adds))
	}
}

// TC: 3s soft / 10s hard timeout — verify hard timeout fires.
func TestDispatcher_HardTimeoutFires(t *testing.T) {
	menu := &fakeMenu{}
	cart := &fakeCart{
		state:   ports.CartState{},
		addOnce: func() { time.Sleep(300 * time.Millisecond) }, // > hard timeout (200ms)
	}
	d := dispatch.New(dispatch.Config{
		Menu:        menu,
		Cart:        cart,
		SoftTimeout: 50 * time.Millisecond,
		HardTimeout: 200 * time.Millisecond,
	})

	res := d.Dispatch(context.Background(), dispatch.Identity{SessionID: "s"}, provider.FunctionCall{
		Name:      dispatch.FnCartAdd,
		CallID:    "c1",
		Arguments: map[string]any{"item_id": "i1"},
	})
	if res.Success {
		t.Fatalf("expected timeout failure")
	}
	if res.Error == "" {
		t.Fatal("expected non-empty error message on timeout")
	}
	if code, _ := res.Payload["error_code"].(string); code != "timeout" {
		t.Errorf("expected error_code=timeout, got %v", res.Payload["error_code"])
	}
}

// TC: all six canonical names route to their port (ADR-0005 D4).
func TestDispatcher_AllSixCanonicalFunctions(t *testing.T) {
	menu := &fakeMenu{}
	cart := &fakeCart{state: ports.CartState{SessionID: "s"}}
	d := newDispatcher(t, menu, cart)
	id := dispatch.Identity{TenantID: "t", OutletID: "o", TableCode: "T", SessionID: "s"}

	cases := []struct {
		name string
		args map[string]any
	}{
		{dispatch.FnMenuSearch, map[string]any{"query": "noodles"}},
		{dispatch.FnCartAdd, map[string]any{"item_id": "i", "qty": 1}},
		{dispatch.FnCartRemove, map[string]any{"line_id": "L"}},
		{dispatch.FnOrderSubmit, map[string]any{"payment_method": "demo"}},
		{dispatch.FnLanguageSwitch, map[string]any{"bcp47": "en-US"}},
		{dispatch.FnCallStaff, map[string]any{"reason": "help"}},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			res := d.Dispatch(context.Background(), id, provider.FunctionCall{
				Name:      c.name,
				CallID:    "c-" + c.name,
				Arguments: c.args,
			})
			if !res.Success {
				t.Fatalf("function %q failed: %s", c.name, res.Error)
			}
			if res.CallID != "c-"+c.name {
				t.Errorf("call_id mismatch: got %q", res.CallID)
			}
		})
	}
}

// TC: unknown function name returns a clean failure (does not panic).
func TestDispatcher_UnknownFunctionReturnsFailure(t *testing.T) {
	d := newDispatcher(t, &fakeMenu{}, &fakeCart{})
	res := d.Dispatch(context.Background(), dispatch.Identity{SessionID: "s"}, provider.FunctionCall{
		Name: "make_coffee", CallID: "x",
	})
	if res.Success {
		t.Fatal("expected failure for unknown function")
	}
	if code, _ := res.Payload["error_code"].(string); code != "unknown_function" {
		t.Errorf("expected error_code=unknown_function, got %v", res.Payload["error_code"])
	}
}

// TC: missing session_id is a hard error (ADR-0005 D4: missing tenant context
// is a hard error, not a fallback). We extend the same rule to session id.
func TestDispatcher_RejectsMissingSession(t *testing.T) {
	d := newDispatcher(t, &fakeMenu{}, &fakeCart{})
	res := d.Dispatch(context.Background(), dispatch.Identity{TenantID: "t"}, provider.FunctionCall{
		Name: dispatch.FnMenuSearch, CallID: "x",
	})
	if res.Success {
		t.Fatal("expected failure when session_id is missing")
	}
}

// TC: a port that errors maps cleanly to a FunctionResult failure (no panic).
func TestDispatcher_PortError(t *testing.T) {
	cart := &fakeCart{addErr: errors.New("db down")}
	d := newDispatcher(t, &fakeMenu{}, cart)
	res := d.Dispatch(context.Background(),
		dispatch.Identity{TenantID: "t", SessionID: "s"},
		provider.FunctionCall{
			Name:      dispatch.FnCartAdd,
			CallID:    "x",
			Arguments: map[string]any{"item_id": "i", "qty": 1},
		},
	)
	if res.Success {
		t.Fatal("expected failure when port returns error")
	}
}
