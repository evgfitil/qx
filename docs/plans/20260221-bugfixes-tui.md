# TUI Bugfixes

## Overview

Сводный план багфиксов TUI в ветке `fzf-style-tui-redesign`. Обновляется по мере обнаружения новых проблем.

**Branch**: `fzf-style-tui-redesign` (текущая ветка, все изменения реализуются здесь).

## Bug 1: Shell integration с существующим буфером сразу запускает генерацию

**Проблема**: когда пользователь вызывает qx хоткеем (`Ctrl+G`) при непустой командной строке (например, `find . -type f -name '*.go' -exec grep -l 'for ' {} +`), текст передаётся через `--query` и qx сразу переходит в `stateLoading` — нет возможности дописать/отредактировать запрос.

**Ожидаемое поведение**: текст из буфера предзаполняет поле ввода, курсор в конце, пользователь может дописать контекст (например, "- допиши обработку ошибок") и нажать Enter для генерации.

**Root cause**: `internal/ui/model.go:95-100` — при `InitialQuery != ""` состояние сразу меняется на `stateLoading`:

```go
initialState := stateInput
if opts.InitialQuery != "" {
    ta.SetValue(opts.InitialQuery)
    ta.CursorEnd()
    initialState = stateLoading  // ← баг: пропускает stateInput
}
```

**Затронутые файлы:**

- `internal/ui/model.go` — убрать переход в `stateLoading` при `InitialQuery`
- `internal/ui/model_test.go` — обновить тесты

**Фикс**: убрать `initialState = stateLoading`. Предзаполнять textarea, но оставаться в `stateInput`. Пользователь жмёт Enter — генерация запускается через штатный `handleEnter()`.

## Bug 2: `qx --last` показывает action menu при `action_menu: false`

**Проблема**: `qx --last` всегда показывает action menu (execute/copy/revise/quit), даже когда `action_menu: false` в конфиге. На скриншоте видно: пользователь вызывает `qx --last` и видит меню, хотя ожидает просто получить последнюю команду в stdout.

**Ожидаемое поведение**: при `action_menu: false` — `qx --last` просто выводит последнюю команду в stdout (shell integration вставит в буфер). При `action_menu: true` — показывает меню как сейчас.

**Root cause**: `cmd/root.go:204` — `runLast()` хардкодит `true` для actionMenu:

```go
return handleSelectedCommand(entry.Selected, entry.Query, entry.PipeContext, true)
```

Все остальные пути (`generateCommands()`, `runInteractive()`) используют `cfg.ActionMenu`, а `runLast()` — нет.

**Затронутые файлы:**

- `cmd/root.go` — `runLast()` должен загружать конфиг и передавать `cfg.ActionMenu`
- `cmd/root_test.go` — обновить тесты

**Фикс**: в `runLast()` загрузить конфиг через `config.Load()` и передать `cfg.ActionMenu` вместо `true`.

## Bug 3: Textarea всегда занимает 3 строки вместо auto-resize

**Проблема**: поле ввода (`stateInput`) всегда отображает 3 строки с prompt `>`, даже когда текст пустой или помещается в одну строку. Выглядит неопрятно — 2 лишних `>` без текста.

**Ожидаемое поведение**: textarea начинается с 1 строки. Когда текст переносится (word-wrap) — расширяется до 2, потом до 3 строк (MaxHeight). При удалении текста — сжимается обратно.

**Root cause**: `internal/ui/model.go:88-89` — фиксированная высота:

```go
ta.MaxHeight = 3
ta.SetHeight(3)  // ← всегда 3 строки
```

Textarea из bubbles НЕ поддерживает auto-expand — высота фиксируется через `SetHeight()`, контент скроллится внутри viewport. Но `LineInfo().Height` возвращает количество визуальных (wrapped) строк, что позволяет реализовать auto-resize вручную.

**Затронутые файлы:**

- `internal/ui/model.go` — начальная высота 1, динамический `SetHeight()` в `Update()`
- `internal/ui/model_test.go` — тесты auto-resize

**Фикс**: `SetHeight(1)`, оставить `MaxHeight = 3`. В `Update()` после обновления textarea вызывать `SetHeight(clamp(LineInfo().Height, 1, MaxHeight))`.

