package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EvidenceRoot points at docs/harvest-monti/sdlc/05-tests/evidence/ relative
// to the worktree root. The harness walks up from cwd until it finds the
// `docs/harvest-monti/sdlc/05-tests` directory so it works whether the test
// is invoked from the worktree root or the e2e/ directory.
func EvidenceRoot() (string, error) {
	cur, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	for i := 0; i < 8; i++ {
		probe := filepath.Join(cur, "docs", "harvest-monti", "sdlc", "05-tests", "evidence")
		if _, err := os.Stat(filepath.Dir(probe)); err == nil {
			if err := os.MkdirAll(probe, 0o755); err != nil {
				return "", err
			}
			return probe, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	// Fall back to a temp dir so tests still pass when docs/ is unreachable
	// (e.g. running in a fresh container without the harvest-god checkout).
	tmp, err := os.MkdirTemp("", "monti-e2e-evidence-*")
	if err != nil {
		return "", err
	}
	return tmp, nil
}

// CaseDir creates `evidence/<test_id>/case-<n>/` and returns its path.
func CaseDir(testID string, n int) (string, error) {
	root, err := EvidenceRoot()
	if err != nil {
		return "", err
	}
	d := filepath.Join(root, testID, fmt.Sprintf("case-%d", n))
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return d, nil
}

// WriteJSONL appends a structured record to a .jsonl file inside the
// case directory.
type Recorder struct {
	dir string
}

func NewRecorder(dir string) *Recorder { return &Recorder{dir: dir} }

func (r *Recorder) Append(filename string, v any) error {
	f, err := os.OpenFile(filepath.Join(r.dir, filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}

func (r *Recorder) WriteFile(filename, body string) error {
	return os.WriteFile(filepath.Join(r.dir, filename), []byte(body), 0o644)
}

// Stamp returns an evidence-stamped time string.
func Stamp() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
