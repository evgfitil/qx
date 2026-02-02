package main

import (
	"errors"
	"os"

	"github.com/evgfitil/qx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		var cancelErr *cmd.CancelledError
		if errors.As(err, &cancelErr) {
			os.Exit(cancelErr.ExitCode)
		}
		os.Exit(1)
	}
}
