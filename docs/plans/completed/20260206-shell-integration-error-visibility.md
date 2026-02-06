# Shell Integration Error Visibility Fix

## Overview

When qx is invoked via shell hotkey (Ctrl+G) without `OPENAI_API_KEY` configured, the error
message is invisible to the user. The terminal prompt redraw overwrites the error output.

**Problem**: Shell integration scripts redirect stderr to `/dev/tty` (`2>/dev/tty`), but after
qx exits with an error, the shell immediately redraws the prompt (`zle reset-prompt` in zsh,
readline redraw in bash, `commandline -f repaint` in fish), which overwrites the error message.

**Solution**: Modify shell scripts to capture stderr to a temp file, and on error (exit code != 0
and != 130), display the error to `/dev/tty` before the prompt redraws. Go code stays unchanged.

## Context

- Files involved: `internal/shell/scripts/bash.sh`, `internal/shell/scripts/zsh.zsh`,
  `internal/shell/scripts/fish.fish`
- Tests: `internal/shell/integration_test.go`
- Error originates in `config.Load()` (returns error when API key is missing) before TUI starts

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility

## Testing Strategy

- **Unit tests**: update `integration_test.go` to verify new patterns in shell scripts
- Shell scripts themselves are embedded via `//go:embed`, tested through `TestScriptContent`

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix

## Implementation Steps

### Task 1: Update bash.sh to capture and display stderr on error

- [x] modify `bash.sh` to capture stderr to temp file instead of `/dev/tty`
- [x] add error display logic: on exit code != 0 and != 130, print captured stderr to `/dev/tty`
- [x] clean up temp file after use
- [x] update `TestScriptContent` for bash: verify new patterns (e.g., `mktemp`, error handling)
- [x] run tests - must pass before next task

### Task 2: Update zsh.zsh to capture and display stderr on error

- [x] modify `zsh.zsh` to capture stderr to temp file instead of `/dev/tty`
- [x] add error display logic before `zle reset-prompt`
- [x] clean up temp file after use
- [x] update `TestScriptContent` for zsh: verify new patterns
- [x] run tests - must pass before next task

### Task 3: Update fish.fish to capture and display stderr on error

- [x] modify `fish.fish` to capture stderr to temp file instead of `/dev/tty`
- [x] add error display logic before `commandline -f repaint`
- [x] clean up temp file after use
- [x] update `TestScriptContent` for fish: verify new patterns
- [x] run tests - must pass before next task

### Task 4: Verify acceptance criteria

- [x] verify: missing API key error is visible when invoked via hotkey
- [x] verify: normal operation (with API key) works unchanged
- [x] verify: cancellation (Esc/Ctrl+C) still works correctly (exit 130)
- [x] run full test suite (`go test ./...`)
- [x] run linter (`golangci-lint run`)

### Task 5: [Final] Update documentation

- [x] update README.md if needed
- [x] do NOT commit `docs/` directory — it is local-only, excluded from git

## Technical Details

### Bash script change

```bash
__qx_widget() {
    local qx_cmd="${QX_PATH:-qx}"
    local result err_file
    err_file=$(mktemp)
    result=$("$qx_cmd" --query "$READLINE_LINE" </dev/tty 2>"$err_file")
    local exit_code=$?
    if [[ ($exit_code -eq 0 || $exit_code -eq 130) && -n "$result" ]]; then
        READLINE_LINE="$result"
        READLINE_POINT=${#result}
    elif [[ $exit_code -ne 0 && $exit_code -ne 130 ]]; then
        echo "" >/dev/tty
        cat "$err_file" >/dev/tty
    fi
    rm -f "$err_file"
}
```

### Zsh script change

```zsh
__qx_widget() {
    local qx_cmd="${QX_PATH:-qx}"
    local current_buffer="$LBUFFER$RBUFFER"
    local result err_file
    err_file=$(mktemp)
    result=$("$qx_cmd" --query "$current_buffer" 2>"$err_file" </dev/tty)
    local exit_code=$?
    if [[ ($exit_code -eq 0 || $exit_code -eq 130) && -n "$result" ]]; then
        LBUFFER="$result"
        RBUFFER=""
    elif [[ $exit_code -ne 0 && $exit_code -ne 130 ]]; then
        echo "" >/dev/tty
        cat "$err_file" >/dev/tty
    fi
    rm -f "$err_file"
    zle reset-prompt
}
```

### Fish script change

```fish
function __qx_widget
    set -l qx_cmd qx
    if set -q QX_PATH; and test -n "$QX_PATH"
        set qx_cmd $QX_PATH
    end
    set -l current_buffer (commandline)
    set -l err_file (mktemp)
    set -l result ($qx_cmd --query "$current_buffer" 2>$err_file </dev/tty | string collect)
    set -l exit_code $pipestatus[1]
    if test \( $exit_code -eq 0 -o $exit_code -eq 130 \) -a -n "$result"
        commandline -r -- "$result"
    else if test $exit_code -ne 0 -a $exit_code -ne 130
        echo "" >/dev/tty
        cat $err_file >/dev/tty
    end
    rm -f $err_file
    commandline -f repaint
end
```

### Key design decisions

- **Temp file vs variable capture**: temp file is simplest and most portable across all three
  shells; created only on hotkey press, immediately cleaned up
- **`echo ""` before error**: ensures error appears on a new line, not on the prompt line that
  will be redrawn
- **Exit code 130 excluded**: cancellation should not show errors, consistent with current behavior
