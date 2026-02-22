package action

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
)

// ExitError wraps a subprocess exit code so callers can propagate it.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("exit status %d", e.Code)
}

// detectShell returns the user's shell from $SHELL env var,
// falling back to /bin/sh if unset.
func detectShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}

// Execute runs the given command string in the user's shell.
// In shell integration mode (stdout=pipe, stderr=TTY), output is routed to
// /dev/tty so it appears on the terminal instead of being captured by $().
// If stdin is a terminal it is passed through; otherwise /dev/tty is opened
// so the subprocess can receive interactive input.
func Execute(command string) error {
	shell := detectShell()
	cmd := exec.Command(shell, "-c", command)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if inShellIntegration() {
		if ttyOut, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
			defer func() { _ = ttyOut.Close() }()
			cmd.Stdout = ttyOut
			cmd.Stderr = ttyOut
		}
	}

	cmd.Stdin = os.Stdin
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		if tty, err := os.Open("/dev/tty"); err == nil {
			defer func() { _ = tty.Close() }()
			cmd.Stdin = tty
		}
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return &ExitError{Code: exitErr.ExitCode()}
		}
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}
