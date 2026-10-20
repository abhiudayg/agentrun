package record

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"agentrun/internal/mode"
)

const digestPrefix = "sha256:"

type Provenance struct {
	Mode         mode.Mode `json:"mode"`
	Host         string    `json:"host"`
	PolicyDigest string    `json:"policy_digest"`
}

func NewProvenance(m mode.Mode, host, policyDigest string) (Provenance, error) {
	if _, err := mode.Parse(string(m)); err != nil {
		return Provenance{}, err
	}
	if host == "" {
		return Provenance{}, errors.New("provenance host is required")
	}
	if !validDigest(policyDigest) {
		return Provenance{}, errors.New("provenance policy digest must be sha256 followed by 64 hex characters")
	}
	return Provenance{Mode: m, Host: host, PolicyDigest: policyDigest}, nil
}

func validDigest(d string) bool {
	hexPart, ok := strings.CutPrefix(d, digestPrefix)
	if !ok {
		return false
	}
	b, err := hex.DecodeString(hexPart)
	return err == nil && len(b) == sha256.Size
}
