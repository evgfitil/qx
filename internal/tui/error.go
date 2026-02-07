package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// errorModel is a minimal bubbletea model that displays a startup error
// and waits for any key press to dismiss.
type errorModel struct {
	err          error
	initialQuery string
}

func (m errorModel) Init() tea.Cmd {
	return nil
}

func (m errorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, tea.Quit
	}
	return m, nil
}

func (m errorModel) View() string {
	return errorStyle().Render(fmt.Sprintf("Error: %v", m.err)) +
		"\n" +
		counterStyle().Render("Press any key to dismiss") +
		"\n"
}

// ShowError displays a startup error inside a TUI and waits for the user to dismiss it.
// Returns CancelledResult with the initial query so the shell can restore the prompt.
func ShowError(err error, initialQuery string) (Result, error) {
	m := errorModel{err: err, initialQuery: initialQuery}

	tty, ttyErr := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if ttyErr != nil {
		tty = os.Stdout
	} else {
		defer tty.Close() //nolint:errcheck
	}

	p := tea.NewProgram(m, tea.WithOutput(tty), tea.WithInputTTY())

	if _, runErr := p.Run(); runErr != nil {
		return nil, runErr
	}

	return CancelledResult{Query: initialQuery}, nil
}
