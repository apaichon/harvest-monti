// Package router is the harvest-monti HTTP surface — a thin Fiber-v3-shaped
// router built on net/http.
//
// We carry our own Router rather than pulling Fiber v3 directly in the
// initial scaffold so `go build ./...` is hermetic on first checkout
// (cmd/monti-api wires the routes; swapping the implementation to
// fiber.App is a single file change in main.go).
//
// The API surface deliberately mirrors Fiber v3 — Get/Post/Delete,
// Group, Ctx.Params / Ctx.Bind, Ctx.Locals — so plugin handlers do not
// need to change once the swap happens.
package router

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
)

// HandlerFunc is the Fiber-style handler signature.
type HandlerFunc func(c *Ctx) error

// Middleware wraps a handler.
type Middleware func(HandlerFunc) HandlerFunc

// Router is the registration surface.
type Router struct {
	mu        sync.RWMutex
	routes    []route
	mws       []Middleware
	prefix    string
	notFound  HandlerFunc
	upgradeWS WSUpgrader
	parent    *Router // set on Group children; routes accrete on the root
}

type route struct {
	method  string
	pattern string
	parts   []string
	handler HandlerFunc
	mws     []Middleware
}

// New returns a fresh Router.
func New() *Router {
	return &Router{
		notFound: func(c *Ctx) error { return c.Status(http.StatusNotFound).JSON(map[string]string{"error": "not found"}) },
	}
}

// Use appends middleware to the chain applied to every subsequent route.
func (r *Router) Use(m ...Middleware) { r.mws = append(r.mws, m...) }

// Group returns a sub-router with the prefix applied to every registration.
// Children write back to the parent so we keep one route table.
func (r *Router) Group(prefix string, mws ...Middleware) *Router {
	return &Router{
		prefix:    r.prefix + prefix,
		mws:       append(append([]Middleware{}, r.mws...), mws...),
		notFound:  r.notFound,
		upgradeWS: r.upgradeWS,
		parent:    r,
	}
}

// root walks up to the top-most Router so all routes share one table.
func (r *Router) root() *Router {
	cur := r
	for cur.parent != nil {
		cur = cur.parent
	}
	return cur
}

// Get registers a GET handler.
func (r *Router) Get(pattern string, h HandlerFunc)    { r.add(http.MethodGet, pattern, h) }
// Post registers a POST handler.
func (r *Router) Post(pattern string, h HandlerFunc)   { r.add(http.MethodPost, pattern, h) }
// Delete registers a DELETE handler.
func (r *Router) Delete(pattern string, h HandlerFunc) { r.add(http.MethodDelete, pattern, h) }
// Put registers a PUT handler.
func (r *Router) Put(pattern string, h HandlerFunc)    { r.add(http.MethodPut, pattern, h) }
// Patch registers a PATCH handler.
func (r *Router) Patch(pattern string, h HandlerFunc)  { r.add(http.MethodPatch, pattern, h) }

// SetWebSocketUpgrader sets the WS upgrader used by Ctx.UpgradeWS.
// Wired by main() so plugins do not import a specific WS implementation.
func (r *Router) SetWebSocketUpgrader(u WSUpgrader) { r.root().upgradeWS = u }

func (r *Router) add(method, pattern string, h HandlerFunc) {
	full := r.prefix + pattern
	parts := splitPath(full)
	root := r.root()
	root.mu.Lock()
	defer root.mu.Unlock()
	root.routes = append(root.routes, route{
		method:  method,
		pattern: full,
		parts:   parts,
		handler: h,
		mws:     append([]Middleware{}, r.mws...),
	})
}

// Routes returns every registered (method, pattern) — used by tests to
// prove the 11 endpoints required by TASK-0009 are wired.
func (r *Router) Routes() []string {
	root := r.root()
	root.mu.RLock()
	defer root.mu.RUnlock()
	out := make([]string, 0, len(root.routes))
	for _, rt := range root.routes {
		out = append(out, rt.method+" "+rt.pattern)
	}
	return out
}