## Refactor 1: Заменить ручную проверку mutually exclusive флагов на cobra native

**Проблема**: в `cmd/root.go:93-105` ручной подсчёт `flagCount++` для проверки взаимной исключительности `--last`, `--history`, `--continue`. Boilerplate-код, который cobra умеет делать нативно.

**Текущий код:**

```go
flagCount := 0
if lastFlag { flagCount++ }
if historyFlag { flagCount++ }
if continueFlag { flagCount++ }
if flagCount > 1 {
    return fmt.Errorf("--last, --history, and --continue are mutually exclusive")
}
```

**Затронутые файлы:**

- `cmd/root.go` — добавить `MarkFlagsMutuallyExclusive` в `init()`, удалить ручную проверку из `run()`
- `cmd/root_test.go` — обновить тесты (cobra выдаёт свой формат ошибки)

**Фикс**: добавить в `init()`:

```go
rootCmd.MarkFlagsMutuallyExclusive("last", "history", "continue")
```

Удалить блок `flagCount` из `run()`. Cobra сама валидирует и выдаёт ошибку: `"if any flags in the group [last history continue] are set none of the others can be; [last history] were all set"`.

## Refactor 2: Добавить короткие флаги `-l` и `-c`

**Проблема**: `--last` и `--continue` — часто используемые флаги без коротких алиасов. В CLAUDE.md уже задокументированы как `(-l)` и `(-c)`, но не реализованы. `--history` оставляем без короткого флага — `-h` занят cobra для `--help`.

**Затронутые файлы:**

- `cmd/root.go` — изменить `BoolVar` на `BoolVarP` для `--last` и `--continue`

**Фикс**: в `init()`:

```go
rootCmd.Flags().BoolVarP(&lastFlag, "last", "l", false, "show last selected command and open action menu")
rootCmd.Flags().BoolVarP(&continueFlag, "continue", "c", false, "refine the last command with a new query")
```

## Development Approach

- **Testing approach**: Regular (code first, then tests)
- complete each task fully before moving to the next
- make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- run tests after each change

## Testing Strategy

- **Unit tests**: required for every task
- Bug 1: model с `InitialQuery` должен начинаться в `stateInput` с предзаполненным текстом
- Bug 1: `Init()` должен возвращать `textarea.Blink` (не spinner + generate)
- Bug 1: Enter в `stateInput` с предзаполненным текстом запускает генерацию
- Bug 1: обратная совместимость: пустой `InitialQuery` — поведение без изменений
- Bug 2: `runLast()` с `action_menu: false` — выводит команду в stdout без меню
- Bug 2: `runLast()` с `action_menu: true` — показывает меню (без регрессии)
- Bug 3: textarea начинается с height=1, растёт при wrap, сжимается при удалении
- Bug 3: placeholder на 1 строке, без лишних `>`
- Refactor 1: `--last --history` вместе — cobra возвращает ошибку (без ручного `flagCount`)
- Refactor 1: одиночные флаги работают как раньше

## Progress Tracking

- mark completed items with `[x]` immediately when done
- add newly discovered tasks with + prefix
- document issues/blockers with ! prefix
- update plan if implementation deviates from original scope

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): code changes, tests
- **Post-Completion** (no checkboxes): manual testing in shell integration mode

## Implementation Steps

### Bug 1, Task 1: Fix initial state when InitialQuery is provided (`internal/ui/model.go`)

- [x] в `newModel()`: убрать `initialState = stateLoading` из блока `if opts.InitialQuery != ""`; оставить `ta.SetValue()` и `ta.CursorEnd()` — textarea предзаполнена, но state остаётся `stateInput`
- [x] убрать присвоение `originalQuery: opts.InitialQuery` в конструкторе Model — `originalQuery` теперь устанавливается только в `handleEnter()` при нажатии Enter
- [x] update `model_test.go`: тест что `newModel()` с `InitialQuery` создаёт модель в `stateInput`
- [x] update `model_test.go`: тест что `Init()` с предзаполненным query возвращает `textarea.Blink`, а не spinner+generate
- [x] update `model_test.go`: тест что Enter в `stateInput` с предзаполненным текстом переводит в `stateLoading` и запускает генерацию
- [x] проверить и обновить существующие тесты, которые полагаются на `InitialQuery` → `stateLoading`
- [x] run tests (`go test ./...`) — must pass

