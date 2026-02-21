package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/evgfitil/qx/internal/action"
	"github.com/evgfitil/qx/internal/guard"
	"github.com/evgfitil/qx/internal/history"
	"github.com/evgfitil/qx/internal/llm"
	"github.com/evgfitil/qx/internal/ui"
)

func TestErrCancelled_CanBeExtracted(t *testing.T) {
	wrapped := fmt.Errorf("run failed: %w", ErrCancelled)

	if !errors.Is(wrapped, ErrCancelled) {
		t.Fatal("expected errors.Is to find ErrCancelled in wrapped error")
	}
}

func TestGenerateCommands_QueryWithSecrets(t *testing.T) {
	origForceSend := forceSend
	defer func() { forceSend = origForceSend }()
	forceSend = false

	err := generateCommands("use key AKIAIOSFODNN7EXAMPLE", "some safe context", nil)
	if err == nil {
		t.Fatal("expected error for query with secrets")
	}

	var secretsErr *guard.SecretsError
	if !errors.As(err, &secretsErr) {
		t.Fatalf("expected SecretsError, got %T: %v", err, err)
	}
}

func TestGenerateCommands_EmptyPipeContextSkipsGuard(t *testing.T) {
	origForceSend := forceSend
	defer func() { forceSend = origForceSend }()
	forceSend = false

	// Isolate from developer's real config so config.Load() always fails.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "")

	// With empty pipe context, only the query is checked.
	// Should pass guard check and fail later at config.Load().
	err := generateCommands("list files", "", nil)

	var secretsErr *guard.SecretsError
	if errors.As(err, &secretsErr) {
		t.Fatal("empty pipe context should not trigger secrets error")
	}
	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}
}

