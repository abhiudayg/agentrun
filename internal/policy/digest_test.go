package policy_test

import (
	"strings"
	"testing"

	"agentrun/internal/policy"
)

func TestDigestMatchesKnownSHA256Vector(t *testing.T) {
	const emptyInput = "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

	if got := policy.Digest(nil); got != emptyInput {
		t.Errorf("Digest(empty) = %q, want %q", got, emptyInput)
	}
}

func TestDigestIsPrefixedSHA256Hex(t *testing.T) {
	got := policy.Digest([]byte("limits: {max_concurrent: 3}"))

	hexPart, ok := strings.CutPrefix(got, "sha256:")
	if !ok {
		t.Fatalf("Digest() = %q, missing sha256: prefix", got)
	}
	if len(hexPart) != 64 {
		t.Errorf("Digest() hex length = %d, want 64", len(hexPart))
	}
}

func TestDigestIsStableForIdenticalContent(t *testing.T) {
	first := policy.Digest([]byte("network: none"))
	second := policy.Digest([]byte("network: none"))

	if first != second {
		t.Errorf("Digest changed for identical input: %q vs %q", first, second)
	}
}

func TestDigestChangesWhenAnyByteChanges(t *testing.T) {
	base := policy.Digest([]byte("max_concurrent: 3"))
	changed := policy.Digest([]byte("max_concurrent: 4"))

	if base == changed {
		t.Error("Digest did not change when policy content changed")
	}
}
