package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/evgfitil/qx/internal/guard"
	"github.com/evgfitil/qx/internal/llm"
)

type state int

const (
	stateInput state = iota
	stateLoading
	stateSelect
	stateDone
)

const (
	maxHeightPercent = 40
	minHeight        = 5
	reservedLines    = 4 // border top + textarea + counter + border bottom
	generateTimeout  = 60 * time.Second
)

// commandsMsg is sent when LLM returns generated commands.
type commandsMsg struct {
	commands []string
	err      error
}

// Model represents the unified TUI state machine.
type Model struct {
	state         state
	theme         Theme
	textArea      textarea.Model
	spinner       spinner.Model
	commands      []string
	filtered      []string
	filteredIdx   []int
	cursor        int
	scrollOffset  int
	selected      string
	prevFilter    string
	err           error
	llmConfig     llm.Config
	forceSend     bool
	pipeContext   string
	width         int
	height        int
	maxHeight     int
	originalQuery string
	quitting      bool

	// selector-only mode
	selectorMode  bool
	items         []string
	displayFn     func(int) string
	selectedIndex int
}

func newTextArea(prompt string, promptStyle lipgloss.Style) textarea.Model {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.CharLimit = 256
	ta.Prompt = prompt
	ta.FocusedStyle.Prompt = promptStyle
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys())
	ta.KeyMap.LineNext = key.NewBinding(key.WithKeys())
	ta.KeyMap.LinePrevious = key.NewBinding(key.WithKeys())
	ta.Focus()
	return ta
}

func newModel(opts RunOptions) Model {
	ta := newTextArea(opts.Theme.Prompt, opts.Theme.PromptStyle())
	ta.Placeholder = "describe the command you need..."
	ta.MaxHeight = 3
	ta.SetHeight(3)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = opts.Theme.MutedStyle()

	initialState := stateInput
	if opts.InitialQuery != "" {
		ta.SetValue(opts.InitialQuery)
		ta.CursorEnd()
		initialState = stateLoading
	}

	return Model{
		state:         initialState,
		theme:         opts.Theme,
		textArea:      ta,
		spinner:       s,
		llmConfig:     opts.LLMConfig,
		forceSend:     opts.ForceSend,
		pipeContext:   opts.PipeContext,
		maxHeight:     minHeight,
		originalQuery: opts.InitialQuery,
		selectedIndex: -1,
	}
}

