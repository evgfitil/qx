package tui

// Result represents the outcome of TUI interaction.
type Result interface {
	isResult()
}

// SelectedResult indicates user selected a command.
type SelectedResult struct {
	Command string
	Query   string
}

func (SelectedResult) isResult() {}

// CancelledResult indicates user cancelled the operation (Esc/Ctrl+C).
type CancelledResult struct {
	Query string
}

func (CancelledResult) isResult() {}
