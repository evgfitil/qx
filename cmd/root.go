package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/evgfitil/qx/internal/config"
	"github.com/evgfitil/qx/internal/guard"
	"github.com/evgfitil/qx/internal/llm"
	"github.com/evgfitil/qx/internal/picker"
	"github.com/evgfitil/qx/internal/shell"
	"github.com/evgfitil/qx/internal/tui"
)

const ExitCodeCancelled = 130

var (
	Version          = "dev"
	shellIntegration string
	showConfig       bool
	queryFlag        string
	forceSend        bool
)

// ErrCancelled indicates user cancelled the operation.
var ErrCancelled = errors.New("operation cancelled")

var rootCmd = &cobra.Command{
	Use:   "qx [query]",
	Short: "Generate shell commands using LLM",
	Long: `qx is a CLI tool that generates shell commands from natural language descriptions.
It uses LLM to generate multiple command variants and presents them in a fzf-style picker.`,
	Version:       Version,
	Args:          cobra.MaximumNArgs(1),
	RunE:          run,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.Flags().StringVar(&shellIntegration, "shell-integration", "", "output shell integration script (bash|zsh|fish)")
	rootCmd.Flags().BoolVar(&showConfig, "config", false, "show config file path")
	rootCmd.Flags().StringVarP(&queryFlag, "query", "q", "", "initial query for TUI input (pre-fills the input field)")
	rootCmd.Flags().BoolVar(&forceSend, "force-send", false, "send query even if secrets detected")
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

	if len(args) == 0 {
		return runInteractive(queryFlag)
	}

	query := args[0]
	return generateCommands(query)
}

func runInteractive(initialQuery string) error {
	cfg, err := config.Load()
	if err != nil {
		if _, showErr := tui.ShowError(err, initialQuery); showErr != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if initialQuery != "" {
			fmt.Println(initialQuery)
		}
		return ErrCancelled
	}

	result, err := tui.Run(cfg.LLM.ToLLMConfig(), initialQuery, forceSend)
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	switch r := result.(type) {
	case tui.CancelledResult:
		if r.Query != "" {
			fmt.Println(r.Query)
		}
		return ErrCancelled
	case tui.SelectedResult:
		if r.Command != "" {
			fmt.Println(r.Command)
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

// generateCommands generates shell commands using LLM based on user query.
func generateCommands(query string) error {
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

	commands, err := provider.Generate(ctx, query, cfg.LLM.Count)
	if err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	for i, cmd := range commands {
		commands[i] = guard.SanitizeOutput(cmd)
	}

	if len(commands) == 0 {
		return fmt.Errorf("no commands generated")
	}

	selected, err := picker.Pick(commands)
	if err != nil {
		if errors.Is(err, picker.ErrAborted) {
			fmt.Println(query)
			return ErrCancelled
		}
		return fmt.Errorf("failed to pick command: %w", err)
	}

	if selected != "" {
		fmt.Println(selected)
	}

	return nil
}
