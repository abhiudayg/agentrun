package verify_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentrun/internal/decisionlog"
	"agentrun/internal/verify"
)

func TestLoopEvidenceDirectoryMatchesAttempts(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "decisions.jsonl")
	checks := []verify.Check{{Name: "artifact", Command: "test -f done.txt || { echo 'missing done.txt'; exit 1; }"}}

	outcome, err := verify.Loop(context.Background(), verify.Config{
		Dir:         dir,
		Task:        "create done.txt",
		Checks:      checks,
		MaxAttempts: 3,
		Agent: func(ctx context.Context, prompt string) error {
			if strings.Contains(prompt, "missing done.txt") {
				return os.WriteFile(filepath.Join(dir, "done.txt"), []byte("ok"), 0o600)
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Loop() error = %v", err)
	}
	if err := verify.SaveOutcome(dir, outcome); err != nil {
		t.Fatalf("SaveOutcome() error = %v", err)
	}
	for _, a := range outcome.Attempts {
		entry := decisionlog.Entry{
			Decision: fmt.Sprintf("attempt %d", a.Number),
			Reason:   "fake agent run",
			Evidence: "outcome.json",
			Outcome:  attemptOutcome(a),
		}
		if err := decisionlog.Append(logPath, entry); err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, "outcome.json"))
	if err != nil {
		t.Fatalf("read outcome.json: %v", err)
	}
	var saved struct {
		Status   string `json:"status"`
		Attempts []struct {
			Number int `json:"number"`
		} `json:"attempts"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("outcome.json is not valid JSON: %v", err)
	}
	if saved.Status != string(verify.Passed) || len(saved.Attempts) != len(outcome.Attempts) {
		t.Fatalf("outcome.json = status %q with %d attempts; want %q with %d", saved.Status, len(saved.Attempts), verify.Passed, len(outcome.Attempts))
	}

	f, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("open decision log: %v", err)
	}
	defer f.Close()
	var got []decisionlog.Entry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e decisionlog.Entry
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("decision log line is not valid JSON: %v", err)
		}
		got = append(got, e)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan decision log: %v", err)
	}
	if len(got) != len(outcome.Attempts) {
		t.Fatalf("decision log has %d entries, want %d", len(got), len(outcome.Attempts))
	}
	if got[0].Outcome != "fail" || got[1].Outcome != "pass" {
		t.Errorf("decision outcomes = %q, %q; want fail, pass", got[0].Outcome, got[1].Outcome)
	}
}

func attemptOutcome(a verify.Attempt) string {
	for _, r := range a.Results {
		if !r.Passed {
			return "fail"
		}
	}
	return "pass"
}
