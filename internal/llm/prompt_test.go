package llm

import (
	"testing"
)

func TestSystemPrompt(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "single command",
			count: 1,
			want:  "exactly 1 different command variants",
		},
		{
			name:  "multiple commands",
			count: 5,
			want:  "exactly 5 different command variants",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SystemPrompt(tt.count)
			if got == "" {
				t.Error("SystemPrompt returned empty string")
			}
			if !contains(got, tt.want) {
				t.Errorf("SystemPrompt does not contain expected text: %q", tt.want)
			}
		})
	}
}

func TestSystemPromptContainsStdinInstruction(t *testing.T) {
	got := SystemPrompt(3)

	if !contains(got, "stdin") {
		t.Error("SystemPrompt should mention stdin context")
	}
	if !contains(got, "ONLY executable shell commands") {
		t.Error("SystemPrompt should contain instruction about generating only executable commands")
	}
}

func TestUserPrompt(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		stdinContent string
		wantQuery    bool
		wantContext  bool
	}{
		{
			name:         "query without stdin",
			query:        "list all files",
			stdinContent: "",
			wantQuery:    true,
			wantContext:  false,
		},
		{
			name:         "query with stdin context",
			query:        "stop the nginx container",
			stdinContent: "CONTAINER ID   IMAGE   STATUS\nabc123   nginx   Up 5 hours",
			wantQuery:    true,
			wantContext:  true,
		},
		{
			name:         "stdin with multiline content",
			query:        "filter errors",
			stdinContent: "line1\nline2\nline3",
			wantQuery:    true,
			wantContext:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserPrompt(tt.query, tt.stdinContent)
			if got == "" {
				t.Error("UserPrompt returned empty string")
			}
			if tt.wantQuery && !contains(got, tt.query) {
				t.Errorf("UserPrompt does not contain query: %q", tt.query)
			}
			if tt.wantContext && !contains(got, "Context from stdin") {
				t.Error("UserPrompt does not contain stdin context header")
			}
			if tt.wantContext && !contains(got, tt.stdinContent) {
				t.Errorf("UserPrompt does not contain stdin content: %q", tt.stdinContent)
			}
			if !tt.wantContext && contains(got, "Context from stdin") {
				t.Error("UserPrompt contains stdin context header when it shouldn't")
			}
		})
	}
}

func TestUserPromptFormat(t *testing.T) {
	query := "use this data"
	stdinContent := "data here"

	got := UserPrompt(query, stdinContent)

	// Verify the format structure
	if !contains(got, "---") {
		t.Error("UserPrompt missing delimiter markers")
	}
	if !contains(got, "User query:") {
		t.Error("UserPrompt missing 'User query:' label")
	}
	if !contains(got, "Generate shell commands") {
		t.Error("UserPrompt missing generation instruction")
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
