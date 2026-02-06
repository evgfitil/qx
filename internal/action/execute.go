package action

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/mattn/go-isatty"
)

// detectShell returns the user's shell from $SHELL env var,
// falling back to /bin/sh if unset.
func detectShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/sh"
}

// Execute runs the given command string in the user's shell
// with inherited stdout and stderr. If stdin is a terminal it is passed through;
// otherwise (pipe already consumed) /dev/tty is opened so the subprocess can
// receive interactive input.
func Execute(command string) error {
	shell := detectShell()
	cmd := exec.Command(shell, "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		cmd.Stdin = os.Stdin
	} else {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			cmd.Stdin = os.Stdin
		} else {
			defer func() { _ = tty.Close() }()
			cmd.Stdin = tty
		}
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}
