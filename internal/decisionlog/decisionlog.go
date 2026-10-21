package decisionlog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

var ErrEmptyDecision = errors.New("decision is empty")

type Entry struct {
	Time     time.Time `json:"time"`
	Decision string    `json:"decision"`
	Reason   string    `json:"reason"`
	Evidence string    `json:"evidence"`
	Outcome  string    `json:"outcome"`
}

func Append(path string, e Entry) error {
	if strings.TrimSpace(e.Decision) == "" {
		return ErrEmptyDecision
	}
	e.Time = time.Now().UTC()

	line, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal decision entry: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open decision log: %w", err)
	}
	_, err = f.Write(append(line, '\n'))
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("write decision log: %w", err)
	}
	return nil
}
