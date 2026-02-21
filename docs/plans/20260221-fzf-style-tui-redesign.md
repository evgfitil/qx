# fzf-style TUI Redesign

## Overview

Replace the fragmented TUI system (bubbletea TUI + go-fuzzyfinder picker + raw /dev/tty action menu) with a single unified bubbletea model that behaves like fzf's `--height` mode: inline rendering in the lower half of the screen, clean disappearance after selection.

**Problem**: three separate TUI systems with inconsistent UX — bubbletea for input, go-fuzzyfinder for selection, raw terminal for action menu.

**Solution**: one bubbletea model with states `input → loading → selecting → done`, inline rendering limited to ~40% terminal height, configurable fzf-like theme, clean ANSI cleanup on exit.

## Context (from discovery)

**Files/components to replace:**

- `internal/tui/` (model.go, tui.go, styles.go, result.go, error.go) — bubbletea-based input/selection
- `internal/picker/` (picker.go) — go-fuzzyfinder wrapper
- `go-fuzzyfinder` dependency in go.mod

**Files to modify:**

- `cmd/root.go` — rewire all modes to use new `internal/ui/` package
- `internal/config/config.go` — add theme and action\_menu settings
- `internal/action/menu.go` — action menu becomes optional (controlled by config)

**Files unchanged:**

- `internal/llm/`, `internal/guard/`, `internal/history/`, `internal/shell/`
- `internal/action/execute.go`, `internal/action/clipboard.go`, `internal/action/revise.go`

**Dependencies to add:** none (bubbletea + lipgloss already in go.mod).
**Dependencies to remove:** `github.com/ktr0731/go-fuzzyfinder`.

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility for config.yaml (new fields are optional with defaults)

## Testing Strategy

- **Unit tests**: required for every task
- Theme loading, state transitions, view rendering, result extraction
- Test model via `tea.Msg` injection (standard bubbletea testing pattern)

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix
- Update plan if implementation deviates from original scope

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): code changes, tests, documentation updates
- **Post-Completion** (no checkboxes): manual testing, verification

## Architecture

### New package structure

```text
internal/ui/
├── theme.go       # Theme struct, defaults, config loading
├── model.go       # Main model: states, Init, Update
├── view.go        # View rendering for all states
├── run.go         # Entry points: Run(), RunSelector()
├── result.go      # Result types: SelectedResult, CancelledResult
└── *_test.go      # Tests
```

### State machine

```text
┌─────────┐  Enter   ┌──────────┐  commands   ┌───────────┐  Enter   ┌──────┐
│  input   │ ───────> │ loading  │ ──────────> │ selecting │ ───────> │ done │
└─────────┘          └──────────┘             └───────────┘          └──────┘
     │                     │                       │                     │
     └─── Esc/Ctrl+C ─────┴─── Esc/Ctrl+C ────────┘                     │
                                                                         v
                                                                   empty View()
                                                                   → cleanup
```

### Two entry points

1. **`Run(opts RunOptions) (Result, error)`** — full flow for interactive/direct/shell-integration modes
   - `opts.InitialQuery` set → skip input, go to loading immediately
   - `opts.InitialQuery` empty → start with input state
2. **`RunSelector(items []string, display func(int) string) (int, error)`** — selector-only for `--history`

### Inline rendering and cleanup

- bubbletea runs without `WithAltScreen()` (inline mode by default)
- Height limited to `maxHeightPercent` of terminal (default 40%)
- On exit: final `View()` returns `""` → bubbletea's renderer uses `EraseScreenBelow` to clear all rendered lines
- Selected command printed to stdout after bubbletea exits

### Theme system

Config example:

```yaml
llm:
  base_url: "https://api.openai.com/v1"
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini"

theme:
  prompt: "> "
  pointer: "▌"           # cursor indicator for selected item
  selected_fg: "170"     # ANSI color or hex (#ff87d7)
  match_fg: "205"        # highlight color for matches
  text_fg: "252"         # normal text color
  muted_fg: "241"        # counter, spinner text
  border: "rounded"      # rounded | normal | thick | hidden
  border_fg: "240"       # border color

action_menu: false       # show action menu after selection (default: false)
```

All fields optional with fzf-like defaults.

## Implementation Steps

### Task 1: Create theme system (`internal/ui/theme.go`)