### Bug 2, Task 1: Fix `runLast()` to respect `action_menu` config (`cmd/root.go`)

- [x] в `runLast()`: загрузить конфиг через `config.Load()` и передать `cfg.ActionMenu` вместо хардкоженного `true` в вызов `handleSelectedCommand()`
- [x] update `root_test.go`: тест что `runLast()` с `action_menu: false` выводит команду в stdout без вызова action menu
- [x] update `root_test.go`: тест что `runLast()` с `action_menu: true` показывает action menu (без регрессии)
- [x] проверить и обновить существующие тесты `runLast`
- [x] run tests (`go test ./...`) — must pass

### Bug 3, Task 1: Auto-resize textarea в stateInput (`internal/ui/model.go`)

- [x] в `newModel()`: изменить `ta.SetHeight(3)` на `ta.SetHeight(1)` — начинаем с 1 строки; `ta.MaxHeight = 3` оставить
- [x] в `Update()`: после обновления textarea в блоке `stateInput` — вычислить нужную высоту через `m.textArea.LineInfo().Height` и вызвать `m.textArea.SetHeight()` если высота изменилась
- [x] обработать edge case: пустой textarea (Value() == "") — `LineInfo().Height` может вернуть 0, нужно `max(height, 1)`
- [x] update `model_test.go`: тест что начальная высота textarea = 1
- [x] update `model_test.go`: тест что после ввода длинного текста (> width) высота увеличивается
- [x] update `model_test.go`: тест что высота не превышает MaxHeight (3)
- [x] run tests (`go test ./...`) — must pass

### Refactor 1, Task 1: Cobra native mutually exclusive flags (`cmd/root.go`)

- [x] в `init()`: добавить `rootCmd.MarkFlagsMutuallyExclusive("last", "history", "continue")` после определения флагов
- [x] в `run()`: удалить весь блок `flagCount` (строки 93-105)
- [x] update `root_test.go`: обновить тесты взаимной исключительности — проверить что cobra возвращает ошибку при `--last --history` (формат ошибки cobra отличается от текущего)
- [x] update `root_test.go`: проверить что одиночные флаги (`--last`, `--history`, `--continue`) работают без ошибок
- [x] run tests (`go test ./...`) — must pass

### Refactor 2, Task 1: Добавить короткие флаги (`cmd/root.go`)

- [ ] в `init()`: `--last` — заменить `BoolVar` на `BoolVarP` с коротким флагом `"l"`
- [ ] в `init()`: `--continue` — заменить `BoolVar` на `BoolVarP` с коротким флагом `"c"`
- [ ] update `root_test.go`: тест что `qx -l` работает как `qx --last`
- [ ] update `root_test.go`: тест что `qx -c "query"` работает как `qx --continue "query"`
- [ ] run tests (`go test ./...`) — must pass

### Verify acceptance criteria (all bugs)

- [ ] verify: Bug 1 — `newModel()` с `InitialQuery` → `stateInput` с предзаполненным текстом
- [ ] verify: Bug 1 — `newModel()` без `InitialQuery` → `stateInput` с пустым полем (без регрессии)
- [ ] verify: Bug 1 — Enter на предзаполненном тексте → `stateLoading` → генерация
- [ ] verify: Bug 2 — `qx --last` с `action_menu: false` просто выводит команду
- [ ] verify: Bug 2 — `qx --last` с `action_menu: true` показывает меню
- [ ] verify: Bug 3 — textarea начинается с 1 строки (1 `>`)
- [ ] verify: Bug 3 — при длинном тексте textarea расширяется до 2-3 строк
- [ ] verify: Bug 3 — при удалении текста textarea сжимается обратно
- [ ] verify: Refactor 1 — `--last --history` → ошибка от cobra
- [ ] verify: Refactor 1 — `--last` отдельно работает
- [ ] verify: Refactor 1 — блок `flagCount` удалён из `run()`
- [ ] verify: Refactor 2 — `qx -l` работает как `qx --last`
- [ ] verify: Refactor 2 — `qx -c "query"` работает как `qx --continue "query"`
- [ ] run full test suite (`go test ./...`)
- [ ] run linter (`golangci-lint run`)
- [ ] build binary (`go build -o qx .`)

