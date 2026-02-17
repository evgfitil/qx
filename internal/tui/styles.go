package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// renderer holds the lipgloss renderer. The tty file descriptor is intentionally
// not closed - it must remain open for the renderer to work throughout the application lifetime.
var renderer *lipgloss.Renderer

func getRenderer() *lipgloss.Renderer {
	if renderer != nil {
		return renderer
	}

	renderer = lipgloss.DefaultRenderer()
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		renderer = lipgloss.NewRenderer(tty)
	}
	return renderer
}

func promptStyle() lipgloss.Style {
	return getRenderer().NewStyle().Foreground(lipgloss.Color("205"))
}

func selectedStyle() lipgloss.Style {
	return getRenderer().NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
}

func normalStyle() lipgloss.Style {
	return getRenderer().NewStyle().Foreground(lipgloss.Color("252"))
}

func loadingStyle() lipgloss.Style {
	return getRenderer().NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
}

func counterStyle() lipgloss.Style {
	return getRenderer().NewStyle().Foreground(lipgloss.Color("241"))
}

func errorStyle() lipgloss.Style {
	return getRenderer().NewStyle().Foreground(lipgloss.Color("196"))
}
