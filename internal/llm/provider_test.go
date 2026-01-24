package llm

import "testing"

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		wantType string // "openai" or "eliza"
	}{
		{
			name: "creates OpenAI provider by default",
			cfg: Config{
				BaseURL: "https://api.openai.com/v1",
				APIKey:  "test-key",
				Model:   "gpt-4o-mini",
			},
			wantType: "openai",
		},
		{
			name: "creates OpenAI provider for Groq",
			cfg: Config{
				BaseURL: "https://api.groq.com/openai/v1",
				APIKey:  "test-key",
				Model:   "llama-3.1-70b",
			},
			wantType: "openai",
		},
		{
			name: "creates OpenAI provider with explicit provider=openai",
			cfg: Config{
				BaseURL:  "https://api.openai.com/v1",
				APIKey:   "test-key",
				Model:    "gpt-4o-mini",
				Provider: "openai",
			},
			wantType: "openai",
		},
		{
			name: "creates Eliza provider with explicit provider=eliza",
			cfg: Config{
				BaseURL:  "https://api.eliza.yandex.net/openai/v1",
				APIKey:   "test-key",
				Model:    "gpt-4o-mini",
				Provider: "eliza",
			},
			wantType: "eliza",
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

			// Check provider type
			switch tt.wantType {
			case "openai":
				if _, ok := provider.(*OpenAIProvider); !ok {
					t.Errorf("NewProvider() returned %T, want *OpenAIProvider", provider)
				}
			case "eliza":
				if _, ok := provider.(*ElizaProvider); !ok {
					t.Errorf("NewProvider() returned %T, want *ElizaProvider", provider)
				}
			}
		})
	}
}

