// tags_prefix_test.go — build-time guard required by TASK-0009.
// Walks every .go file under the project (excluding vendor, .worktrees,
// the test file itself) and fails if any cache tag literal begins with a
// reserved harvest-core namespace (analytics:, duckdb:, clickhouse:,
// usage:). Also asserts at least one "monti:" tag exists to catch the
// inverse mistake (someone deleting the whole package).
package cache

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// reservedTagRE matches Go string literals that look like cache tags
// (lowercase identifiers separated by colons, no spaces) and start with a
// reserved harvest-core analytics namespace. Restricting the character
// class to [a-z0-9_:{}.-] keeps human-readable usage strings (e.g.
// "usage: migrate up|down") out of the false-positive set.
var reservedTagRE = regexp.MustCompile(`"(?:analytics|duckdb|clickhouse|usage):[a-zA-Z0-9_:{}.\-]+"`)

// montiTagRE confirms the positive case: at least one monti: tag exists.
var montiTagRE = regexp.MustCompile(`"monti:[a-zA-Z0-9_:{}.\-]+"`)

func TestNoReservedCacheTagNamespaces(t *testing.T) {
	root := projectRoot(t)

	var foundMontiTag bool
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == ".git" || base == ".worktrees" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// Skip this file itself — its regex literals are not real tags.
		if strings.HasSuffix(path, "tags_prefix_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if hits := reservedTagRE.FindAllString(string(data), -1); len(hits) > 0 {
			t.Errorf("reserved cache namespace found in %s: %v", path, hits)
		}
		if montiTagRE.MatchString(string(data)) {
			foundMontiTag = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if !foundMontiTag {
		t.Fatalf("no monti: cache tags found anywhere in the source tree — cache.tags.go missing?")
	}
}

// projectRoot walks up from the current package directory until it finds
// a go.mod so the test runs from any check-out depth.
func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found upward from %s", wd)
		}
		dir = parent
	}
}
