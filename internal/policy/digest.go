package policy

import (
	"crypto/sha256"
	"encoding/hex"
)

func Digest(canonical []byte) string {
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:])
}
