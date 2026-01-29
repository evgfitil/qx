package llm

import "testing"

func TestFormatCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple command",
			input:    "ls -la",
			expected: "ls -la",
		},
		{
			name:     "single pipe",
			input:    "ps aux | grep nginx",
			expected: "ps aux \\\n\t| grep nginx",
		},
		{
			name:     "multiple pipes",
			input:    "cat file | grep pattern | awk '{print $1}'",
			expected: "cat file \\\n\t| grep pattern \\\n\t| awk '{print $1}'",
		},
		{
			name:     "logical and",
			input:    "mkdir dir && cd dir",
			expected: "mkdir dir \\\n\t&& cd dir",
		},
		{
			name:     "logical or",
			input:    "test -f file || touch file",
			expected: "test -f file \\\n\t|| touch file",
		},
		{
			name:     "pipe in quotes preserved",
			input:    "echo 'hello | world'",
			expected: "echo 'hello | world'",
		},
		{
			name:     "invalid syntax returns original",
			input:    "echo 'unclosed",
			expected: "echo 'unclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCommand(tt.input)
			if result != tt.expected {
				t.Errorf("FormatCommand(%q)\ngot:  %q\nwant: %q", tt.input, result, tt.expected)
			}
		})
	}
}
