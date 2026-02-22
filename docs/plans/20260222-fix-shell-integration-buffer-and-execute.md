# Fix shell integration: buffer not cleared and execute output invisible

## Overview

Two bugs in shell integration mode (`Ctrl+G`) with `action_menu: true`:

1. **LBUFFER not cleared after Execute/Copy**: after any action except Quit, the original command that was in the input line before `Ctrl+G` remains. Copy works (command is in clipboard), but the buffer stays dirty.
2. **Execute output not visible**: command output written to `/dev/tty` may not be properly displayed because zsh's line editor doesn't know the terminal was modified externally.

### Root cause analysis

**Bug 1 — LBUFFER stays unchanged:**

Shell scripts capture stdout: `result=$(qx --query ...)`. After action menu:

- **Quit**: `fmt.Println(command)` writes to stdout → `$result` non-empty → LBUFFER updated
- **Execute**: `Execute(command)` runs command, returns nil → nothing on stdout → `$result` empty
- **Copy**: copies to clipboard, prints to stderr → nothing on stdout → `$result` empty

Shell scripts check `-n "$result"` for ALL exit codes. When result is empty (Execute/Copy), the condition fails and LBUFFER is NOT updated:

```bash
# current (broken) — all three shells
if [[ ($exit_code -eq 0 || $exit_code -eq 130) && -n "$result" ]]; then
    LBUFFER="$result"  # never reached for Execute/Copy
fi
```

**Bug 2 — Execute output invisible:**

When `Execute()` runs inside `$()`, command output goes to `/dev/tty`. After qx exits, zsh's `zle reset-prompt` redraws the prompt but doesn't know the terminal display changed. This can cause the prompt to overwrite or misalign with the command output. Missing `zle -I` (invalidate display) call.

### Data flow (current, broken)

```text
Shell: result=$("qx" --query "$LBUFFER$RBUFFER" 2>/dev/tty </dev/tty)

Execute path:
  qx → action menu → [e]xecute → Execute(command)
    → cmd output to /dev/tty (appears on terminal)
    → returns nil, nothing on stdout
  $result = "" (empty)
  condition: (0 -eq 0) && -n "" → FALSE
  LBUFFER unchanged → old command stays         ← BUG

Copy path:
  qx → action menu → [c]opy → CopyToClipboard(command)
    → clipboard updated, stderr message
    → returns nil, nothing on stdout
  $result = "" (empty)
  condition: FALSE
  LBUFFER unchanged → old command stays         ← BUG
```

### Data flow (fixed)

```text
Shell script: split condition by exit code

  exit 0 (success — Quit/Execute/Copy):
    LBUFFER="$result"
    Quit:    result="command" → LBUFFER="command"
    Execute: result=""        → LBUFFER="" (cleared)
    Copy:    result=""        → LBUFFER="" (cleared)

  exit 130 (cancelled):
    Only update if result non-empty (preserve query in buffer)
    Cancel from picker: result="query" → LBUFFER="query"
    Cancel from menu:   result=""      → LBUFFER unchanged

  zle -I before zle reset-prompt (zsh only):
    Tells zle the terminal display was modified externally
```

## Context

- **Affected files**: `internal/shell/scripts/bash.sh`, `internal/shell/scripts/zsh.zsh`, `internal/shell/scripts/fish.fish` (and their tests if any), `internal/shell/embed_test.go` (embedded script tests)
- **Branch**: `fzf-style-tui-redesign` (all changes MUST be made on this branch only)
- **Go code NOT changed**: `internal/action/menu.go`, `internal/action/execute.go` — these work correctly, the fix is entirely in shell scripts
- **Related fix**: `20260222-fix-action-menu-shell-integration-bugs.md` fixed Execute output routing and menu cleanup; this fixes the shell script side

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- complete each task fully before moving to the next
- make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- run tests after each change

## Testing Strategy