## Technical Details

### Изменение в `newModel()` (model.go:95-100)

До:

```go
initialState := stateInput
if opts.InitialQuery != "" {
    ta.SetValue(opts.InitialQuery)
    ta.CursorEnd()
    initialState = stateLoading
}
```

После:

```go
if opts.InitialQuery != "" {
    ta.SetValue(opts.InitialQuery)
    ta.CursorEnd()
}
```

`initialState` всегда `stateInput`. Переход в `stateLoading` происходит штатно через `handleEnter()` при нажатии Enter.

### Поле `originalQuery`

Сейчас `originalQuery` устанавливается в конструкторе из `opts.InitialQuery`. После фикса — только в `handleEnter()` (строка 266: `m.originalQuery = query`). Это корректно, т.к. пользователь может отредактировать запрос перед отправкой.

### Изменение в `runLast()` (root.go:189-205)

До:

```go
func runLast() error {
    store, err := newHistoryStore()
    // ...
    entry, err := store.Last()
    // ...
    return handleSelectedCommand(entry.Selected, entry.Query, entry.PipeContext, true)
}
```

После:

```go
func runLast() error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    store, err := newHistoryStore()
    // ...
    entry, err := store.Last()
    // ...
    return handleSelectedCommand(entry.Selected, entry.Query, entry.PipeContext, cfg.ActionMenu)
}
```

### Auto-resize логика в Update() (model.go)

В блоке `if m.state == stateInput || m.state == stateSelect` после обновления textarea:

```go
if m.state == stateInput {
    needed := m.textArea.LineInfo().Height
    if needed < 1 {
        needed = 1
    }
    if needed != m.textArea.Height() {
        m.textArea.SetHeight(needed)
    }
}
```

`MaxHeight = 3` остаётся — `SetHeight()` внутри textarea уже clamp-ит значение в диапазон `[1, MaxHeight]`.

### Начальная высота (model.go:88-89)

До:

```go
ta.MaxHeight = 3
ta.SetHeight(3)
```

После:

```go
ta.MaxHeight = 3
ta.SetHeight(1)
```

### Cobra MarkFlagsMutuallyExclusive (root.go)

В `init()` после определения флагов:

```go
func init() {
    generateCommandsFn = generateCommands

    rootCmd.Flags().StringVar(&shellIntegration, "shell-integration", "", "...")
    // ... остальные флаги ...
    rootCmd.Flags().BoolVar(&continueFlag, "continue", false, "...")

    rootCmd.MarkFlagsMutuallyExclusive("last", "history", "continue")
}
```

В `run()` удалить строки 93-105 (блок `flagCount`). Логика `if lastFlag` / `if historyFlag` / `if continueFlag` остаётся — cobra проверяет взаимную исключительность ДО вызова `RunE`.

## Post-Completion

**Manual verification:**

- Bug 1: test shell integration (`Ctrl+G`) с непустой строкой — должно показать поле ввода с текстом, можно дописать и нажать Enter
- Bug 1: test shell integration (`Ctrl+G`) с пустой строкой — поле ввода пустое, как раньше
- Bug 1: test прямой вызов `qx "query"` — не затронут (позиционные аргументы, другой путь)
- Bug 2: test `qx --last` с `action_menu: false` — выводит последнюю команду без меню
- Bug 2: test `qx --last` с `action_menu: true` — показывает action menu как раньше
- Bug 3: пустое поле — 1 строка с placeholder, без лишних `>`
- Bug 3: набрать длинный текст — textarea плавно расширяется до 2-3 строк
- Bug 3: стереть текст — textarea сжимается обратно до 1 строки
- Bug 3: предзаполненный запрос из shell integration — высота соответствует длине текста
- Refactor 1: `qx --last --history` — cobra выдаёт ошибку, не доходя до `run()`
- Refactor 1: `qx --last` / `qx --history` / `qx --continue "query"` — работают как раньше
- Refactor 2: `qx -l` → показывает последнюю команду
- Refactor 2: `qx -c "уточнение"` → refine последней команды
