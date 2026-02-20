package action

import (
	"errors"
	"strings"
	"testing"
)

func TestReadRefinementFrom_NormalInput(t *testing.T) {
	r := strings.NewReader("make it recursive\n")
	got, err := readRefinementFrom(r)
	if err != nil {
		t.Fatalf("readRefinementFrom() returned error: %v", err)
	}
	if got != "make it recursive" {
		t.Errorf("readRefinementFrom() = %q, want %q", got, "make it recursive")
	}
}

func TestReadRefinementFrom_TrimWhitespace(t *testing.T) {
	r := strings.NewReader("  add -v flag  \n")
	got, err := readRefinementFrom(r)
	if err != nil {
		t.Fatalf("readRefinementFrom() returned error: %v", err)
	}
	if got != "add -v flag" {
		t.Errorf("readRefinementFrom() = %q, want %q", got, "add -v flag")
	}
}

func TestReadRefinementFrom_EmptyInput(t *testing.T) {
	r := strings.NewReader("\n")
	_, err := readRefinementFrom(r)
	if !errors.Is(err, ErrEmptyRefinement) {
		t.Errorf("readRefinementFrom(empty) = %v, want ErrEmptyRefinement", err)
	}
}

func TestReadRefinementFrom_WhitespaceOnly(t *testing.T) {
	r := strings.NewReader("   \n")
	_, err := readRefinementFrom(r)
	if !errors.Is(err, ErrEmptyRefinement) {
		t.Errorf("readRefinementFrom(whitespace) = %v, want ErrEmptyRefinement", err)
	}
}

func TestReadRefinementFrom_EOF(t *testing.T) {
	r := strings.NewReader("")
	_, err := readRefinementFrom(r)
	if !errors.Is(err, ErrEmptyRefinement) {
		t.Errorf("readRefinementFrom(EOF) = %v, want ErrEmptyRefinement", err)
	}
}
