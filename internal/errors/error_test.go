package errors_test

import (
	"strings"
	"testing"

	"agentrun/internal/errors"
)

func TestStandardErrorFormatting(t *testing.T) {
	// Arrange
	err := errors.New(
		"cannot start run: Docker daemon is not reachable",
		"the Docker socket at /var/run/docker.sock does not exist or is not accessible",
		"start Docker Desktop, or add your user to the docker group, then run `agentrun doctor`",
		"https://agentrun.dev/docs/troubleshooting#docker",
	)

	// Act
	formatted := err.Error()

	// Assert
	expectedLines := []string{
		"error: cannot start run: Docker daemon is not reachable",
		"  cause: the Docker socket at /var/run/docker.sock does not exist or is not accessible",
		"  fix:   start Docker Desktop, or add your user to the docker group, then run `agentrun doctor`",
		"  docs:  https://agentrun.dev/docs/troubleshooting#docker",
	}

	actualLines := strings.Split(strings.TrimSpace(formatted), "\n")
	if len(actualLines) != len(expectedLines) {
		t.Fatalf("expected %d lines, got %d: %q", len(expectedLines), len(actualLines), formatted)
	}

	for i, expected := range expectedLines {
		if actualLines[i] != expected {
			t.Errorf("line %d mismatch: got %q, want %q", i, actualLines[i], expected)
		}
	}
}

func TestCatalogErrorValidation(t *testing.T) {
	// Arrange: each error in the catalog must have all 4 non-empty fields
	catalog := errors.Catalog()

	// Act & Assert
	if len(catalog) == 0 {
		t.Fatal("error catalog should not be empty")
	}

	for code, item := range catalog {
		if item.What == "" {
			t.Errorf("catalog error %s missing 'what'", code)
		}
		if item.Cause == "" {
			t.Errorf("catalog error %s missing 'cause'", code)
		}
		if item.Fix == "" {
			t.Errorf("catalog error %s missing 'fix'", code)
		}
		if item.Docs == "" {
			t.Errorf("catalog error %s missing 'docs'", code)
		}
	}
}

func TestWithDetailsImmutability(t *testing.T) {
	// Arrange
	original := errors.New("base error", "base cause", "base fix", "base docs")

	// Act
	customized := original.WithCause("updated cause")

	// Assert
	if original.Cause != "base cause" {
		t.Errorf("original error was mutated: got %q, want 'base cause'", original.Cause)
	}
	if customized.Cause != "updated cause" {
		t.Errorf("customized error did not update cause: got %q", customized.Cause)
	}
	if customized.What != original.What || customized.Fix != original.Fix || customized.Docs != original.Docs {
		t.Errorf("other fields should remain identical")
	}
}
