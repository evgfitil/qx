# Fix theme colors missing in shell integration mode

## Overview

Theme colors (selected item highlighting, pointer color, muted text, border color) are not rendered when qx runs through shell integration (`Ctrl+G`) or when stdout is not a native TTY. Colors work correctly in `qx --last` -> revise flow because stdout/stderr are natural TTYs.

**Root cause**: lipgloss v1.1.0 uses a global default renderer that lazily detects color profile from `os.Stderr`. When qx runs via shell integration, stdout is captured by `$()` (pipe), and even though `2>/dev/tty` redirects stderr, the `colorprofile.Detect` may not correctly determine terminal capabilities. All theme style methods use `lipgloss.NewStyle()` which relies on this potentially broken default renderer, while bubbletea outputs to `/dev/tty` via `tea.WithOutput(tty)`.

**Fix**: create an explicit `*lipgloss.Renderer` from the same `/dev/tty` file descriptor used for bubbletea output, inject it into `Theme`, and use `renderer.NewStyle()` in all style getters instead of `lipgloss.NewStyle()`.

**Branch**: `fzf-style-tui-redesign` (this is a bugfix within the ongoing TUI redesign, not a separate feature branch).

## Context (from discovery)

**Files to modify:**

- `internal/ui/theme.go` — add renderer field, update style getters
- `internal/ui/run.go` — create tty-based renderer, inject into theme before model creation
- `internal/ui/model.go` — use theme renderer for textarea/spinner styling
- `internal/ui/theme_test.go` — update tests for renderer-aware theme
- `internal/ui/model_test.go` — update tests to provide renderer in theme

**Files unchanged:**

- `internal/config/config.go` — `ThemeConfig.ToTheme()` stays the same (no renderer at config level)
- `cmd/root.go` — no changes needed (renderer is injected inside `Run()`/`RunSelector()`)
- `internal/ui/view.go` — style getters already called on theme, no direct `lipgloss.NewStyle()` calls
- `internal/ui/result.go` — no styling

**Dependencies:** none new (lipgloss v1.1.0 already provides `lipgloss.NewRenderer()`)

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change

## Testing Strategy

- **Unit tests**: required for every task
- Theme with renderer: verify style getters produce styles via the provided renderer
- Theme without renderer (nil): verify fallback to `lipgloss.NewStyle()` for backward compatibility
- Model construction: verify renderer flows through to textarea/spinner

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix
- Update plan if implementation deviates from original scope

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): code changes, tests
- **Post-Completion** (no checkboxes): manual testing in shell integration mode

## Implementation Steps

### Task 1: Add renderer to Theme and update style getters (`internal/ui/theme.go`)

- [x] add `renderer *lipgloss.Renderer` field to `Theme` struct
- [x] add `WithRenderer(r *lipgloss.Renderer) Theme` method that returns a copy with renderer set
- [x] update `SelectedStyle()` to use `t.renderer.NewStyle()` when renderer is non-nil, fallback to `lipgloss.NewStyle()`
- [x] update `NormalStyle()`, `MutedStyle()`, `PromptStyle()`, `BorderStyle()` the same way
- [x] extract a helper `newStyle()` on Theme to avoid repeating nil-check in every getter
- [x] update `theme_test.go`: test style creation with explicit renderer
- [x] update `theme_test.go`: test style creation without renderer (nil fallback)
- [x] run tests — must pass before next task

### Task 2: Create renderer from `/dev/tty` in Run/RunSelector (`internal/ui/run.go`)

- [x] in `Run()`: after opening `/dev/tty`, create `lipgloss.NewRenderer(tty)` and call `opts.Theme.WithRenderer(renderer)` before passing to `newModel()`
- [x] in `RunSelector()`: same — create renderer from tty, call `theme.WithRenderer(renderer)` before `newSelectorModel()`
- [x] handle fallback: if `/dev/tty` open fails (non-unix), use `lipgloss.DefaultRenderer()`
- [x] run tests — must pass before next task

### Task 3: Use theme renderer for textarea and spinner styling (`internal/ui/model.go`)

- [ ] in `newTextArea()`: use theme's `newStyle()` for `FocusedStyle.Text` and `FocusedStyle.CursorLine` instead of `lipgloss.NewStyle()`
- [ ] update `newModel()` and `newSelectorModel()` to pass theme to `newTextArea()` so it can use the renderer
- [ ] update `model_test.go`: verify model construction works with renderer-aware theme
- [ ] run tests — must pass before next task

### Task 4: Verify acceptance criteria

- [ ] verify all style getters use the renderer when provided
- [ ] verify fallback works when renderer is nil (backward compat, tests without TTY)
- [ ] run full test suite (`go test ./...`)
- [ ] run linter (`golangci-lint run`) — all issues must be fixed
- [ ] build binary (`go build -o qx .`)

## Technical Details

### Theme struct change

```go
type Theme struct {
    Prompt     string
    Pointer    string
    SelectedFg string
    MatchFg    string
    TextFg     string
    MutedFg    string
    Border     string
    BorderFg   string
    renderer   *lipgloss.Renderer // unexported, set by WithRenderer()
}

func (t Theme) WithRenderer(r *lipgloss.Renderer) Theme {
    t.renderer = r
    return t
}

func (t Theme) newStyle() lipgloss.Style {
    if t.renderer != nil {
        return t.renderer.NewStyle()
    }
    return lipgloss.NewStyle()
}
```

### Run() change

```go
func Run(opts RunOptions) (Result, error) {
    tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
    if err != nil {
        tty = os.Stdout
    } else {
        defer tty.Close()
    }

    renderer := lipgloss.NewRenderer(tty)
    opts.Theme = opts.Theme.WithRenderer(renderer)

    m := newModel(opts)
    p := tea.NewProgram(m, tea.WithOutput(tty), tea.WithInputTTY())
    // ...
}
```

### Style getter pattern

```go
func (t Theme) SelectedStyle() lipgloss.Style {
    return t.newStyle().Foreground(lipgloss.Color(t.SelectedFg)).Bold(true)
}
```

## Post-Completion

**Manual verification:**

- test shell integration mode (`Ctrl+G`): TUI should render with colors (selected item highlighted, border colored)
- test direct mode (`qx "query"`): colors work
- test `qx --last` -> revise: colors still work (regression check)
- test interactive mode (`qx`): colors in all states (input prompt, spinner, selector)
- verify on terminal without `/dev/tty` support (fallback to default renderer)
