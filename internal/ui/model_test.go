package ui

import (
	"bytes"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	if m.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d)", m.state, stateInput)
	}
	if m.originalQuery != "" {
		t.Errorf("originalQuery = %q, want empty (set only on Enter)", m.originalQuery)
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

func TestNewModelInitReturnsBlinkWithQuery(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "test",
		Theme:        DefaultTheme(),
	})
	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() returned nil, want textarea.Blink command")
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
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.state = stateLoading

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

func TestEnterWithNonEmptyQuery(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.textArea.SetValue("list files")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateLoading {
		t.Errorf("state = %d, want stateLoading (%d)", model.state, stateLoading)
	}
	if model.originalQuery != "list files" {
		t.Errorf("originalQuery = %q, want %q", model.originalQuery, "list files")
	}
	if cmd == nil {
		t.Error("cmd = nil, want batch of spinner.Tick + generateCommands")
	}
}

func TestEnterWithPrefilledQuery(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "list files",
		Theme:        DefaultTheme(),
	})

	if m.state != stateInput {
		t.Fatalf("precondition: state = %d, want stateInput", m.state)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateLoading {
		t.Errorf("state = %d, want stateLoading (%d)", model.state, stateLoading)
	}
	if model.originalQuery != "list files" {
		t.Errorf("originalQuery = %q, want %q", model.originalQuery, "list files")
	}
	if cmd == nil {
		t.Error("cmd = nil, want batch of spinner.Tick + generateCommands")
	}
}

func TestEnterWithEmptyQuery(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d)", model.state, stateInput)
	}
	if cmd != nil {
		t.Error("cmd should be nil for empty query")
	}
}

func TestEnterWithWhitespaceOnlyQuery(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.textArea.SetValue("   ")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d)", model.state, stateInput)
	}
	if cmd != nil {
		t.Error("cmd should be nil for whitespace-only query")
	}
}

func TestEnterWithSecretDetected(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.textArea.SetValue("use api_key=abcdefghijklmnopqrstuvwxyz1234")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d) when secret detected", model.state, stateInput)
	}
	if model.err == nil {
		t.Error("err = nil, want guard error for detected secret")
	}
}

func TestEnterWithSecretDetectedForceSend(t *testing.T) {
	m := newModel(RunOptions{
		Theme:     DefaultTheme(),
		ForceSend: true,
	})
	m.textArea.SetValue("use api_key=abcdefghijklmnopqrstuvwxyz1234")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateLoading {
		t.Errorf("state = %d, want stateLoading (%d) with forceSend", model.state, stateLoading)
	}
	if model.err != nil {
		t.Errorf("err = %v, want nil with forceSend", model.err)
	}
	if cmd == nil {
		t.Error("cmd = nil, want batch command for loading")
	}
}

func TestEnterClearsError(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.err = errTest
	m.textArea.SetValue("list files")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.err != nil {
		t.Errorf("err = %v, want nil after successful submit", model.err)
	}
	if model.state != stateLoading {
		t.Errorf("state = %d, want stateLoading (%d)", model.state, stateLoading)
	}
}

func TestInputViewShowsTextarea(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.textArea.SetValue("test query")

	view := m.View()

	if view == "" {
		t.Error("View() returned empty string, want textarea content")
	}
	if !strings.Contains(view, "test query") {
		t.Errorf("View() does not contain query text")
	}
}

func TestInputViewShowsError(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.err = errTest

	view := m.View()

	if !strings.Contains(view, "test error") {
		t.Errorf("View() does not contain error message")
	}
}

func TestInputViewHidesErrorWhenNil(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	view := m.View()

	if strings.Contains(view, "Error:") {
		t.Errorf("View() should not contain Error: when err is nil")
	}
}

func TestEnterDuringLoadingIgnored(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.state = stateLoading

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateLoading {
		t.Errorf("state = %d, want stateLoading (%d) — Enter should be ignored", model.state, stateLoading)
	}
	if cmd != nil {
		t.Error("cmd should be nil when Enter is ignored during loading")
	}
}

func TestLoadingViewShowsSpinnerText(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.state = stateLoading

	view := m.View()

	if !strings.Contains(view, "Generating commands...") {
		t.Errorf("loading View() should contain 'Generating commands...', got %q", view)
	}
}

func TestLoadingViewNotEmpty(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.state = stateLoading

	view := m.View()

	if view == "" {
		t.Error("loading View() should not be empty")
	}
}

