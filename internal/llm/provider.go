package llm

// Config contains configuration for LLM provider.
type Config struct {
	BaseURL  string
	APIKey   string
	Model    string
	Provider string
	Count    int // number of command variants to generate
}

// NewProvider creates a new LLM provider with the given configuration.
func NewProvider(cfg Config) (*OpenAIProvider, error) {
	return newOpenAIProvider(cfg)
}
