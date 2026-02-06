package tui

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"

	"github.com/evgfitil/qx/internal/guard"
	"github.com/evgfitil/qx/internal/llm"
)

// newMockLLMServer creates a test HTTP server that captures the request body
// and returns a predefined LLM response.
func newMockLLMServer(t *testing.T, capturedBody *string, responseJSON string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		*capturedBody = string(body)

		resp := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Content: responseJSON}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestModel_Result_CancelledOnEsc(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}
	initialQuery := "list files"

	m := NewModel(cfg, initialQuery, false, "")

	// Simulate Esc press
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected result to be CancelledResult after Esc")
	}
	if cancelled.Query != initialQuery {
		t.Errorf("expected Query = %q, got %q", initialQuery, cancelled.Query)
	}
}

func TestModel_Result_SelectedOnEnter(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")

	// Simulate receiving commands from LLM
	m.commands = []string{"ls -la", "ls -lah", "ls -l"}
	m.filtered = m.commands
	m.state = stateSelect
	m.cursor = 1 // select second command

	// Simulate Enter press to select
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	result := model.Result()
	selected, ok := result.(SelectedResult)
	if !ok {
		t.Fatal("expected result to be SelectedResult after selection")
	}
	if selected.Command != "ls -lah" {
		t.Errorf("expected Command = %q, got %q", "ls -lah", selected.Command)
	}
}

func TestModel_Result_CancelledWithCtrlC(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}
	initialQuery := "show processes"

	m := NewModel(cfg, initialQuery, false, "")

	// Simulate Ctrl+C press
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected result to be CancelledResult after Ctrl+C")
	}
	if cancelled.Query != initialQuery {
		t.Errorf("expected Query = %q, got %q", initialQuery, cancelled.Query)
	}
}

func TestModel_Result_NoActionYet(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "test query", false, "")

	// No user action yet - model in initial state
	result := m.Result()

	// Should return CancelledResult with current query when no action taken
	_, ok := result.(CancelledResult)
	if !ok {
		t.Error("expected result to be CancelledResult when no action taken")
	}
}

func TestModel_Result_EmptyQueryOnEsc(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	// Start with empty query
	m := NewModel(cfg, "", false, "")

	// Simulate Esc press
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected result to be CancelledResult after Esc")
	}
	if cancelled.Query != "" {
		t.Errorf("expected Query = %q, got %q", "", cancelled.Query)
	}
}

func TestModel_Result_ModifiedQueryOnEsc(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	initialQuery := "list files"
	m := NewModel(cfg, initialQuery, false, "")

	// Simulate typing additional text - send individual key messages
	for _, r := range " with details" {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	// Now the text input should contain "list files with details"
	expectedQuery := "list files with details"
	if m.textInput.Value() != expectedQuery {
		t.Fatalf("precondition failed: expected input value %q, got %q", expectedQuery, m.textInput.Value())
	}

	// Simulate Esc press
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected result to be CancelledResult after Esc")
	}

	// Should return the modified query, not the initial one
	if cancelled.Query != expectedQuery {
		t.Errorf("expected Query = %q, got %q", expectedQuery, cancelled.Query)
	}
}

func TestModel_Result_CancelledFromSelectState(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	originalQuery := "list files"
	m := NewModel(cfg, originalQuery, false, "")

	// Simulate the flow: user enters query, presses Enter, receives commands
	m.originalQuery = originalQuery
	m.commands = []string{"ls -la", "ls -lah", "ls -l"}
	m.filtered = m.commands
	m.state = stateSelect
	m.textInput.SetValue("") // This happens when transitioning to stateSelect

	// Simulate Esc press while in select state
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected result to be CancelledResult after Esc in select state")
	}

	// Should return the original query, not empty string
	if cancelled.Query != originalQuery {
		t.Errorf("expected Query = %q, got %q", originalQuery, cancelled.Query)
	}
}

