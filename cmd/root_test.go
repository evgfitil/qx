package cmd

import (
	"errors"
	"fmt"
	"testing"
)

func TestCancelledError_Error(t *testing.T) {
	err := &CancelledError{}
	if err.Error() != "operation cancelled" {
		t.Errorf("expected 'operation cancelled', got %q", err.Error())
	}
}

func TestCancelledError_CanBeExtracted(t *testing.T) {
	original := &CancelledError{}
	wrapped := fmt.Errorf("run failed: %w", original)

	var cancelErr *CancelledError
	if !errors.As(wrapped, &cancelErr) {
		t.Fatal("expected errors.As to find CancelledError in wrapped error")
	}
}

func TestExitCodeCancelled(t *testing.T) {
	if ExitCodeCancelled != 130 {
		t.Errorf("expected ExitCodeCancelled to be 130, got %d", ExitCodeCancelled)
	}
}
