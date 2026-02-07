package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evgfitil/qx/internal/llm"
)

// Run starts the interactive TUI and returns the result of user interaction.
// Returns SelectedResult if user chose a command, CancelledResult if cancelled.
// initialQuery is optional and pre-fills the input field.
// forceSend bypasses secret detection if true.
// pipeContext is optional stdin content to use as context for LLM generation.
func Run(cfg llm.Config, initialQuery string, forceSend bool, pipeContext string) (Result, error) {
	m := NewModel(cfg, initialQuery, forceSend, pipeContext)

	// Open /dev/tty for TUI output so it works even when stdout is redirected
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		tty = os.Stdout
	} else {
		defer tty.Close() //nolint:errcheck
	}

	p := tea.NewProgram(m, tea.WithOutput(tty), tea.WithInputTTY())

	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := result.(Model); ok {
		return m.Result(), nil
	}

	return CancelledResult{Query: initialQuery}, nil
}
