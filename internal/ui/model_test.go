package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModelWithoutQuery(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	if m.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d)", m.state, stateInput)
	}
	if m.originalQuery != "" {
		t.Errorf("originalQuery = %q, want empty", m.originalQuery)
	}
	if m.selectedIndex != -1 {
		t.Errorf("selectedIndex = %d, want -1", m.selectedIndex)
	}
	if m.maxHeight != minHeight {
		t.Errorf("maxHeight = %d, want %d", m.maxHeight, minHeight)
	}
}

func TestNewModelWithQuery(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "list files",
		Theme:        DefaultTheme(),
	})

	if m.state != stateLoading {
		t.Errorf("state = %d, want stateLoading (%d)", m.state, stateLoading)
	}
	if m.originalQuery != "list files" {
		t.Errorf("originalQuery = %q, want %q", m.originalQuery, "list files")
	}
	if m.textArea.Value() != "list files" {
		t.Errorf("textArea value = %q, want %q", m.textArea.Value(), "list files")
	}
}

func TestNewModelInitReturnsBlinkForInput(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() returned nil, want textarea.Blink command")
	}
}

func TestNewModelInitReturnsTickForLoading(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "test",
		Theme:        DefaultTheme(),
	})
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() returned nil, want spinner.Tick command")
	}
}

func TestNewSelectorModel(t *testing.T) {
	items := []string{"cmd1", "cmd2", "cmd3"}
	display := func(i int) string { return items[i] }
	m := newSelectorModel(items, display, DefaultTheme())

	if m.state != stateSelect {
		t.Errorf("state = %d, want stateSelect (%d)", m.state, stateSelect)
	}
	if !m.selectorMode {
		t.Error("selectorMode = false, want true")
	}
	if len(m.items) != 3 {
		t.Errorf("items count = %d, want 3", len(m.items))
	}
	if len(m.filtered) != 3 {
		t.Errorf("filtered count = %d, want 3", len(m.filtered))
	}
	if m.selectedIndex != -1 {
		t.Errorf("selectedIndex = %d, want -1", m.selectedIndex)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)

	if model.width != 120 {
		t.Errorf("width = %d, want 120", model.width)
	}
	if model.height != 40 {
		t.Errorf("height = %d, want 40", model.height)
	}

	wantMaxHeight := max(40*maxHeightPercent/100, minHeight)
	if model.maxHeight != wantMaxHeight {
		t.Errorf("maxHeight = %d, want %d", model.maxHeight, wantMaxHeight)
	}
}

func TestWindowSizeMsgSmallTerminal(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	model := updated.(Model)

	wantMaxHeight := max(10*maxHeightPercent/100, minHeight)
	if model.maxHeight != wantMaxHeight {
		t.Errorf("maxHeight = %d, want %d (minHeight should apply)", model.maxHeight, wantMaxHeight)
	}
	if model.maxHeight < minHeight {
		t.Errorf("maxHeight = %d, should not be less than minHeight (%d)", model.maxHeight, minHeight)
	}
}

func TestEscQuits(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	if model.state != stateDone {
		t.Errorf("state = %d, want stateDone (%d)", model.state, stateDone)
	}
	if model.selected != "" {
		t.Errorf("selected = %q, want empty (cancelled)", model.selected)
	}
	if cmd == nil {
		t.Error("cmd = nil, want tea.Quit")
	}
}

func TestCtrlCQuits(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)

	if model.state != stateDone {
		t.Errorf("state = %d, want stateDone (%d)", model.state, stateDone)
	}
	if model.selected != "" {
		t.Errorf("selected = %q, want empty (cancelled)", model.selected)
	}
	if cmd == nil {
		t.Error("cmd = nil, want tea.Quit")
	}
}

func TestEscFromLoadingState(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "test query",
		Theme:        DefaultTheme(),
	})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	if model.state != stateDone {
		t.Errorf("state = %d, want stateDone (%d)", model.state, stateDone)
	}
	if cmd == nil {
		t.Error("cmd = nil, want tea.Quit")
	}
}

func TestEscFromSelectState(t *testing.T) {
	items := []string{"cmd1", "cmd2"}
	m := newSelectorModel(items, func(i int) string { return items[i] }, DefaultTheme())

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	if model.state != stateDone {
		t.Errorf("state = %d, want stateDone (%d)", model.state, stateDone)
	}
	if cmd == nil {
		t.Error("cmd = nil, want tea.Quit")
	}
}

func TestCommandsMsgSuccess(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "test",
		Theme:        DefaultTheme(),
	})

	cmds := []string{"ls -la", "find . -name '*.go'", "tree"}
	updated, _ := m.Update(commandsMsg{commands: cmds})
	model := updated.(Model)

	if model.state != stateSelect {
		t.Errorf("state = %d, want stateSelect (%d)", model.state, stateSelect)
	}
	if len(model.commands) != 3 {
		t.Errorf("commands count = %d, want 3", len(model.commands))
	}
	if len(model.filtered) != 3 {
		t.Errorf("filtered count = %d, want 3", len(model.filtered))
	}
	if model.cursor != 0 {
		t.Errorf("cursor = %d, want 0", model.cursor)
	}
}

func TestCommandsMsgError(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "test",
		Theme:        DefaultTheme(),
	})
	m.state = stateLoading

	updated, _ := m.Update(commandsMsg{err: errTest})
	model := updated.(Model)

	if model.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d)", model.state, stateInput)
	}
	if model.err == nil {
		t.Error("err = nil, want error")
	}
}

func TestMaxHeightCalculation(t *testing.T) {
	tests := []struct {
		name       string
		termHeight int
		want       int
	}{
		{"large terminal", 100, 40},
		{"medium terminal", 30, 12},
		{"small terminal uses minHeight", 10, minHeight},
		{"tiny terminal uses minHeight", 5, minHeight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newModel(RunOptions{Theme: DefaultTheme()})
			updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: tt.termHeight})
			model := updated.(Model)

			if model.maxHeight != tt.want {
				t.Errorf("maxHeight = %d, want %d", model.maxHeight, tt.want)
			}
		})
	}
}

var errTest = testError("test error")

type testError string

func (e testError) Error() string { return string(e) }
