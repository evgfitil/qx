package shell

import (
	_ "embed"
	"fmt"
)

//go:embed scripts/bash.sh
var bashScript string

//go:embed scripts/zsh.zsh
var zshScript string

// Script returns the shell integration script for the specified shell.
func Script(shell string) (string, error) {
	switch shell {
	case "bash":
		return bashScript, nil
	case "zsh":
		return zshScript, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (supported: bash, zsh)", shell)
	}
}
