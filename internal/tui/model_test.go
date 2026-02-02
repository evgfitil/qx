package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evgfitil/qx/internal/llm"
)

func TestModel_Result_CancelledOnEsc(t *testing.T) {
	cfg := llm.Config{
		BaseURL: "http://localhost",
		APIKey:  "test",
		Model:   "test",
		Count:   3,
	}
	initialQuery := "list files"

	m := NewModel(cfg, initialQuery, false)

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

	m := NewModel(cfg, "", false)

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

	m := NewModel(cfg, initialQuery, false)

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

	m := NewModel(cfg, "test query", false)

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
	m := NewModel(cfg, "", false)

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
	m := NewModel(cfg, initialQuery, false)

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
	m := NewModel(cfg, originalQuery, false)

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
	m := NewModel(cfg, originalQuery, false)

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
