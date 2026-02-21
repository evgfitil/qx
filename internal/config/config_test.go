package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"

	"github.com/evgfitil/qx/internal/ui"
)

// resetViper clears all viper state between tests to avoid cross-contamination.
func resetViper() {
	viper.Reset()
}

func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	cfgDir := filepath.Join(dir, ".config", "qx")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	cfgFile := filepath.Join(cfgDir, "config.yaml")
	if err := os.WriteFile(cfgFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return cfgFile
}

func TestLoadConfigWithTheme(t *testing.T) {
	resetViper()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("OPENAI_API_KEY", "test-key")

	cfgContent := `
llm:
  model: "gpt-4o-mini"
  count: 3
theme:
  prompt: "$ "
  pointer: ">"
  selected_fg: "196"
  match_fg: "46"
  text_fg: "255"
  muted_fg: "245"
  border: "thick"
  border_fg: "250"
action_menu: true
`
	writeConfig(t, tmpDir, cfgContent)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Theme.Prompt != "$ " {
		t.Errorf("Theme.Prompt = %q, want %q", cfg.Theme.Prompt, "$ ")
	}
	if cfg.Theme.Pointer != ">" {
		t.Errorf("Theme.Pointer = %q, want %q", cfg.Theme.Pointer, ">")
	}
	if cfg.Theme.SelectedFg != "196" {
		t.Errorf("Theme.SelectedFg = %q, want %q", cfg.Theme.SelectedFg, "196")
	}
	if cfg.Theme.MatchFg != "46" {
		t.Errorf("Theme.MatchFg = %q, want %q", cfg.Theme.MatchFg, "46")
	}
	if cfg.Theme.TextFg != "255" {
		t.Errorf("Theme.TextFg = %q, want %q", cfg.Theme.TextFg, "255")
	}
	if cfg.Theme.MutedFg != "245" {
		t.Errorf("Theme.MutedFg = %q, want %q", cfg.Theme.MutedFg, "245")
	}
	if cfg.Theme.Border != "thick" {
		t.Errorf("Theme.Border = %q, want %q", cfg.Theme.Border, "thick")
	}
	if cfg.Theme.BorderFg != "250" {
		t.Errorf("Theme.BorderFg = %q, want %q", cfg.Theme.BorderFg, "250")
	}
	if !cfg.ActionMenu {
		t.Error("ActionMenu = false, want true")
	}
}

func TestLoadConfigWithoutTheme(t *testing.T) {
	resetViper()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("OPENAI_API_KEY", "test-key")

	cfgContent := `
llm:
  model: "gpt-4o-mini"
  count: 5
`
	writeConfig(t, tmpDir, cfgContent)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	defaults := ui.DefaultTheme()
	if cfg.Theme.Prompt != defaults.Prompt {
		t.Errorf("Theme.Prompt = %q, want default %q", cfg.Theme.Prompt, defaults.Prompt)
	}
	if cfg.Theme.Pointer != defaults.Pointer {
		t.Errorf("Theme.Pointer = %q, want default %q", cfg.Theme.Pointer, defaults.Pointer)
	}
	if cfg.Theme.SelectedFg != defaults.SelectedFg {
		t.Errorf("Theme.SelectedFg = %q, want default %q", cfg.Theme.SelectedFg, defaults.SelectedFg)
	}
	if cfg.Theme.MatchFg != defaults.MatchFg {
		t.Errorf("Theme.MatchFg = %q, want default %q", cfg.Theme.MatchFg, defaults.MatchFg)
	}
	if cfg.Theme.TextFg != defaults.TextFg {
		t.Errorf("Theme.TextFg = %q, want default %q", cfg.Theme.TextFg, defaults.TextFg)
	}
	if cfg.Theme.MutedFg != defaults.MutedFg {
		t.Errorf("Theme.MutedFg = %q, want default %q", cfg.Theme.MutedFg, defaults.MutedFg)
	}
	if cfg.Theme.Border != defaults.Border {
		t.Errorf("Theme.Border = %q, want default %q", cfg.Theme.Border, defaults.Border)
	}
	if cfg.Theme.BorderFg != defaults.BorderFg {
		t.Errorf("Theme.BorderFg = %q, want default %q", cfg.Theme.BorderFg, defaults.BorderFg)
	}
	if cfg.ActionMenu {
		t.Error("ActionMenu = true, want false (default)")
	}
}

func TestLoadConfigActionMenu(t *testing.T) {
	resetViper()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("OPENAI_API_KEY", "test-key")

	cfgContent := `
llm:
  model: "gpt-4o-mini"
action_menu: true
`
	writeConfig(t, tmpDir, cfgContent)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if !cfg.ActionMenu {
		t.Error("ActionMenu = false, want true")
	}
}

func TestThemeConfigToTheme(t *testing.T) {
	tc := ThemeConfig{
		Prompt:     "$ ",
		Pointer:    ">",
		SelectedFg: "196",
		MatchFg:    "46",
		TextFg:     "255",
		MutedFg:    "245",
		Border:     "normal",
		BorderFg:   "250",
	}

	theme := tc.ToTheme()

	if theme.Prompt != tc.Prompt {
		t.Errorf("Prompt = %q, want %q", theme.Prompt, tc.Prompt)
	}
	if theme.Pointer != tc.Pointer {
		t.Errorf("Pointer = %q, want %q", theme.Pointer, tc.Pointer)
	}
	if theme.SelectedFg != tc.SelectedFg {
		t.Errorf("SelectedFg = %q, want %q", theme.SelectedFg, tc.SelectedFg)
	}
	if theme.MatchFg != tc.MatchFg {
		t.Errorf("MatchFg = %q, want %q", theme.MatchFg, tc.MatchFg)
	}
	if theme.TextFg != tc.TextFg {
		t.Errorf("TextFg = %q, want %q", theme.TextFg, tc.TextFg)
	}
	if theme.MutedFg != tc.MutedFg {
		t.Errorf("MutedFg = %q, want %q", theme.MutedFg, tc.MutedFg)
	}
	if theme.Border != tc.Border {
		t.Errorf("Border = %q, want %q", theme.Border, tc.Border)
	}
	if theme.BorderFg != tc.BorderFg {
		t.Errorf("BorderFg = %q, want %q", theme.BorderFg, tc.BorderFg)
	}
}
