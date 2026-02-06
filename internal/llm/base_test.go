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

	commands, err := provider.Generate(context.Background(), "stop nginx", 1, "CONTAINER ID\nabc123 nginx")
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

	commands, err := provider.Generate(context.Background(), "list files", 1, "")
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

	_, err := provider.Generate(context.Background(), "", 1, "")
	if err == nil {
		t.Fatal("Generate() expected error for empty query")
	}
	if !strings.Contains(err.Error(), "query cannot be empty") {
		t.Errorf("expected 'query cannot be empty' error, got: %v", err)
	}
}
