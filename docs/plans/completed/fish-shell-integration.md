# Fish Shell Integration

## Overview

Add Fish shell support for qx shell integration. Currently qx supports bash and zsh via `--shell-integration` flag and Ctrl+G hotkey. Fish shell users cannot use this integration.

This feature adds a Fish shell function that binds Ctrl+G to invoke qx, following the same pattern as existing bash/zsh implementations. It also updates the CLI flag handler, error message, and README documentation.

## Context

- Files involved:
  - `internal/shell/integration.go` — shell script dispatcher (embed + switch)
  - `internal/shell/scripts/bash.sh` — bash reference implementation
  - `internal/shell/scripts/zsh.zsh` — zsh reference implementation
  - `cmd/root.go` — CLI flag handler for `--shell-integration`
  - `README.md` — usage documentation
- Pattern: scripts embedded via `go:embed`, returned by `Script()` function based on shell name
- No existing tests for `internal/shell/` package

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

- **Unit tests**: required for every task
- Test `Script()` function for all shells including fish and error case
- Test fish script content contains expected Fish shell constructs

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Create Fish shell script

- [x] create `internal/shell/scripts/fish.fish` with `__qx_widget` function
- [x] implement Fish key binding for Ctrl+G using `bind \cg`
- [x] handle command buffer capture via `commandline` builtin
- [x] handle exit code and result replacement logic
- [x] verify script manually against Fish shell syntax documentation
- [x] run tests (`go test ./...`) - must pass before next task

### Task 2: Add Fish support to integration.go

- [x] add `go:embed scripts/fish.fish` variable in `internal/shell/integration.go`
- [x] add `case "fish"` to `Script()` switch statement
- [x] update error message to include fish in supported shells list
- [x] write tests for `Script()` — success cases for bash, zsh, fish
- [x] write tests for `Script()` — error case for unsupported shell
- [x] write test verifying fish script contains expected Fish constructs (`bind`, `commandline`)
- [x] run tests (`go test ./...`) - must pass before next task

### Task 3: Verify acceptance criteria

- [x] verify `qx --shell-integration fish` outputs valid Fish script
- [x] verify `qx --shell-integration bash` and `zsh` still work unchanged
- [x] verify unsupported shell returns error with updated message
- [x] run full test suite (`go test ./...`)
- [x] run linter (`golangci-lint run`) - all issues must be fixed

### Task 4: [Final] Update documentation

- [x] add Fish shell instructions to README.md shell integration section
- [x] update README to mention Fish alongside bash/zsh where relevant

## Technical Details

### Fish script structure

Fish uses different primitives than bash/zsh:

| Aspect | Bash | Zsh | Fish |
|--------|------|-----|------|
| Buffer capture | `$READLINE_LINE` | `$LBUFFER$RBUFFER` | `commandline` |
| Buffer update | Set `READLINE_LINE` | Set `LBUFFER` | `commandline -r` |
| Key binding | `bind -x '"\C-g": ...'` | `bindkey '^G' ...` | `bind \cg ...` |
| Function def | `func() { }` | `func() { }` | `function func ... end` |

### Fish script expected behavior

1. Capture current command buffer via `commandline`
2. Call `qx --query "$buffer"` with I/O redirected to `/dev/tty`
3. On success (exit 0 or 130) with non-empty result: replace buffer via `commandline -r`
4. On cancellation: restore original query to buffer
5. Repaint prompt via `commandline -f repaint`

## Post-Completion

**Manual verification:**

- test Ctrl+G in Fish shell with empty buffer
- test Ctrl+G with pre-typed text
- test Esc cancellation preserves buffer
- test `QX_PATH` override works in Fish
