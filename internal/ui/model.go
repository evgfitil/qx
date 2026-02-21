package ui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
	cursor        int
	scrollOffset  int
	selected      string
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

	return Model{
		state:         stateSelect,
		theme:         theme,
		textArea:      ta,
		spinner:       s,
		maxHeight:     minHeight,
		selectorMode:  true,
		items:         items,
		filtered:      items,
		displayFn:     display,
		selectedIndex: -1,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if m.state == stateLoading {
		return m.spinner.Tick
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
		}

	case commandsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateInput
			return m, nil
		}
		m.commands = msg.commands
		m.filtered = msg.commands
		m.cursor = 0
		m.scrollOffset = 0
		m.state = stateSelect
		return m, nil

	case spinner.TickMsg:
		if m.state == stateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.
func (m Model) View() string {
	return ""
}

// Result returns the outcome of TUI interaction.
func (m Model) Result() Result {
	if m.selected != "" {
		return SelectedResult{Command: m.selected, Query: m.originalQuery}
	}
	return CancelledResult{Query: m.originalQuery}
}
