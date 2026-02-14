package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/evgfitil/qx/internal/guard"
)

func TestErrCancelled_CanBeExtracted(t *testing.T) {
	wrapped := fmt.Errorf("run failed: %w", ErrCancelled)

	if !errors.Is(wrapped, ErrCancelled) {
		t.Fatal("expected errors.Is to find ErrCancelled in wrapped error")
	}
}

func TestGenerateCommands_QueryWithSecrets(t *testing.T) {
	origForceSend := forceSend
	defer func() { forceSend = origForceSend }()
	forceSend = false

	err := generateCommands("use key AKIAIOSFODNN7EXAMPLE", "some safe context")
	if err == nil {
		t.Fatal("expected error for query with secrets")
	}

	var secretsErr *guard.SecretsError
	if !errors.As(err, &secretsErr) {
		t.Fatalf("expected SecretsError, got %T: %v", err, err)
	}
}

func TestGenerateCommands_EmptyPipeContextSkipsGuard(t *testing.T) {
	origForceSend := forceSend
	defer func() { forceSend = origForceSend }()
	forceSend = false

	// Point config to a nonexistent directory so config.Load() always fails,
	// regardless of the developer's local environment.
	t.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")

	// With empty pipe context, only the query is checked.
	// Should pass guard check and fail later at config.Load().
	err := generateCommands("list files", "")

	var secretsErr *guard.SecretsError
	if errors.As(err, &secretsErr) {
		t.Fatal("empty pipe context should not trigger secrets error")
	}
	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}
}

func TestHandleSelectedCommand_NonTTY_PrintsToStdout(t *testing.T) {
	// When stdout is a pipe (non-TTY), handleSelectedCommand should print
	// the command to stdout without showing the action menu.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	handleErr := handleSelectedCommand("echo hello")
	_ = w.Close()

	if handleErr != nil {
		t.Errorf("handleSelectedCommand returned error: %v", handleErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if string(out) != "echo hello\n" {
		t.Errorf("handleSelectedCommand output = %q, want %q", string(out), "echo hello\n")
	}
}

func TestRunInteractive_MultilineQueryDoesNotPanic(t *testing.T) {
	// Smoke test: runInteractive with multiline query (line continuations)
	// should not panic. Config error is expected in test environment.
	t.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")

	multilineQuery := "ps aux \\\n\t| grep nginx \\\n\t| sort"

	err := runInteractive(multilineQuery, "")
	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}
}

func TestRunInteractive_SimpleQueryDoesNotPanic(t *testing.T) {
	// Smoke test: runInteractive with a simple query should not panic.
	t.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")

	err := runInteractive("list all running containers", "")
	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}
}

func TestHandleSelectedCommand_NonTTY_EmptyCommand(t *testing.T) {
	// Even with empty command, non-TTY path should print and return nil.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	handleErr := handleSelectedCommand("")
	_ = w.Close()

	if handleErr != nil {
		t.Errorf("handleSelectedCommand returned error: %v", handleErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if string(out) != "\n" {
		t.Errorf("handleSelectedCommand output = %q, want %q", string(out), "\n")
	}
}
