package tui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestErrorModelView(t *testing.T) {
	err := errors.New("API key not configured")
	m := errorModel{err: err, initialQuery: "test"}

	view := m.View()

	if !strings.Contains(view, "API key not configured") {
		t.Errorf("View should contain error message, got: %s", view)
	}
	if !strings.Contains(view, "Press any key to dismiss") {
		t.Errorf("View should contain dismiss hint, got: %s", view)
	}
}

func TestErrorModelQuitOnKeyPress(t *testing.T) {
	err := errors.New("test error")
	m := errorModel{err: err, initialQuery: "hello"}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}

	// Verify model is unchanged
	em := updated.(errorModel)
	if em.initialQuery != "hello" {
		t.Errorf("expected initialQuery 'hello', got %q", em.initialQuery)
	}
}

func TestErrorModelIgnoresNonKeyMessages(t *testing.T) {
	err := errors.New("test error")
	m := errorModel{err: err, initialQuery: "hello"}

	_, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	if cmd != nil {
		t.Error("expected nil command for non-key message")
	}
}

func TestErrorModelInit(t *testing.T) {
	m := errorModel{err: errors.New("test"), initialQuery: "q"}

	cmd := m.Init()
	if cmd != nil {
		t.Error("expected nil init command")
	}
}
