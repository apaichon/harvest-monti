// Package db is a thin abstraction over *sql.DB so plugin repositories
// stay storage-agnostic. The monti-api binary supplies a *sql.DB at boot;
// the test suite supplies an in-memory fake.
//
// Lives under internal/infra/db per DES-0007 §1 + harvest-monti SBA layout.
package db

import (
	"context"
	"database/sql"
	"errors"
)

// ErrNotFound is the canonical "not in store" error. Repositories return it
// when a tenant-scoped query returns zero rows so the handler can reply 404
// without leaking cross-tenant existence (TEST-0008 TC-10).
var ErrNotFound = errors.New("monti/db: not found")

// Executor matches both *sql.DB and *sql.Tx — repositories accept this so
// they can be composed inside a service-level transaction (cart submit ->
// order place is one tx in DES-0007 §1).
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Pool is the boot-time handle held by main(). A nil Pool is legal in unit
// tests; plugin services degrade to in-memory mode in that case.
type Pool struct {
	DB *sql.DB
}

// New wraps a *sql.DB so callers can pass a single struct around.
func New(d *sql.DB) *Pool { return &Pool{DB: d} }

// Exec runs against the underlying pool. Plugin services use a tx Executor
// when they need atomicity; everything else uses the pool.
func (p *Pool) Exec() Executor {
	if p == nil || p.DB == nil {
		return nil
	}
	return p.DB
}
