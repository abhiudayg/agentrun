package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"agentrun/internal/errors"
)

type command func(stdout io.Writer, jsonOut bool, args []string) error

var commands = map[string]command{}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, commands))
}

func run(args []string, stdout, stderr io.Writer, cmds map[string]command) int {
	fs := flag.NewFlagSet("agentrun-dev", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "print results as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	rest := fs.Args()
	if len(rest) == 0 {
		fmt.Fprintln(stderr, errors.New(
			"no command given",
			"no subcommand name was passed after the global flags",
			"run agentrun-dev <command> [args]",
			"https://agentrun.dev/docs/troubleshooting#commands",
		))
		return 2
	}

	name := rest[0]
	cmd, ok := cmds[name]
	if !ok {
		fmt.Fprintln(stderr, errors.New(
			fmt.Sprintf("unknown command %q", name),
			fmt.Sprintf("no subcommand is named %q", name),
			"check the spelling of the command name",
			"https://agentrun.dev/docs/troubleshooting#commands",
		))
		return 2
	}

	if err := cmd(stdout, *jsonOut, rest[1:]); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
