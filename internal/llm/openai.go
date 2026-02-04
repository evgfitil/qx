package llm

import (
	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider is an LLM provider for standard OpenAI-compatible APIs.
type OpenAIProvider struct {
	baseProvider
}

// newOpenAIProvider creates a new OpenAI-compatible provider.
func newOpenAIProvider(cfg Config) (*OpenAIProvider, error) {
	config := openai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}

	return &OpenAIProvider{
		baseProvider: baseProvider{
			client: openai.NewClientWithConfig(config),
			model:  cfg.Model,
		},
	}, nil
}