func TestCommandsMsgSuccessResetsScrollOffset(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "test",
		Theme:        DefaultTheme(),
	})
	m.scrollOffset = 5

	cmds := []string{"cmd1", "cmd2"}
	updated, _ := m.Update(commandsMsg{commands: cmds})
	model := updated.(Model)

	if model.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want 0 after commandsMsg", model.scrollOffset)
	}
}

func TestCommandsMsgClearsTextarea(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "list files",
		Theme:        DefaultTheme(),
	})

	cmds := []string{"ls -la", "find . -type f"}
	updated, _ := m.Update(commandsMsg{commands: cmds})
	model := updated.(Model)

	if model.textArea.Value() != "" {
		t.Errorf("textArea value = %q, want empty (should be cleared for filtering)", model.textArea.Value())
	}
	if model.prevFilter != "" {
		t.Errorf("prevFilter = %q, want empty", model.prevFilter)
	}
}

func TestCommandsMsgErrorPreservesQuery(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "my query",
		Theme:        DefaultTheme(),
	})

	// Press Enter to set originalQuery and transition to stateLoading
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	// Simulate error from LLM
	updated, _ = model.Update(commandsMsg{err: errTest})
	model = updated.(Model)

	if model.originalQuery != "my query" {
		t.Errorf("originalQuery = %q, want %q after error", model.originalQuery, "my query")
	}
}

// --- Selector state tests (Task 7) ---

func newSelectModel(commands []string) Model {
	m := newModel(RunOptions{
		InitialQuery: "test",
		Theme:        DefaultTheme(),
	})
	updated, _ := m.Update(commandsMsg{commands: commands})
	return updated.(Model)
}

func TestSelectorNavigationDown(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(Model)

	if model.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after Down", model.cursor)
	}
}

func TestSelectorNavigationUp(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})
	m.cursor = 2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	model := updated.(Model)

	if model.cursor != 1 {
		t.Errorf("cursor = %d, want 1 after Up", model.cursor)
	}
}

func TestSelectorNavigationUpAtTop(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	model := updated.(Model)

	if model.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (should stay at top)", model.cursor)
	}
}

func TestSelectorNavigationDownAtBottom(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})
	m.cursor = 2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(Model)

	if model.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (should stay at bottom)", model.cursor)
	}
}

func TestSelectorNavigationMultipleDown(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})

	for i := 0; i < 5; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = result.(Model)
	}

	if m.cursor != 2 {
		t.Errorf("cursor = %d, want 2 (clamped to last item)", m.cursor)
	}
}

func TestSelectorFilter(t *testing.T) {
	m := newSelectModel([]string{"ls -la", "find . -name '*.go'", "grep error log.txt"})

	m.textArea.SetValue("find")
	// Trigger filter by simulating a text change
	m.prevFilter = ""
	m.applyFilter()

	if len(m.filtered) != 1 {
		t.Errorf("filtered count = %d, want 1 for filter 'find'", len(m.filtered))
	}
	if m.filtered[0] != "find . -name '*.go'" {
		t.Errorf("filtered[0] = %q, want %q", m.filtered[0], "find . -name '*.go'")
	}
}

func TestSelectorFilterCaseInsensitive(t *testing.T) {
	m := newSelectModel([]string{"LS -LA", "find files", "GREP pattern"})

	m.textArea.SetValue("grep")
	m.applyFilter()

	if len(m.filtered) != 1 {
		t.Errorf("filtered count = %d, want 1 for case-insensitive filter 'grep'", len(m.filtered))
	}
	if m.filtered[0] != "GREP pattern" {
		t.Errorf("filtered[0] = %q, want %q", m.filtered[0], "GREP pattern")
	}
}

func TestSelectorFilterResetsCursor(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})
	m.cursor = 2

	m.textArea.SetValue("cmd")
	m.applyFilter()

	if m.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after filter change", m.cursor)
	}
}

func TestSelectorFilterEmptyShowsAll(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})

	m.textArea.SetValue("")
	m.applyFilter()

	if len(m.filtered) != 3 {
		t.Errorf("filtered count = %d, want 3 for empty filter", len(m.filtered))
	}
}

func TestSelectorFilterNoMatch(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})

	m.textArea.SetValue("zzz")
	m.applyFilter()

	if len(m.filtered) != 0 {
		t.Errorf("filtered count = %d, want 0 for non-matching filter", len(m.filtered))
	}
}

