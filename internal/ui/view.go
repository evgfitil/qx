package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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
		b.WriteString(m.viewSelector())

	case stateDone:
		// TODO: Task 8
	}

	return b.String()
}

func (m Model) viewSelector() string {
	var content strings.Builder

	content.WriteString(m.textArea.View())
	content.WriteString("\n")

	visible := m.visibleItemCount()
	end := m.scrollOffset + visible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	pointerWidth := lipgloss.Width(m.theme.Pointer)
	padding := strings.Repeat(" ", pointerWidth)

	for i := m.scrollOffset; i < end; i++ {
		displayText := m.getDisplayText(i)
		if i == m.cursor {
			content.WriteString(m.theme.Pointer + " " + m.theme.SelectedStyle().Render(displayText))
		} else {
			content.WriteString(padding + " " + m.theme.NormalStyle().Render(displayText))
		}
		content.WriteString("\n")
	}

	total := len(m.commands)
	if m.selectorMode {
		total = len(m.items)
	}
	content.WriteString(m.theme.MutedStyle().Render(fmt.Sprintf("%d/%d", len(m.filtered), total)))

	borderStyle := m.theme.BorderStyle()
	if m.width > 0 {
		borderStyle = borderStyle.Width(m.width - 2)
	}

	return borderStyle.Render(content.String()) + "\n"
}
