package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/evgfitil/qx/internal/action"
	"github.com/evgfitil/qx/internal/config"
	"github.com/evgfitil/qx/internal/guard"
	"github.com/evgfitil/qx/internal/history"
	"github.com/evgfitil/qx/internal/llm"
	"github.com/evgfitil/qx/internal/shell"
	"github.com/evgfitil/qx/internal/ui"
)

const ExitCodeCancelled = 130

var (
	Version          = "dev"
	shellIntegration string
	showConfig       bool
	queryFlag        string
	forceSend        bool
	lastFlag         bool
	historyFlag      bool
	continueFlag     bool
)

// ErrCancelled indicates user cancelled the operation.
var ErrCancelled = errors.New("operation cancelled")

// Overridable function references for testing.
var (
	shouldPromptFn     = action.ShouldPrompt
	promptActionFn     = action.PromptAction
	readRefinementFn   = action.ReadRefinement
	generateCommandsFn func(query string, pipeContext string, followUp *llm.FollowUpContext) error
	uiRunFn            = ui.Run
	uiRunSelectorFn    = ui.RunSelector
)

var rootCmd = &cobra.Command{
	Use:   "qx [query]",
	Short: "Generate shell commands using LLM",
	Long: `qx is a CLI tool that generates shell commands from natural language descriptions.
It uses LLM to generate multiple command variants and presents them in a fzf-style picker.

After selecting a command, choose an action: execute it, copy to clipboard, or print to stdout.

Pipe command output into qx to provide context for more precise command generation:
  ls -la | qx "delete files larger than 1GB"
  docker ps | qx "stop all nginx containers"`,
	Version:       Version,
	Args:          cobra.MaximumNArgs(1),
	RunE:          run,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	generateCommandsFn = generateCommands

	rootCmd.Flags().StringVar(&shellIntegration, "shell-integration", "", "output shell integration script (bash|zsh|fish)")
	rootCmd.Flags().BoolVar(&showConfig, "config", false, "show config file path")
	rootCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "initial query for TUI input (pre-fills the input field)")
	rootCmd.Flags().BoolVar(&forceSend, "force-send", false, "send query even if secrets detected")
	rootCmd.Flags().BoolVar(&lastFlag, "last", false, "show last selected command and open action menu")
	rootCmd.Flags().BoolVar(&historyFlag, "history", false, "browse command history with interactive picker")
	rootCmd.Flags().BoolVar(&continueFlag, "continue", false, "refine the last command with a new query")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	if showConfig {
		fmt.Println(config.Path())
		return nil
	}

	if shellIntegration != "" {
		return handleShellIntegration(shellIntegration)
	}

	flagCount := 0
	if lastFlag {
		flagCount++
	}
	if historyFlag {
		flagCount++
	}
	if continueFlag {
		flagCount++
	}
	if flagCount > 1 {
		return fmt.Errorf("--last, --history, and --continue are mutually exclusive")
	}

	if lastFlag {
		return runLast()
	}

	if historyFlag {
		return runHistory()
	}

	pipeContext, err := readStdin()
	if err != nil {
		return err
	}

	if pipeContext != "" {
		if err := guard.CheckQuery(pipeContext, forceSend); err != nil {
			return err
		}
	}

	if continueFlag {
		if len(args) == 0 {
			return fmt.Errorf("--continue requires a query argument")
		}
		return runContinue(args[0], pipeContext)
	}

	if len(args) == 0 {
		return runInteractive(queryFlag, pipeContext)
	}

	query := args[0]
	return generateCommands(query, pipeContext, nil)
}

func runInteractive(initialQuery string, pipeContext string) error {
	initialQuery = llm.UnformatCommand(initialQuery)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if initialQuery != "" {
			fmt.Println(initialQuery)
		}
		return ErrCancelled
	}

	result, err := uiRunFn(ui.RunOptions{
		InitialQuery: initialQuery,
		LLMConfig:    cfg.LLM.ToLLMConfig(),
		ForceSend:    forceSend,
		PipeContext:  pipeContext,
		Theme:        cfg.Theme.ToTheme(),
	})
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	switch r := result.(type) {
	case ui.CancelledResult:
		if r.Query != "" {
			fmt.Println(r.Query)
		}
		return ErrCancelled
	case ui.SelectedResult:
		if r.Command != "" {
			return handleSelectedCommand(r.Command, r.Query, pipeContext, cfg.ActionMenu)
		}
		return nil
	default:
		return fmt.Errorf("unexpected result type: %T", result)
	}
}

func handleShellIntegration(shellName string) error {
	script, err := shell.Script(shellName)
	if err != nil {
		return err
	}
	fmt.Print(script)
	return nil
}

// runLast loads the most recent history entry and opens the action menu on it.
func runLast() error {
	store, err := newHistoryStore()
	if err != nil {
		return fmt.Errorf("failed to access history: %w", err)
	}

	entry, err := store.Last()
	if err != nil {
		if errors.Is(err, history.ErrEmpty) {
			return fmt.Errorf("no history yet — run a query first")
		}
		return fmt.Errorf("failed to read history: %w", err)
	}

	return handleSelectedCommand(entry.Selected, entry.Query, entry.PipeContext, true)
}

