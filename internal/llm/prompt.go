package llm

import (
	"encoding/json"
	"fmt"
)

// commandsResponse represents the expected JSON structure from LLM
type commandsResponse struct {
	Commands []string `json:"commands"`
}

// SystemPrompt generates the system prompt for command generation.
// count specifies how many command variants should be generated.
func SystemPrompt(count int) string {
	return fmt.Sprintf(`You are a shell command generator. Generate shell commands based on user descriptions.

Rules:
- Generate ONLY executable shell commands, no explanations or descriptions
- Generate POSIX-compatible commands that work in bash, zsh, and fish
- Return exactly %d different command variants
- Commands should be practical and safe
- Prefer common Unix utilities (find, grep, awk, sed, etc.)
- Never include explanations, only raw commands
- Each command should solve the same task in a different way
- If context from stdin is provided, use it to generate more relevant commands

Response format (JSON):
{
  "commands": ["command1", "command2", ...]
}`, count)
}

// UserPrompt formats the user query with optional stdin context.
// If stdinContent is non-empty, it's included as context for the LLM.
func UserPrompt(query, stdinContent string) string {
	if stdinContent == "" {
		return query
	}

	return fmt.Sprintf(`Context from stdin:
---
%s
---

User query: %s

Generate shell commands that use the context above.`, stdinContent, query)
}

// DescribeSystemPrompt generates the system prompt for command description mode.
func DescribeSystemPrompt() string {
	return `You are a shell command expert. Explain shell commands clearly and concisely.

Rules:
- Explain what each part of the command does
- Describe any flags or options used
- Explain the expected behavior and output
- Keep explanations clear and practical
- Do not suggest alternative commands unless the command is incorrect`
}

// DescribeUserPrompt formats the user query for command description mode.
func DescribeUserPrompt(command string) string {
	return fmt.Sprintf(`Explain what the following shell command does:

Command: %s

Provide a clear, concise explanation of:
1. What each part of the command does
2. Any flags or options used
3. Expected behavior and output`, command)
}

// ParseCommands parses JSON response from LLM into a list of commands
func ParseCommands(jsonResponse []byte) ([]string, error) {
	if len(jsonResponse) == 0 {
		return nil, fmt.Errorf("empty response from LLM")
	}

	var response commandsResponse
	if err := json.Unmarshal(jsonResponse, &response); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(response.Commands) == 0 {
		return nil, fmt.Errorf("LLM returned no commands")
	}

	validCommands := make([]string, 0, len(response.Commands))
	for _, cmd := range response.Commands {
		if cmd != "" {
			validCommands = append(validCommands, FormatCommand(cmd))
		}
	}

	if len(validCommands) == 0 {
		return nil, fmt.Errorf("LLM returned only empty commands")
	}

	return validCommands, nil
}
