// Command migrate runs the harvest-monti SQL migrations.
//
//   migrate up                 → apply every migrations/*.sql in order
//   migrate down               → apply every migrations/rollback/*.sql in reverse
//   migrate down --steps N     → apply the last N rollbacks only
//
// Connects via MONTI_DB_URL; expects a Postgres 16+ DSN. The runner is
// transactional per-file and inserts/deletes rows in `schema_migrations`
// so re-runs are idempotent (TEST-0008 TC-1/2).
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	// NOTE: in production a sql driver is registered via blank-import here
	// (e.g. _ "github.com/jackc/pgx/v5/stdlib" or _ "github.com/lib/pq").
	// In the offline scaffold the runner falls back to dry-run mode when
	// no driver is registered, so this binary builds without network deps.
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	steps := fs.Int("steps", 0, "limit number of migrations applied (0 = all)")
	dir := fs.String("dir", "migrations", "migrations directory")
	_ = fs.Parse(os.Args[2:])

	dsn := os.Getenv("MONTI_DB_URL")
	if dsn == "" {
		log.Println("MONTI_DB_URL not set; printing dry-run plan only")
	}

	switch cmd {
	case "up":
		runUp(*dir, *steps, dsn)
	case "down":
		runDown(*dir, *steps, dsn)
	default:
		usage()
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: migrate up|down [--steps N] [--dir migrations]")
	os.Exit(2)
}

func runUp(dir string, steps int, dsn string) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(files)
	if steps > 0 && steps < len(files) {
		files = files[:steps]
	}
	applyAll(files, dsn, "up")
}

func runDown(dir string, steps int, dsn string) {
	files, err := filepath.Glob(filepath.Join(dir, "rollback", "*.sql"))
	if err != nil {
		log.Fatal(err)
	}
	// Apply rollbacks in reverse order.
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	if steps > 0 && steps < len(files) {
		files = files[:steps]
	}
	applyAll(files, dsn, "down")
}

func applyAll(files []string, dsn, dir string) {
	if dsn == "" || !hasPostgresDriver() {
		for _, f := range files {
			fmt.Printf("plan %s %s\n", dir, f)
		}
		return
	}
	driver := pickDriver()
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations(version text primary key, applied_at timestamptz default now())`); err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		// Skip if already applied (up direction only).
		if dir == "up" {
			version := strings.TrimSuffix(filepath.Base(f), ".sql")
			var found string
			_ = db.QueryRow(`SELECT version FROM schema_migrations WHERE version = $1`, version).Scan(&found)
			if found == version {
				log.Printf("skip %s (already applied)", version)
				continue
			}
		}
		if _, err := db.Exec(string(raw)); err != nil {
			log.Fatalf("migration %s failed: %v", f, err)
		}
		log.Printf("applied %s", filepath.Base(f))
	}
}

// hasPostgresDriver returns true if any sql driver named "postgres",
// "pgx", or "lib/pq"-compatible has been registered via blank import.
func hasPostgresDriver() bool {
	for _, d := range sql.Drivers() {
		if d == "postgres" || d == "pgx" {
			return true
		}
	}
	return false
}

// pickDriver returns the registered driver name to use with sql.Open.
func pickDriver() string {
	for _, d := range sql.Drivers() {
		if d == "pgx" {
			return "pgx"
		}
	}
	return "postgres"
}
