# Fix action_menu in shell integration mode

## Overview

`action_menu: true` в конфиге не работает в shell integration mode — вместо показа меню команда просто вставляется в буфер терминала.

## Root Cause

Shell integration скрипты захватывают stdout через `$()`:

```bash
result=$("$qx_cmd" --query "$READLINE_LINE" </dev/tty 2>/dev/tty)
```

`handleSelectedCommand()` (root.go:350) проверяет `shouldPromptFn()` → `action.ShouldPrompt()` → проверяет **только stdout**. Stdout — pipe → false → меню не показывается.

При этом stderr перенаправлен на `/dev/tty` и IS a TTY. Меню уже рендерится на stderr (`fmt.Fprintf(os.Stderr, ...)`), поэтому достаточно добавить fallback-проверку stderr.

## Context

- **Затронутые файлы**: `internal/action/menu.go`, `cmd/root.go`, `cmd/root_test.go`
- **Branch**: `fzf-style-tui-redesign` (все изменения реализуются здесь)
- **Справка**: на ветке `short-flags-and-action-menu-config` (коммит `e59e49a`) этот фикс уже был реализован, но не попал в текущую ветку

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- complete each task fully before moving to the next
- make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- run tests after each change

## Testing Strategy

- `ShouldPromptStderr()` — unit test
- `handleSelectedCommand` — тесты для комбинаций: stdout TTY, stdout pipe + stderr TTY + actionMenu true/false

## Progress Tracking

- mark completed items with `[x]` immediately when done
- add newly discovered tasks with + prefix
- document issues/blockers with ! prefix

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): code changes, tests
- **Post-Completion** (no checkboxes): manual testing in shell integration mode

## Implementation Steps

### Task 1: Add `ShouldPromptStderr()` (`internal/action/menu.go`)

- [x] добавить `ShouldPromptStderr()` — проверяет `os.Stderr.Fd()` через `isatty`, аналогично `ShouldPrompt()`
- [x] write test for `ShouldPromptStderr()`
- [x] run tests (`go test ./internal/action/...`) — must pass

### Task 2: Wire stderr check in `handleSelectedCommand()` (`cmd/root.go`)

- [ ] добавить `shouldPromptStderrFn = action.ShouldPromptStderr` в overridable vars (строка 40)
- [ ] в `handleSelectedCommand()` (строка 350): заменить `if !actionMenu || !shouldPromptFn()` на двухступенчатую проверку — если stdout не TTY, но actionMenu и stderr TTY → показать меню
- [ ] write tests: actionMenu=true, stdout=pipe, stderr=TTY → меню показывается
- [ ] write tests: actionMenu=false, stdout=pipe, stderr=TTY → меню НЕ показывается
- [ ] write tests: actionMenu=true, stdout=TTY → меню показывается (без регрессии)
- [ ] update existing tests that mock `shouldPromptFn`
- [ ] run tests (`go test ./...`) — must pass

### Task 3: Verify acceptance criteria

- [ ] verify: все комбинации stdout/stderr/actionMenu работают корректно
- [ ] run full test suite (`go test ./...`)
- [ ] build binary (`go build -o qx .`)

## Technical Details

### Новая функция в `menu.go`

```go
// ShouldPromptStderr returns true if stderr is a TTY. Used in shell
// integration mode where stdout is captured but stderr goes to /dev/tty.
func ShouldPromptStderr() bool {
    return isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
}
```

### Новая переменная в `root.go`

```go
shouldPromptStderrFn = action.ShouldPromptStderr
```

### Изменение логики в `handleSelectedCommand()`

До:

```go
if !actionMenu || !shouldPromptFn() {
    // print command, no menu
}
```

После:

```go
showMenu := shouldPromptFn()
if !showMenu && actionMenu {
    showMenu = shouldPromptStderrFn()
}
if !showMenu {
    // print command, no menu
}
```

Логика: сначала проверяем stdout (прямой вызов `qx`). Если stdout — pipe, но `actionMenu: true` — проверяем stderr (shell integration mode, где `2>/dev/tty`).

## Post-Completion

**Manual verification:**

- `action_menu: true` в конфиге + `Ctrl+G` → после выбора команды показывается меню (execute/copy/revise/quit)
- `action_menu: false` в конфиге + `Ctrl+G` → команда вставляется в буфер без меню
- прямой вызов `qx "query"` с `action_menu: true` → меню показывается (stdout = TTY)
- `qx --last` с `action_menu: true` → меню показывается
- `qx --last` с `action_menu: false` → команда выводится без меню