func TestSelectorEnterSelectsCommand(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})
	m.cursor = 1

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateDone {
		t.Errorf("state = %d, want stateDone (%d)", model.state, stateDone)
	}
	if model.selected != "cmd2" {
		t.Errorf("selected = %q, want %q", model.selected, "cmd2")
	}
	if cmd == nil {
		t.Error("cmd = nil, want tea.Quit")
	}
}

func TestSelectorEnterWithEmptyFilteredList(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2"})
	m.filtered = nil
	m.filteredIdx = nil

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateSelect {
		t.Errorf("state = %d, want stateSelect (should stay when nothing to select)", model.state)
	}
	if model.selected != "" {
		t.Errorf("selected = %q, want empty", model.selected)
	}
	if cmd != nil {
		t.Error("cmd should be nil when nothing to select")
	}
}

func TestSelectorEnterAfterFilter(t *testing.T) {
	m := newSelectModel([]string{"ls -la", "find . -name '*.go'", "grep error log.txt"})

	m.textArea.SetValue("grep")
	m.applyFilter()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateDone {
		t.Errorf("state = %d, want stateDone (%d)", model.state, stateDone)
	}
	if model.selected != "grep error log.txt" {
		t.Errorf("selected = %q, want %q", model.selected, "grep error log.txt")
	}
	if cmd == nil {
		t.Error("cmd = nil, want tea.Quit")
	}
}

func TestAutoSelectSingleCommand(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "test",
		Theme:        DefaultTheme(),
	})

	updated, cmd := m.Update(commandsMsg{commands: []string{"only-cmd"}})
	model := updated.(Model)

	if model.state != stateDone {
		t.Errorf("state = %d, want stateDone (%d) for auto-select", model.state, stateDone)
	}
	if model.selected != "only-cmd" {
		t.Errorf("selected = %q, want %q", model.selected, "only-cmd")
	}
	if cmd == nil {
		t.Error("cmd = nil, want tea.Quit for auto-select")
	}
}

func TestAutoSelectDoesNotTriggerForMultipleCommands(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "test",
		Theme:        DefaultTheme(),
	})

	updated, cmd := m.Update(commandsMsg{commands: []string{"cmd1", "cmd2"}})
	model := updated.(Model)

	if model.state != stateSelect {
		t.Errorf("state = %d, want stateSelect (%d)", model.state, stateSelect)
	}
	if model.selected != "" {
		t.Errorf("selected = %q, want empty", model.selected)
	}
	if cmd != nil {
		t.Error("cmd should be nil (no auto-quit)")
	}
}

func TestSelectorResultExtraction(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2"})
	m.cursor = 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	result := model.Result()
	selected, ok := result.(SelectedResult)
	if !ok {
		t.Fatalf("result type = %T, want SelectedResult", result)
	}
	if selected.Command != "cmd2" {
		t.Errorf("Command = %q, want %q", selected.Command, "cmd2")
	}
}

func TestSelectorCancelledResult(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2"})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	result := model.Result()
	if _, ok := result.(CancelledResult); !ok {
		t.Fatalf("result type = %T, want CancelledResult", result)
	}
}

func TestSelectorModeEnterSetsIndex(t *testing.T) {
	items := []string{"item-a", "item-b", "item-c"}
	m := newSelectorModel(items, func(i int) string { return items[i] }, DefaultTheme())
	m.cursor = 2

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.selectedIndex != 2 {
		t.Errorf("selectedIndex = %d, want 2", model.selectedIndex)
	}
	if model.selected != "item-c" {
		t.Errorf("selected = %q, want %q", model.selected, "item-c")
	}
	if cmd == nil {
		t.Error("cmd = nil, want tea.Quit")
	}
}

