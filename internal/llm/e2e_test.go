//go:build e2e

package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// requireEnv skips the test if the environment variable is not set
func requireEnv(t *testing.T, key string) string {
	t.Helper()
	val := os.Getenv(key)
	if val == "" {
		t.Skipf("skipping: %s not set", key)
	}
	return val
}

func newTestProvider(t *testing.T) Provider {
	t.Helper()
	apiKey := requireEnv(t, "OPENAI_API_KEY")
	baseURL := requireEnv(t, "LLM_BASE_URL")
	model := os.Getenv("LLM_MODEL")

	if model == "" {
		model = "gpt-4o-mini"
	}

	provider, err := NewProvider(Config{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	return provider
}

func TestE2E_SingleToolPreference_NoPipe(t *testing.T) {
	provider := newTestProvider(t)

	tests := []struct {
		name     string
		query    string
		wantHint string // substring expected in at least one command
	}{
		{
			name:     "find go files and count lines",
			query:    "find all go files and count lines in each",
			wantHint: "find",
		},
		{
			name:     "extract name from package.json",
			query:    "extract name field from package.json",
			wantHint: "jq",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			commands, err := provider.Generate(ctx, tt.query, 3, "")
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			t.Logf("Query: %s", tt.query)
			for i, cmd := range commands {
				t.Logf("  [%d] %s (pipes: %d)", i, cmd, strings.Count(cmd, "|"))
			}

			hasHint := false
			for _, cmd := range commands {
				if strings.Contains(strings.ToLower(cmd), tt.wantHint) {
					hasHint = true
					break
				}
			}
			if !hasHint {
				t.Errorf("expected at least one command containing %q", tt.wantHint)
			}

			hasMinimalPipe := false
			for _, cmd := range commands {
				if strings.Count(cmd, "|") <= 1 {
					hasMinimalPipe = true
					break
				}
			}
			if !hasMinimalPipe {
				t.Errorf("expected at least one command with 0-1 pipes, all have more")
			}
		})
	}
}

func TestE2E_SingleToolPreference_WithPipe(t *testing.T) {
	provider := newTestProvider(t)

	tests := []struct {
		name        string
		query       string
		pipeContext string
		wantHint    string
	}{
		{
			name:        "pipe JSON get active users",
			query:       "get active users",
			pipeContext: `{"users":[{"name":"alice","active":true},{"name":"bob","active":false}]}`,
			wantHint:    "jq",
		},
		{
			name:        "pipe tabular data find high CPU",
			query:       "find processes using more than 1% CPU",
			pipeContext: "USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND\nroot         1  0.0  0.1 169360 13288 ?        Ss   Jan01   0:15 /sbin/init\nwww-data  1234  5.2  1.3 456789 98765 ?        Sl   10:00   1:23 nginx: worker\npostgres  5678  2.1  0.8 234567 54321 ?        Ss   09:00   0:45 postgres",
			wantHint:    "awk",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			commands, err := provider.Generate(ctx, tt.query, 3, tt.pipeContext)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			t.Logf("Query: %s (pipe context provided)", tt.query)
			for i, cmd := range commands {
				t.Logf("  [%d] %s (pipes: %d)", i, cmd, strings.Count(cmd, "|"))
			}

			hasHint := false
			for _, cmd := range commands {
				if strings.Contains(strings.ToLower(cmd), tt.wantHint) {
					hasHint = true
					break
				}
			}
			if !hasHint {
				t.Errorf("expected at least one command containing %q", tt.wantHint)
			}

			hasMinimalPipe := false
			for _, cmd := range commands {
				if strings.Count(cmd, "|") <= 1 {
					hasMinimalPipe = true
					break
				}
			}
			if !hasMinimalPipe {
				t.Errorf("expected at least one command with 0-1 pipes, all have more")
			}
		})
	}
}
