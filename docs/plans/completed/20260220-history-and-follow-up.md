# History and Follow-up Context

## Overview

Implement query history persistence and single-level follow-up context for qx. Currently every qx invocation is completely independent — no state is saved between runs. This plan adds:

- **History storage** — save every query + generated commands + selected command to `~/.config/qx/history.json`
- **`--last`** — show the last selected command and open action menu
- **`--history`** — fzf-style interactive picker over past queries, selected entry goes to action menu
- **`--continue "refinement"`** — read last query + selected command, inject as context into LLM prompt for refinement

These features share a common storage layer: follow-up reads the latest history entry, `--last` displays it, `--history` lists all entries.

## Context

### Files/components involved

- `internal/history/` — new package (storage layer)
- `internal/llm/base.go` — prompt construction, inject follow-up context
- `internal/llm/prompt.go` — system prompt extension for follow-up mode
- `internal/llm/provider.go` — `Provider` interface extension
- `internal/tui/result.go` — add `Query` field to `SelectedResult`
- `internal/tui/model.go` — propagate query to result
- `internal/picker/picker.go` — reuse for history picker
- `cmd/root.go` — new flags, save-to-history, history/last/continue flows

### Related patterns

- Stdin pipe context (`cmd/stdin.go`, `base.go:55-57`) — follow-up context injection is architecturally similar
- `SelectedResult` (`tui/result.go:17`) currently returns only `Command string`; `originalQuery` lives in TUI model but is not propagated up
- `go-fuzzyfinder` picker (`internal/picker/picker.go`) — reusable for `--history`

### Dependencies

- No new external dependencies needed; `encoding/json` + `os` for storage
- Existing `go-fuzzyfinder` for history picker

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
- Focus on `internal/history/` package — pure logic, easy to test
- LLM prompt changes — test message construction, not actual API calls
- Integration points (`cmd/root.go`) — test flag parsing and flow routing

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with warning prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Create history storage package

- [x] create `internal/history/history.go` with `Entry` struct: `Query string`, `Commands []string`, `Selected string`, `PipeContext string`, `Timestamp time.Time`
- [x] implement `Store` type with `filePath string` field and `NewStore(dir string) *Store` constructor
- [x] implement `Store.Add(entry Entry) error` — read file, append entry, rotate to last 100 entries, write back atomically (write to temp file, rename)
- [x] implement `Store.Last() (Entry, error)` — return most recent entry
- [x] implement `Store.List() ([]Entry, error)` — return all entries, newest first
- [x] write tests for `Add` (empty file, append, rotation at limit)
- [x] write tests for `Last` (empty history, single entry, multiple entries)
- [x] write tests for `List` (empty, ordering)
- [x] run tests — must pass before next task

### Task 2: Propagate query through SelectedResult

- [x] add `Query string` field to `SelectedResult` in `internal/tui/result.go`
- [x] set `Query` from `m.originalQuery` when constructing `SelectedResult` in `internal/tui/model.go`
- [x] update `cmd/root.go` non-interactive path (`generateCommands`) — propagate query alongside selected command
- [x] verify existing tests pass with the new field
- [x] write tests for `SelectedResult` query propagation
- [x] run tests — must pass before next task

### Task 3: Save to history after command selection

- [x] initialize `history.Store` in `cmd/root.go` using config dir path
- [x] after successful command selection in `runInteractive()`, save entry with query, all commands, selected command, and pipe context
- [x] after successful command selection in non-interactive path, save entry similarly
- [x] write tests for save-on-selection flow (mock or use temp dir)
- [x] run tests — must pass before next task

### Task 4: Implement --last flag

- [x] add `--last` flag to cobra command in `cmd/root.go`
- [x] when `--last` is set: load `Store.Last()`, print the selected command, invoke `action.PromptAction()` on it
- [x] handle empty history case — print informative error message
- [x] write tests for `--last` flag (with history, without history)
- [x] run tests — must pass before next task

### Task 5: Implement --history flag with fzf picker

- [x] add `--history` flag to cobra command in `cmd/root.go`
- [x] when `--history` is set: load `Store.List()`, format entries for display (query + selected command + timestamp)
- [x] use `internal/picker` (go-fuzzyfinder) to present entries for selection
- [x] selected entry goes to `action.PromptAction()` with the stored command
- [x] handle empty history — print informative error message
- [x] write tests for history picker flow
- [x] run tests — must pass before next task

### Task 6: Implement --continue flag with follow-up context

- [x] add `--continue` flag to cobra command in `cmd/root.go`
- [x] extend `Provider` interface: add `previousQuery string` and `previousCommand string` parameters to `Generate()`, or add a `FollowUpContext` struct parameter
- [x] in `internal/llm/base.go`, when follow-up context is present: build messages as `[system, user(prev query), assistant(prev command), user(new query)]` instead of `[system, user(query)]`
- [x] in `internal/llm/prompt.go`, extend `SystemPrompt()` to include follow-up rules when in continue mode (e.g., "refine the previous command based on user's new request")
- [x] in `cmd/root.go`, when `--continue` is set: load `Store.Last()`, pass previous context through to `Generate()`
- [x] handle empty history — print informative error message
- [x] write tests for follow-up prompt construction (verify message structure)
- [x] write tests for system prompt follow-up rules
- [x] run tests — must pass before next task

### Task 7: Verify acceptance criteria

- [x] verify `--last` shows last command and opens action menu
- [x] verify `--history` opens fzf picker with past queries
- [x] verify `--continue "refinement"` sends previous context to LLM
- [x] verify history rotation works at 100 entries
- [x] verify empty history is handled gracefully for all three flags
- [x] verify pipe context is preserved in history entries
- [x] run full test suite (`go test ./...`)
- [x] run linter (`golangci-lint run`)
- [x] verify test coverage for `internal/history/` package

### Task 8: Update documentation

- [x] update README.md with new flags: `--last`, `--history`, `--continue`
- [x] add usage examples for each flag
- [x] update roadmap (`docs/plans/20260203-product-roadmap-research.md`) — mark History and Follow-up as implemented in checklist

## Technical Details

### History entry structure

```go
type Entry struct {
    Query       string    `json:"query"`
    Selected    string    `json:"selected"`
    PipeContext string    `json:"pipe_context,omitempty"`
    Timestamp   time.Time `json:"timestamp"`
}
```

Note: `Commands []string` was removed in the revise-action-and-cleanup follow-up work.

### Storage file

- Location: `~/.config/qx/history.json`
- Format: JSON array of Entry objects, newest last
- Rotation: keep last 100 entries, trim on write
- Atomic writes: write to `history.json.tmp`, then `os.Rename`

### Follow-up prompt structure

When `--continue` is used, LLM messages become:

```text
[system]: base system prompt + follow-up rules
[user]:   {previous query}
[assistant]: {previous selected command}
[user]:   {new refinement query}
```

Follow-up system prompt addition:

```text
The user is refining a previous command. Consider the conversation history
and generate commands that address the user's refinement request.
```

### Flag interactions

- `--last` and `--history` are mutually exclusive with regular query mode
- `--continue` requires a query argument (the refinement)
- `--continue` combined with pipe is valid — both previous context and new pipe context are included

## Post-Completion

**Manual verification:**

- test `--continue` with real LLM to verify refinement quality
- test history picker UX with 50+ entries for usability
- test shell integration (Ctrl+G) still works correctly after changes
- verify `--last` / `--history` work in non-TTY environments
