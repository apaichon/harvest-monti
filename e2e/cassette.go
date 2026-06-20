// Package e2e is the TASK-0014 end-to-end harness. It runs against either
// the docker-compose stack from compose.e2e.yml or, when Docker is
// unavailable, against the in-memory fakes shipped by TASK-0009 / TASK-0010.
//
// The cassette schema here mirrors cmd/mock-gemini exactly — the JSON files
// are the single source of truth.
package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Frame is one entry in a cassette. Identical to cmd/mock-gemini.Frame.
type Frame struct {
	AfterMS   int            `json:"after_ms"`
	Type      string         `json:"type"`
	Role      string         `json:"role,omitempty"`
	Text      string         `json:"text,omitempty"`
	Final     bool           `json:"final,omitempty"`
	Language  string         `json:"language,omitempty"`
	Name      string         `json:"name,omitempty"`
	CallID    string         `json:"call_id,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
	PCM       string         `json:"pcm,omitempty"`
	Code      string         `json:"code,omitempty"`
	Message   string         `json:"message,omitempty"`
}

// Cassette is the JSON file shape.
type Cassette struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Frames      []Frame `json:"frames"`
}

// LoadCassette reads a cassette JSON file by short name from e2e/fixtures/voice.
func LoadCassette(name string) (Cassette, error) {
	dir, err := fixturesDir()
	if err != nil {
		return Cassette{}, err
	}
	path := filepath.Join(dir, name+".cassette.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return Cassette{}, fmt.Errorf("load cassette %s: %w", name, err)
	}
	var c Cassette
	if err := json.Unmarshal(b, &c); err != nil {
		return Cassette{}, fmt.Errorf("parse cassette %s: %w", name, err)
	}
	return c, nil
}

// fixturesDir returns the absolute path to e2e/fixtures/voice. Tests run
// from inside e2e/ when invoked via `go test ./e2e/...`.
func fixturesDir() (string, error) {
	candidates := []string{
		"fixtures/voice",
		"e2e/fixtures/voice",
		"../e2e/fixtures/voice",
	}
	for _, c := range candidates {
		if abs, err := filepath.Abs(c); err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs, nil
			}
		}
	}
	return "", fmt.Errorf("e2e fixtures dir not found from cwd")
}
