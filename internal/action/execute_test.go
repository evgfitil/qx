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
	t.Setenv("SHELL", "")
	os.Unsetenv("SHELL") //nolint:errcheck // always succeeds for valid key

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

	// Redirect stdout to capture output.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	oldStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = oldStdout })

	// Redirect stderr to pipe to prevent shell integration detection
	// (stdout=pipe + stderr=TTY would route output to /dev/tty).
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}
	defer func() { _ = stderrR.Close() }()

	origStderr := os.Stderr
	os.Stderr = stderrW
	t.Cleanup(func() { os.Stderr = origStderr })

	execErr := Execute("echo hello")
	_ = w.Close()
	_ = stderrW.Close()

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

func TestExecute_ShellIntegration_NoOutputOnStdoutPipe(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	// Open /dev/tty to simulate shell integration mode (stderr=TTY).
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		t.Skip("skipping: /dev/tty not available")
	}
	defer func() { _ = tty.Close() }()

	// Set stderr to /dev/tty (TTY) and stdout to a pipe.
	origStderr := os.Stderr
	os.Stderr = tty
	t.Cleanup(func() { os.Stderr = origStderr })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	execErr := Execute("echo hello")
	_ = w.Close()

	if execErr != nil {
		t.Fatalf("Execute returned error: %v", execErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()
	if len(out) != 0 {
		t.Errorf("expected no output on stdout pipe in shell integration mode, got %q", string(out))
	}
}

func TestExecute_ShellIntegration_FailingCommand(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		t.Skip("skipping: /dev/tty not available")
	}
	defer func() { _ = tty.Close() }()

	origStderr := os.Stderr
	os.Stderr = tty
	t.Cleanup(func() { os.Stderr = origStderr })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	execErr := Execute("echo fail >&2; exit 1")
	_ = w.Close()

	if execErr == nil {
		t.Fatal("expected error from failing command, got nil")
	}

	var exitErr *ExitError
	if !errors.As(execErr, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", execErr, execErr)
	}
	if exitErr.Code != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.Code)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()
	if len(out) != 0 {
		t.Errorf("expected no output on stdout pipe for failing command in shell integration mode, got %q", string(out))
	}
}

func TestExecute_NormalMode_OutputOnStdout(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	// Both stdout and stderr are pipes — not shell integration mode.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}
	defer func() { _ = stderrR.Close() }()

	origStderr := os.Stderr
	os.Stderr = stderrW
	t.Cleanup(func() { os.Stderr = origStderr })

	execErr := Execute("echo hello")
	_ = w.Close()
	_ = stderrW.Close()

	if execErr != nil {
		t.Fatalf("Execute returned error: %v", execErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()
	if string(out) != "hello\n" {
		t.Errorf("expected output on stdout in normal mode, got %q", string(out))
	}
}
