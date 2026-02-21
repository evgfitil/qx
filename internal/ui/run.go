package ui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evgfitil/qx/internal/llm"
)

// RunOptions configures the TUI run behavior.
type RunOptions struct {
	InitialQuery string
	LLMConfig    llm.Config
	ForceSend    bool
	PipeContext  string
	Theme        Theme
}

// Run starts the full TUI flow (input -> loading -> selecting -> done)
// and returns the result of user interaction.
func Run(opts RunOptions) (Result, error) {
	m := newModel(opts)

	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		tty = os.Stdout
	} else {
		defer tty.Close() //nolint:errcheck
	}

	p := tea.NewProgram(m, tea.WithOutput(tty), tea.WithInputTTY())

	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	if model, ok := result.(Model); ok {
		return model.Result(), nil
	}

	return CancelledResult{Query: opts.InitialQuery}, nil
}

// RunSelector starts a selector-only TUI for picking from a list of items.
// Returns the selected index or -1 if cancelled.
func RunSelector(items []string, display func(int) string, theme Theme) (int, error) {
	m := newSelectorModel(items, display, theme)

	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		tty = os.Stdout
	} else {
		defer tty.Close() //nolint:errcheck
	}

	p := tea.NewProgram(m, tea.WithOutput(tty), tea.WithInputTTY())

	result, err := p.Run()
	if err != nil {
		return -1, fmt.Errorf("selector error: %w", err)
	}

	if model, ok := result.(Model); ok {
		return model.selectedIndex, nil
	}

	return -1, nil
}
