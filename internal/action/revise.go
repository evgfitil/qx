package action

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// ErrEmptyRefinement indicates the user submitted an empty refinement query.
var ErrEmptyRefinement = errors.New("empty refinement query")

// ReadRefinement opens /dev/tty, sets raw mode, and uses term.NewTerminal
// to read one line with built-in echo. This avoids relying on the terminal
// driver's ECHO flag which may be left disabled after bubbletea exits.
func ReadRefinement() (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer func() { _ = tty.Close() }()

	oldState, err := term.MakeRaw(int(tty.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(tty.Fd()), oldState) }()

	// Newline before prompt to separate from the cleared action menu.
	_, _ = fmt.Fprint(tty, "\r\n")

	t := term.NewTerminal(tty, "  > ")
	line, err := t.ReadLine()
	if err != nil {
		return "", fmt.Errorf("failed to read refinement: %w", err)
	}

	// Erase the refinement prompt from the screen.
	// After ReadLine the cursor is 2 lines below the leading \r\n:
	// line N (blank), line N+1 (prompt+input), line N+2 (cursor).
	// Move up 2, go to column 0, clear to end of screen.
	_, _ = fmt.Fprint(tty, "\033[2A\r\033[J")

	return readRefinementLine(line)
}

// readRefinementFrom reads one line from the provided reader and returns
// the trimmed input. Returns ErrEmptyRefinement if the input is blank.
// Used in tests where /dev/tty is not available.
func readRefinementFrom(r io.Reader) (string, error) {
	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil || n == 0 {
		return "", ErrEmptyRefinement
	}
	return readRefinementLine(string(buf[:n]))
}

// readRefinementLine trims and validates a refinement input line.
func readRefinementLine(line string) (string, error) {
	text := strings.TrimSpace(line)
	if text == "" {
		return "", ErrEmptyRefinement
	}
	return text, nil
}
