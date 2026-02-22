# Fix action menu bugs in shell integration mode

## Overview

Two bugs in action menu when used via shell integration (`Ctrl+G`) with `action_menu: true`:

1. **Execute output capture**: selecting `[e]xecute` causes the executed command's stdout to be inserted into the terminal input line instead of being displayed normally
2. **Menu residue**: after any action (execute/copy/quit), the action menu text and the selected command remain visible on the terminal

### Bug 1: Execute output capture

Root cause: `Execute()` sets `cmd.Stdout = os.Stdout`, but in shell integration mode `os.Stdout` is the capture pipe from `$()` in the shell script. The command's output gets captured and assigned to `READLINE_LINE`/`LBUFFER`/`commandline`.

### Bug 2: Menu residue

Root cause: `promptActionWith()` prints the menu to stderr (`/dev/tty` in shell integration mode) but never clears it after the user makes a choice. The menu text and command preview remain on the terminal after qx exits.

## Context

- **Affected files**: `internal/action/execute.go`, `internal/action/menu.go` (and their tests)
- **Branch**: `fzf-style-tui-redesign` (all changes MUST be made on this branch only)
- **Related fix**: `20260221-fix-action-menu-shell-integration.md` fixed action menu display; this fixes execution output routing and menu cleanup
- **Shell integration scripts** (`internal/shell/scripts/`): capture stdout via `$()`, redirect stderr to `/dev/tty`

### Data flow (current, broken)

```text
Shell: result=$("qx" --query ... 2>/dev/tty </dev/tty)
  qx stdout = capture pipe
  qx stderr = /dev/tty

Bug 1 — Execute output goes to capture pipe:
  Action menu → [e]xecute → Execute(command)
    cmd.Stdout = os.Stdout = capture pipe  ← BUG
    command output captured by $()
    READLINE_LINE = command output          ← BUG MANIFESTS

Bug 2 — Menu text stays on terminal:
  promptActionWith() prints to stderr (/dev/tty):
    \n  {command}\n\n  [e]xecute  [c]opy  [r]evise  [q]uit
  After action chosen: no cleanup           ← BUG
  Menu text remains visible in terminal     ← BUG MANIFESTS
```

### Data flow (fixed)

```text
Bug 1 fix — redirect Execute() output to /dev/tty:
  Execute() detects shell integration mode:
    stdout is NOT TTY AND stderr IS TTY → open /dev/tty for output
    cmd.Stdout = /dev/tty  ← output goes to terminal
    cmd.Stderr = /dev/tty
    $() captures empty string
    shell script: result="" → no READLINE_LINE update

Bug 2 fix — clear menu after action chosen:
  After readAction() returns, in shell integration mode:
    use ANSI escape sequences to erase menu lines from terminal
    \033[A (move up) + \033[2K (clear line) for each menu line
  In normal mode: keep current behavior (newlines for spacing)
```

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- complete each task fully before moving to the next
- make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- run tests after each change

## Testing Strategy

- `Execute()` with stdout=pipe + stderr=TTY: output should NOT appear on stdout
- `Execute()` with stdout=TTY: output should appear on stdout (no regression)
- `Execute()` with /dev/tty open failure: graceful fallback
- `promptActionWith()` in shell integration mode: output includes ANSI clear sequences
- `promptActionWith()` in normal mode: output does NOT include clear sequences (no regression)

## Progress Tracking

- mark completed items with `[x]` immediately when done
- add newly discovered tasks with + prefix
- document issues/blockers with ! prefix

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): code changes, tests
- **Post-Completion** (no checkboxes): manual testing in shell integration mode

## Implementation Steps

### Task 1: Redirect Execute() output to /dev/tty in shell integration mode

- [x] in `Execute()` (`internal/action/execute.go`): detect shell integration mode — `!isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stderr.Fd())`
- [x] when in shell integration mode: open `/dev/tty` for writing (`os.OpenFile("/dev/tty", os.O_WRONLY, 0)`) and use it for `cmd.Stdout` and `cmd.Stderr`
- [x] defer close of the opened tty file
- [x] when `/dev/tty` open fails: fall back to inherited os.Stdout/os.Stderr (current behavior)
- [x] write test: stdout=pipe, stderr=TTY → `Execute("echo hello")` produces no output on stdout pipe
- [x] write test: stdout=TTY → `Execute("echo hello")` output appears on stdout (no regression)
- [x] run tests (`go test ./internal/action/...`) — must pass

