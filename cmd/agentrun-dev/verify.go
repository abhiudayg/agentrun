package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"agentrun/internal/errors"
	"agentrun/internal/verify"
)

type checkJSON struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
}

func runVerify(stdout io.Writer, jsonOut bool, args []string) error {
	if len(args) < 2 {
		return errors.New(
			"verify needs a directory and at least one check",
			"no directory or check was given after verify",
			"run agentrun-dev verify <dir> name=cmd ...",
			"https://agentrun.dev/docs/troubleshooting#commands",
		)
	}
	dir, specs := args[0], args[1:]

	checks := make([]verify.Check, 0, len(specs))
	for _, spec := range specs {
		name, command, ok := strings.Cut(spec, "=")
		if !ok || name == "" || command == "" {
			return errors.New(
				fmt.Sprintf("invalid check %q", spec),
				"a check must look like name=cmd with both parts non-empty",
				"run agentrun-dev verify <dir> name=cmd ...",
				"https://agentrun.dev/docs/troubleshooting#commands",
			)
		}
		checks = append(checks, verify.Check{Name: name, Command: command})
	}

	results, err := verify.RunChecks(context.Background(), dir, checks)
	if err != nil {
		return err
	}
	if err := printResults(stdout, jsonOut, results); err != nil {
		return err
	}

	for _, r := range results {
		if !r.Passed {
			return errors.New(
				fmt.Sprintf("check %q failed", r.Name),
				fmt.Sprintf("the command exited with code %d", r.ExitCode),
				"fix the failing check, then run agentrun-dev verify again",
				"https://agentrun.dev/docs/troubleshooting#commands",
			)
		}
	}
	return nil
}

func printResults(stdout io.Writer, jsonOut bool, results []verify.Result) error {
	if jsonOut {
		out := make([]checkJSON, 0, len(results))
		for _, r := range results {
			out = append(out, checkJSON{Name: r.Name, Passed: r.Passed, ExitCode: r.ExitCode, Output: r.Output})
		}
		return json.NewEncoder(stdout).Encode(out)
	}

	for _, r := range results {
		if r.Passed {
			fmt.Fprintf(stdout, "ok    %s\n", r.Name)
			continue
		}
		fmt.Fprintf(stdout, "FAIL  %s (exit %d)\n%s", r.Name, r.ExitCode, r.Output)
	}
	return nil
}
