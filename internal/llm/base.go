package llm

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// categorizeAPIError returns a user-friendly error message based on the API error type
func categorizeAPIError(err error) error {
	var apiErr *openai.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.HTTPStatusCode {
		case 401:
			return fmt.Errorf("authentication failed: check OPENAI_API_KEY")
		case 429:
			return fmt.Errorf("rate limit exceeded")
		case 500, 502, 503:
			return fmt.Errorf("API server error: try again later")
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("request timed out")
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("request canceled")
	}
	return err
}

// Generate creates shell commands based on user query.
// stdinContent is optional context from stdin that helps generate more relevant commands.
func (p *baseProvider) Generate(ctx context.Context, query string, count int, stdinContent string) ([]string, error) {
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
				Content: UserPrompt(query, stdinContent),
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
		return nil, categorizeAPIError(err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	content := resp.Choices[0].Message.Content
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("LLM returned empty response")
	}

	commands, err := ParseCommands([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM output: %w", err)
	}

	return commands, nil
}

// Describe provides explanation for a shell command.
func (p *baseProvider) Describe(ctx context.Context, command string) (string, error) {
	if command == "" {
		return "", fmt.Errorf("command cannot be empty")
	}

	req := openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: DescribeSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: DescribeUserPrompt(command),
			},
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
		return "", categorizeAPIError(err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("LLM returned empty response")
	}

	return content, nil
}