- [x] define `Theme` struct with fields: Prompt, Pointer, SelectedFg, MatchFg, TextFg, MutedFg, Border, BorderFg
- [x] implement `DefaultTheme()` returning fzf-like defaults
- [x] implement `lipgloss.Style` getters on Theme: `SelectedStyle()`, `NormalStyle()`, `MutedStyle()`, `PromptStyle()`, `BorderStyle()`
- [x] write tests for DefaultTheme field values
- [x] write tests for style getters (verify colors match theme fields)
- [x] run tests — must pass before next task

### Task 2: Add theme and action\_menu to config (`internal/config/config.go`)

- [x] add `ThemeConfig` struct with mapstructure tags matching config keys
- [x] add `ActionMenu bool` field to Config
- [x] add `Theme ThemeConfig` field to Config
- [x] set viper defaults for all theme fields (matching DefaultTheme)
- [x] set viper default for `action_menu: false`
- [x] implement `ThemeConfig.ToTheme()` converter to `ui.Theme`
- [x] write tests for config loading with theme section present
- [x] write tests for config loading without theme section (defaults used)
- [x] write tests for action\_menu field loading
- [x] run tests — must pass before next task

### Task 3: Create result types and run scaffolding (`internal/ui/result.go`, `internal/ui/run.go`)

- [x] define `Result` interface with `isResult()` marker method
- [x] define `SelectedResult{Command, Query string}` implementing Result
- [x] define `CancelledResult{Query string}` implementing Result
- [x] define `RunOptions` struct: InitialQuery, LLMConfig, ForceSend, PipeContext, Theme
- [x] implement `Run(opts RunOptions) (Result, error)` skeleton — creates model, runs bubbletea Program with `tea.WithOutput(tty)`, `tea.WithInputTTY()`, NO `tea.WithAltScreen()`
- [x] implement `RunSelector(items []string, display func(int) string, theme Theme) (int, error)` skeleton
- [x] write tests for SelectedResult and CancelledResult construction
- [x] run tests — must pass before next task

### Task 4: Core model with state machine (`internal/ui/model.go`)

- [x] define state enum: `stateInput`, `stateLoading`, `stateSelect`, `stateDone`
- [x] define `Model` struct with fields: state, theme, textarea, spinner, commands, filtered, cursor, selected, err, opts, width, height, maxHeight, originalQuery
- [x] implement `newModel(opts RunOptions)` constructor
- [x] implement `newSelectorModel(items []string, display func(int) string, theme Theme)` constructor (starts in stateSelect)
- [x] implement `Init()` returning textarea.Blink
- [x] implement `Update()` skeleton: handle WindowSizeMsg, KeyMsg dispatch per state, commandsMsg, spinner.TickMsg
- [x] implement height management: `maxHeight = max(height * 40 / 100, minHeight)`
- [x] write tests for newModel initial state (stateInput when no query, stateLoading when query provided)
- [x] write tests for WindowSizeMsg handling (maxHeight calculation)
- [x] write tests for Esc/Ctrl+C → stateDone with empty selected (cancelled)
- [x] run tests — must pass before next task

### Task 5: Input state (`internal/ui/model.go` Update, `internal/ui/view.go`)

- [x] implement input state key handling in Update: Enter submits query, transitions to stateLoading
- [x] implement textarea setup: single-line, themed prompt, placeholder "describe the command you need..."
- [x] implement LLM generation as tea.Cmd (same pattern as current tui/model.go)
- [x] implement input state View: textarea + optional error message
- [x] write tests for Enter with non-empty query → stateLoading transition
- [x] write tests for Enter with empty query → stays in stateInput
- [x] write tests for guard.CheckQuery error display
- [x] run tests — must pass before next task

### Task 6: Loading state (`internal/ui/model.go` Update, `internal/ui/view.go`)

- [x] implement loading state: spinner ticks, ignore Enter
- [x] implement commandsMsg handler: transition to stateSelect, populate commands/filtered
- [x] implement commandsMsg error handler: transition back to stateInput with error
- [x] implement loading state View: spinner + "Generating commands..." text
- [x] write tests for commandsMsg success → stateSelect with correct commands
- [x] write tests for commandsMsg error → stateInput with error set
- [x] run tests — must pass before next task

### Task 7: Selector state (`internal/ui/view.go`, `internal/ui/model.go`)

