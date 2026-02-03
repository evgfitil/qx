package stdin

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// MaxStdinSize is the maximum size of stdin content that will be read.
// Larger inputs are rejected to prevent memory exhaustion.
const MaxStdinSize = 1024 * 1024 // 1MB

// Reader provides methods for reading stdin content.
type Reader struct {
	input io.Reader
}

// New creates a new Reader with the provided input.
// Pass os.Stdin for normal operation.
func New(input io.Reader) *Reader {
	return &Reader{input: input}
}

// IsPiped returns true if stdin contains piped data (not a terminal).
func (r *Reader) IsPiped() bool {
	if f, ok := r.input.(*os.File); ok {
		return !term.IsTerminal(int(f.Fd()))
	}
	return true
}

// Read reads all content from stdin if it's piped.
// Returns empty string and nil error if stdin is a terminal.
// Returns error if stdin content exceeds MaxStdinSize.
func (r *Reader) Read() (string, error) {
	if !r.IsPiped() {
		return "", nil
	}

	limitedReader := io.LimitReader(r.input, MaxStdinSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	if len(data) > MaxStdinSize {
		return "", fmt.Errorf("stdin content exceeds maximum size of %d bytes", MaxStdinSize)
	}

	return string(data), nil
}

// ReadFromStdin is a convenience function that reads piped stdin content.
// Returns empty string if stdin is a terminal.
func ReadFromStdin() (string, error) {
	reader := New(os.Stdin)
	return reader.Read()
}
