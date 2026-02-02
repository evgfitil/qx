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
	if !result.IsCancelled() {
		t.Error("expected Result().IsCancelled() to be true after Esc")
	}

	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected result to be CancelledResult")
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
	if result.IsCancelled() {
		t.Error("expected Result().IsCancelled() to be false after selection")
	}

	selected, ok := result.(SelectedResult)
	if !ok {
		t.Fatal("expected result to be SelectedResult")
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
	if !result.IsCancelled() {
		t.Error("expected Result().IsCancelled() to be true after Ctrl+C")
	}

	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatal("expected result to be CancelledResult")
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
	if !result.IsCancelled() {
		t.Error("expected Result().IsCancelled() to be true when no action taken")
	}
}

func TestCancelledResult_IsCancelled(t *testing.T) {
	r := CancelledResult{Query: "test"}
	if !r.IsCancelled() {
		t.Error("CancelledResult.IsCancelled() should return true")
	}
}

func TestSelectedResult_IsCancelled(t *testing.T) {
	r := SelectedResult{Command: "ls"}
	if r.IsCancelled() {
		t.Error("SelectedResult.IsCancelled() should return false")
	}
}
