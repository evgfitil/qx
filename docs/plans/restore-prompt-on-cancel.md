# Restore Prompt on Cancel

## Overview

При отмене выбора команды (Esc) в qx теряется исходный промпт пользователя. Нужно восстанавливать текст в командной строке, чтобы можно было его подправить и попробовать снова.

**Текущее поведение:**
1. Пользователь нажимает `Ctrl+G`, вводит промпт
2. Смотрит варианты — не подходят
3. Нажимает Esc
4. Текст потерян, командная строка пустая

**Желаемое поведение:**
1. При Esc → исходный промпт возвращается в командную строку
2. Можно сразу отредактировать и повторить

## Context

**Файлы затронуты:**
- `internal/tui/model.go` — обработка Esc, возврат результата
- `cmd/root.go` — обработка результата TUI, exit code
- `internal/shell/scripts/bash.sh` — логика восстановления
- `internal/shell/scripts/zsh.zsh` — логика восстановления
- `internal/shell/scripts/fish.fish` — логика восстановления (если есть)

**Текущий flow:**
```
Shell (Ctrl+G) → qx --query "prompt" → TUI → Esc → stdout пустой, exit 0
                                                  → Shell не обновляет READLINE_LINE
                                                  → текст потерян
```

**Решение:**
```
Shell (Ctrl+G) → qx --query "prompt" → TUI → Esc → stdout = "prompt", exit 130
                                                  → Shell видит exit 130
                                                  → восстанавливает READLINE_LINE = stdout
```

## Development Approach

- **Testing approach**: TDD
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run tests after each change

## Testing Strategy

- **Unit tests**: required for every task
- Run `go test ./...` after each change

## Progress Tracking

- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with ➕ prefix
- Document issues/blockers with ⚠️ prefix

## Implementation Steps

### Task 1: Add CancelledResult type to TUI

Нужен способ различать "выбрал команду", "отменил" и "ничего не выбрал".

- [x] write test: `Model.Result()` returns `CancelledResult{Query: "..."}` when Esc pressed
- [x] write test: `Model.Result()` returns `SelectedResult{Command: "..."}` when command selected
- [x] add `Result` interface with `IsCancelled()` method in `internal/tui/result.go`
- [x] add `CancelledResult` struct with `Query` field
- [x] add `SelectedResult` struct with `Command` field
- [x] update `Model` to track cancellation state and store initial query
- [x] update Esc handler to set cancellation state
- [x] add `Model.Result()` method returning appropriate result type
- [x] run tests - must pass before next task

### Task 2: Update cmd/root.go to handle cancellation

- [ ] write test: `runInteractive` outputs original query and exits 130 on cancel
- [ ] write test: `runInteractive` outputs selected command and exits 0 on success
- [ ] update `runInteractive` to call `model.Result()` instead of `model.Selected()`
- [ ] handle `CancelledResult`: print query to stdout, return error with exit code 130
- [ ] handle `SelectedResult`: print command to stdout, return nil
- [ ] add custom error type for cancellation with exit code
- [ ] update main error handling to use correct exit code
- [ ] run tests - must pass before next task

### Task 3: Update shell integration scripts

- [ ] update `bash.sh`: handle exit code 130, restore READLINE_LINE from stdout
- [ ] update `zsh.zsh`: handle exit code 130, restore LBUFFER from stdout
- [ ] update `fish.fish` (if exists): handle exit code 130
- [ ] manual test: verify prompt restoration works in bash
- [ ] manual test: verify prompt restoration works in zsh
- [ ] run tests - must pass before next task

### Task 4: Handle edge cases

- [ ] write test: empty initial query + Esc → exit 130, stdout empty
- [ ] write test: modified query in TUI + Esc → return current (modified) query, not initial
- [ ] update TUI to return current input value on cancel, not initial query
- [ ] run tests - must pass before next task

### Task 5: Verify acceptance criteria

- [ ] verify: Esc restores prompt in bash
- [ ] verify: Esc restores prompt in zsh
- [ ] verify: modified prompt in TUI is restored (not original)
- [ ] verify: selecting command still works correctly
- [ ] run full test suite
- [ ] run linter - all issues must be fixed

### Task 6: [Final] Update documentation

- [ ] update README.md if shell integration section mentions behavior

## Technical Details

**Exit codes:**
- `0` — команда выбрана успешно
- `130` — пользователь отменил (Esc/Ctrl+C), промпт в stdout для восстановления

**Result types:**
```go
type Result interface {
    IsCancelled() bool
}

type CancelledResult struct {
    Query string // текущий текст из input field
}

type SelectedResult struct {
    Command string
}
```

**Shell script logic (bash):**
```bash
if [[ $exit_code -eq 0 && -n "$result" ]]; then
    READLINE_LINE="$result"  # выбранная команда
elif [[ $exit_code -eq 130 ]]; then
    READLINE_LINE="$result"  # восстановленный промпт
fi
```

## Post-Completion

**Manual verification:**
- Протестировать в реальном терминале bash и zsh
- Убедиться, что курсор встаёт в конец восстановленной строки
