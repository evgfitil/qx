package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evgfitil/qx/internal/llm"
)

// Run starts the interactive TUI and returns the selected command.
// Returns empty string if user cancelled (Esc/Ctrl+C).
// initialQuery is optional and pre-fills the input field.
func Run(cfg llm.Config, initialQuery string) (string, error) {
	m := NewModel(cfg, initialQuery)

	// Open /dev/tty for TUI output so it works even when stdout is redirected
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		// Fallback to stdout if /dev/tty is not available
		tty = os.Stdout
	} else {
		defer tty.Close()
	}

	p := tea.NewProgram(m, tea.WithOutput(tty))

	result, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := result.(Model); ok {
		return m.Selected(), nil
	}

	return "", nil
}
