package cmd

import (
	"errors"
	"testing"
)

func TestCancelledError_Error(t *testing.T) {
	err := &CancelledError{Query: "test query", ExitCode: ExitCodeCancelled}
	if err.Error() != "operation cancelled" {
		t.Errorf("expected 'operation cancelled', got %q", err.Error())
	}
}

func TestCancelledError_ExitCode(t *testing.T) {
	err := &CancelledError{Query: "test query", ExitCode: ExitCodeCancelled}
	if err.ExitCode != 130 {
		t.Errorf("expected exit code 130, got %d", err.ExitCode)
	}
}

func TestCancelledError_CanBeExtracted(t *testing.T) {
	err := &CancelledError{Query: "my query", ExitCode: ExitCodeCancelled}
	var wrappedErr error = err

	var cancelErr *CancelledError
	if !errors.As(wrappedErr, &cancelErr) {
		t.Fatal("expected errors.As to find CancelledError")
	}
	if cancelErr.Query != "my query" {
		t.Errorf("expected query 'my query', got %q", cancelErr.Query)
	}
	if cancelErr.ExitCode != 130 {
		t.Errorf("expected exit code 130, got %d", cancelErr.ExitCode)
	}
}
