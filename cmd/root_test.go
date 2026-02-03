package cmd

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrCancelled_CanBeExtracted(t *testing.T) {
	wrapped := fmt.Errorf("run failed: %w", ErrCancelled)

	if !errors.Is(wrapped, ErrCancelled) {
		t.Fatal("expected errors.Is to find ErrCancelled in wrapped error")
	}
}

func TestDescribeFlagRegistered(t *testing.T) {
	flag := rootCmd.Flags().Lookup("describe")
	if flag == nil {
		t.Fatal("--describe flag not registered")
	}
	if flag.Shorthand != "d" {
		t.Errorf("--describe flag shorthand = %q, want 'd'", flag.Shorthand)
	}
}

func TestDescribeModeRequiresArgument(t *testing.T) {
	oldDescribeMode := describeMode
	defer func() { describeMode = oldDescribeMode }()

	describeMode = true
	err := run(rootCmd, []string{})
	if err == nil {
		t.Fatal("expected error when describe mode has no argument")
	}
	if err.Error() != "describe mode requires a command argument" {
		t.Errorf("unexpected error: %v", err)
	}
}
