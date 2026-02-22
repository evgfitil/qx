package tui

import "github.com/charmbracelet/lipgloss"

// Theme defines the visual appearance of the TUI.
type Theme struct {
	Prompt     string
	Pointer    string
	SelectedFg string
	MatchFg    string
	TextFg     string
	MutedFg    string
	Border     string
	BorderFg   string
	renderer   *lipgloss.Renderer
}

// WithRenderer returns a copy of the theme with the given renderer set.
// The renderer determines which output the styles render to, ensuring colors
// work correctly even when stdout is captured (e.g. shell integration mode).
func (t Theme) WithRenderer(r *lipgloss.Renderer) Theme {
	t.renderer = r
	return t
}

// newStyle creates a new lipgloss.Style using the theme's renderer if set,
// falling back to the default renderer otherwise.
func (t Theme) newStyle() lipgloss.Style {
	if t.renderer != nil {
		return t.renderer.NewStyle()
	}
	return lipgloss.NewStyle()
}

// DefaultTheme returns an fzf-like theme with sensible defaults.
func DefaultTheme() Theme {
	return Theme{
		Prompt:     "> ",
		Pointer:    "▌",
		SelectedFg: "170",
		MatchFg:    "205",
		TextFg:     "252",
		MutedFg:    "241",
		Border:     "rounded",
		BorderFg:   "240",
	}
}

// SelectedStyle returns the style for the currently selected item.
func (t Theme) SelectedStyle() lipgloss.Style {
	return t.newStyle().Foreground(lipgloss.Color(t.SelectedFg)).Bold(true)
}

// NormalStyle returns the style for unselected items.
func (t Theme) NormalStyle() lipgloss.Style {
	return t.newStyle().Foreground(lipgloss.Color(t.TextFg))
}

// MutedStyle returns the style for secondary text like counters and spinners.
func (t Theme) MutedStyle() lipgloss.Style {
	return t.newStyle().Foreground(lipgloss.Color(t.MutedFg))
}

// PromptStyle returns the style for the input prompt symbol.
func (t Theme) PromptStyle() lipgloss.Style {
	return t.newStyle().Foreground(lipgloss.Color(t.MatchFg))
}

// BorderStyle returns the lipgloss border style based on the theme's border type.
func (t Theme) BorderStyle() lipgloss.Style {
	return t.newStyle().
		Border(t.borderType()).
		BorderForeground(lipgloss.Color(t.BorderFg))
}

func (t Theme) borderType() lipgloss.Border {
	switch t.Border {
	case "normal":
		return lipgloss.NormalBorder()
	case "thick":
		return lipgloss.ThickBorder()
	case "hidden":
		return lipgloss.HiddenBorder()
	default:
		return lipgloss.RoundedBorder()
	}
}