// runHistory loads all history entries and presents an interactive picker.
func runHistory() error {
	store, err := newHistoryStore()
	if err != nil {
		return fmt.Errorf("failed to access history: %w", err)
	}

	entries, err := store.List()
	if err != nil {
		return fmt.Errorf("failed to read history: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no history yet — run a query first")
	}

	items := make([]string, len(entries))
	for i := range entries {
		items[i] = formatHistoryEntry(entries[i])
	}

	theme := ui.DefaultTheme()
	idx, err := uiRunSelectorFn(items, func(i int) string {
		return items[i]
	}, theme)
	if err != nil {
		return fmt.Errorf("failed to pick from history: %w", err)
	}

	if idx < 0 {
		return ErrCancelled
	}

	return handleSelectedCommand(entries[idx].Selected, entries[idx].Query, entries[idx].PipeContext, true)
}

// runContinue loads the last history entry and uses it as follow-up context
// for refining the previous command with a new query.
func runContinue(query string, pipeContext string) error {
	store, err := newHistoryStore()
	if err != nil {
		return fmt.Errorf("failed to access history: %w", err)
	}

	entry, err := store.Last()
	if err != nil {
		if errors.Is(err, history.ErrEmpty) {
			return fmt.Errorf("no history yet — run a query first")
		}
		return fmt.Errorf("failed to read history: %w", err)
	}

	followUp := &llm.FollowUpContext{
		PreviousQuery:   entry.Query,
		PreviousCommand: entry.Selected,
	}

	return generateCommands(query, pipeContext, followUp)
}

// formatHistoryEntry formats a history entry for display in the picker.
func formatHistoryEntry(e history.Entry) string {
	ts := e.Timestamp.Format("Jan 02 15:04")
	return fmt.Sprintf("[%s] %s → %s", ts, e.Query, e.Selected)
}

// generateCommands generates shell commands using LLM based on user query.
func generateCommands(query string, pipeContext string, followUp *llm.FollowUpContext) error {
	if err := guard.CheckQuery(query, forceSend); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, err := llm.NewProvider(cfg.LLM.ToLLMConfig())
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.DefaultTimeout)
	defer cancel()

	commands, err := provider.Generate(ctx, query, cfg.LLM.Count, pipeContext, followUp)
	if err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	for i, cmd := range commands {
		commands[i] = guard.SanitizeOutput(cmd)
	}

	if len(commands) == 0 {
		return fmt.Errorf("no commands generated")
	}

	if len(commands) == 1 {
		return handleSelectedCommand(commands[0], query, pipeContext, cfg.ActionMenu)
	}

	idx, err := uiRunSelectorFn(commands, func(i int) string {
		return commands[i]
	}, cfg.Theme.ToTheme())
	if err != nil {
		return fmt.Errorf("failed to pick command: %w", err)
	}

	if idx < 0 {
		fmt.Println(query)
		return ErrCancelled
	}

	return handleSelectedCommand(commands[idx], query, pipeContext, cfg.ActionMenu)
}

// newHistoryStore creates a history store using the default config directory.
// Overridden in tests to use a temp directory.
var newHistoryStore = func() (*history.Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return history.NewStore(filepath.Join(home, config.Dir)), nil
}

// saveToHistory persists a history entry. Errors are silently ignored
// because history is a convenience feature that should not break the main flow.
func saveToHistory(entry history.Entry) {
	store, err := newHistoryStore()
	if err != nil {
		return
	}
	_ = store.Add(entry)
}

// handleSelectedCommand either shows the post-selection action menu or
// prints the command to stdout. The action menu is shown only when
// actionMenu is true AND stdout is a TTY. When the user chooses "revise",
// it reads a refinement query and starts a new generation cycle with
// follow-up context. History is saved only on the final action
// (execute/copy/quit), not on intermediate revisions.
func handleSelectedCommand(command, query, pipeContext string, actionMenu bool) error {
	if !actionMenu || !shouldPromptFn() {
		saveToHistory(history.Entry{
			Query:       query,
			Selected:    command,
			PipeContext: pipeContext,
			Timestamp:   time.Now(),
		})
		fmt.Println(command)
		return nil
	}

	err := promptActionFn(command)
	if errors.Is(err, action.ErrCancelled) {
		return ErrCancelled
	}

	var reviseErr *action.ReviseRequestedError
	if errors.As(err, &reviseErr) {
		refinement, readErr := readRefinementFn()
		if readErr != nil {
			if errors.Is(readErr, action.ErrEmptyRefinement) {
				return ErrCancelled
			}
			return readErr
		}
		followUp := &llm.FollowUpContext{
			PreviousQuery:   query,
			PreviousCommand: command,
		}
		return generateCommandsFn(refinement, pipeContext, followUp)
	}

	saveToHistory(history.Entry{
		Query:       query,
		Selected:    command,
		PipeContext: pipeContext,
		Timestamp:   time.Now(),
	})
	return err
}
