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
// hasPipeContext indicates whether stdin context is provided with the request.
// hasFollowUp indicates whether this is a follow-up refinement of a previous command.
func SystemPrompt(count int, hasPipeContext bool, hasFollowUp bool) string {
	pipeRules := ""
	if hasPipeContext {
		pipeRules = `
- When stdin context is provided, use the concrete values from it to generate precise commands
- Generate commands that reference actual data from the context (file names, container IDs, process IDs, etc.)
- Identify the source tool from the context and prefer using its built-in capabilities for filtering, sorting, and formatting over adding separate tools to the pipeline`
	}

	followUpRules := ""
	if hasFollowUp {
		followUpRules = `
- The user is refining a previous command. Consider the conversation history and generate commands that address the user's refinement request`
	}

	return fmt.Sprintf(`You are a shell command generator. Generate shell commands based on user descriptions.

Rules:
- Generate POSIX-compatible commands that work in bash, zsh, and fish
- Return exactly %d different command variants
- Commands should be practical and safe
- Prefer using a single tool's full capabilities over chaining multiple tools
- Use built-in filtering, selection, and formatting options of tools (e.g., jq select instead of grep, kubectl --field-selector instead of pipe to grep, find -exec instead of find | xargs)
- When a tool supports structured output (JSON, YAML, CSV), use its native query capabilities rather than text processing with grep/awk/sed
- Minimize pipe chains: fewer pipes = better
- Never include explanations, only raw commands
- Each command should solve the same task in a different way%s%s

Response format (JSON):
{
  "commands": ["command1", "command2", ...]
}`, count, pipeRules, followUpRules)
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