func TestSelectorModeFilteredEnterSetsCorrectIndex(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	m := newSelectorModel(items, func(i int) string { return items[i] }, DefaultTheme())

	m.textArea.SetValue("cherry")
	m.applyFilter()

	if len(m.filtered) != 1 {
		t.Fatalf("filtered count = %d, want 1", len(m.filtered))
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.selectedIndex != 2 {
		t.Errorf("selectedIndex = %d, want 2 (original index of cherry)", model.selectedIndex)
	}
}

func TestSelectorViewContainsCommands(t *testing.T) {
	m := newSelectModel([]string{"ls -la", "find .", "tree"})
	m.width = 80
	m.maxHeight = 10

	view := m.View()

	if !strings.Contains(view, "ls -la") {
		t.Error("View() should contain first command")
	}
	if !strings.Contains(view, "find .") {
		t.Error("View() should contain second command")
	}
	if !strings.Contains(view, "tree") {
		t.Error("View() should contain third command")
	}
}

func TestSelectorViewContainsPointer(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2"})
	m.width = 80
	m.maxHeight = 10

	view := m.View()

	if !strings.Contains(view, DefaultTheme().Pointer) {
		t.Errorf("View() should contain pointer %q", DefaultTheme().Pointer)
	}
}

func TestSelectorViewContainsCounter(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})
	m.width = 80
	m.maxHeight = 10

	view := m.View()

	if !strings.Contains(view, "3/3") {
		t.Error("View() should contain counter '3/3'")
	}
}

func TestSelectorViewFilteredCounter(t *testing.T) {
	m := newSelectModel([]string{"ls -la", "find . -name '*.go'", "grep error"})
	m.width = 80
	m.maxHeight = 10
	m.textArea.SetValue("find")
	m.applyFilter()

	view := m.View()

	if !strings.Contains(view, "1/3") {
		t.Error("View() should contain counter '1/3' after filtering")
	}
}

func TestSelectorScrollOffset(t *testing.T) {
	commands := make([]string, 20)
	for i := range commands {
		commands[i] = "cmd" + string(rune('a'+i))
	}
	m := newSelectModel(commands)
	m.maxHeight = 6 // reservedLines=4, so visible=2

	// Move cursor down past visible area
	for i := 0; i < 5; i++ {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = result.(Model)
	}

	if m.cursor != 5 {
		t.Errorf("cursor = %d, want 5", m.cursor)
	}
	if m.scrollOffset < 1 {
		t.Errorf("scrollOffset = %d, should be > 0 when cursor exceeds visible area", m.scrollOffset)
	}
}

func TestSelectorScrollOffsetUp(t *testing.T) {
	commands := make([]string, 20)
	for i := range commands {
		commands[i] = "cmd" + string(rune('a'+i))
	}
	m := newSelectModel(commands)
	m.maxHeight = 6 // visible = 2
	m.cursor = 10
	m.scrollOffset = 9

	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = result.(Model)

	if m.cursor != 9 {
		t.Errorf("cursor = %d, want 9", m.cursor)
	}
	if m.scrollOffset != 9 {
		t.Errorf("scrollOffset = %d, want 9", m.scrollOffset)
	}
}

func TestVisibleItemCount(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.maxHeight = 10

	got := m.visibleItemCount()
	want := 10 - reservedLines
	if got != want {
		t.Errorf("visibleItemCount() = %d, want %d", got, want)
	}
}

func TestVisibleItemCountMinimumOne(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.maxHeight = 3 // less than reservedLines

	got := m.visibleItemCount()
	if got < 1 {
		t.Errorf("visibleItemCount() = %d, want at least 1", got)
	}
}

func TestNavigationWithEmptyFilteredList(t *testing.T) {
	m := newSelectModel([]string{"cmd1"})
	m.filtered = nil
	m.filteredIdx = nil

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(Model)

	if model.cursor != 0 {
		t.Errorf("cursor = %d, want 0 with empty filtered list", model.cursor)
	}
}

func TestSelectorFilterWithDisplayFn(t *testing.T) {
	items := []string{"a", "b", "c"}
	display := func(i int) string {
		names := []string{"alpha-item", "beta-item", "charlie-item"}
		return names[i]
	}
	m := newSelectorModel(items, display, DefaultTheme())

	m.textArea.SetValue("beta")
	m.applyFilter()

	if len(m.filtered) != 1 {
		t.Errorf("filtered count = %d, want 1 for filter 'beta'", len(m.filtered))
	}
	if m.filteredIdx[0] != 1 {
		t.Errorf("filteredIdx[0] = %d, want 1 (original index of 'b')", m.filteredIdx[0])
	}
}

func TestGetDisplayText(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2", "cmd3"})

	got := m.getDisplayText(1)
	if got != "cmd2" {
		t.Errorf("getDisplayText(1) = %q, want %q", got, "cmd2")
	}
}

