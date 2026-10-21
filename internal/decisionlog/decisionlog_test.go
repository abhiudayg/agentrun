package decisionlog_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentrun/internal/decisionlog"
)

func TestAppendWritesOneJSONLinePerEntryInOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "decisions.jsonl")
	entries := []decisionlog.Entry{
		{Decision: "use JSON lines", Reason: "append-friendly", Evidence: "backlog item 2", Outcome: "accepted"},
		{Decision: "reject empty decisions", Reason: "entries must say what changed", Evidence: "item 2 done-when", Outcome: "accepted"},
		{Decision: "stamp time in Append", Reason: "callers should not forge time", Evidence: "review", Outcome: "pending"},
	}

	for _, e := range entries {
		if err := decisionlog.Append(path, e); err != nil {
			t.Fatalf("Append(%q) error = %v", e.Decision, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read decision log: %v", err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != len(entries) {
		t.Fatalf("decision log has %d lines, want %d", len(lines), len(entries))
	}
	for i, line := range lines {
		var got decisionlog.Entry
		if err := json.Unmarshal([]byte(line), &got); err != nil {
			t.Fatalf("line %d is not valid JSON: %v", i, err)
		}
		want := entries[i]
		if got.Decision != want.Decision || got.Reason != want.Reason || got.Evidence != want.Evidence || got.Outcome != want.Outcome {
			t.Errorf("line %d = %+v, want fields of %+v", i, got, want)
		}
		if got.Time.IsZero() {
			t.Errorf("line %d has zero time", i)
		}
	}
}

func TestAppendReturnsErrorWhenLogCannotBeOpened(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-dir", "decisions.jsonl")

	err := decisionlog.Append(path, decisionlog.Entry{Decision: "d"})

	if err == nil {
		t.Fatal("Append() error = nil, want error for a path in a missing directory")
	}
}

func TestAppendRejectsEmptyDecision(t *testing.T) {
	for _, decision := range []string{"", "   "} {
		path := filepath.Join(t.TempDir(), "decisions.jsonl")

		err := decisionlog.Append(path, decisionlog.Entry{Decision: decision, Reason: "r"})

		if !errors.Is(err, decisionlog.ErrEmptyDecision) {
			t.Errorf("Append(decision=%q) error = %v, want ErrEmptyDecision", decision, err)
		}
		if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("Append(decision=%q) wrote to the log; stat error = %v", decision, statErr)
		}
	}
}
