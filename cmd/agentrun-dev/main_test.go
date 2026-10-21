package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"testing"
)

func recorder(name string, seen *[]string) command {
	return func(stdout io.Writer, jsonOut bool, args []string) error {
		*seen = append(*seen, fmt.Sprintf("%s json=%t args=%v", name, jsonOut, args))
		return nil
	}
}

func TestRunDispatchesToNamedCommand(t *testing.T) {
	var seen []string
	cmds := map[string]command{
		"verify": recorder("verify", &seen),
		"doctor": recorder("doctor", &seen),
	}
	var stdout, stderr bytes.Buffer

	code := run([]string{"verify", "dir", "lint=golint"}, &stdout, &stderr, cmds)

	want := []string{"verify json=false args=[dir lint=golint]"}
	if code != 0 || !slices.Equal(seen, want) {
		t.Fatalf("run() = %d, calls %q; want 0, calls %q (stderr %q)", code, seen, want, stderr.String())
	}
}

func TestRunPassesCommandOutputToStdout(t *testing.T) {
	cmds := map[string]command{
		"doctor": func(stdout io.Writer, jsonOut bool, args []string) error {
			_, err := io.WriteString(stdout, "all tools found\n")
			return err
		},
	}
	var stdout, stderr bytes.Buffer

	code := run([]string{"doctor"}, &stdout, &stderr, cmds)

	if code != 0 || stdout.String() != "all tools found\n" {
		t.Fatalf("run() = %d, stdout %q; want 0, stdout %q", code, stdout.String(), "all tools found\n")
	}
}

func TestRunParsesJSONFlagBeforeSubcommand(t *testing.T) {
	var seen []string
	cmds := map[string]command{"doctor": recorder("doctor", &seen)}
	var stdout, stderr bytes.Buffer

	code := run([]string{"--json", "doctor", "extra"}, &stdout, &stderr, cmds)

	want := []string{"doctor json=true args=[extra]"}
	if code != 0 || !slices.Equal(seen, want) {
		t.Fatalf("run() = %d, calls %q; want 0, calls %q", code, seen, want)
	}
}

func TestRunLeavesJSONFlagAfterSubcommandAsArgument(t *testing.T) {
	var seen []string
	cmds := map[string]command{"doctor": recorder("doctor", &seen)}
	var stdout, stderr bytes.Buffer

	code := run([]string{"doctor", "--json"}, &stdout, &stderr, cmds)

	want := []string{"doctor json=false args=[--json]"}
	if code != 0 || !slices.Equal(seen, want) {
		t.Fatalf("run() = %d, calls %q; want 0, calls %q", code, seen, want)
	}
}

func TestRunRejectsUnknownGlobalFlag(t *testing.T) {
	var seen []string
	cmds := map[string]command{"doctor": recorder("doctor", &seen)}
	var stdout, stderr bytes.Buffer

	code := run([]string{"--bogus", "doctor"}, &stdout, &stderr, cmds)

	if code != 2 || len(seen) != 0 {
		t.Fatalf("run() = %d, calls %q; want 2 and no calls", code, seen)
	}
}

func TestRunUnknownCommandPrintsFourLineError(t *testing.T) {
	cmds := map[string]command{"doctor": recorder("doctor", new([]string))}
	var stdout, stderr bytes.Buffer

	code := run([]string{"nope"}, &stdout, &stderr, cmds)

	want := "error: unknown command \"nope\"\n" +
		"  cause: no subcommand is named \"nope\"\n" +
		"  fix:   check the spelling of the command name\n" +
		"  docs:  https://agentrun.dev/docs/troubleshooting#commands\n"
	if code != 2 || stderr.String() != want || stdout.Len() != 0 {
		t.Fatalf("run() = %d, stderr %q; want 2, stderr %q", code, stderr.String(), want)
	}
}

func TestRunMissingCommandPrintsFourLineError(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := run([]string{"--json"}, &stdout, &stderr, map[string]command{})

	want := "error: no command given\n" +
		"  cause: no subcommand name was passed after the global flags\n" +
		"  fix:   run agentrun-dev <command> [args]\n" +
		"  docs:  https://agentrun.dev/docs/troubleshooting#commands\n"
	if code != 2 || stderr.String() != want {
		t.Fatalf("run() = %d, stderr %q; want 2, stderr %q", code, stderr.String(), want)
	}
}

func TestRunReturnsOneWhenCommandFails(t *testing.T) {
	cmds := map[string]command{
		"verify": func(stdout io.Writer, jsonOut bool, args []string) error {
			return errors.New("checks failed")
		},
	}
	var stdout, stderr bytes.Buffer

	code := run([]string{"verify"}, &stdout, &stderr, cmds)

	if code != 1 || stderr.String() != "checks failed\n" {
		t.Fatalf("run() = %d, stderr %q; want 1, stderr %q", code, stderr.String(), "checks failed\n")
	}
}