func TestGetDisplayTextSelectorMode(t *testing.T) {
	items := []string{"a", "b", "c"}
	display := func(i int) string {
		names := []string{"alpha", "beta", "charlie"}
		return names[i]
	}
	m := newSelectorModel(items, display, DefaultTheme())

	got := m.getDisplayText(1)
	if got != "beta" {
		t.Errorf("getDisplayText(1) = %q, want %q", got, "beta")
	}
}

// --- Done state and result extraction tests (Task 8) ---

func TestDoneStateViewReturnsEmpty(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.state = stateDone
	m.quitting = true

	view := m.View()

	if view != "" {
		t.Errorf("View() = %q, want empty string in done state", view)
	}
}

func TestDoneStateViewAfterSelection(t *testing.T) {
	m := newSelectModel([]string{"cmd1", "cmd2"})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	view := model.View()

	if view != "" {
		t.Errorf("View() = %q, want empty string after selection", view)
	}
}

func TestDoneStateViewAfterCancel(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	view := model.View()

	if view != "" {
		t.Errorf("View() = %q, want empty string after cancel", view)
	}
}

func TestResultSelectedFromInput(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "list files",
		Theme:        DefaultTheme(),
	})

	// Press Enter to submit the pre-filled query (sets originalQuery)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	// Simulate receiving commands and selecting one
	updated, _ = model.Update(commandsMsg{commands: []string{"ls -la", "find ."}})
	model = updated.(Model)
	model.cursor = 0
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	result := model.Result()
	selected, ok := result.(SelectedResult)
	if !ok {
		t.Fatalf("result type = %T, want SelectedResult", result)
	}
	if selected.Command != "ls -la" {
		t.Errorf("Command = %q, want %q", selected.Command, "ls -la")
	}
	if selected.Query != "list files" {
		t.Errorf("Query = %q, want %q", selected.Query, "list files")
	}
}

func TestResultCancelledFromInput(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})
	m.textArea.SetValue("test query")

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatalf("result type = %T, want CancelledResult", result)
	}
	if cancelled.Query != "test query" {
		t.Errorf("Query = %q, want %q (should fall back to textarea value)", cancelled.Query, "test query")
	}
}

func TestResultCancelledFromInputWithInitialQuery(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "find . -type f",
		Theme:        DefaultTheme(),
	})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatalf("result type = %T, want CancelledResult", result)
	}
	if cancelled.Query != "find . -type f" {
		t.Errorf("Query = %q, want %q (should preserve pre-filled query for shell buffer restore)", cancelled.Query, "find . -type f")
	}
}

func TestResultCancelledFromLoading(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "my query",
		Theme:        DefaultTheme(),
	})

	// Press Enter to submit query (sets originalQuery, transitions to stateLoading)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	// Cancel during loading
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)

	result := model.Result()
	cancelled, ok := result.(CancelledResult)
	if !ok {
		t.Fatalf("result type = %T, want CancelledResult", result)
	}
	if cancelled.Query != "my query" {
		t.Errorf("Query = %q, want %q", cancelled.Query, "my query")
	}
}

func TestResultAutoSelectSetsQuery(t *testing.T) {
	m := newModel(RunOptions{
		InitialQuery: "find go files",
		Theme:        DefaultTheme(),
	})

	// Press Enter to submit query (sets originalQuery)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	// Simulate single command result (auto-select)
	updated, _ = model.Update(commandsMsg{commands: []string{"find . -name '*.go'"}})
	model = updated.(Model)

	result := model.Result()
	selected, ok := result.(SelectedResult)
	if !ok {
		t.Fatalf("result type = %T, want SelectedResult", result)
	}
	if selected.Command != "find . -name '*.go'" {
		t.Errorf("Command = %q, want %q", selected.Command, "find . -name '*.go'")
	}
	if selected.Query != "find go files" {
		t.Errorf("Query = %q, want %q", selected.Query, "find go files")
	}
}

func TestRunSelectorResultIndex(t *testing.T) {
	items := []string{"item-a", "item-b", "item-c"}
	m := newSelectorModel(items, func(i int) string { return items[i] }, DefaultTheme())
	m.cursor = 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.selectedIndex != 1 {
		t.Errorf("selectedIndex = %d, want 1", model.selectedIndex)
	}
}

func TestRunSelectorCancelledIndex(t *testing.T) {
	items := []string{"item-a", "item-b"}
	m := newSelectorModel(items, func(i int) string { return items[i] }, DefaultTheme())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	if model.selectedIndex != -1 {
		t.Errorf("selectedIndex = %d, want -1 (cancelled)", model.selectedIndex)
	}
}

