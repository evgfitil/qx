package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"

	"github.com/evgfitil/qx/internal/llm"
)

const (
	Dir            = ".config/qx"
	File           = "config.yaml"
	DefaultBaseURL = "https://api.openai.com/v1"
	DefaultModel   = "gpt-4o-mini"
	DefaultCount   = 5
	DefaultTimeout = 60 * time.Second
)

// Config represents the application configuration
type Config struct {
	LLM LLMConfig `mapstructure:"llm"`
}

// LLMConfig contains LLM-related configuration
type LLMConfig struct {
	BaseURL   string `mapstructure:"base_url"`
	Model     string `mapstructure:"model"`
	Count     int    `mapstructure:"count"`
	Provider  string `mapstructure:"provider"`
	APIKey    string `mapstructure:"apikey"`
	ForceSend bool   `mapstructure:"-"`
}

// ToLLMConfig converts LLMConfig to llm.Config for provider creation
func (c LLMConfig) ToLLMConfig() llm.Config {
	return llm.Config{
		BaseURL:   c.BaseURL,
		APIKey:    c.APIKey,
		Model:     c.Model,
		Provider:  c.Provider,
		Count:     c.Count,
		ForceSend: c.ForceSend,
	}
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

	viper.MustBindEnv("llm.apikey", "OPENAI_API_KEY")

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
		return nil, fmt.Errorf("llm.model is required in %s", path)
	}
	if cfg.LLM.Count < 1 {
		return nil, fmt.Errorf("llm.count must be at least 1, got %d (in %s)", cfg.LLM.Count, path)
	}

	if cfg.LLM.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable or llm.apikey in %s are required", path)
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
