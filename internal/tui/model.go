package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/evgfitil/qx/internal/config"
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
	reservedLines    = 3  // lines reserved for prompt and counter
	minHeight        = 5  // minimum TUI height
)

// commandsMsg is sent when LLM returns commands
type commandsMsg struct {
	commands []string
	err      error
}

// Model represents the TUI state
type Model struct {
	state     state
	textInput textinput.Model
	spinner   spinner.Model
	commands []string
	cursor   int
	filtered []string
	selected  string
	err       error
	llmConfig llm.Config
	width     int
	height    int
	maxHeight int
	quitting  bool
}

// NewModel creates a new TUI model with optional initial query
func NewModel(cfg llm.Config, initialQuery string) Model {
	ti := textinput.New()
	ti.Placeholder = "describe the command you need..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.Prompt = "> "
	ti.PromptStyle = promptStyle()
	ti.Cursor.Style = cursorStyle()

	if initialQuery != "" {
		ti.SetValue(initialQuery)
		ti.CursorEnd()
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = loadingStyle()

	return Model{
		state:     stateInput,
		textInput: ti,
		spinner:   s,
		llmConfig: cfg,
		maxHeight: 10,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.maxHeight = max(msg.Height*maxHeightPercent/100, minHeight)
		m.textInput.Width = msg.Width - 4

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
		m.textInput.SetValue("")
		m.textInput.Placeholder = "filter results..."
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
		m.textInput, cmd = m.textInput.Update(msg)
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
		query := strings.TrimSpace(m.textInput.Value())
		if query == "" {
			return m, nil
		}
		m.state = stateLoading
		return m, tea.Batch(
			m.spinner.Tick,
			generateCommands(query, m.llmConfig),
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
	filter := strings.ToLower(m.textInput.Value())
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
	b.WriteString(m.textInput.View())
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

// Selected returns the selected command
func (m Model) Selected() string {
	return m.selected
}

func generateCommands(query string, cfg llm.Config) tea.Cmd {
	return func() tea.Msg {
		provider, err := llm.NewProvider(cfg)
		if err != nil {
			return commandsMsg{err: err}
		}

		ctx, cancel := context.WithTimeout(context.Background(), config.DefaultTimeout)
		defer cancel()

		commands, err := provider.Generate(ctx, query, cfg.Count)
		return commandsMsg{commands: commands, err: err}
	}
}
