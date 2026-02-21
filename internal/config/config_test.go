package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// setupConfigFile creates a temporary config directory structure and sets HOME.
// Returns a cleanup function that restores the original HOME.
func setupConfigFile(t *testing.T, content string) {
	t.Helper()

	dir := t.TempDir()
	configDir := filepath.Join(dir, Dir)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if content != "" {
		if err := os.WriteFile(filepath.Join(configDir, File), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("HOME", dir)
}

func TestLoadPreferences_MissingFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	prefs := LoadPreferences()

	if prefs.ActionMenu {
		t.Error("expected ActionMenu to be false when config file is missing")
	}
}

func TestLoadPreferences_WithoutActionMenu(t *testing.T) {
	setupConfigFile(t, `llm:
  base_url: "https://api.example.com/v1"
  model: "test-model"
`)

	prefs := LoadPreferences()

	if prefs.ActionMenu {
		t.Error("expected ActionMenu to be false when not specified in config")
	}
}

func TestLoadPreferences_ActionMenuTrue(t *testing.T) {
	setupConfigFile(t, `action_menu: true
llm:
  model: "test-model"
`)

	prefs := LoadPreferences()

	if !prefs.ActionMenu {
		t.Error("expected ActionMenu to be true")
	}
}

func TestLoadPreferences_ActionMenuFalse(t *testing.T) {
	setupConfigFile(t, `action_menu: false
llm:
  model: "test-model"
`)

	prefs := LoadPreferences()

	if prefs.ActionMenu {
		t.Error("expected ActionMenu to be false")
	}
}

func TestLoad_PopulatesActionMenu(t *testing.T) {
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })

	setupConfigFile(t, `action_menu: true
llm:
  base_url: "https://api.example.com/v1"
  model: "test-model"
  apikey: "test-key"
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if !cfg.ActionMenu {
		t.Error("expected ActionMenu to be true after Load()")
	}
}

func TestLoad_ActionMenuDefaultsFalse(t *testing.T) {
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })

	setupConfigFile(t, `llm:
  base_url: "https://api.example.com/v1"
  model: "test-model"
  apikey: "test-key"
`)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.ActionMenu {
		t.Error("expected ActionMenu to default to false")
	}
}
