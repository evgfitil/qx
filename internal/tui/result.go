package tui

// Result represents the outcome of TUI interaction.
type Result interface {
	IsCancelled() bool
}

// CancelledResult indicates user cancelled the operation (Esc/Ctrl+C).
// Query contains the current input text for restoration.
type CancelledResult struct {
	Query string
}

// IsCancelled returns true for CancelledResult.
func (r CancelledResult) IsCancelled() bool {
	return true
}

// SelectedResult indicates user selected a command.
type SelectedResult struct {
	Command string
}

// IsCancelled returns false for SelectedResult.
func (r SelectedResult) IsCancelled() bool {
	return false
}
