package cmd

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestReadFromReader_PipedInput(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	content := "total 24\ndrwxr-xr-x  5 user staff  160 Jan  1 12:00 dir1\n-rw-r--r--  1 user staff 1024 Jan  1 12:00 file.txt"
	go func() {
		_, _ = w.WriteString(content)
		_ = w.Close()
	}()

	got, err := readFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestReadFromReader_PipedEmptyInput(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	go func() {
		_ = w.Close()
	}()

	got, err := readFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "" {
		t.Errorf("expected empty string for empty pipe, got %q", got)
	}
}

func TestReadFromReader_PipedWhitespaceOnly(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	go func() {
		_, _ = w.WriteString("  \n\t\n  ")
		_ = w.Close()
	}()

	got, err := readFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "" {
		t.Errorf("expected empty string for whitespace-only pipe, got %q", got)
	}
}

func TestReadFromReader_TTYInput(t *testing.T) {
	f, err := os.Open("/dev/tty")
	if err != nil {
		t.Skip("cannot open /dev/tty (no TTY available)")
	}
	defer func() { _ = f.Close() }()

	got, err := readFromReader(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "" {
		t.Errorf("expected empty string for TTY, got %q", got)
	}
}

func TestReadFromReader_OversizedInput(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	go func() {
		data := strings.Repeat("x", maxStdinSize+1)
		_, _ = w.WriteString(data)
		_ = w.Close()
	}()

	_, err = readFromReader(r)
	if !errors.Is(err, ErrStdinTooLarge) {
		t.Errorf("expected ErrStdinTooLarge, got %v", err)
	}
}

func TestReadFromReader_ExactLimitInput(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	data := strings.Repeat("x", maxStdinSize)
	go func() {
		_, _ = w.WriteString(data)
		_ = w.Close()
	}()

	got, err := readFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != data {
		t.Errorf("expected data of length %d, got length %d", len(data), len(got))
	}
}
