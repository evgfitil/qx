package llm

import (
	"fmt"
	"testing"
)

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

func TestUnformatCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple command without continuations",
			input:    "ls -la",
			expected: "ls -la",
		},
		{
			name:     "single pipe continuation",
			input:    "ps aux \\\n\t| grep nginx",
			expected: "ps aux | grep nginx",
		},
		{
			name:     "multiple pipe continuations",
			input:    "cat file \\\n\t| grep pattern \\\n\t| awk '{print $1}'",
			expected: "cat file | grep pattern | awk '{print $1}'",
		},
		{
			name:     "logical and continuation",
			input:    "mkdir dir \\\n\t&& cd dir",
			expected: "mkdir dir && cd dir",
		},
		{
			name:     "logical or continuation",
			input:    "test -f file \\\n\t|| touch file",
			expected: "test -f file || touch file",
		},
		{
			name:     "mixed operators",
			input:    "cmd1 \\\n\t| cmd2 \\\n\t&& cmd3 \\\n\t|| cmd4",
			expected: "cmd1 | cmd2 && cmd3 || cmd4",
		},
		{
			name:     "literal backslash preserved",
			input:    `echo "hello\nworld"`,
			expected: `echo "hello\nworld"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "spaces before backslash",
			input:    "ps aux   \\\n\t| grep nginx",
			expected: "ps aux | grep nginx",
		},
		{
			name:     "spaces instead of tab after newline",
			input:    "ps aux \\\n    | grep nginx",
			expected: "ps aux | grep nginx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnformatCommand(tt.input)
			if result != tt.expected {
				t.Errorf("UnformatCommand(%q)\ngot:  %q\nwant: %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatUnformatRoundTrip(t *testing.T) {
	commands := []string{
		"ls -la",
		"ps aux | grep nginx",
		"cat file | grep pattern | awk '{print $1}'",
		"mkdir dir && cd dir",
		"test -f file || touch file",
		"echo 'hello | world'",
		"echo 'unclosed",
	}

	for _, cmd := range commands {
		t.Run(fmt.Sprintf("roundtrip_%s", cmd), func(t *testing.T) {
			formatted := FormatCommand(cmd)
			result := UnformatCommand(formatted)
			if result != cmd {
				t.Errorf("round-trip failed for %q\nFormatCommand: %q\nUnformatCommand: %q", cmd, formatted, result)
			}
		})
	}
}
