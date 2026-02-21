package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultThemeFieldValues(t *testing.T) {
	theme := DefaultTheme()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Prompt", theme.Prompt, "> "},
		{"Pointer", theme.Pointer, "▌"},
		{"SelectedFg", theme.SelectedFg, "170"},
		{"MatchFg", theme.MatchFg, "205"},
		{"TextFg", theme.TextFg, "252"},
		{"MutedFg", theme.MutedFg, "241"},
		{"Border", theme.Border, "rounded"},
		{"BorderFg", theme.BorderFg, "240"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("DefaultTheme().%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestSelectedStyleUsesThemeColor(t *testing.T) {
	theme := Theme{SelectedFg: "170"}
	style := theme.SelectedStyle()

	fg := style.GetForeground()
	if fg != lipgloss.Color("170") {
		t.Errorf("SelectedStyle foreground = %v, want Color(170)", fg)
	}
	if !style.GetBold() {
		t.Error("SelectedStyle should be bold")
	}
}

func TestNormalStyleUsesThemeColor(t *testing.T) {
	theme := Theme{TextFg: "252"}
	style := theme.NormalStyle()

	fg := style.GetForeground()
	if fg != lipgloss.Color("252") {
		t.Errorf("NormalStyle foreground = %v, want Color(252)", fg)
	}
}

func TestMutedStyleUsesThemeColor(t *testing.T) {
	theme := Theme{MutedFg: "241"}
	style := theme.MutedStyle()

	fg := style.GetForeground()
	if fg != lipgloss.Color("241") {
		t.Errorf("MutedStyle foreground = %v, want Color(241)", fg)
	}
}

func TestPromptStyleUsesMatchColor(t *testing.T) {
	theme := Theme{MatchFg: "205"}
	style := theme.PromptStyle()

	fg := style.GetForeground()
	if fg != lipgloss.Color("205") {
		t.Errorf("PromptStyle foreground = %v, want Color(205)", fg)
	}
}

func TestBorderStyleUsesThemeColor(t *testing.T) {
	theme := Theme{Border: "rounded", BorderFg: "240"}
	style := theme.BorderStyle()

	fg := style.GetBorderTopForeground()
	if fg != lipgloss.Color("240") {
		t.Errorf("BorderStyle border foreground = %v, want Color(240)", fg)
	}
}

func TestBorderTypeMapping(t *testing.T) {
	tests := []struct {
		name   string
		border string
		want   lipgloss.Border
	}{
		{"rounded", "rounded", lipgloss.RoundedBorder()},
		{"normal", "normal", lipgloss.NormalBorder()},
		{"thick", "thick", lipgloss.ThickBorder()},
		{"hidden", "hidden", lipgloss.HiddenBorder()},
		{"unknown defaults to rounded", "unknown", lipgloss.RoundedBorder()},
		{"empty defaults to rounded", "", lipgloss.RoundedBorder()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := Theme{Border: tt.border, BorderFg: "240"}
			got := theme.borderType()
			if got != tt.want {
				t.Errorf("borderType(%q) = %v, want %v", tt.border, got, tt.want)
			}
		})
	}
}

func TestCustomThemeColors(t *testing.T) {
	theme := Theme{
		SelectedFg: "#ff87d7",
		MatchFg:    "#00ff00",
		TextFg:     "#ffffff",
		MutedFg:    "#666666",
		BorderFg:   "#333333",
	}

	if fg := theme.SelectedStyle().GetForeground(); fg != lipgloss.Color("#ff87d7") {
		t.Errorf("SelectedStyle with hex color: got %v, want #ff87d7", fg)
	}
	if fg := theme.NormalStyle().GetForeground(); fg != lipgloss.Color("#ffffff") {
		t.Errorf("NormalStyle with hex color: got %v, want #ffffff", fg)
	}
	if fg := theme.MutedStyle().GetForeground(); fg != lipgloss.Color("#666666") {
		t.Errorf("MutedStyle with hex color: got %v, want #666666", fg)
	}
	if fg := theme.PromptStyle().GetForeground(); fg != lipgloss.Color("#00ff00") {
		t.Errorf("PromptStyle with hex color: got %v, want #00ff00", fg)
	}
}