func TestHandleSelectedCommand_NonTTY_PrintsToStdout(t *testing.T) {
	// When stdout is a pipe (non-TTY), handleSelectedCommand should print
	// the command to stdout without showing the action menu.
	withTempHistoryStore(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	handleErr := handleSelectedCommand("echo hello", "test query", "", true)
	_ = w.Close()

	if handleErr != nil {
		t.Errorf("handleSelectedCommand returned error: %v", handleErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if string(out) != "echo hello\n" {
		t.Errorf("handleSelectedCommand output = %q, want %q", string(out), "echo hello\n")
	}
}

func TestRunInteractive_MultilineQueryDoesNotPanic(t *testing.T) {
	// Smoke test: runInteractive with multiline query (line continuations)
	// should not panic. Config error is expected in test environment.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "")

	multilineQuery := "ps aux \\\n\t| grep nginx \\\n\t| sort"

	err := runInteractive(multilineQuery, "")
	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}
}

func TestRunInteractive_SimpleQueryDoesNotPanic(t *testing.T) {
	// Smoke test: runInteractive with a simple query should not panic.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "")

	err := runInteractive("list all running containers", "")
	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}
}

func TestHandleSelectedCommand_NonTTY_EmptyCommand(t *testing.T) {
	// Even with empty command, non-TTY path should print and return nil.
	withTempHistoryStore(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	handleErr := handleSelectedCommand("", "test query", "", true)
	_ = w.Close()

	if handleErr != nil {
		t.Errorf("handleSelectedCommand returned error: %v", handleErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if string(out) != "\n" {
		t.Errorf("handleSelectedCommand output = %q, want %q", string(out), "\n")
	}
}

func withTempHistoryStore(t *testing.T) *history.Store {
	t.Helper()
	dir := t.TempDir()
	store := history.NewStore(dir)
	orig := newHistoryStore
	newHistoryStore = func() (*history.Store, error) { return store, nil }
	t.Cleanup(func() { newHistoryStore = orig })
	return store
}

// withTestConfig sets up a test config environment with a config file.
// actionMenu controls the action_menu setting in the generated config.
func withTestConfig(t *testing.T, actionMenu bool) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("OPENAI_API_KEY", "test-key")

	configDir := filepath.Join(dir, ".config", "qx")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	content := fmt.Sprintf("action_menu: %v\n", actionMenu)
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

func TestSaveToHistory_PersistsEntry(t *testing.T) {
	store := withTempHistoryStore(t)

	entry := history.Entry{
		Query:       "list files",
		Selected:    "ls -la",
		PipeContext: "some context",
		Timestamp:   time.Now(),
	}
	saveToHistory(entry)

	got, err := store.Last()
	if err != nil {
		t.Fatalf("Last() error = %v", err)
	}
	if got.Query != "list files" {
		t.Errorf("Query = %q, want %q", got.Query, "list files")
	}
	if got.Selected != "ls -la" {
		t.Errorf("Selected = %q, want %q", got.Selected, "ls -la")
	}
	if got.PipeContext != "some context" {
		t.Errorf("PipeContext = %q, want %q", got.PipeContext, "some context")
	}
}

func TestSaveToHistory_MultipleSaves(t *testing.T) {
	store := withTempHistoryStore(t)

	for _, q := range []string{"first", "second", "third"} {
		saveToHistory(history.Entry{
			Query:     q,
			Selected:  "cmd1",
			Timestamp: time.Now(),
		})
	}

	entries, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	if entries[0].Query != "third" {
		t.Errorf("newest entry = %q, want %q", entries[0].Query, "third")
	}
}

func TestSaveToHistory_StoreCreationError(t *testing.T) {
	orig := newHistoryStore
	newHistoryStore = func() (*history.Store, error) {
		return nil, fmt.Errorf("no home directory")
	}
	t.Cleanup(func() { newHistoryStore = orig })

	// Should not panic when store creation fails
	saveToHistory(history.Entry{
		Query:     "test",
		Selected:  "cmd",
		Timestamp: time.Now(),
	})
}

func TestRunLast_WithHistory(t *testing.T) {
	store := withTempHistoryStore(t)
	withTestConfig(t, false)

	_ = store.Add(history.Entry{
		Query:     "find large files",
		Selected:  "find . -size +100M",
		Timestamp: time.Now(),
	})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	runErr := runLast()
	_ = w.Close()

	if runErr != nil {
		t.Fatalf("runLast() error = %v", runErr)
	}

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if string(out) != "find . -size +100M\n" {
		t.Errorf("runLast output = %q, want %q", string(out), "find . -size +100M\n")
	}
}

func TestRunLast_EmptyHistory(t *testing.T) {
	withTempHistoryStore(t)
	withTestConfig(t, false)

	err := runLast()
	if err == nil {
		t.Fatal("expected error for empty history")
	}
	if got := err.Error(); got != "no history yet — run a query first" {
		t.Errorf("error = %q, want %q", got, "no history yet — run a query first")
	}
}

func TestRunLast_StoreCreationError(t *testing.T) {
	withTestConfig(t, false)

	orig := newHistoryStore
	newHistoryStore = func() (*history.Store, error) {
		return nil, fmt.Errorf("no home directory")
	}
	t.Cleanup(func() { newHistoryStore = orig })

	err := runLast()
	if err == nil {
		t.Fatal("expected error when store creation fails")
	}
	want := "failed to access history: no home directory"
	if got := err.Error(); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestRunLast_ActionMenuFalse_PrintsWithoutMenu(t *testing.T) {
	withMockFns(t)
	store := withTempHistoryStore(t)
	withTestConfig(t, false)

	_ = store.Add(history.Entry{
		Query:     "find large files",
		Selected:  "find . -size +100M",
		Timestamp: time.Now(),
	})

	shouldPromptFn = func() bool { return true }
	menuCalled := false
	promptActionFn = func(cmd string) error {
		menuCalled = true
		return nil
	}

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	err := runLast()
	_ = w.Close()

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if menuCalled {
		t.Error("action menu should not be called when action_menu is false in config")
	}
	if string(out) != "find . -size +100M\n" {
		t.Errorf("output = %q, want %q", string(out), "find . -size +100M\n")
	}
}

func TestRunLast_ActionMenuTrue_ShowsMenu(t *testing.T) {
	withMockFns(t)
	store := withTempHistoryStore(t)
	withTestConfig(t, true)

	_ = store.Add(history.Entry{
		Query:     "find large files",
		Selected:  "find . -size +100M",
		Timestamp: time.Now(),
	})

	shouldPromptFn = func() bool { return true }
	menuCalled := false
	promptActionFn = func(cmd string) error {
		menuCalled = true
		return nil
	}

	err := runLast()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !menuCalled {
		t.Error("action menu should be called when action_menu is true in config")
	}
}

func TestRunHistory_EmptyHistory(t *testing.T) {
	withTempHistoryStore(t)

	err := runHistory()
	if err == nil {
		t.Fatal("expected error for empty history")
	}
	if got := err.Error(); got != "no history yet — run a query first" {
		t.Errorf("error = %q, want %q", got, "no history yet — run a query first")
	}
}

func TestRunHistory_StoreCreationError(t *testing.T) {
	orig := newHistoryStore
	newHistoryStore = func() (*history.Store, error) {
		return nil, fmt.Errorf("no home directory")
	}
	t.Cleanup(func() { newHistoryStore = orig })

	err := runHistory()
	if err == nil {
		t.Fatal("expected error when store creation fails")
	}
	want := "failed to access history: no home directory"
	if got := err.Error(); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestFormatHistoryEntry(t *testing.T) {
	ts := time.Date(2026, 2, 20, 14, 30, 0, 0, time.UTC)
	entry := history.Entry{
		Query:     "list running containers",
		Selected:  "docker ps",
		Timestamp: ts,
	}

	got := formatHistoryEntry(entry)
	want := "[Feb 20 14:30] list running containers → docker ps"
	if got != want {
		t.Errorf("formatHistoryEntry() = %q, want %q", got, want)
	}
}

func TestFormatHistoryEntry_LongQuery(t *testing.T) {
	ts := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	entry := history.Entry{
		Query:     "find all files larger than 1GB and delete them",
		Selected:  "find . -size +1G -delete",
		Timestamp: ts,
	}

	got := formatHistoryEntry(entry)
	want := "[Jan 05 09:00] find all files larger than 1GB and delete them → find . -size +1G -delete"
	if got != want {
		t.Errorf("formatHistoryEntry() = %q, want %q", got, want)
	}
}

func TestSaveToHistory_EmptyPipeContext(t *testing.T) {
	store := withTempHistoryStore(t)

	saveToHistory(history.Entry{
		Query:     "list files",
		Selected:  "ls",
		Timestamp: time.Now(),
	})

	got, err := store.Last()
	if err != nil {
		t.Fatalf("Last() error = %v", err)
	}
	if got.PipeContext != "" {
		t.Errorf("PipeContext = %q, want empty", got.PipeContext)
	}
}

func TestRunContinue_EmptyHistory(t *testing.T) {
	withTempHistoryStore(t)

	err := runContinue("make it recursive", "")
	if err == nil {
		t.Fatal("expected error for empty history")
	}
	if got := err.Error(); got != "no history yet — run a query first" {
		t.Errorf("error = %q, want %q", got, "no history yet — run a query first")
	}
}

func TestRunContinue_StoreCreationError(t *testing.T) {
	orig := newHistoryStore
	newHistoryStore = func() (*history.Store, error) {
		return nil, fmt.Errorf("no home directory")
	}
	t.Cleanup(func() { newHistoryStore = orig })

	err := runContinue("refine this", "")
	if err == nil {
		t.Fatal("expected error when store creation fails")
	}
	want := "failed to access history: no home directory"
	if got := err.Error(); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestRun_ContinueWithoutQueryArg(t *testing.T) {
	withTempHistoryStore(t)

	origContinueFlag := continueFlag
	continueFlag = true
	t.Cleanup(func() { continueFlag = origContinueFlag })

	err := run(rootCmd, []string{})
	if err == nil {
		t.Fatal("expected error for --continue without query")
	}
	want := "--continue requires a query argument"
	if got := err.Error(); got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestRun_MutuallyExclusiveFlags(t *testing.T) {
	withTempHistoryStore(t)
	resetRootCmdFlags(t)

	rootCmd.SetArgs([]string{"--last", "--history"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for combined --last and --history")
	}
	if !strings.Contains(err.Error(), "if any flags in the group") {
		t.Errorf("expected cobra mutual exclusion error, got: %q", err.Error())
	}
}

func TestRun_SingleFlags_NoMutualExclusionError(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"last only", []string{"--last"}},
		{"history only", []string{"--history"}},
		{"continue with query", []string{"--continue", "refine"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTempHistoryStore(t)
			resetRootCmdFlags(t)
			t.Setenv("HOME", t.TempDir())
			t.Setenv("OPENAI_API_KEY", "")

			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			if err != nil && strings.Contains(err.Error(), "if any flags in the group") {
				t.Errorf("single flag should not trigger mutual exclusion error: %v", err)
			}
		})
	}
}

func TestRun_ShortFlagL_WorksAsLast(t *testing.T) {
	store := withTempHistoryStore(t)
	resetRootCmdFlags(t)
	withTestConfig(t, false)

	_ = store.Add(history.Entry{
		Query:     "find large files",
		Selected:  "find . -size +100M",
		Timestamp: time.Now(),
	})

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	rootCmd.SetArgs([]string{"-l"})
	err := rootCmd.Execute()
	_ = w.Close()

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if err != nil {
		t.Fatalf("qx -l error = %v", err)
	}
	if string(out) != "find . -size +100M\n" {
		t.Errorf("output = %q, want %q", string(out), "find . -size +100M\n")
	}
}

func TestRun_ShortFlagC_WorksAsContinue(t *testing.T) {
	store := withTempHistoryStore(t)
	resetRootCmdFlags(t)

	_ = store.Add(history.Entry{
		Query:     "find files",
		Selected:  "find . -name '*.go'",
		Timestamp: time.Now(),
	})

	// runContinue will fail at config.Load() in test env — that's expected.
	// We just verify the flag is recognized and triggers the continue path.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "")

	rootCmd.SetArgs([]string{"-c", "only large files"})
	err := rootCmd.Execute()

	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}
	// Should reach generateCommands (config.Load fails), not flag parsing error
	if strings.Contains(err.Error(), "unknown shorthand flag") {
		t.Errorf("short flag -c not recognized: %v", err)
	}
}

// resetRootCmdFlags saves and restores flag state for rootCmd.
// Must be called in tests that use rootCmd.Execute() with SetArgs.
func resetRootCmdFlags(t *testing.T) {
	t.Helper()
	origLast := lastFlag
	origHistory := historyFlag
	origContinue := continueFlag
	t.Cleanup(func() {
		rootCmd.SetArgs(nil)
		lastFlag = origLast
		historyFlag = origHistory
		continueFlag = origContinue
		for _, name := range []string{"last", "history", "continue"} {
			if f := rootCmd.Flags().Lookup(name); f != nil {
				f.Changed = false
			}
		}
	})
}

func TestRunContinue_WithHistory(t *testing.T) {
	store := withTempHistoryStore(t)

	_ = store.Add(history.Entry{
		Query:     "find large files",
		Selected:  "find . -size +100M",
		Timestamp: time.Now(),
	})

	// runContinue will try to load config and create LLM provider,
	// which will fail in test environment. That's expected.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "")

	err := runContinue("only go files", "")
	if err == nil {
		t.Fatal("expected error from config.Load() in test environment")
	}

	// The error should come from generateCommands (config.Load), not from history access
	got := err.Error()
	if got == "no history yet — run a query first" {
		t.Error("should not get empty history error when history has entries")
	}
	if !strings.Contains(got, "failed to load config") && !strings.Contains(got, "failed to generate") {
		t.Errorf("expected config/generate error, got: %q", got)
	}
}

// withMockFns saves and restores overridable function references used by
// handleSelectedCommand. Call at the start of any test that overrides them.
func withMockFns(t *testing.T) {
	t.Helper()
	origShouldPrompt := shouldPromptFn
	origPromptAction := promptActionFn
	origReadRefinement := readRefinementFn
	origGenerateCommands := generateCommandsFn
	origUiRun := uiRunFn
	origUiRunSelector := uiRunSelectorFn
	t.Cleanup(func() {
		shouldPromptFn = origShouldPrompt
		promptActionFn = origPromptAction
		readRefinementFn = origReadRefinement
		generateCommandsFn = origGenerateCommands
		uiRunFn = origUiRun
		uiRunSelectorFn = origUiRunSelector
	})
}

func TestHandleSelectedCommand_Revise_FollowUpContext(t *testing.T) {
	withMockFns(t)
	store := withTempHistoryStore(t)

	shouldPromptFn = func() bool { return true }
	promptActionFn = func(cmd string) error {
		return &action.ReviseRequestedError{}
	}
	readRefinementFn = func() (string, error) {
		return "make it recursive", nil
	}

	var capturedQuery string
	var capturedPipeContext string
	var capturedFollowUp *llm.FollowUpContext
	generateCommandsFn = func(query string, pipeContext string, followUp *llm.FollowUpContext) error {
		capturedQuery = query
		capturedPipeContext = pipeContext
		capturedFollowUp = followUp
		return nil
	}

	err := handleSelectedCommand("find .", "find files", "some context", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedQuery != "make it recursive" {
		t.Errorf("query = %q, want %q", capturedQuery, "make it recursive")
	}
	if capturedPipeContext != "some context" {
		t.Errorf("pipeContext = %q, want %q", capturedPipeContext, "some context")
	}
	if capturedFollowUp == nil {
		t.Fatal("followUp is nil")
	}
	if capturedFollowUp.PreviousQuery != "find files" {
		t.Errorf("PreviousQuery = %q, want %q", capturedFollowUp.PreviousQuery, "find files")
	}
	if capturedFollowUp.PreviousCommand != "find ." {
		t.Errorf("PreviousCommand = %q, want %q", capturedFollowUp.PreviousCommand, "find .")
	}

	// No history should be saved on revise (intermediate step)
	entries, _ := store.List()
	if len(entries) != 0 {
		t.Errorf("expected 0 history entries on revise, got %d", len(entries))
	}
}

func TestHandleSelectedCommand_Revise_EmptyRefinement(t *testing.T) {
	withMockFns(t)

	shouldPromptFn = func() bool { return true }
	promptActionFn = func(cmd string) error {
		return &action.ReviseRequestedError{}
	}
	readRefinementFn = func() (string, error) {
		return "", action.ErrEmptyRefinement
	}

	err := handleSelectedCommand("find .", "find files", "", true)
	if !errors.Is(err, ErrCancelled) {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func TestHandleSelectedCommand_Revise_ReadError(t *testing.T) {
	withMockFns(t)

	shouldPromptFn = func() bool { return true }
	promptActionFn = func(cmd string) error {
		return &action.ReviseRequestedError{}
	}
	readRefinementFn = func() (string, error) {
		return "", fmt.Errorf("failed to read from tty")
	}

	err := handleSelectedCommand("find .", "find files", "", true)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "failed to read from tty" {
		t.Errorf("error = %q, want %q", got, "failed to read from tty")
	}
}

func TestHandleSelectedCommand_Revise_HistorySavedOnFinalAction(t *testing.T) {
	withMockFns(t)
	store := withTempHistoryStore(t)

	callCount := 0
	shouldPromptFn = func() bool { return true }
	promptActionFn = func(cmd string) error {
		callCount++
		if callCount == 1 {
			return &action.ReviseRequestedError{}
		}
		// Simulate quit: print command, return nil
		fmt.Println(cmd)
		return nil
	}
	readRefinementFn = func() (string, error) {
		return "make it recursive", nil
	}
	generateCommandsFn = func(query string, pipeContext string, followUp *llm.FollowUpContext) error {
		// Simulate the second pick: call handleSelectedCommand with new command
		return handleSelectedCommand("find . -r", query, pipeContext, true)
	}

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	err := handleSelectedCommand("find .", "find files", "ctx", true)
	_ = w.Close()
	_, _ = io.ReadAll(r)
	_ = r.Close()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// History should be saved once (for the final selection only)
	entries, _ := store.List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(entries))
	}
	if entries[0].Query != "make it recursive" {
		t.Errorf("Query = %q, want %q", entries[0].Query, "make it recursive")
	}
	if entries[0].Selected != "find . -r" {
		t.Errorf("Selected = %q, want %q", entries[0].Selected, "find . -r")
	}
	if entries[0].PipeContext != "ctx" {
		t.Errorf("PipeContext = %q, want %q", entries[0].PipeContext, "ctx")
	}
}

func TestHandleSelectedCommand_NonTTY_SavesHistory(t *testing.T) {
	store := withTempHistoryStore(t)

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	_ = handleSelectedCommand("echo hello", "greet", "pipe data", true)
	_ = w.Close()
	_, _ = io.ReadAll(r)
	_ = r.Close()

	entries, _ := store.List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(entries))
	}
	if entries[0].Query != "greet" {
		t.Errorf("Query = %q, want %q", entries[0].Query, "greet")
	}
	if entries[0].Selected != "echo hello" {
		t.Errorf("Selected = %q, want %q", entries[0].Selected, "echo hello")
	}
	if entries[0].PipeContext != "pipe data" {
		t.Errorf("PipeContext = %q, want %q", entries[0].PipeContext, "pipe data")
	}
}

func TestHandleSelectedCommand_TTY_CancelledReturnsErrCancelled(t *testing.T) {
	withMockFns(t)

	shouldPromptFn = func() bool { return true }
	promptActionFn = func(cmd string) error {
		return action.ErrCancelled
	}

	err := handleSelectedCommand("echo hello", "test", "", true)
	if !errors.Is(err, ErrCancelled) {
		t.Errorf("expected ErrCancelled, got %v", err)
	}
}

func TestHandleSelectedCommand_TTY_ActionErrorSavesHistory(t *testing.T) {
	withMockFns(t)
	store := withTempHistoryStore(t)

	actionErr := fmt.Errorf("execution failed: exit status 1")
	shouldPromptFn = func() bool { return true }
	promptActionFn = func(cmd string) error {
		return actionErr
	}

	err := handleSelectedCommand("bad-cmd", "run thing", "ctx", true)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != actionErr.Error() {
		t.Errorf("error = %q, want %q", err.Error(), actionErr.Error())
	}

	entries, _ := store.List()
	if len(entries) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(entries))
	}
	if entries[0].Selected != "bad-cmd" {
		t.Errorf("Selected = %q, want %q", entries[0].Selected, "bad-cmd")
	}
}

func TestHandleSelectedCommand_ActionMenuFalse_PrintsWithoutMenu(t *testing.T) {
	withMockFns(t)
	withTempHistoryStore(t)

	shouldPromptFn = func() bool { return true }
	menuCalled := false
	promptActionFn = func(cmd string) error {
		menuCalled = true
		return nil
	}

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	err := handleSelectedCommand("echo test", "query", "", false)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if menuCalled {
		t.Error("action menu should not be called when actionMenu is false")
	}
	if string(out) != "echo test\n" {
		t.Errorf("output = %q, want %q", string(out), "echo test\n")
	}
}

func TestRunInteractive_ConfigError_PrintsToStderr(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "")

	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	err := runInteractive("my query", "")
	_ = w.Close()

	stderrOut, _ := io.ReadAll(r)
	_ = r.Close()

	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
	if !strings.Contains(string(stderrOut), "Error:") {
		t.Errorf("stderr should contain 'Error:', got %q", string(stderrOut))
	}
}

func TestRunInteractive_WithMockedUI_SelectedResult(t *testing.T) {
	withMockFns(t)
	withTempHistoryStore(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	uiRunFn = func(opts ui.RunOptions) (ui.Result, error) {
		return ui.SelectedResult{Command: "ls -la", Query: "list files"}, nil
	}

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	err := runInteractive("list files", "")
	_ = w.Close()

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "ls -la\n" {
		t.Errorf("output = %q, want %q", string(out), "ls -la\n")
	}
}

func TestRunInteractive_WithMockedUI_CancelledResult(t *testing.T) {
	withMockFns(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	uiRunFn = func(opts ui.RunOptions) (ui.Result, error) {
		return ui.CancelledResult{Query: "list files"}, nil
	}

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	err := runInteractive("list files", "")
	_ = w.Close()

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
	if string(out) != "list files\n" {
		t.Errorf("output = %q, want %q", string(out), "list files\n")
	}
}

func TestRunHistory_WithMockedSelector_SelectsEntry(t *testing.T) {
	withMockFns(t)
	store := withTempHistoryStore(t)

	_ = store.Add(history.Entry{
		Query:     "find files",
		Selected:  "find . -name '*.go'",
		Timestamp: time.Now(),
	})

	uiRunSelectorFn = func(items []string, display func(int) string, theme ui.Theme) (int, error) {
		return 0, nil
	}

	r, w, _ := os.Pipe()
	origStdout := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	err := runHistory()
	_ = w.Close()

	out, _ := io.ReadAll(r)
	_ = r.Close()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(out) != "find . -name '*.go'\n" {
		t.Errorf("output = %q, want %q", string(out), "find . -name '*.go'\n")
	}
}

func TestRunHistory_WithMockedSelector_Cancelled(t *testing.T) {
	withMockFns(t)
	store := withTempHistoryStore(t)

	_ = store.Add(history.Entry{
		Query:     "find files",
		Selected:  "find . -name '*.go'",
		Timestamp: time.Now(),
	})

	uiRunSelectorFn = func(items []string, display func(int) string, theme ui.Theme) (int, error) {
		return -1, nil
	}

	err := runHistory()
	if !errors.Is(err, ErrCancelled) {
		t.Fatalf("expected ErrCancelled, got %v", err)
	}
}
