package mode

import (
	"fmt"

	"agentrun/internal/errors"
)

type Mode string

const (
	Lite Mode = "lite"
	Full Mode = "full"
)

func Parse(s string) (Mode, error) {
	switch m := Mode(s); m {
	case Lite, Full:
		return m, nil
	default:
		return "", errors.Catalog()[errors.ErrInvalidMode].WithWhat(fmt.Sprintf("invalid mode %q", s))
	}
}

func Resolve(flag, organization string) (Mode, error) {
	effective := Lite
	for _, requested := range []string{flag, organization} {
		if requested == "" {
			continue
		}
		m, err := Parse(requested)
		if err != nil {
			return "", err
		}
		if m == Full {
			effective = Full
		}
	}
	return effective, nil
}

func RequireFull(m Mode, command string) error {
	if m == Full {
		return nil
	}
	return errors.Catalog()[errors.ErrFullModeRequired].WithWhat(command + " requires Full mode")
}
