package action

import (
	"errors"
	"io"
	"os"
	"testing"
)

func TestDetectShell_FromEnv(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")

	got := detectShell()
	if got != "/bin/zsh" {
		t.Errorf("detectShell() = %q, want %q", got, "/bin/zsh")
	}
}

func TestDetectShell_Fallback(t *testing.T) {
	t.Setenv("SHELL", "")

	got := detectShell()
	if got != "/bin/sh" {
		t.Errorf("detectShell() = %q, want %q", got, "/bin/sh")
	}
}

func TestDetectShell_CustomShell(t *testing.T) {
	t.Setenv("SHELL", "/usr/local/bin/fish")

	got := detectShell()
	if got != "/usr/local/bin/fish" {
		t.Errorf("detectShell() = %q, want %q", got, "/usr/local/bin/fish")
	}
}

func TestDetectShell_UnsetEnv(t *testing.T) {
	orig, existed := os.LookupEnv("SHELL")
	os.Unsetenv("SHELL")
	t.Cleanup(func() {
		if existed {
			os.Setenv("SHELL", orig)
		}
	})

	got := detectShell()
	if got != "/bin/sh" {
		t.Errorf("detectShell() = %q, want %q", got, "/bin/sh")
	}
}

func TestExecute_SimpleCommand(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	err := Execute("true")
	if err != nil {
		t.Errorf("Execute(\"true\") returned error: %v", err)
	}
}

func TestExecute_FailingCommand(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	err := Execute("false")
	if err == nil {
		t.Fatal("Execute(\"false\") expected error, got nil")
	}

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.Code)
	}
}

func TestExecute_ExitCode2(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	err := Execute("exit 2")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 2 {
		t.Errorf("expected exit code 2, got %d", exitErr.Code)
	}
}

func TestExecute_EchoCommand(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	// Redirect stdout to capture output
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = oldStdout })

	execErr := Execute("echo hello")
	_ = w.Close()

	if execErr != nil {
		t.Fatalf("Execute(\"echo hello\") returned error: %v", execErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()
	if string(out) != "hello\n" {
		t.Errorf("Execute(\"echo hello\") output = %q, want %q", string(out), "hello\n")
	}
}

func TestExecute_InvalidShell(t *testing.T) {
	t.Setenv("SHELL", "/nonexistent/shell")

	err := Execute("echo test")
	if err == nil {
		t.Error("Execute with invalid shell expected error, got nil")
	}
}