- embed tests: verify shell scripts contain expected patterns (split conditions, `zle -I`)
- **TUI integration tests via `mcp__tui-test`**: automated verification of shell integration in real zsh session (see Task 3)
- manual testing: verify behavior in actual shell integration mode (post-completion)

## TUI Testing Setup

Use the `tui-test` MCP server to test shell integration in a real zsh session.

### Environment preparation

```text
1. Build binary:     go build -o qx .
2. Launch session:   mcp__tui-test__launch_tui("zsh --no-rcs", mode="buffer", dimensions="120x30", timeout=120)
3. Set env vars (send each + Ctrl+M as Enter):
   - export QX_PATH=/Users/erakhmetzyan/personal-dev/qx/qx
   - export OPENAI_API_KEY=$(cat ~/.eliza_token)
4. Load integration: eval "$($QX_PATH --shell-integration zsh)"
```

### Key interaction pattern

```text
1. Type query text on prompt
2. Ctrl+G → qx TUI appears ("> query" on screen)
3. Ctrl+M (Enter) → submit query, wait for "3/3" (results loaded)
4. Ctrl+M (Enter) → select first command, action menu appears ([e]xecute [c]opy [r]evise [q]uit)
5. Send action key (e/c/q)
6. Capture screen → check LBUFFER state on prompt line
```

### Important notes

- use `send_ctrl("m")` for Enter (not `\n` — it sends literal characters)
- use `send_ctrl("g")` for Ctrl+G
- wait for LLM response with `expect_text("3/3", timeout=30)` before proceeding
- API token: `export OPENAI_API_KEY=$(cat ~/.eliza_token)` (eliza yandex API)
- **test queries**: use only read-only, non-modifying unix commands as queries (e.g. `list files in current dir`, `show disk usage`, `show current date`). Never use queries that could generate destructive commands (`rm`, `kill`, `dd`, etc.)

## Progress Tracking

- mark completed items with `[x]` immediately when done
- add newly discovered tasks with + prefix
- document issues/blockers with ! prefix

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): code changes, tests
- **Post-Completion** (no checkboxes): manual testing in shell integration mode

## Implementation Steps

### Task 1: Fix shell scripts to clear buffer on exit 0

- [x] update `internal/shell/scripts/bash.sh`: split the condition — exit 0 always updates `READLINE_LINE` (even if `$result` is empty), exit 130 only updates if `$result` is non-empty
- [x] update `internal/shell/scripts/zsh.zsh`: same split condition for `LBUFFER`; add `zle -I` before `zle reset-prompt` to invalidate display after external terminal writes
- [x] update `internal/shell/scripts/fish.fish`: same split condition for `commandline`
- [x] write/update embed tests: verify bash script contains split condition pattern (exit 0 without `-n` check)
- [x] write/update embed tests: verify zsh script contains `zle -I` before `zle reset-prompt`
- [x] write/update embed tests: verify fish script contains split condition pattern
- [x] run tests (`go test ./internal/shell/...`) — must pass

### Task 2: Verify acceptance criteria

- [x] verify: shell scripts handle exit 0 with empty result (clears buffer)
- [x] verify: shell scripts handle exit 0 with non-empty result (sets buffer to command)
- [x] verify: shell scripts handle exit 130 with non-empty result (preserves query)
- [x] verify: shell scripts handle exit 130 with empty result (buffer unchanged)
- [x] verify: zsh script has `zle -I` for display invalidation
- [x] run full test suite (`go test ./...`)
- [x] build binary (`go build -o qx .`)

### Task 3: TUI integration test via `tui-test` MCP

Automated verification in real zsh session using `mcp__tui-test`. See "TUI Testing Setup" section above for environment details.

