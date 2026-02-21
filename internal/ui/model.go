package ui

import tea "github.com/charmbracelet/bubbletea"

// Model represents the TUI state. Full implementation in Task 4.
type Model struct {
	selected      string
	originalQuery string
	selectedIndex int
}

func newModel(_ RunOptions) Model {
	return Model{}
}

func newSelectorModel(_ []string, _ func(int) string, _ Theme) Model {
	return Model{selectedIndex: -1}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return ""
}

// Result returns the outcome of TUI interaction.
func (m Model) Result() Result {
	if m.selected != "" {
		return SelectedResult{Command: m.selected, Query: m.originalQuery}
	}
	return CancelledResult{Query: m.originalQuery}
}
