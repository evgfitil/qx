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

// PickIndex displays fzf-style picker for a slice of items with a custom
// display function and returns the selected index. Returns ErrAborted if
// the user cancels selection.
func PickIndex(n int, display func(i int) string) (int, error) {
	if n <= 0 {
		return -1, errors.New("no items to pick from")
	}

	items := make([]int, n)
	for i := range items {
		items[i] = i
	}

	idx, err := fuzzyfinder.Find(items, func(i int) string {
		return display(items[i])
	})
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return -1, ErrAborted
		}
		return -1, err
	}

	return items[idx], nil
}