- [ ] build fresh binary (`go build -o qx .`)
- [ ] launch zsh session, set env vars (`QX_PATH`, `OPENAI_API_KEY` via `cat ~/.eliza_token`), load shell integration
- [ ] **test Quit**: type text → `Ctrl+G` → submit → select → press `q` → capture screen → verify prompt contains selected command (LBUFFER updated)
- [ ] **test Copy**: type text → `Ctrl+G` → submit → select → press `c` → capture screen → verify prompt is **empty** (LBUFFER cleared, old text gone)
- [ ] **test Execute**: type text → `Ctrl+G` → submit → select → press `e` → capture screen → verify prompt is **empty** (LBUFFER cleared) and command output visible above prompt
- [ ] **test Cancel**: type text → `Ctrl+G` → submit → select → press `Escape` → capture screen → verify prompt still contains **original text** (LBUFFER unchanged)
- [ ] close tui-test session

## Technical Details

### Shell script changes

**bash.sh** (before):

```bash
if [[ ($exit_code -eq 0 || $exit_code -eq 130) && -n "$result" ]]; then
    READLINE_LINE="$result"
    READLINE_POINT=${#result}
fi
```

**bash.sh** (after):

```bash
if [[ $exit_code -eq 0 ]]; then
    READLINE_LINE="$result"
    READLINE_POINT=${#result}
elif [[ $exit_code -eq 130 && -n "$result" ]]; then
    READLINE_LINE="$result"
    READLINE_POINT=${#result}
fi
```

**zsh.zsh** (before):

```zsh
if [[ ($exit_code -eq 0 || $exit_code -eq 130) && -n "$result" ]]; then
    LBUFFER="$result"
    RBUFFER=""
fi
zle reset-prompt
```

**zsh.zsh** (after):

```zsh
if [[ $exit_code -eq 0 ]]; then
    LBUFFER="$result"
    RBUFFER=""
elif [[ $exit_code -eq 130 && -n "$result" ]]; then
    LBUFFER="$result"
    RBUFFER=""
fi
zle -I
zle reset-prompt
```

**fish.fish** (before):

```fish
if test \( $exit_code -eq 0 -o $exit_code -eq 130 \) -a -n "$result"
    commandline -r -- "$result"
end
```

**fish.fish** (after):

```fish
if test $exit_code -eq 0
    commandline -r -- "$result"
else if test $exit_code -eq 130 -a -n "$result"
    commandline -r -- "$result"
end
```

### Behavior matrix after fix

| Action   | exit code | stdout    | Buffer result         |
|----------|-----------|-----------|----------------------|
| Quit     | 0         | command   | LBUFFER = command    |
| Execute  | 0         | (empty)   | LBUFFER = "" (clear) |
| Copy     | 0         | (empty)   | LBUFFER = "" (clear) |
| Cancel (menu) | 130  | (empty)   | LBUFFER unchanged    |
| Cancel (picker) | 130 | query    | LBUFFER = query      |
| Error    | 1         | (empty)   | LBUFFER unchanged    |

### Why `zle -I`

`zle -I` (invalidate) tells zsh's line editor that the terminal display was modified outside of zle's control. Without it, `zle reset-prompt` may not properly account for content written to `/dev/tty` by `Execute()`, leading to display artifacts or the prompt overwriting command output. Safe to call unconditionally — if no external changes happened, it just forces a full redraw.

## Post-Completion

**Manual verification:**

- `action_menu: true` + `Ctrl+G` + select command + `[e]xecute` → command runs, output displayed, input line **cleared**, no old command residue
- `action_menu: true` + `Ctrl+G` + select command + `[c]opy` → clipboard works, input line **cleared**
- `action_menu: true` + `Ctrl+G` + select command + `[q]uit` → selected command inserted into input line
- `action_menu: true` + `Ctrl+G` + select command + `[r]evise` → refinement works, final action clears/sets buffer correctly
- `action_menu: true` + `Ctrl+G` + `Escape`/`Ctrl+C` in action menu → input line **unchanged** (original text preserved)
- `action_menu: true` + `Ctrl+G` + `Escape`/`Ctrl+C` in picker → original query preserved in input line
- `action_menu: false` + `Ctrl+G` + select command → command inserted into input line (no regression)
- direct `qx "query"` (without shell integration) → all actions work as before (no regression)
