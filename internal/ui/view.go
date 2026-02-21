package ui

import (
	"fmt"
	"strings"
)

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	switch m.state {
	case stateInput:
		b.WriteString(m.textArea.View())
		b.WriteString("\n")
		if m.err != nil {
			b.WriteString(m.theme.MutedStyle().Render(fmt.Sprintf("Error: %v", m.err)))
			b.WriteString("\n")
		}

	case stateLoading:
		b.WriteString(m.spinner.View())
		b.WriteString(m.theme.MutedStyle().Render(" Generating commands..."))
		b.WriteString("\n")

	case stateSelect:
		// TODO: Task 7

	case stateDone:
		// TODO: Task 8
	}

	return b.String()
}
