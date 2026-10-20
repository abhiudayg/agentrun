package mode_test

import (
	"fmt"
	"strings"
	"testing"

	"agentrun/internal/mode"
)

func TestParseAcceptsLiteAndFull(t *testing.T) {
	cases := map[string]mode.Mode{
		"lite": mode.Lite,
		"full": mode.Full,
	}

	for input, want := range cases {
		got, err := mode.Parse(input)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Errorf("Parse(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseRejectsAnythingElseWithStandardError(t *testing.T) {
	inputs := []string{"", "Full", "FULL", " full", "full ", "lite\n", "enterprise"}

	for _, input := range inputs {
		t.Run(fmt.Sprintf("%q", input), func(t *testing.T) {
			_, err := mode.Parse(input)
			if err == nil {
				t.Fatalf("Parse(%q) accepted an invalid mode", input)
			}

			msg := err.Error()
			if !strings.HasPrefix(msg, "error: invalid mode") {
				t.Errorf("error does not start with the invalid mode message: %q", msg)
			}
			for _, field := range []string{"  cause: ", "  fix:   ", "  docs:  "} {
				if !strings.Contains(msg, field) {
					t.Errorf("error is missing %q: %q", field, msg)
				}
			}
		})
	}
}

func TestResolveDefaultsToLiteWhenNothingIsSet(t *testing.T) {
	got, err := mode.Resolve("", "")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got != mode.Lite {
		t.Errorf("Resolve(\"\", \"\") = %q, want %q", got, mode.Lite)
	}
}

func TestResolveSelectsFullWhenEitherSourceRequestsIt(t *testing.T) {
	cases := []struct {
		name         string
		flag         string
		organization string
	}{
		{name: "flag requests full", flag: "full"},
		{name: "organization requests full", organization: "full"},
		{name: "flag and organization both request full", flag: "full", organization: "full"},
		{name: "lite flag cannot disable organization full", flag: "lite", organization: "full"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := mode.Resolve(tc.flag, tc.organization)
			if err != nil {
				t.Fatalf("Resolve(%q, %q) returned error: %v", tc.flag, tc.organization, err)
			}
			if got != mode.Full {
				t.Errorf("Resolve(%q, %q) = %q, want %q", tc.flag, tc.organization, got, mode.Full)
			}
		})
	}
}

func TestResolveStaysLiteWhenOnlyLiteIsRequested(t *testing.T) {
	got, err := mode.Resolve("lite", "lite")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got != mode.Lite {
		t.Errorf("Resolve(\"lite\", \"lite\") = %q, want %q", got, mode.Lite)
	}
}

func TestResolveRejectsInvalidValueEvenWhenOtherSourceIsValid(t *testing.T) {
	cases := []struct {
		name         string
		flag         string
		organization string
	}{
		{name: "invalid flag", flag: "bogus"},
		{name: "invalid organization", organization: "bogus"},
		{name: "invalid organization hidden behind full flag", flag: "full", organization: "bogus"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := mode.Resolve(tc.flag, tc.organization); err == nil {
				t.Errorf("Resolve(%q, %q) accepted an invalid mode", tc.flag, tc.organization)
			}
		})
	}
}

func TestRequireFullAllowsFullMode(t *testing.T) {
	if err := mode.RequireFull(mode.Full, "host init"); err != nil {
		t.Errorf("RequireFull(full) returned error: %v", err)
	}
}

func TestRequireFullBlocksLiteAndNamesTheModeSwitch(t *testing.T) {
	err := mode.RequireFull(mode.Lite, "host init")
	if err == nil {
		t.Fatal("RequireFull(lite) allowed a Full-only command")
	}

	msg := err.Error()
	for _, want := range []string{
		"error: host init requires Full mode",
		"--mode full",
		"mode: full",
		"  docs:  ",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q does not contain %q", msg, want)
		}
	}
}

func TestRequireFullBlocksAnyValueThatIsNotFull(t *testing.T) {
	if err := mode.RequireFull(mode.Mode("enterprise"), "host init"); err == nil {
		t.Error("RequireFull allowed an unknown mode value")
	}
}
