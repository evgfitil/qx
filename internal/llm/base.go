package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

const (
	DefaultRequestTimeout = 30 * time.Second

	// DefaultTemperature defines the temperature for command generation
	DefaultTemperature = 0.7
)

// baseProvider contains common logic for all LLM providers
type baseProvider struct {
	client *openai.Client
	model  string
}

// Generate creates shell commands based on user query
func (p *baseProvider) Generate(ctx context.Context, query string, count int) ([]string, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	req := openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: SystemPrompt(count),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: query,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Temperature: DefaultTemperature,
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultRequestTimeout)
		defer cancel()
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	content := resp.Choices[0].Message.Content
	commands, err := ParseCommands([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM output: %w", err)
	}

	return commands, nil
}