func TestEmptyCommandList(t *testing.T) {
	m := newModel(RunOptions{InitialQuery: "test query", Theme: DefaultTheme()})
	m.state = stateLoading

	updated, _ := m.Update(commandsMsg{commands: []string{}})
	model := updated.(Model)

	if model.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d)", model.state, stateInput)
	}
	if model.err == nil {
		t.Error("err = nil, want 'no commands generated' error")
	}
}

func TestEmptyCommandListNil(t *testing.T) {
	m := newModel(RunOptions{InitialQuery: "test query", Theme: DefaultTheme()})
	m.state = stateLoading

	updated, _ := m.Update(commandsMsg{commands: nil})
	model := updated.(Model)

	if model.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d)", model.state, stateInput)
	}
	if model.err == nil {
		t.Error("err = nil, want 'no commands generated' error")
	}
}

func TestEmptyCommandListEnterDoesNothing(t *testing.T) {
	m := newModel(RunOptions{InitialQuery: "test query", Theme: DefaultTheme()})
	m.state = stateSelect
	m.commands = []string{}
	m.filtered = []string{}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.state != stateSelect {
		t.Errorf("state = %d, want stateSelect (%d) — Enter should be no-op with no items", model.state, stateSelect)
	}
	if cmd != nil {
		t.Errorf("cmd = %v, want nil", cmd)
	}
}

func TestEmptyCommandListEscCancels(t *testing.T) {
	m := newModel(RunOptions{InitialQuery: "test query", Theme: DefaultTheme()})
	m.state = stateSelect
	m.commands = []string{}
	m.filtered = []string{}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)

	if model.state != stateDone {
		t.Errorf("state = %d, want stateDone (%d)", model.state, stateDone)
	}
	if model.selected != "" {
		t.Errorf("selected = %q, want empty (cancelled)", model.selected)
	}
}

func TestNewModelWithRendererAwareTheme(t *testing.T) {
	r := lipgloss.NewRenderer(&bytes.Buffer{})
	theme := DefaultTheme().WithRenderer(r)
	m := newModel(RunOptions{Theme: theme})

	if m.state != stateInput {
		t.Errorf("state = %d, want stateInput (%d)", m.state, stateInput)
	}
	if m.theme.renderer != r {
		t.Error("model theme should preserve the renderer")
	}
}

func TestNewSelectorModelWithRendererAwareTheme(t *testing.T) {
	r := lipgloss.NewRenderer(&bytes.Buffer{})
	theme := DefaultTheme().WithRenderer(r)
	items := []string{"cmd1", "cmd2"}
	m := newSelectorModel(items, func(i int) string { return items[i] }, theme)

	if m.state != stateSelect {
		t.Errorf("state = %d, want stateSelect (%d)", m.state, stateSelect)
	}
	if m.theme.renderer != r {
		t.Error("selector model theme should preserve the renderer")
	}
}

// --- Textarea auto-resize tests (Bug 3) ---

func TestTextareaInitialHeight(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	if m.textArea.Height() != 1 {
		t.Errorf("initial Height() = %d, want 1", m.textArea.Height())
	}
}

func TestTextareaGrowsOnLongInput(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	// Set width so wrapping is predictable (textarea effective width ~15 chars)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 20, Height: 40})
	m = updated.(Model)

	// Set text that exceeds the effective width and triggers wrapping
	m.textArea.SetValue("this is a long text that should wrap to multiple lines")

	// Trigger update to run auto-resize
	updated, _ = m.Update(nil)
	m = updated.(Model)

	if m.textArea.Height() < 2 {
		t.Errorf("Height() = %d after long text, want >= 2", m.textArea.Height())
	}
}

func TestTextareaHeightCappedAtMaxHeight(t *testing.T) {
	m := newModel(RunOptions{Theme: DefaultTheme()})

	// Set narrow width so text wraps aggressively
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 10, Height: 40})
	m = updated.(Model)

	// Set very long text that would wrap to many lines
	m.textArea.SetValue("this is a very very very long text that should wrap to many many lines well beyond three")

	// Trigger update to run auto-resize
	updated, _ = m.Update(nil)
	m = updated.(Model)

	if m.textArea.Height() > 3 {
		t.Errorf("Height() = %d, want <= 3 (MaxHeight)", m.textArea.Height())
	}
}

var errTest = testError("test error")

type testError string

func (e testError) Error() string { return string(e) }
