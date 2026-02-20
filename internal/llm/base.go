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
// pipeContext contains optional stdin data piped into qx for additional context.
// followUp, when non-nil, injects previous query/command as conversation history for refinement.
func (p *baseProvider) Generate(ctx context.Context, query string, count int, pipeContext string, followUp *FollowUpContext) ([]string, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	userMessage := query
	if pipeContext != "" {
		userMessage = fmt.Sprintf("Context:\n<stdin>\n%s\n</stdin>\n\nTask: %s", pipeContext, query)
	}

	messages := buildMessages(count, pipeContext, userMessage, followUp)

	req := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: messages,
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

// buildMessages constructs the chat message list for the LLM request.
// When followUp is non-nil, inserts previous query/command as conversation history.
func buildMessages(count int, pipeContext string, userMessage string, followUp *FollowUpContext) []openai.ChatCompletionMessage {
	hasFollowUp := followUp != nil
	systemMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: SystemPrompt(count, pipeContext != "", hasFollowUp),
	}

	if !hasFollowUp {
		return []openai.ChatCompletionMessage{
			systemMsg,
			{Role: openai.ChatMessageRoleUser, Content: userMessage},
		}
	}

	return []openai.ChatCompletionMessage{
		systemMsg,
		{Role: openai.ChatMessageRoleUser, Content: followUp.PreviousQuery},
		{Role: openai.ChatMessageRoleAssistant, Content: followUp.PreviousCommand},
		{Role: openai.ChatMessageRoleUser, Content: userMessage},
	}
}
