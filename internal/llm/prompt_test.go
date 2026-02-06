package llm

import (
	"testing"
)

func TestSystemPrompt(t *testing.T) {
	tests := []struct {
		name           string
		count          int
		hasPipeContext bool
		want           string
		wantAlso       string
		wantAbsent     string
	}{
		{
			name:           "single command without pipe context",
			count:          1,
			hasPipeContext: false,
			want:           "exactly 1 different command variants",
			wantAbsent:     "stdin context",
		},
		{
			name:           "multiple commands without pipe context",
			count:          5,
			hasPipeContext: false,
			want:           "exactly 5 different command variants",
			wantAbsent:     "stdin context",
		},
		{
			name:           "with pipe context includes stdin instructions and count",
			count:          3,
			hasPipeContext: true,
			want:           "stdin context is provided",
			wantAlso:       "exactly 3 different command variants",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SystemPrompt(tt.count, tt.hasPipeContext)
			if got == "" {
				t.Error("SystemPrompt returned empty string")
			}
			if !contains(got, tt.want) {
				t.Errorf("SystemPrompt does not contain expected text: %q", tt.want)
			}
			if tt.wantAlso != "" && !contains(got, tt.wantAlso) {
				t.Errorf("SystemPrompt does not contain expected text: %q", tt.wantAlso)
			}
			if tt.wantAbsent != "" && contains(got, tt.wantAbsent) {
				t.Errorf("SystemPrompt should not contain: %q", tt.wantAbsent)
			}
		})
	}
}

func TestParseCommands(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "valid single command",
			input: `{"commands": ["ls -la"]}`,
			want:  []string{"ls -la"},
		},
		{
			name:  "valid multiple commands",
			input: `{"commands": ["ls -la", "find . -name '*.go'"]}`,
			want:  []string{"ls -la", "find . -name '*.go'"},
		},
		{
			name:    "invalid JSON",
			input:   `{"commands": [}`,
			wantErr: true,
		},
		{
			name:    "empty response",
			input:   "",
			wantErr: true,
		},
		{
			name:    "empty commands array",
			input:   `{"commands": []}`,
			wantErr: true,
		},
		{
			name:  "commands with empty strings",
			input: `{"commands": ["ls -la", "", "pwd"]}`,
			want:  []string{"ls -la", "pwd"},
		},
		{
			name:    "all empty commands",
			input:   `{"commands": ["", ""]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCommands([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommands() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equalSlices(got, tt.want) {
				t.Errorf("ParseCommands() = %v, want %v", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
