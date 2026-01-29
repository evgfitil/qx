package llm

import "testing"

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "creates OpenAI provider by default",
			cfg: Config{
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-key",
				Model:   "gpt-4o-mini",
			},
		},
		{
			name: "creates OpenAI provider for Groq",
			cfg: Config{
				BaseURL: "https://api.groq.com/openai/v1",
				APIKey:  "test-key",
				Model:   "llama-3.1-70b",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.cfg)
			if err != nil {
				t.Errorf("NewProvider() unexpected error = %v", err)
				return
			}

			if provider == nil {
				t.Error("NewProvider() returned nil provider")
				return
			}

			if _, ok := provider.(*OpenAIProvider); !ok {
				t.Errorf("NewProvider() returned %T, want *OpenAIProvider", provider)
			}
		})
	}
}
