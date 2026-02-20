package action

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ErrEmptyRefinement indicates the user submitted an empty refinement query.
var ErrEmptyRefinement = errors.New("empty refinement query")

// ReadRefinement opens /dev/tty, restores cooked mode if needed,
// prints a "> " prompt to stderr, and reads one line of input.
func ReadRefinement() (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	// Terminal may still be in raw mode from the action menu keypress reader.
	// Restore cooked mode so line editing (backspace, etc.) works normally.
	oldState, err := term.GetState(int(tty.Fd()))
	if err == nil {
		_ = term.Restore(int(tty.Fd()), oldState)
		defer func() { _ = term.Restore(int(tty.Fd()), oldState) }()
	}

	fmt.Fprint(os.Stderr, "\n  > ")

	return readRefinementFrom(tty)
}

// readRefinementFrom reads one line from the provided reader and returns
// the trimmed input. Returns ErrEmptyRefinement if the input is blank.
func readRefinementFrom(r io.Reader) (string, error) {
	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read refinement: %w", err)
		}
		return "", ErrEmptyRefinement
	}

	text := strings.TrimSpace(scanner.Text())
	if text == "" {
		return "", ErrEmptyRefinement
	}

	return text, nil
}
