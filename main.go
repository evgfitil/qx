package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/evgfitil/qx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, cmd.ErrCancelled) {
			os.Exit(cmd.ExitCodeCancelled)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
