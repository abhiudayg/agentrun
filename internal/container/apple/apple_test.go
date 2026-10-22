package apple_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"agentrun/internal/container"
	"agentrun/internal/container/apple"
)

var _ container.Backend = (*apple.Apple)(nil)

func validSpec() container.Spec {
	return container.Spec{
		Image:    "ghcr.io/example/agent:1.2.3",
		User:     "1000:1000",
		ReadOnly: true,
		Tmpfs:    []string{"/tmp"},
		Network:  "agentrun-internal",
		Cmd:      []string{"sh", "-c", "echo hi"},
	}
}

func fakeContainer(t *testing.T, exitCode int) (binary string, argsFile string) {
	t.Helper()
	dir := t.TempDir()
	binary = filepath.Join(dir, "container")
	argsFile = filepath.Join(dir, "args.txt")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$@\" > '" + argsFile + "'\n" +
		"echo \"container stdout\"\n" +
		"echo \"container stderr\" >&2\n" +
		"exit " + strconv.Itoa(exitCode) + "\n"
	if err := os.WriteFile(binary, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake container: %v", err)
	}
	return binary, argsFile
}

func recordedArgs(t *testing.T, argsFile string) []string {
	t.Helper()
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("read recorded args: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func TestRunPassesRunArgsAfterRunSubcommandWithoutDetach(t *testing.T) {
	binary, argsFile := fakeContainer(t, 0)
	backend := &apple.Apple{Binary: binary, Output: &bytes.Buffer{}}

	if _, err := backend.Run(context.Background(), validSpec()); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	want := []string{
		"run",
		"--user", "1000:1000",
		"--read-only",
		"--tmpfs", "/tmp",
		"--network", "agentrun-internal",
		"ghcr.io/example/agent:1.2.3",
		"sh", "-c", "echo hi",
	}
	got := recordedArgs(t, argsFile)
	if !slices.Equal(got, want) {
		t.Errorf("container args = %q, want %q", got, want)
	}
	if slices.Contains(got, "-d") {
		t.Errorf("container args = %q, must not contain -d", got)
	}
}

func TestRunReturnsExitCodeWithNilError(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
	}{
		{name: "zero exit", exitCode: 0},
		{name: "nonzero exit", exitCode: 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binary, _ := fakeContainer(t, tt.exitCode)
			backend := &apple.Apple{Binary: binary, Output: &bytes.Buffer{}}

			got, err := backend.Run(context.Background(), validSpec())

			if err != nil {
				t.Fatalf("Run() error = %v, want nil", err)
			}
			if got.ExitCode != tt.exitCode {
				t.Errorf("Run() exit code = %d, want %d", got.ExitCode, tt.exitCode)
			}
		})
	}
}

func TestRunWritesStdoutAndStderrToOutput(t *testing.T) {
	binary, _ := fakeContainer(t, 0)
	var out bytes.Buffer
	backend := &apple.Apple{Binary: binary, Output: &out}

	if _, err := backend.Run(context.Background(), validSpec()); err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	for _, line := range []string{"container stdout", "container stderr"} {
		if !strings.Contains(out.String(), line) {
			t.Errorf("Output = %q, want it to contain %q", out.String(), line)
		}
	}
}

func TestRunInvalidSpecReturnsErrRootUserWithoutStartingBinary(t *testing.T) {
	binary, argsFile := fakeContainer(t, 0)
	backend := &apple.Apple{Binary: binary, Output: &bytes.Buffer{}}
	spec := validSpec()
	spec.User = "0:0"

	_, err := backend.Run(context.Background(), spec)

	if !errors.Is(err, container.ErrRootUser) {
		t.Fatalf("Run() error = %v, want %v", err, container.ErrRootUser)
	}
	if _, statErr := os.Stat(argsFile); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("container binary ran for invalid spec; stat args file error = %v", statErr)
	}
}

func TestRunMissingBinaryReturnsError(t *testing.T) {
	backend := &apple.Apple{Binary: filepath.Join(t.TempDir(), "missing"), Output: &bytes.Buffer{}}

	if _, err := backend.Run(context.Background(), validSpec()); err == nil {
		t.Fatal("Run() error = nil, want error for missing binary")
	}
}

func TestRunCancelledContextReturnsError(t *testing.T) {
	binary, argsFile := fakeContainer(t, 0)
	backend := &apple.Apple{Binary: binary, Output: &bytes.Buffer{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := backend.Run(ctx, validSpec())

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want %v", err, context.Canceled)
	}
	if _, statErr := os.Stat(argsFile); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("container binary ran after cancellation; stat args file error = %v", statErr)
	}
}
