package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestGenerate_WithPipeContext(t *testing.T) {
	var capturedRequest openai.ChatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: `{"commands": ["docker stop abc123"]}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := openai.DefaultConfig("test-key")
	cfg.BaseURL = server.URL + "/v1"
	provider := &baseProvider{
		client: openai.NewClientWithConfig(cfg),
		model:  "test-model",
	}

	commands, err := provider.Generate(context.Background(), "stop nginx", 1, "CONTAINER ID\nabc123 nginx", nil)
	if err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	// Verify user message contains pipe context with proper format
	userMsg := capturedRequest.Messages[1].Content
	if !strings.Contains(userMsg, "<stdin>") {
		t.Error("user message should contain <stdin> tag")
	}
	if !strings.Contains(userMsg, "abc123 nginx") {
		t.Error("user message should contain pipe context data")
	}
	if !strings.Contains(userMsg, "Task: stop nginx") {
		t.Error("user message should contain 'Task: ' prefix for query")
	}

	// Verify system prompt contains stdin instructions
	sysMsg := capturedRequest.Messages[0].Content
	if !strings.Contains(sysMsg, "stdin context") {
		t.Error("system prompt should mention stdin context when pipe context is present")
	}
}

func TestGenerate_WithoutPipeContext(t *testing.T) {
	var capturedRequest openai.ChatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: `{"commands": ["ls -la"]}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := openai.DefaultConfig("test-key")
	cfg.BaseURL = server.URL + "/v1"
	provider := &baseProvider{
		client: openai.NewClientWithConfig(cfg),
		model:  "test-model",
	}

	commands, err := provider.Generate(context.Background(), "list files", 1, "", nil)
	if err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	// Verify user message is just the query without any context wrapping
	userMsg := capturedRequest.Messages[1].Content
	if userMsg != "list files" {
		t.Errorf("expected user message %q, got %q", "list files", userMsg)
	}
	if strings.Contains(userMsg, "<stdin>") {
		t.Error("user message should not contain <stdin> tag when no pipe context")
	}

	// Verify system prompt does not contain stdin instructions
	sysMsg := capturedRequest.Messages[0].Content
	if strings.Contains(sysMsg, "stdin context") {
		t.Error("system prompt should not mention stdin context when no pipe context")
	}
}

func TestGenerate_EmptyQuery(t *testing.T) {
	provider := &baseProvider{model: "test"}

	_, err := provider.Generate(context.Background(), "", 1, "", nil)
	if err == nil {
		t.Fatal("Generate() expected error for empty query")
	}
	if !strings.Contains(err.Error(), "query cannot be empty") {
		t.Errorf("expected 'query cannot be empty' error, got: %v", err)
	}
}

func TestGenerate_WithFollowUpContext(t *testing.T) {
	var capturedRequest openai.ChatCompletionRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&capturedRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: `{"commands": ["find . -name '*.go' -size +1M"]}`}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cfg := openai.DefaultConfig("test-key")
	cfg.BaseURL = server.URL + "/v1"
	provider := &baseProvider{
		client: openai.NewClientWithConfig(cfg),
		model:  "test-model",
	}

	followUp := &FollowUpContext{
		PreviousQuery:   "find large files",
		PreviousCommand: "find . -size +100M",
	}

	commands, err := provider.Generate(context.Background(), "only go files", 1, "", followUp)
	if err != nil {
		t.Fatalf("Generate() unexpected error: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}

	// Verify message structure: [system, user(prev), assistant(prev cmd), user(new)]
	if len(capturedRequest.Messages) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(capturedRequest.Messages))
	}

	if capturedRequest.Messages[0].Role != openai.ChatMessageRoleSystem {
		t.Errorf("message[0] role = %q, want system", capturedRequest.Messages[0].Role)
	}
	if capturedRequest.Messages[1].Role != openai.ChatMessageRoleUser {
		t.Errorf("message[1] role = %q, want user", capturedRequest.Messages[1].Role)
	}
	if capturedRequest.Messages[1].Content != "find large files" {
		t.Errorf("message[1] content = %q, want previous query", capturedRequest.Messages[1].Content)
	}
	if capturedRequest.Messages[2].Role != openai.ChatMessageRoleAssistant {
		t.Errorf("message[2] role = %q, want assistant", capturedRequest.Messages[2].Role)
	}
	if capturedRequest.Messages[2].Content != "find . -size +100M" {
		t.Errorf("message[2] content = %q, want previous command", capturedRequest.Messages[2].Content)
	}
	if capturedRequest.Messages[3].Role != openai.ChatMessageRoleUser {
		t.Errorf("message[3] role = %q, want user", capturedRequest.Messages[3].Role)
	}
	if capturedRequest.Messages[3].Content != "only go files" {
		t.Errorf("message[3] content = %q, want new query", capturedRequest.Messages[3].Content)
	}

	// Verify system prompt contains follow-up rules
	sysMsg := capturedRequest.Messages[0].Content
	if !strings.Contains(sysMsg, "refining a previous command") {
		t.Error("system prompt should contain follow-up refinement rules")
	}
}

func TestBuildMessages_WithoutFollowUp(t *testing.T) {
	msgs := buildMessages(3, "", "list files", nil)
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != openai.ChatMessageRoleSystem {
		t.Errorf("message[0] role = %q, want system", msgs[0].Role)
	}
	if msgs[1].Content != "list files" {
		t.Errorf("message[1] content = %q, want %q", msgs[1].Content, "list files")
	}
}

func TestBuildMessages_WithFollowUp(t *testing.T) {
	followUp := &FollowUpContext{
		PreviousQuery:   "find files",
		PreviousCommand: "find . -type f",
	}
	msgs := buildMessages(3, "", "make it recursive", followUp)
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}
	if msgs[1].Content != "find files" {
		t.Errorf("message[1] = %q, want previous query", msgs[1].Content)
	}
	if msgs[2].Content != "find . -type f" {
		t.Errorf("message[2] = %q, want previous command", msgs[2].Content)
	}
	if msgs[3].Content != "make it recursive" {
		t.Errorf("message[3] = %q, want new query", msgs[3].Content)
	}
}
