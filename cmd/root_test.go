package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/evgfitil/qx/internal/guard"
)

func TestErrCancelled_CanBeExtracted(t *testing.T) {
	wrapped := fmt.Errorf("run failed: %w", ErrCancelled)

	if !errors.Is(wrapped, ErrCancelled) {
		t.Fatal("expected errors.Is to find ErrCancelled in wrapped error")
	}
}

func TestGenerateCommands_PipeContextWithSecrets(t *testing.T) {
	origForceSend := forceSend
	defer func() { forceSend = origForceSend }()
	forceSend = false

	pipeContext := "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"
	err := generateCommands("list files", pipeContext)
	if err == nil {
		t.Fatal("expected error for pipe context with secrets")
	}

	var secretsErr *guard.SecretsError
	if !errors.As(err, &secretsErr) {
		t.Fatalf("expected SecretsError, got %T: %v", err, err)
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

func TestGenerateCommands_PipeContextSecretsForceSendBypass(t *testing.T) {
	origForceSend := forceSend
	defer func() { forceSend = origForceSend }()
	forceSend = true

	// With forceSend=true, secrets in pipe context should be bypassed.
	// Should pass guard check and fail later at config.Load().
	err := generateCommands("list files", "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE")

	var secretsErr *guard.SecretsError
	if errors.As(err, &secretsErr) {
		t.Fatal("forceSend=true should bypass secrets detection in pipe context")
	}
	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}
}
