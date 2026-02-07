# Post-Selection Actions Menu

## Overview

- Add a post-selection action menu that appears after user picks a command from the picker or TUI
- Menu offers three actions: Execute (run via subprocess), Copy (to clipboard), Quit (print to stdout)
- Menu only appears when stdout is a TTY (direct terminal invocation); when stdout is redirected (shell integration Ctrl+G), the command is printed to stdout as before — preserving backward compatibility

## Git Context

- **Branch**: `stdin-pipe-support` (existing)
- **PR**: #24 — дополняем текущий PR новыми коммитами
- Не создавать новую ветку, все изменения коммитить в `stdin-pipe-support`
- Этот план НЕ коммитить — `docs/` в `.gitignore`

## Context (from discovery)

- Files/components involved:
  - `cmd/root.go` — both `generateCommands()` and `runInteractive()` print selected command via `fmt.Println`
  - `internal/picker/picker.go` — `Pick()` returns selected command string
  - `internal/tui/tui.go` — `Run()` returns `Result` interface with `SelectedResult.Command`
  - `internal/tui/model.go` — after selection in `stateSelect`, immediately quits with `tea.Quit`
  - `cmd/stdin.go` — already uses `mattn/go-isatty` for TTY detection (reusable pattern)
- Related patterns: `mattn/go-isatty` already a direct dependency; `atotto/clipboard` is indirect dependency in go.mod
- Dependencies: `atotto/clipboard` needs to be promoted to direct dependency

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
  - tests are not optional - they are a required part of the checklist
  - write unit tests for new functions/methods
  - write unit tests for modified functions/methods
  - add new test cases for new code paths
  - update existing test cases if behavior changes
  - tests cover both success and error scenarios
- **CRITICAL: all tests must pass before starting next task** - no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility

## Testing Strategy

- **Unit tests**: required for every task (see Development Approach above)

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix
- Update plan if implementation deviates from original scope
- Keep plan in sync with actual work done

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): tasks achievable within this codebase - code changes, tests, documentation updates
- **Post-Completion** (no checkboxes): items requiring external action - manual testing, changes in consuming projects, deployment configs, third-party verifications

## Implementation Steps

### Task 1: Create command execution utility

- [x] create `internal/action/execute.go` with `Execute(command string) error`
- [x] detect user shell from `$SHELL` env var, fallback to `/bin/sh`
- [x] run command via `exec.Command(shell, "-c", command)` with inherited stdin/stdout/stderr
- [x] write tests for shell detection logic (with/without $SHELL)
- [x] write tests for Execute with a simple command (e.g., `echo test`)
- [x] run tests - must pass before next task

### Task 2: Create clipboard copy utility

- [x] create `internal/action/clipboard.go` with `CopyToClipboard(command string) error`
- [x] use `atotto/clipboard` directly — promote from indirect to direct dependency
- [x] write tests for CopyToClipboard (mock or skip on CI if no display)
- [x] run tests - must pass before next task

### Task 3: Create post-selection action menu

- [x] create `internal/action/menu.go` with `PromptAction(command string) error`
- [x] display selected command and prompt: `[e]xecute  [c]opy  [q]uit`
- [x] read single keypress from `/dev/tty` (not stdin — stdin may be a pipe)
- [x] on `e`: call `Execute(command)`, return its error
- [x] on `c`: call `CopyToClipboard(command)`, print confirmation, return nil
- [x] on `q` or Enter: return nil (caller prints command to stdout)
- [x] create `ShouldPrompt() bool` — returns true if stdout is a TTY (using `go-isatty` on `os.Stdout.Fd()`)
- [x] write tests for ShouldPrompt with TTY vs non-TTY
- [x] write tests for menu action dispatch logic (mock reader for keypress)
- [x] run tests - must pass before next task

### Task 4: Wire post-selection menu into cmd/root.go

- [x] in `generateCommands()`: after `picker.Pick()`, if `action.ShouldPrompt()` call `action.PromptAction(selected)`; otherwise `fmt.Println(selected)` as before
- [x] in `runInteractive()`: after `tui.Run()` returns `SelectedResult`, if `action.ShouldPrompt()` call `action.PromptAction(r.Command)`; otherwise `fmt.Println(r.Command)` as before
- [x] write tests for generateCommands path with TTY stdout (verify PromptAction is called)
- [x] write tests for generateCommands path with non-TTY stdout (verify fmt.Println behavior preserved)
- [x] run tests - must pass before next task

### Task 5: Verify acceptance criteria

- [x] verify: `docker images | qx "delete large images"` → picker → action menu appears
- [x] verify: `qx` interactive TUI → select command → action menu appears
- [x] verify: shell integration (Ctrl+G) → command goes to readline, no menu
- [x] verify: `e` executes command in subprocess
- [x] verify: `c` copies to clipboard
- [x] verify: `q` prints to stdout
- [x] run full test suite (unit tests)
- [x] run linter - all issues must be fixed

### Task 6: [Final] Update documentation

- [x] update README.md with post-selection actions documentation
- [x] update `rootCmd.Long` description in `cmd/root.go` if needed

## Technical Details

### Action Menu Display

```
  docker stop $(docker ps -q --filter ancestor=nginx)

  [e]xecute  [c]opy  [q]uit
```

- Command displayed with indentation
- Action prompt below
- Single keypress — no Enter required

### TTY Detection for stdout

```go
func ShouldPrompt() bool {
    return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}
```

When shell integration captures stdout, `os.Stdout.Fd()` is not a TTY → menu skipped → command printed to stdout → shell widget places it on command line.

### Command Execution

```go
func Execute(command string) error {
    shell := os.Getenv("SHELL")
    if shell == "" {
        shell = "/bin/sh"
    }
    cmd := exec.Command(shell, "-c", command)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

### Reading keypress from /dev/tty

```go
tty, _ := os.Open("/dev/tty")
// set raw mode to read single keypress without Enter
// read 1 byte, dispatch action
```

## Post-Completion

**Manual verification:**

- test Execute with destructive command confirmation (e.g., `rm` — should work)
- test Copy on macOS (pbcopy) and Linux (xclip/xsel)
- test shell integration (Ctrl+G) still works without menu appearing
- test pipe mode end-to-end: `docker ps | qx "stop nginx" → [e]xecute`
