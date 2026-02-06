package action

import (
	"fmt"
	"os"
	"os/exec"
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
// with inherited stdin, stdout, and stderr.
func Execute(command string) error {
	shell := detectShell()
	cmd := exec.Command(shell, "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}
