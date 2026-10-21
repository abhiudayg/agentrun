package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"agentrun/internal/errors"
)

var requiredTools = []string{"git", "go", "sh"}

func doctor(stdout io.Writer, _ bool, _ []string) error {
	var missing []string
	for _, tool := range requiredTools {
		if _, err := exec.LookPath(tool); err != nil {
			fmt.Fprintf(stdout, "%s: missing\n", tool)
			missing = append(missing, tool)
			continue
		}
		fmt.Fprintf(stdout, "%s: ok\n", tool)
	}
	if len(missing) == 0 {
		return nil
	}
	names := strings.Join(missing, ", ")
	return errors.New(
		"missing required tools: "+names,
		"not found by exec.LookPath: "+names,
		"install the missing tools and put them on PATH",
		"https://agentrun.dev/docs/troubleshooting#commands",
	)
}