- [x] implement selector navigation: Up/Down arrow keys move cursor
- [x] implement type-to-filter: textarea in selector filters commands (case-insensitive substring match)
- [x] implement Enter → select command, transition to stateDone
- [x] implement auto-select when only 1 command
- [x] implement selector View: bordered box with pointer indicator, scrollable list, counter "N/M"
- [x] implement height-aware pagination: visible items = maxHeight - reserved lines (border + input + counter)
- [x] implement scroll offset tracking for lists longer than visible area
- [x] write tests for navigation (Up/Down within bounds)
- [x] write tests for filtering (type text → filtered list updates)
- [x] write tests for Enter → stateDone with correct selected command
- [x] write tests for auto-select with single command
- [x] run tests — must pass before next task

### Task 8: Done state and cleanup (`internal/ui/view.go`, `internal/ui/run.go`)

- [x] implement stateDone View: return `""` (empty string triggers bubbletea cleanup)
- [x] implement `Model.Result()` method extracting SelectedResult or CancelledResult
- [x] complete `Run()`: after `tea.Program.Run()`, extract result from model
- [x] complete `RunSelector()`: after `tea.Program.Run()`, extract selected index
- [x] write tests for done state View returning empty string
- [x] write tests for Result() extraction for selected and cancelled cases
- [x] run tests — must pass before next task

### Task 9: Integrate into cmd/root.go

- [x] replace `tui.Run()` call in `runInteractive()` with `ui.Run()`
- [x] replace `picker.Pick()` call in `generateCommands()` with `ui.RunSelector()` (wrap commands as items)
- [x] replace `picker.PickIndex()` call in `runHistory()` with `ui.RunSelector()`
- [x] wire action\_menu config: if `cfg.ActionMenu` true and stdout is TTY, call `action.PromptAction()`; otherwise print to stdout
- [x] update imports: remove `internal/tui`, remove `internal/picker`, add `internal/ui`
- [x] update `tui.ShowError()` references — move error display into `ui.Run()` or handle in cmd
- [x] write tests for runInteractive with new ui.Run integration
- [x] write tests for generateCommands with new ui.RunSelector integration
- [x] update existing cmd/root\_test.go for changed flow
- [x] run tests — must pass before next task

### Task 10: Remove old packages and dependencies

- [x] delete `internal/tui/` directory entirely
- [x] delete `internal/picker/` directory entirely
- [x] run `go mod tidy` to remove `go-fuzzyfinder` and its transitive deps (`tcell`, `termbox-go`, etc.)
- [x] verify no remaining imports of `internal/tui` or `internal/picker`
- [x] run tests — must pass before next task

### Task 11: Verify acceptance criteria

- [ ] verify all requirements from Overview are implemented
- [ ] verify edge cases: empty query, LLM error, no commands generated, single command
- [ ] run full test suite (`go test ./...`)
- [ ] run linter (`golangci-lint run`) — all issues must be fixed
- [ ] verify cleanup behavior: TUI content disappears after selection

### Task 12: [Final] Update documentation

- [ ] update CLAUDE.md architecture section if package names changed
- [ ] update README.md if config format changed (new theme section)

## Technical Details

### Model fields

```go
type Model struct {
    state         state
    theme         Theme
    textArea      textarea.Model
    spinner       spinner.Model
    commands      []string     // all generated commands
    filtered      []string     // after filter applied
    cursor        int          // selected index in filtered
    scrollOffset  int          // first visible item index
    selected      string       // chosen command (empty = cancelled)
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
```

### View layout (selector state)

```text
╭────────────────────────────────────────────╮
│ > filter text_                             │   ← textarea (1 line in selector mode)
│ ▌ find . -name "*.go" | xargs grep -l err │   ← selected item (themed)
│   grep -rl "error" --include="*.go" .      │   ← normal item
│   fd -e go -x grep -l "error"              │   ← normal item
│   rg -l "error" -t go                      │   ← normal item
│ 4/4                                        │   ← counter
╰────────────────────────────────────────────╯
```

### Config backward compatibility

All new fields have defaults:

- `theme.*` fields default to fzf-like colors
- `action_menu` defaults to `false`
- Existing configs without these fields work unchanged

## Post-Completion

**Manual verification:**

- test shell integration mode (Ctrl+G): TUI appears inline, selection replaces command line, TUI disappears
- test direct mode (`qx "query"`): results appear, selection works, cleanup is clean
- test interactive mode (`qx`): input → loading → selection flow
- test history mode (`qx --history`): entries displayed, selection works
- test pipe mode (`echo data | qx "query"`): pipe context passed to LLM
- test action menu with `action_menu: true` in config
- verify terminal cleanup: no artifacts left after exit
- verify theme customization: change colors in config, verify visual changes
