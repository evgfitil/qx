package llm

import "context"

// Config contains configuration for LLM provider
type Config struct {
	BaseURL  string
	APIKey   string
	Model    string
	Provider string // optional: "eliza" to use Eliza, empty or "openai" for OpenAI-compatible API
	Count    int    // number of command variants to generate
}

// Provider generates shell commands using LLM
type Provider interface {
	Generate(ctx context.Context, query string, count int) ([]string, error)
}

// NewProvider creates appropriate provider based on configuration
func NewProvider(cfg Config) (Provider, error) {
	return newOpenAIProvider(cfg)
}
