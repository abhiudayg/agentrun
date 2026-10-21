package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agentrun/internal/verify"
)

func saveOutcome(t *testing.T, dir string, outcome verify.Outcome) {
	t.Helper()
	if err := verify.SaveOutcome(dir, outcome); err != nil {
		t.Fatalf("SaveOutcome() error = %v", err)
	}
}

func TestEvidencePrintsStatusAndAttemptResults(t *testing.T) {
	dir := t.TempDir()
	saveOutcome(t, dir, verify.Outcome{
		Status: verify.Exhausted,
		Attempts: []verify.Attempt{
			{Number: 1, Results: []verify.Result{
				{Name: "lint", Passed: true},
				{Name: "unit", Passed: false, ExitCode: 1},
			}},
			{Number: 2, AgentErr: errors.New("agent exited 1"), Results: []verify.Result{
				{Name: "unit", Passed: false, ExitCode: 1},
			}},
		},
	})
	var stdout, stderr bytes.Buffer

	code := run([]string{"evidence", dir}, &stdout, &stderr, commands)

	want := "status: exhausted\n" +
		"attempt 1\n" +
		"  lint: pass (exit 0)\n" +
		"  unit: fail (exit 1)\n" +
		"attempt 2\n" +
		"  agent: agent exited 1\n" +
		"  unit: fail (exit 1)\n"
	if code != 0 || stdout.String() != want || stderr.Len() != 0 {
		t.Fatalf("run() = %d, stdout %q, stderr %q; want 0, stdout %q", code, stdout.String(), stderr.String(), want)
	}
}

func TestEvidenceMissingOutcomePrintsFourLineError(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := run([]string{"evidence", dir}, &stdout, &stderr, commands)

	lines := strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n")
	if code != 1 || stdout.Len() != 0 || len(lines) != 4 {
		t.Fatalf("run() = %d, stdout %q, stderr %q; want 1 and a four-line error", code, stdout.String(), stderr.String())
	}
	prefixes := []string{"error: ", "  cause: ", "  fix:   ", "  docs:  "}
	for i, prefix := range prefixes {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Errorf("line %d = %q; want prefix %q", i, lines[i], prefix)
		}
	}
}

func TestEvidenceMalformedOutcomeReturnsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "outcome.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer

	code := run([]string{"evidence", dir}, &stdout, &stderr, commands)

	if code != 1 || stdout.Len() != 0 || !strings.HasPrefix(stderr.String(), "decode outcome.json:") {
		t.Fatalf("run() = %d, stdout %q, stderr %q; want 1 and a decode error", code, stdout.String(), stderr.String())
	}
}

func TestEvidenceWithoutDirectoryPrintsFourLineError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"evidence"}, &stdout, &stderr, commands)

	lines := strings.Split(strings.TrimSuffix(stderr.String(), "\n"), "\n")
	if code != 1 || stdout.Len() != 0 || len(lines) != 4 || !strings.HasPrefix(lines[0], "error: ") {
		t.Fatalf("run() = %d, stdout %q, stderr %q; want 1 and a four-line error", code, stdout.String(), stderr.String())
	}
}

func TestEvidenceIsRegisteredInCommands(t *testing.T) {
	if _, ok := commands["evidence"]; !ok {
		t.Fatal(`commands has no "evidence" entry`)
	}
}
