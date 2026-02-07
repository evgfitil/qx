package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/evgfitil/qx/cmd"
	"github.com/evgfitil/qx/internal/action"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, cmd.ErrCancelled) {
			os.Exit(cmd.ExitCodeCancelled)
		}
		var exitErr *action.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
