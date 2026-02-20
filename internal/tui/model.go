package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/evgfitil/qx/internal/config"
	"github.com/evgfitil/qx/internal/guard"
	"github.com/evgfitil/qx/internal/llm"
)

type state int

const (
	stateInput state = iota
	stateLoading
	stateSelect
)

const (
	maxHeightPercent = 40 // percentage of terminal height for TUI
	reservedLines    = 3  // lines reserved for textarea (1 line in filter mode) and counter
	minHeight        = 5  // minimum TUI height
)

// commandsMsg is sent when LLM returns commands
type commandsMsg struct {
	commands []string
	err      error
}

// Model represents the TUI state
type Model struct {
	state         state
	textArea      textarea.Model
	spinner       spinner.Model
	commands      []string
	cursor        int
	filtered      []string
	selected      string
	err           error
	llmConfig     llm.Config
	forceSend     bool
	pipeContext   string
	width         int
	height        int
	maxHeight     int
	quitting      bool
	originalQuery string
}

// NewModel creates a new TUI model with optional initial query and pipe context
func NewModel(cfg llm.Config, initialQuery string, forceSend bool, pipeContext string) Model {
	ta := textarea.New()
	ta.Placeholder = "describe the command you need..."
	ta.ShowLineNumbers = false
	ta.MaxHeight = 3
	ta.SetHeight(3)
	ta.CharLimit = 256
	ta.Prompt = "> "
	ta.FocusedStyle.Prompt = promptStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.KeyMap.InsertNewline = key.NewBinding(key.WithKeys())
	ta.KeyMap.LineNext = key.NewBinding(key.WithKeys())
	ta.KeyMap.LinePrevious = key.NewBinding(key.WithKeys())
	ta.Focus()

	if initialQuery != "" {
		ta.SetValue(initialQuery)
		ta.CursorEnd()
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = loadingStyle()

	return Model{
		state:       stateInput,
		textArea:    ta,
		spinner:     s,
		llmConfig:   cfg,
		forceSend:   forceSend,
		pipeContext: pipeContext,
		maxHeight:   10,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// Update implements tea.Model
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
			m.quitting = true
			return m, tea.Quit

		case tea.KeyEnter:
			return m.handleEnter()

		case tea.KeyUp:
			if m.state == stateSelect && m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.state == stateSelect && m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
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
		m.state = stateSelect
		m.textArea.SetValue("")
		m.textArea.Placeholder = "filter results..."
		m.textArea.MaxHeight = 1
		m.textArea.SetHeight(1)
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

		if m.state == stateSelect {
			m = m.updateFilter()
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
		return m, tea.Batch(
			m.spinner.Tick,
			generateCommands(query, m.llmConfig, m.pipeContext),
		)

	case stateSelect:
		if len(m.filtered) > 0 {
			m.selected = m.filtered[m.cursor]
			m.quitting = true
			return m, tea.Quit
		}

	case stateLoading:
		// ignore Enter during loading
	}
	return m, nil
}

func (m Model) updateFilter() Model {
	filter := strings.ToLower(m.textArea.Value())
	if filter == "" {
		m.filtered = m.commands
	} else {
		m.filtered = nil
		for _, cmd := range m.commands {
			if strings.Contains(strings.ToLower(cmd), filter) {
				m.filtered = append(m.filtered, cmd)
			}
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	return m
}

func truncateWithEllipsis(s string, maxWidth int) string {
	if maxWidth <= 3 {
		return s
	}
	if len(s) <= maxWidth {
		return s
	}
	return s[:maxWidth-3] + "..."
}

func wrapCommand(s string, width int) string {
	if width <= 0 {
		return s
	}
	return lipgloss.NewStyle().Width(width).Render(s)
}

// View implements tea.Model
func (m Model) View() string {
	if m.quitting && m.selected == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.textArea.View())
	b.WriteString("\n")

	switch m.state {
	case stateInput:
		if m.err != nil {
			b.WriteString(errorStyle().Render(fmt.Sprintf("Error: %v", m.err)))
			b.WriteString("\n")
		}

	case stateLoading:
		b.WriteString(m.spinner.View())
		b.WriteString(" Generating commands...")
		b.WriteString("\n")

	case stateSelect:
		maxItems := max(m.maxHeight-reservedLines, 1)

		start := 0
		if m.cursor >= maxItems {
			start = m.cursor - maxItems + 1
		}
		end := min(start+maxItems, len(m.filtered))

		for i := start; i < end; i++ {
			cmd := m.filtered[i]
			if i == m.cursor {
				wrapped := wrapCommand("> "+cmd, m.width)
				b.WriteString(selectedStyle().Render(wrapped))
			} else {
				truncated := truncateWithEllipsis("  "+cmd, m.width)
				b.WriteString(normalStyle().Render(truncated))
			}
			b.WriteString("\n")
		}

		b.WriteString(counterStyle().Render(fmt.Sprintf("  %d/%d", len(m.filtered), len(m.commands))))
		b.WriteString("\n")
	}

	return b.String()
}

// Result returns the outcome of TUI interaction.
// Returns SelectedResult if a command was selected, CancelledResult otherwise.
func (m Model) Result() Result {
	if m.selected != "" {
		return SelectedResult{Command: m.selected, Query: m.originalQuery}
	}
	// In select/loading states, textArea may contain filter text or still has query,
	// but originalQuery always has the submitted query
	if (m.state == stateSelect || m.state == stateLoading) && m.originalQuery != "" {
		return CancelledResult{Query: m.originalQuery}
	}
	return CancelledResult{Query: m.textArea.Value()}
}

func generateCommands(query string, cfg llm.Config, pipeContext string) tea.Cmd {
	return func() tea.Msg {
		provider, err := llm.NewProvider(cfg)
		if err != nil {
			return commandsMsg{err: err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultTimeout)
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
