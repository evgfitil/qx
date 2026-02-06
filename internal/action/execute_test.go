package action

import (
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
		t.Error("Execute(\"false\") expected error, got nil")
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

	buf := make([]byte, 64)
	n, _ := r.Read(buf)
	got := string(buf[:n])
	if got != "hello\n" {
		t.Errorf("Execute(\"echo hello\") output = %q, want %q", got, "hello\n")
	}
}

func TestExecute_InvalidShell(t *testing.T) {
	t.Setenv("SHELL", "/nonexistent/shell")

	err := Execute("echo test")
	if err == nil {
		t.Error("Execute with invalid shell expected error, got nil")
	}
}
