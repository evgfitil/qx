package picker

import (
	"errors"

	"github.com/ktr0731/go-fuzzyfinder"
)

// ErrAborted indicates user cancelled selection
var ErrAborted = errors.New("selection aborted")

// Pick displays fzf-style picker for command selection.
func Pick(commands []string) (string, error) {
	if len(commands) == 0 {
		return "", errors.New("no commands to pick from")
	}

	if len(commands) == 1 {
		return commands[0], nil
	}

	idx, err := fuzzyfinder.Find(commands, func(i int) string {
		return commands[i]
	})
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return "", ErrAborted
		}
		return "", err
	}

	return commands[idx], nil
}
