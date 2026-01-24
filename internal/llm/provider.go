package llm

import "context"

// Config contains configuration for LLM provider
type Config struct {
	BaseURL  string
	APIKey   string
	Model    string
	Provider string // optional: "eliza" to use Eliza, empty or "openai" for OpenAI-compatible API
}

// Provider generates shell commands using LLM
type Provider interface {
	Generate(ctx context.Context, query string, count int) ([]string, error)
}

// NewProvider creates appropriate provider based on configuration.
// Returns ElizaProvider if cfg.Provider is "eliza", otherwise returns OpenAIProvider.
func NewProvider(cfg Config) (Provider, error) {
	if cfg.Provider == "eliza" {
		return newElizaProvider(cfg)
	}
	return newOpenAIProvider(cfg)
}
