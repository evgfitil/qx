package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

const maxStdinSize = 64 * 1024 // 64KB

// ErrStdinTooLarge indicates piped input exceeds the size limit.
var ErrStdinTooLarge = errors.New("stdin input too large (max 64KB)")

// readStdin detects piped input and reads up to 64KB.
// Returns empty string if stdin is a TTY (no pipe).
func readStdin() (string, error) {
	return readFromReader(os.Stdin)
}

// readFromReader reads piped input from the given file descriptor.
// Extracted for testability - readStdin calls this with os.Stdin.
func readFromReader(f *os.File) (string, error) {
	if isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd()) {
		return "", nil
	}

	limited := io.LimitReader(f, int64(maxStdinSize)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}

	if len(data) > maxStdinSize {
		return "", ErrStdinTooLarge
	}

	return strings.TrimSpace(string(data)), nil
}