// ServeHTTP implements http.Handler so the router can be mounted by main.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	root := r.root()
	parts := splitPath(req.URL.Path)
	root.mu.RLock()
	matched, params := matchRoute(root.routes, req.Method, parts)
	root.mu.RUnlock()

	ctx := &Ctx{
		req:       req,
		resp:      w,
		params:    params,
		locals:    map[string]any{},
		status:    http.StatusOK,
		upgradeWS: root.upgradeWS,
	}
	var h HandlerFunc
	if matched == nil {
		h = root.notFound
	} else {
		h = matched.handler
		for i := len(matched.mws) - 1; i >= 0; i-- {
			h = matched.mws[i](h)
		}
	}
	if err := h(ctx); err != nil {
		ctx.handleErr(err)
	}
}

// splitPath cleans a URL path into segments. "/api/v1/menu" -> ["api","v1","menu"].
func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// matchRoute walks routes in declaration order; the first method+pattern
// match wins. Param segments "{name}" or ":name" populate params.
func matchRoute(routes []route, method string, parts []string) (*route, map[string]string) {
	for i := range routes {
		rt := &routes[i]
		if rt.method != method || len(rt.parts) != len(parts) {
			continue
		}
		params := map[string]string{}
		ok := true
		for j, seg := range rt.parts {
			if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
				params[seg[1:len(seg)-1]] = parts[j]
				continue
			}
			if strings.HasPrefix(seg, ":") {
				params[seg[1:]] = parts[j]
				continue
			}
			if seg != parts[j] {
				ok = false
				break
			}
		}
		if ok {
			return rt, params
		}
	}
	return nil, nil
}

// ----- Ctx -----

// Ctx is the per-request context. Fiber v3-shaped surface so plugin code
// reads the same on both router implementations.
type Ctx struct {
	req       *http.Request
	resp      http.ResponseWriter
	params    map[string]string
	locals    map[string]any
	status    int
	body      []byte
	hasBody   bool
	upgradeWS WSUpgrader
}

// Context returns the request context.
func (c *Ctx) Context() context.Context { return c.req.Context() }

// Params returns a URL path parameter.
func (c *Ctx) Params(name string) string { return c.params[name] }

// Query returns a single query value.
func (c *Ctx) Query(name string) string { return c.req.URL.Query().Get(name) }

// Locals reads or writes a per-request value (used by middleware).
func (c *Ctx) Locals(key string, value ...any) any {
	if len(value) > 0 {
		c.locals[key] = value[0]
		return value[0]
	}
	return c.locals[key]
}

// Status sets the HTTP status code; returns the receiver for chaining.
func (c *Ctx) Status(code int) *Ctx {
	c.status = code
	return c
}

// JSON serializes v and writes the response.
func (c *Ctx) JSON(v any) error {
	c.resp.Header().Set("Content-Type", "application/json")
	c.resp.WriteHeader(c.status)
	return json.NewEncoder(c.resp).Encode(v)
}

// Bind decodes the request body into v.
func (c *Ctx) Bind(v any) error {
	if c.req.Body == nil {
		return errors.New("router: empty body")
	}
	b, err := io.ReadAll(c.req.Body)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, v)
}

// Request returns the underlying *http.Request (used by WS upgrader).
func (c *Ctx) Request() *http.Request { return c.req }

// Response returns the underlying ResponseWriter (used by WS upgrader).
func (c *Ctx) Response() http.ResponseWriter { return c.resp }

// UpgradeWS hands the connection to the wired WS upgrader. Returns
// ErrNoWSUpgrader if main() never installed one.
func (c *Ctx) UpgradeWS() (WSConn, error) {
	if c.upgradeWS == nil {
		return nil, ErrNoWSUpgrader
	}
	return c.upgradeWS.Upgrade(c.resp, c.req)
}

// ErrNoWSUpgrader is returned by Ctx.UpgradeWS when none is installed.
var ErrNoWSUpgrader = errors.New("router: no WS upgrader installed")

// handleErr is the fallback when a handler returns a non-nil error.
func (c *Ctx) handleErr(err error) {
	c.resp.Header().Set("Content-Type", "application/json")
	c.resp.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(c.resp).Encode(map[string]string{"error": err.Error()})
}

// ----- WebSocket abstraction -----

// WSConn is what plugins use to write status frames. Implementations live
// outside this package — order/ws.go uses an in-memory fake during tests.
type WSConn interface {
	WriteJSON(v any) error
	Close() error
}

// WSUpgrader upgrades an HTTP request to a WS connection.
type WSUpgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request) (WSConn, error)
}
