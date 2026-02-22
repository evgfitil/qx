package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

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

// saveTermState saves the current terminal state from /dev/tty and returns
// a function that restores it. This guards against bubbletea leaving the
// terminal in raw mode (echo disabled) after exit, which breaks subsequent
// line-oriented input such as the revise prompt.
func saveTermState() func() {
	f, err := os.Open("/dev/tty")
	if err != nil {
		return func() {}
	}
	state, err := term.GetState(int(f.Fd()))
	if err != nil {
		_ = f.Close()
		return func() {}
	}
	return func() {
		_ = term.Restore(int(f.Fd()), state)
		_ = f.Close()
	}
}

// openTTY opens /dev/tty for writing and creates a lipgloss renderer from it.
// Falls back to os.Stdout with default renderer if /dev/tty is unavailable.
// The caller must close the returned file when tty != os.Stdout.
func openTTY(theme Theme) (*os.File, Theme) {
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return os.Stdout, theme.WithRenderer(lipgloss.DefaultRenderer())
	}
	return tty, theme.WithRenderer(lipgloss.NewRenderer(tty))
}

// Run starts the full TUI flow (input -> loading -> selecting -> done)
// and returns the result of user interaction.
func Run(opts RunOptions) (Result, error) {
	tty, theme := openTTY(opts.Theme)
	if tty != os.Stdout {
		defer tty.Close() //nolint:errcheck
	}
	opts.Theme = theme

	restore := saveTermState()
	m := newModel(opts)
	p := tea.NewProgram(m, tea.WithOutput(tty), tea.WithInputTTY())

	result, err := p.Run()
	restore()
	if err != nil {
		return nil, fmt.Errorf("TUI error: %w", err)
	}

	model, ok := result.(Model)
	if !ok {
		return nil, fmt.Errorf("unexpected model type: %T", result)
	}
	return model.Result(), nil
}

// RunSelector starts a selector-only TUI for picking from a list of items.
// Returns the selected index or -1 if cancelled.
func RunSelector(items []string, display func(int) string, theme Theme) (int, error) {
	tty, theme := openTTY(theme)
	if tty != os.Stdout {
		defer tty.Close() //nolint:errcheck
	}

	restore := saveTermState()
	m := newSelectorModel(items, display, theme)
	p := tea.NewProgram(m, tea.WithOutput(tty), tea.WithInputTTY())

	result, err := p.Run()
	restore()
	if err != nil {
		return -1, fmt.Errorf("selector error: %w", err)
	}

	model, ok := result.(Model)
	if !ok {
		return -1, fmt.Errorf("unexpected model type: %T", result)
	}
	return model.selectedIndex, nil
}
