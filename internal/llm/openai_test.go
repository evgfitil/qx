package llm

import "testing"

func TestNewOpenAIProvider(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "valid config",
			cfg: Config{
				APIKey: "test-key",
				Model:  "gpt-4o-mini",
			},
		},
		{
			name: "custom base URL",
			cfg: Config{
				BaseURL: "https://api.groq.com/openai/v1",
				APIKey:  "test-key",
				Model:   "llama-3.1-70b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := newOpenAIProvider(tt.cfg)
			if err != nil {
				t.Errorf("newOpenAIProvider() unexpected error = %v", err)
				return
			}
			if provider == nil {
				t.Error("newOpenAIProvider() returned nil provider")
			}
		})
	}
}