func TestModel_Result_CancelledFromLoadingState(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	originalQuery := "list files"
	m := NewModel(cfg, originalQuery, false, "")

	// Simulate the flow: user enters query and presses Enter, now in loading state
	m.state = stateLoading
	m.originalQuery = originalQuery

	// Simulate Esc press while loading
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected result to be CancelledResult after Esc in loading state")
	}

	// Should return the original query
	if cancelled.Query != originalQuery {
		t.Errorf("expected Query = %q, got %q", originalQuery, cancelled.Query)
	}
}

func TestNewModel_WithPipeContext(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}
	pipeCtx := "total 48\n-rw-r--r-- 1 user staff 1024 Jan 1 file.txt"

	m := NewModel(cfg, "delete large files", false, pipeCtx)

	if m.pipeContext != pipeCtx {
		t.Errorf("expected pipeContext = %q, got %q", pipeCtx, m.pipeContext)
	}
	if m.state != stateInput {
		t.Errorf("expected state = stateInput, got %d", m.state)
	}
}

func TestNewModel_WithoutPipeContext(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "list files", false, "")

	if m.pipeContext != "" {
		t.Errorf("expected pipeContext = %q, got %q", "", m.pipeContext)
	}
}

func TestGenerateCommands_WithPipeContext(t *testing.T) {
	var capturedBody string
	server := newMockLLMServer(t, &capturedBody, `{"commands": ["docker stop abc"]}`)
	defer server.Close()

	cfg := llm.Config{
		BaseURL: server.URL + "/v1",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}
	pipeCtx := "CONTAINER ID\nabc123 nginx"

	cmd := generateCommands("stop nginx", cfg, pipeCtx)
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd")
	}

	msg := cmd()
	cmdMsg, ok := msg.(commandsMsg)
	if !ok {
		t.Fatalf("expected commandsMsg, got %T", msg)
	}
	if cmdMsg.err != nil {
		t.Fatalf("unexpected error: %v", cmdMsg.err)
	}

	// JSON encoding escapes angle brackets, so check for the content itself
	if !strings.Contains(capturedBody, "abc123 nginx") {
		t.Error("request should contain pipe context data")
	}
	if !strings.Contains(capturedBody, "Task: stop nginx") {
		t.Error("request should contain 'Task:' prefix for the query")
	}
}

func TestGenerateCommands_WithoutPipeContext(t *testing.T) {
	var capturedBody string
	server := newMockLLMServer(t, &capturedBody, `{"commands": ["ls -la"]}`)
	defer server.Close()

	cfg := llm.Config{
		BaseURL: server.URL + "/v1",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	cmd := generateCommands("list files", cfg, "")
	if cmd == nil {
		t.Fatal("expected non-nil tea.Cmd")
	}

	msg := cmd()
	cmdMsg, ok := msg.(commandsMsg)
	if !ok {
		t.Fatalf("expected commandsMsg, got %T", msg)
	}
	if cmdMsg.err != nil {
		t.Fatalf("unexpected error: %v", cmdMsg.err)
	}

	if strings.Contains(capturedBody, "Task:") {
		t.Error("request should not contain 'Task:' prefix when no pipe context")
	}
}

func TestHandleEnter_PipeContextSecretDetection(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}
	secretPipeCtx := "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE"

	m := NewModel(cfg, "use this key", false, secretPipeCtx)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.err == nil {
		t.Fatal("expected error for secret in pipe context")
	}

	var secretsErr *guard.SecretsError
	if !errors.As(model.err, &secretsErr) {
		t.Errorf("expected SecretsError, got %T: %v", model.err, model.err)
	}
	if model.state != stateInput {
		t.Errorf("expected state to remain stateInput, got %d", model.state)
	}
}

func TestHandleEnter_PipeContextNoSecret(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}
	safePipeCtx := "total 48\n-rw-r--r-- 1 user staff 1024 Jan 1 file.txt"

	m := NewModel(cfg, "delete large files", false, safePipeCtx)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.err != nil {
		t.Errorf("expected no error, got %v", model.err)
	}
	if model.state != stateLoading {
		t.Errorf("expected state = stateLoading, got %d", model.state)
	}
	if cmd == nil {
		t.Error("expected non-nil command for generating")
	}
}
