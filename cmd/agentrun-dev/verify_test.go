package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type verifyJSONResult struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
}

func runCLI(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut, commands)
	return code, out.String(), errOut.String()
}

func TestVerifyPassingCheckExitsZero(t *testing.T) {
	code, stdout, stderr := runCLI("verify", t.TempDir(), "build=true")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr %q)", code, stderr)
	}
	if !strings.Contains(stdout, "build") {
		t.Errorf("stdout = %q, want it to name the passing check", stdout)
	}
}

func TestVerifyFailingCheckExitsOneWithOutput(t *testing.T) {
	code, stdout, stderr := runCLI("verify", t.TempDir(), "unit=echo boom; exit 3")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(stdout, "unit") || !strings.Contains(stdout, "boom") {
		t.Errorf("stdout = %q, want the failing check name and its output", stdout)
	}
	if !strings.Contains(stderr, "error: check \"unit\" failed") {
		t.Errorf("stderr = %q, want a four-line error naming the failed check", stderr)
	}
}

func TestVerifyRunsChecksInGivenDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "marker"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runCLI("verify", dir, "marker=test -f marker")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 because the check runs inside dir (stderr %q)", code, stderr)
	}
}

func TestVerifyJSONPrintsPassingResults(t *testing.T) {
	code, stdout, stderr := runCLI("--json", "verify", t.TempDir(), "build=true")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr %q)", code, stderr)
	}
	var results []verifyJSONResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("stdout is not a JSON array: %v (stdout %q)", err, stdout)
	}
	if len(results) != 1 || results[0].Name != "build" || !results[0].Passed || results[0].ExitCode != 0 {
		t.Errorf("results = %+v, want one passing build result", results)
	}
}

func TestVerifyJSONPrintsFailingResultAndExitsOne(t *testing.T) {
	code, stdout, _ := runCLI("--json", "verify", t.TempDir(), "unit=echo boom; exit 3")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	var results []verifyJSONResult
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Fatalf("stdout is not a JSON array: %v (stdout %q)", err, stdout)
	}
	if len(results) != 1 {
		t.Fatalf("results = %+v, want one failing result", results)
	}
	failed := results[0]
	if failed.Name != "unit" || failed.Passed || failed.ExitCode != 3 || !strings.Contains(failed.Output, "boom") {
		t.Errorf("failing result = %+v, want unit exit 3 with output containing boom", failed)
	}
}

func TestVerifyBadUsagePrintsFourLineError(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name string
		args []string
	}{
		{name: "no directory or checks", args: []string{"verify"}},
		{name: "directory without checks", args: []string{"verify", dir}},
		{name: "check without equals sign", args: []string{"verify", dir, "build"}},
		{name: "check with empty name", args: []string{"verify", dir, "=true"}},
		{name: "check with empty command", args: []string{"verify", dir, "build="}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, stdout, stderr := runCLI(tt.args...)

			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want empty", stdout)
			}
			if !strings.HasPrefix(stderr, "error: ") || strings.Count(stderr, "\n") != 4 || !strings.Contains(stderr, "  fix:") {
				t.Errorf("stderr = %q, want a four-line error", stderr)
			}
		})
	}
}
