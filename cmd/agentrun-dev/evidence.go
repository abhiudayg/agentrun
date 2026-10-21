package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"agentrun/internal/errors"
)

type evidenceFile struct {
	Status   string            `json:"status"`
	Attempts []evidenceAttempt `json:"attempts"`
}

type evidenceAttempt struct {
	Number   int             `json:"number"`
	AgentErr string          `json:"agent_err"`
	Results  []evidenceCheck `json:"results"`
}

type evidenceCheck struct {
	Name     string `json:"Name"`
	Passed   bool   `json:"Passed"`
	ExitCode int    `json:"ExitCode"`
}

func evidence(stdout io.Writer, _ bool, args []string) error {
	if len(args) != 1 {
		return errors.New(
			"no run directory given",
			"evidence reads outcome.json from exactly one directory",
			"run agentrun-dev evidence <dir>",
			"https://agentrun.dev/docs/troubleshooting#evidence",
		)
	}
	dir := args[0]

	data, err := os.ReadFile(filepath.Join(dir, "outcome.json"))
	if os.IsNotExist(err) {
		return errors.New(
			fmt.Sprintf("no outcome.json in %q", dir),
			"the directory has no saved verify outcome",
			"run the verify loop in that directory, or pass the directory that holds outcome.json",
			"https://agentrun.dev/docs/troubleshooting#evidence",
		)
	}
	if err != nil {
		return fmt.Errorf("read outcome.json: %w", err)
	}

	var file evidenceFile
	if err := json.Unmarshal(data, &file); err != nil {
		return fmt.Errorf("decode outcome.json: %w", err)
	}
	_, err = io.WriteString(stdout, formatEvidence(file))
	return err
}

func formatEvidence(file evidenceFile) string {
	var b strings.Builder
	fmt.Fprintf(&b, "status: %s\n", file.Status)
	for _, a := range file.Attempts {
		fmt.Fprintf(&b, "attempt %d\n", a.Number)
		if a.AgentErr != "" {
			fmt.Fprintf(&b, "  agent: %s\n", a.AgentErr)
		}
		for _, c := range a.Results {
			verdict := "fail"
			if c.Passed {
				verdict = "pass"
			}
			fmt.Fprintf(&b, "  %s: %s (exit %d)\n", c.Name, verdict, c.ExitCode)
		}
	}
	return b.String()
}
