# Revise Action and History Cleanup

## Overview

Add `[r]evise` option to the post-selection action menu for iterative command refinement without leaving the interactive flow. After selecting a command, the user can press `r` to type a refinement query via a simple TTY prompt, generating new commands with follow-up context. This creates a natural loop: generate -> select -> revise -> generate -> select -> execute/copy until the user is satisfied.

Additionally, remove the unused `Commands []string` field from `history.Entry` — no feature (`--last`, `--history`, `--continue`) uses it.

## Context

### Branch

All work is done in the existing branch `history-and-follow-up` (PR #31). Commit on top of existing history/follow-up changes.

### Files/components involved

- `internal/action/menu.go` — action menu: add `ActionRevise`, update display, return `ReviseRequestedError`
- `internal/action/menu_test.go` — tests for revise keypress and dispatch
- `internal/action/revise.go` — new file for `ReadRefinement()` TTY prompt
- `cmd/root.go` — handle revise result, create generation loop with `FollowUpContext`
- `cmd/root_test.go` — tests for revise loop, updated history saving
- `internal/history/history.go` — remove `Commands` field from `Entry`
- `internal/history/history_test.go` — update tests for new `Entry` shape
- `internal/tui/result.go` — remove `Commands` from `SelectedResult`
- `internal/tui/model.go` — stop populating `Commands` in `Result()`
- `internal/tui/model_test.go` — update tests

### Related patterns

- `FollowUpContext` (`internal/llm/provider.go`) — already exists, used by `--continue`
- `promptActionWith()` testability pattern — accepts `io.Reader` for testing without `/dev/tty`
- `readAction()` pattern — opens `/dev/tty` with raw mode, reads keypress

### Dependencies

- No new external dependencies
- `golang.org/x/term` already used for raw-mode TTY reading
- `bufio.Scanner` from stdlib for line reading in refinement prompt

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
  - write unit tests for new functions/methods
  - write unit tests for modified functions/methods
  - tests cover both success and error scenarios
- **CRITICAL: all tests must pass before starting next task** — no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change
- Maintain backward compatibility

## Testing Strategy

- **Unit tests**: required for every task
- Action menu: test `readKeypress` with `'r'/'R'` input, test `dispatchAction` returns `ReviseRequestedError`
- Refinement prompt: test `ReadRefinement` with mock reader
- Revise loop in `cmd/root.go`: test with temp history store and mock generation
- History cleanup: update existing tests to match new `Entry` shape without `Commands`

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with warning prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Remove Commands field from history.Entry and SelectedResult

- [x] remove `Commands []string` from `Entry` struct in `internal/history/history.go`
- [x] remove `Commands []string` from `SelectedResult` in `internal/tui/result.go`
- [x] stop populating `Commands` in `Model.Result()` in `internal/tui/model.go`
- [x] remove `Commands` from all `history.Entry{}` literals in `cmd/root.go` (two places: `runInteractive` and `generateCommands`)
- [x] update tests in `internal/history/history_test.go` — remove `Commands` from test entries
- [x] update tests in `internal/tui/model_test.go` — remove `Commands` assertions
- [x] update tests in `cmd/root_test.go` — remove `Commands` from test entries and assertions
- [x] run tests — must pass before next task

### Task 2: Add ActionRevise to action menu

- [x] add `ActionRevise` to the `Action` enum in `internal/action/menu.go`
- [x] add `'r'/'R'` case to `readKeypress()` returning `ActionRevise`
- [x] add `[r]evise` to the display string in `promptActionWith()`
- [x] create `ReviseRequestedError` type with `Command string` field in `internal/action/menu.go`
- [x] add `ActionRevise` case to `dispatchAction()` returning `&ReviseRequestedError{Command: command}`
- [x] write tests for `readKeypress` with `'r'` and `'R'` input
- [x] write test for `dispatchAction(ActionRevise, ...)` returning `ReviseRequestedError`
- [x] write test for `promptActionWith` with `'r'` input returning `ReviseRequestedError`
- [x] run tests — must pass before next task

### Task 3: Add refinement prompt reader

- [x] create `internal/action/revise.go` with `ReadRefinement() (string, error)` function
- [x] implement TTY reading: open `/dev/tty`, restore terminal from raw mode (if still in raw), print `> ` prompt to stderr, read one line with `bufio.Scanner`, return trimmed input
- [x] add testable variant `readRefinementFrom(r io.Reader) (string, error)` that reads from provided reader
- [x] handle empty input — return descriptive error
- [x] write tests for `readRefinementFrom` (normal input, empty input, EOF)
- [x] run tests — must pass before next task

### Task 4: Implement revise loop in cmd/root.go

- [x] modify `handleSelectedCommand` to detect `ReviseRequestedError` via `errors.As`
- [x] when revise detected: call `ReadRefinement()` to get refinement query, then call `generateCommands(refinement, "", followUp)` with `FollowUpContext{PreviousQuery: originalQuery, PreviousCommand: selectedCommand}`
- [x] the loop naturally continues because `generateCommands` calls `handleSelectedCommand` at the end, which shows the action menu again
- [x] track original query through the loop — first revision uses the query from `SelectedResult` or the CLI arg, subsequent revisions use the previous refinement query
- [x] save to history only on final selection (execute/copy/quit), not on intermediate revise steps
- [x] handle `ReadRefinement` errors and cancellation gracefully
- [x] write tests for revise loop with mock history store (verify `FollowUpContext` is constructed correctly)
- [x] write tests for revise cancellation (empty input, read error)
- [x] write test that history is saved only on final selection, not on intermediate revisions
- [x] run tests — must pass before next task

### Task 5: Verify acceptance criteria

- [x] verify `[r]evise` appears in action menu display
- [x] verify pressing `r` triggers refinement prompt
- [x] verify refinement query generates new commands with follow-up context
- [x] verify multi-step revision works (revise -> revise -> execute)
- [x] verify history is saved only on final selection
- [x] verify `Commands` field is removed from history entries
- [x] verify existing flags (`--last`, `--history`, `--continue`) still work
- [x] run full test suite (`go test ./...`)
- [x] run linter (`golangci-lint run`)

### Task 6: Update documentation

- [ ] update README.md — add revise option to post-selection actions section
- [ ] update CLAUDE.md project overview if needed (action menu description)

## Technical Details

### ReviseRequestedError

```go
type ReviseRequestedError struct {
    Command string
}

func (e *ReviseRequestedError) Error() string {
    return "revise requested"
}
```

### Refinement prompt

Simple TTY prompt printed to stderr, reads one line from `/dev/tty`:

```text
  [e]xecute  [c]opy  [r]evise  [q]uit r

  > make it recursive
```

### Revise loop flow

```text
handleSelectedCommand(command)
  └─ PromptAction(command)
       └─ user presses 'r'
            └─ returns ReviseRequestedError{Command: command}
  └─ errors.As(err, &reviseErr)
       └─ ReadRefinement() → "make it recursive"
       └─ generateCommands("make it recursive", pipeContext, &FollowUpContext{...})
            └─ LLM generates refined commands
            └─ picker.Pick()
            └─ saveToHistory() ← only here, on final pick
            └─ handleSelectedCommand(newCommand) ← loop back
```

### History Entry after cleanup

```go
type Entry struct {
    Query       string    `json:"query"`
    Selected    string    `json:"selected"`
    PipeContext string    `json:"pipe_context,omitempty"`
    Timestamp   time.Time `json:"timestamp"`
}
```

Backward compatible: existing history.json files with `commands` field will be ignored by `json.Unmarshal` since the field no longer exists in the struct.

## Post-Completion

**Manual verification:**

- test revise flow with real LLM: generate -> select -> revise -> verify refined commands are contextually relevant
- test multi-step revision (3+ iterations) for UX smoothness
- test revise in shell integration mode (Ctrl+G) — verify terminal state is clean after revise loop
- test revise in non-TTY environment — verify graceful fallback
- verify existing history.json files with `commands` field still load correctly after cleanup
