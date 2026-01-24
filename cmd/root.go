package cmd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/erakhmetzyan/qx/internal/config"
	"github.com/erakhmetzyan/qx/internal/llm"
	"github.com/erakhmetzyan/qx/internal/picker"
)

var (
	Version = "dev" // устанавливается при сборке

	// флаги командной строки
	inlineMode       bool
	shellIntegration string
	showConfig       bool
)

const defaultTimeout = 60 * time.Second

var rootCmd = &cobra.Command{
	Use:   "qx [query]",
	Short: "Generate shell commands using LLM",
	Long: `qx is a CLI tool that generates shell commands from natural language descriptions.
It uses LLM to generate multiple command variants and presents them in a fzf-style picker.`,
	Version: Version,
	Args:    cobra.MaximumNArgs(1),
	RunE:    run,
}

func init() {
	rootCmd.Flags().BoolVar(&inlineMode, "inline", false, "inline mode for shell integration")
	rootCmd.Flags().StringVar(&shellIntegration, "shell-integration", "", "output shell integration script (bash|zsh)")
	rootCmd.Flags().BoolVar(&showConfig, "config", false, "show config file path")
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
		return fmt.Errorf("query is required")
	}

	query := args[0]
	return generateCommands(query)
}

func handleShellIntegration(shell string) error {
	switch shell {
	case "bash", "zsh":
		return errors.New("shell integration not implemented yet")
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh)", shell)
	}
}

// generateCommands generates shell commands using LLM based on user query.
func generateCommands(query string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	provider, err := llm.NewProvider(llm.Config{
		BaseURL:  cfg.LLM.BaseURL,
		APIKey:   cfg.LLM.APIKey,
		Model:    cfg.LLM.Model,
		Provider: cfg.LLM.Provider,
	})
	if err != nil {
		return fmt.Errorf("failed to create LLM provider: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	commands, err := provider.Generate(ctx, query, cfg.LLM.Count)
	if err != nil {
		return fmt.Errorf("failed to generate commands: %w", err)
	}

	if len(commands) == 0 {
		return fmt.Errorf("no commands generated")
	}

	selected, err := picker.Pick(commands)
	if err != nil {
		if errors.Is(err, picker.ErrAborted) {
			return nil
		}
		return fmt.Errorf("failed to pick command: %w", err)
	}

	if selected != "" {
		fmt.Println(selected)
	}

	return nil
}
