package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sashabaranov/go-openai"

	"github.com/evgfitil/qx/internal/llm"
)

// newMockLLMServer creates a test HTTP server that captures the request body
// and returns a predefined LLM response.
func newMockLLMServer(t *testing.T, capturedBody *string, responseJSON string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
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
	if m.textArea.Value() != expectedQuery {
		t.Fatalf("precondition failed: expected input value %q, got %q", expectedQuery, m.textArea.Value())
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
	m.textArea.SetValue("") // This happens when transitioning to stateSelect

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

func TestNewModel_TextAreaConfig(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")

	if m.textArea.ShowLineNumbers {
		t.Error("expected ShowLineNumbers = false")
	}
	if m.textArea.MaxHeight != 3 {
		t.Errorf("expected MaxHeight = 3, got %d", m.textArea.MaxHeight)
	}
	if m.textArea.CharLimit != 256 {
		t.Errorf("expected CharLimit = 256, got %d", m.textArea.CharLimit)
	}
	if m.textArea.Prompt != "> " {
		t.Errorf("expected Prompt = %q, got %q", "> ", m.textArea.Prompt)
	}
	if m.textArea.Placeholder != "describe the command you need..." {
		t.Errorf("expected default placeholder, got %q", m.textArea.Placeholder)
	}
}

func TestNewModel_TextAreaInitialQuery(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	initialQuery := "list running containers"
	m := NewModel(cfg, initialQuery, false, "")

	if m.textArea.Value() != initialQuery {
		t.Errorf("expected textarea value = %q, got %q", initialQuery, m.textArea.Value())
	}
}

func TestNewModel_TextAreaEmptyInitialQuery(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")

	if m.textArea.Value() != "" {
		t.Errorf("expected empty textarea value, got %q", m.textArea.Value())
	}
}

func TestEnterKey_SubmitsInStateInput(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "list files", false, "")

	// Enter in stateInput should transition to stateLoading (submission)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateLoading {
		t.Errorf("expected state = stateLoading after Enter in stateInput, got %d", model.state)
	}
	if cmd == nil {
		t.Error("expected non-nil command after Enter submission")
	}
	if model.originalQuery != "list files" {
		t.Errorf("expected originalQuery = %q, got %q", "list files", model.originalQuery)
	}
}

func TestEnterKey_SelectsInStateSelect(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")
	m.commands = []string{"ls -la", "ls -lah"}
	m.filtered = m.commands
	m.state = stateSelect
	m.cursor = 0

	// Enter in stateSelect should select the command and quit
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.selected != "ls -la" {
		t.Errorf("expected selected = %q, got %q", "ls -la", model.selected)
	}
	if !model.quitting {
		t.Error("expected quitting = true after Enter in stateSelect")
	}
}

func TestEnterKey_DoesNotInsertNewline(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "test query", false, "")

	// Simulate pressing Enter - should NOT add a newline to textarea value
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	// The model should have transitioned to loading (not stayed in input with newline)
	if model.state != stateLoading {
		t.Errorf("expected stateLoading, got %d", model.state)
	}

	// The textarea should not contain any newlines from Enter key
	if strings.Contains(model.textArea.Value(), "\n") {
		t.Error("Enter key should not insert newline into textarea")
	}
}

func TestEscKey_CancelsInAllStates(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	tests := []struct {
		name  string
		state state
	}{
		{"stateInput", stateInput},
		{"stateLoading", stateLoading},
		{"stateSelect", stateSelect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(cfg, "test", false, "")
			m.state = tt.state

			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
			model := updated.(Model)

			if !model.quitting {
				t.Errorf("expected quitting = true after Esc in %s", tt.name)
			}
		})
	}
}

func TestView_ContainsTextAreaContent(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "list files", false, "")
	// Set width so textarea renders properly
	m.width = 80
	m.textArea.SetWidth(78)

	view := m.View()

	// View should contain the textarea value
	if !strings.Contains(view, "list files") {
		t.Errorf("expected View to contain textarea value 'list files', got:\n%s", view)
	}
}

func TestView_StateSelect_ShowsCommands(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")
	m.commands = []string{"ls -la", "ls -lah", "ls -l"}
	m.filtered = m.commands
	m.state = stateSelect
	m.cursor = 0
	m.width = 80
	m.maxHeight = 10
	m.textArea.SetWidth(78)

	view := m.View()

	// View should contain the commands
	if !strings.Contains(view, "ls -la") {
		t.Errorf("expected View to contain 'ls -la', got:\n%s", view)
	}
	if !strings.Contains(view, "ls -lah") {
		t.Errorf("expected View to contain 'ls -lah', got:\n%s", view)
	}
	// View should contain the counter
	if !strings.Contains(view, "3/3") {
		t.Errorf("expected View to contain counter '3/3', got:\n%s", view)
	}
}

func TestView_StateLoading_ShowsSpinner(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "test", false, "")
	m.state = stateLoading
	m.width = 80
	m.textArea.SetWidth(78)

	view := m.View()

	if !strings.Contains(view, "Generating commands...") {
		t.Errorf("expected View to contain 'Generating commands...', got:\n%s", view)
	}
}

func TestView_Quitting_ReturnsEmpty(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")
	m.quitting = true
	m.selected = ""

	view := m.View()

	if view != "" {
		t.Errorf("expected empty View when quitting with no selection, got:\n%s", view)
	}
}

func TestView_StateInput_ShowsError(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")
	m.state = stateInput
	m.err = fmt.Errorf("test error")
	m.width = 80
	m.textArea.SetWidth(78)

	view := m.View()

	if !strings.Contains(view, "test error") {
		t.Errorf("expected View to contain error message, got:\n%s", view)
	}
}

func TestResult_ReturnsCorrectQueryFromTextArea(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")
	m.textArea.SetValue("find large files")

	result := m.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected CancelledResult when in stateInput with no selection")
	}
	if cancelled.Query != "find large files" {
		t.Errorf("expected Query = %q, got %q", "find large files", cancelled.Query)
	}
}

func TestResult_SelectedResultFromSelection(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")
	m.selected = "docker ps -a"

	result := m.Result()
	selected, ok := result.(SelectedResult)
	if !ok {
		t.Fatal("expected SelectedResult when command is selected")
	}
	if selected.Command != "docker ps -a" {
		t.Errorf("expected Command = %q, got %q", "docker ps -a", selected.Command)
	}
}

func TestResult_OriginalQueryInSelectState(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}

	m := NewModel(cfg, "", false, "")
	m.state = stateSelect
	m.originalQuery = "list containers"
	m.textArea.SetValue("filter text")

	result := m.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected CancelledResult in select state with no selection")
	}
	// Should return original query, not the filter text
	if cancelled.Query != "list containers" {
		t.Errorf("expected Query = %q, got %q", "list containers", cancelled.Query)
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
