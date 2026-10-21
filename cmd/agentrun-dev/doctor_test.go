package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func fakePath(t *testing.T, tools ...string) {
	t.Helper()
	dir := t.TempDir()
	for _, tool := range tools {
		if err := os.WriteFile(filepath.Join(dir, tool), nil, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
}

func TestDoctorReportsEachToolWhenAllFound(t *testing.T) {
	fakePath(t, "git", "go", "sh")
	var stdout bytes.Buffer

	err := doctor(&stdout, false, nil)

	wantOut := "git: ok\ngo: ok\nsh: ok\n"
	if err != nil || stdout.String() != wantOut {
		t.Fatalf("doctor() = %v, stdout %q; want nil, stdout %q", err, stdout.String(), wantOut)
	}
}

func TestDoctorReportsMissingToolWithFourLineError(t *testing.T) {
	fakePath(t, "git", "sh")
	var stdout bytes.Buffer

	err := doctor(&stdout, false, nil)

	wantOut := "git: ok\ngo: missing\nsh: ok\n"
	wantErr := "error: missing required tools: go\n" +
		"  cause: not found by exec.LookPath: go\n" +
		"  fix:   install the missing tools and put them on PATH\n" +
		"  docs:  https://agentrun.dev/docs/troubleshooting#commands"
	if err == nil || err.Error() != wantErr || stdout.String() != wantOut {
		t.Fatalf("doctor() = %v, stdout %q; want error %q, stdout %q", err, stdout.String(), wantErr, wantOut)
	}
}

func TestRunDoctorExitsOneWhenToolMissing(t *testing.T) {
	fakePath(t, "git", "go")
	var stdout, stderr bytes.Buffer

	code := run([]string{"doctor"}, &stdout, &stderr, commands)

	if code != 1 || !bytes.Contains(stderr.Bytes(), []byte("missing required tools: sh")) {
		t.Fatalf("run(doctor) = %d, stderr %q; want 1 with missing sh reported", code, stderr.String())
	}
}
