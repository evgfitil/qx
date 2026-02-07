package action

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// CopyToClipboard copies the given command string to the system clipboard.
func CopyToClipboard(command string) error {
	if err := clipboard.WriteAll(command); err != nil {
		return fmt.Errorf("clipboard copy failed: %w", err)
	}
	return nil
}