### Task 2: Clear action menu after action chosen in shell integration mode

- [x] add `inShellIntegration()` helper in `menu.go`: returns true when stdout is NOT TTY and stderr IS TTY
- [x] in `promptActionWith()`, after `readAction()` returns: if `inShellIntegration()`, emit ANSI sequences to stderr to erase menu lines (move cursor up + clear line for each of the 4 printed lines: empty, command, empty, menu)
- [x] keep current behavior (newlines for spacing) when NOT in shell integration mode
- [x] write test: `promptActionWith()` with stdout=pipe, stderr=pipe → stderr output includes ANSI clear sequences (`\033[A`, `\033[2K`)
- [x] write test: `promptActionWith()` with stdout=TTY → stderr output does NOT include ANSI clear sequences (no regression)
- [x] run tests (`go test ./internal/action/...`) — must pass

### Task 3: Verify acceptance criteria

- [ ] verify: Execute() in shell integration mode doesn't leak output to stdout pipe
- [ ] verify: Execute() in normal mode still outputs to stdout
- [ ] verify: action menu is erased after action chosen in shell integration mode
- [ ] verify: action menu spacing preserved in normal mode
- [ ] run full test suite (`go test ./...`)
- [ ] build binary (`go build -o qx .`)

## Technical Details

### Shell integration detection

```go
func inShellIntegration() bool {
    stdoutIsPipe := !isatty.IsTerminal(os.Stdout.Fd()) &&
        !isatty.IsCygwinTerminal(os.Stdout.Fd())
    stderrIsTTY := isatty.IsTerminal(os.Stderr.Fd()) ||
        isatty.IsCygwinTerminal(os.Stderr.Fd())
    return stdoutIsPipe && stderrIsTTY
}
```

This pattern (stdout=pipe, stderr=TTY) uniquely identifies shell integration mode where `2>/dev/tty` redirects stderr to the terminal while `$()` captures stdout.

### Modified Execute() (conceptual)

```go
func Execute(command string) error {
    shell := detectShell()
    cmd := exec.Command(shell, "-c", command)

    stdoutIsNotTTY := !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd())
    stderrIsTTY := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())

    if stdoutIsNotTTY && stderrIsTTY {
        ttyOut, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
        if err == nil {
            defer func() { _ = ttyOut.Close() }()
            cmd.Stdout = ttyOut
            cmd.Stderr = ttyOut
        } else {
            cmd.Stdout = os.Stdout
            cmd.Stderr = os.Stderr
        }
    } else {
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
    }

    // stdin handling (unchanged)
    ...
}
```

### Menu cleanup ANSI sequences

The menu prints 4 lines to stderr:

```text
\n                                          ← line 1 (empty)
  {command}\n                               ← line 2
\n                                          ← line 3 (empty)
  [e]xecute  [c]opy  [r]evise  [q]uit      ← line 4 (cursor here)
```

Cleanup after action chosen (shell integration mode only):

```go
if inShellIntegration() {
    // Erase menu: move up 3 lines from current position, clear to end of screen
    fmt.Fprintf(os.Stderr, "\r\033[3A\033[J")
}
```

- `\r` — move to beginning of current line
- `\033[3A` — move cursor up 3 lines (to line 1)
- `\033[J` — clear from cursor to end of screen

## Post-Completion

**Manual verification:**

- `action_menu: true` + `Ctrl+G` + select command + `[e]xecute` → command runs, output displayed in terminal, input line stays clean, menu is erased
- `action_menu: true` + `Ctrl+G` + select command + `[c]opy` → clipboard works, menu is erased
- `action_menu: true` + `Ctrl+G` + select command + `[q]uit` → command inserted into input line, menu is erased
- `action_menu: true` + `Ctrl+G` + select command + `[r]evise` → refinement prompt appears, menu is erased
- direct `qx "query"` + `[e]xecute` → command runs, output displayed normally, menu stays visible with spacing (no regression)
- direct `qx "query"` + `[q]uit` → command printed to stdout, menu stays visible (no regression)
