package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	Dir            = ".config/qx"
	File           = "config.yaml"
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "gpt-4o-mini"
	DefaultCount   = 5
)

// Config represents the application configuration
type Config struct {
	LLM LLMConfig `mapstructure:"llm"`
}

// LLMConfig contains LLM-related configuration
type LLMConfig struct {
	BaseURL  string `mapstructure:"base_url"` // default: https://api.openai.com/v1
	Model    string `mapstructure:"model"`    // default: gpt-4o-mini
	Count    int    `mapstructure:"count"`    // default: 5 (number of command variants)
	Provider string `mapstructure:"provider"` // optional: "openai" or "eliza" (auto-detected if not set)
	APIKey   string `mapstructure:"-"`        // from environment variable only
}

// configPath returns the full path to the config file
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, Dir, File), nil
}

// Load reads configuration from ~/.config/qx/config.yaml and environment variables
func Load() (*Config, error) {
	viper.SetDefault("llm.base_url", DefaultBaseURL)
	viper.SetDefault("llm.model", DefaultModel)
	viper.SetDefault("llm.count", DefaultCount)

	path, err := configPath()
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if readErr := viper.ReadInConfig(); readErr != nil {
		if !os.IsNotExist(readErr) {
			return nil, fmt.Errorf("failed to read config file: %w", readErr)
		}
	}

	var cfg Config
	if unmarshalErr := viper.Unmarshal(&cfg); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", unmarshalErr)
	}

	if cfg.LLM.Model == "" {
		return nil, fmt.Errorf("llm.model is required")
	}
	if cfg.LLM.Count < 1 {
		return nil, fmt.Errorf("llm.count must be at least 1, got %d", cfg.LLM.Count)
	}

	cfg.LLM.APIKey = os.Getenv("OPENAI_API_KEY")
	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	return &cfg, nil
}

// Path returns the path to the config file
func Path() string {
	path, err := configPath()
	if err != nil {
		return filepath.Join("~", Dir, File)
	}
	return path
}