func newSelectorModel(items []string, display func(int) string, theme Theme) Model {
	ta := newTextArea(theme.Prompt, theme.PromptStyle())
	ta.Placeholder = "filter..."
	ta.MaxHeight = 1
	ta.SetHeight(1)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.MutedStyle()

	idx := make([]int, len(items))
	for i := range items {
		idx[i] = i
	}

	return Model{
		state:         stateSelect,
		theme:         theme,
		textArea:      ta,
		spinner:       s,
		maxHeight:     minHeight,
		selectorMode:  true,
		items:         items,
		filtered:      items,
		filteredIdx:   idx,
		displayFn:     display,
		selectedIndex: -1,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if m.state == stateLoading {
		return tea.Batch(
			m.spinner.Tick,
			generateCommands(m.originalQuery, m.llmConfig, m.pipeContext),
		)
	}
	return textarea.Blink
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.maxHeight = max(msg.Height*maxHeightPercent/100, minHeight)
		m.textArea.SetWidth(msg.Width - 2)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.state = stateDone
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyUp:
			if m.state == stateSelect {
				m.moveCursor(-1)
				return m, nil
			}

		case tea.KeyDown:
			if m.state == stateSelect {
				m.moveCursor(1)
				return m, nil
			}
		}

	case commandsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateInput
			return m, nil
		}

		if len(msg.commands) == 0 {
			m.err = fmt.Errorf("no commands generated")
			m.state = stateInput
			return m, nil
		}

		m.commands = msg.commands
		m.filtered = msg.commands
		m.filteredIdx = make([]int, len(msg.commands))
		for i := range msg.commands {
			m.filteredIdx[i] = i
		}
		m.cursor = 0
		m.scrollOffset = 0

		if len(msg.commands) == 1 {
			m.selected = msg.commands[0]
			m.state = stateDone
			m.quitting = true
			return m, tea.Quit
		}

		m.textArea.SetValue("")
		m.textArea.Placeholder = "filter..."
		m.textArea.MaxHeight = 1
		m.textArea.SetHeight(1)
		m.prevFilter = ""
		m.state = stateSelect
		return m, nil

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.state == stateInput || m.state == stateSelect {
		var cmd tea.Cmd
		m.textArea, cmd = m.textArea.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.state == stateSelect {
		if current := m.textArea.Value(); current != m.prevFilter {
			m.prevFilter = current
			m.applyFilter()
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case stateInput:
		query := strings.TrimSpace(m.textArea.Value())
		if query == "" {
			return m, nil
		}

		if err := guard.CheckQuery(query, m.forceSend); err != nil {
			m.err = err
			return m, nil
		}

		m.state = stateLoading
		m.originalQuery = query
		m.err = nil
		return m, tea.Batch(
			m.spinner.Tick,
			generateCommands(query, m.llmConfig, m.pipeContext),
		)

	case stateSelect:
		if len(m.filtered) == 0 {
			return m, nil
		}
		if m.selectorMode {
			m.selectedIndex = m.filteredIdx[m.cursor]
		}
		m.selected = m.filtered[m.cursor]
		m.state = stateDone
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func generateCommands(query string, cfg llm.Config, pipeContext string) tea.Cmd {
	return func() tea.Msg {
		provider, err := llm.NewProvider(cfg)
		if err != nil {
			return commandsMsg{err: err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), generateTimeout)
		defer cancel()

		commands, err := provider.Generate(ctx, query, cfg.Count, pipeContext, nil)
		if err != nil {
			return commandsMsg{err: err}
		}

		for i, cmd := range commands {
			commands[i] = guard.SanitizeOutput(cmd)
		}

		return commandsMsg{commands: commands}
	}
}

func (m *Model) moveCursor(delta int) {
	if len(m.filtered) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	m.adjustScroll()
}

func (m *Model) adjustScroll() {
	visible := m.visibleItemCount()
	if visible <= 0 {
		return
	}
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+visible {
		m.scrollOffset = m.cursor - visible + 1
	}
}

func (m Model) visibleItemCount() int {
	visible := m.maxHeight - reservedLines
	if visible < 1 {
		visible = 1
	}
	return visible
}

func (m *Model) applyFilter() {
	query := strings.ToLower(strings.TrimSpace(m.textArea.Value()))
	m.filtered = nil
	m.filteredIdx = nil

	if m.selectorMode {
		for i := range m.items {
			display := m.displayFn(i)
			if query == "" || strings.Contains(strings.ToLower(display), query) {
				m.filtered = append(m.filtered, m.items[i])
				m.filteredIdx = append(m.filteredIdx, i)
			}
		}
	} else {
		for i, cmd := range m.commands {
			if query == "" || strings.Contains(strings.ToLower(cmd), query) {
				m.filtered = append(m.filtered, cmd)
				m.filteredIdx = append(m.filteredIdx, i)
			}
		}
	}

	m.cursor = 0
	m.scrollOffset = 0
}

func (m Model) getDisplayText(filteredIndex int) string {
	if m.selectorMode && m.displayFn != nil {
		return m.displayFn(m.filteredIdx[filteredIndex])
	}
	return m.filtered[filteredIndex]
}

// Result returns the outcome of TUI interaction.
func (m Model) Result() Result {
	if m.selected != "" {
		return SelectedResult{Command: m.selected, Query: m.originalQuery}
	}
	return CancelledResult{Query: m.originalQuery}
}
