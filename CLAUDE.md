# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**qx** — минималистичная CLI-утилита на Go для генерации shell-команд через LLM. Пользователь описывает задачу на естественном языке, получает несколько вариантов команд через fzf-style интерфейс и выбирает нужную.

## Build and Development Commands

```bash
# Build
go build -o qx .

# Run tests
go test ./...

# Run single test
go test -run TestName ./path/to/package

# Lint (if golangci-lint configured)
golangci-lint run
```

## Architecture

### Core Components

- **CLI layer** (`cmd/`) — обработка аргументов через cobra, подкоманды:
  - `qx "описание"` — основной режим генерации;
  - `echo "data" | qx "use this"` — stdin как контекст для генерации;
  - `qx --shell-integration bash|zsh|fish` — вывод скрипта интеграции;
  - `qx --config` — работа с конфигурацией;
  - `qx --last` (`-l`) — показать последнюю выбранную команду (action menu при `action_menu: true` в конфиге);
  - `qx --history` — интерактивный fzf-picker по истории запросов;
  - `qx --continue` (`-c`) `"уточнение"` — уточнить последнюю команду через follow-up контекст.

- **LLM client** (`internal/llm/`) — OpenAI-compatible клиент. Работает с любым провайдером (OpenAI, Groq, Together, Ollama, LM Studio).

- **UI** (`internal/ui/`) — unified bubbletea-based TUI with fzf-style inline rendering. Handles the full flow: input, loading, command selection, and cleanup. Renders without alternate screen (inline mode), limited to ~40% terminal height. Supports type-to-filter, scroll, auto-select for single results. Configurable theme with fzf-like defaults.

- **Post-selection actions** (`internal/action/`) — меню действий после выбора команды (execute/copy/revise/quit) с TTY-детекцией и raw-mode вводом через `/dev/tty`. Меню показывается когда `action_menu: true` в конфиге И доступен TTY: сначала проверяется stdout (`ShouldPrompt()`), затем stderr (`ShouldPromptStderr()`) как fallback для shell integration mode. Детекция shell integration mode (stdout=pipe, stderr=TTY) вынесена в общий `inShellIntegration()` (var func в `menu.go`), используется и в `menu.go` (ANSI-очистка меню), и в `execute.go` (маршрутизация вывода на `/dev/tty`). Revise позволяет итеративно уточнять команду через follow-up контекст.

- **Shell integration** (`internal/shell/`) — генерация скриптов для bash, zsh, fish. Скрипты встроены через `embed` (`internal/shell/scripts/`). Контракт с Go-кодом по exit code: exit 0 — безусловное обновление буфера (пустой результат = очистка после Execute/Copy), exit 130 — обновление только при непустом результате (сохранение query при отмене). zsh-скрипт использует `zle -I` перед `zle reset-prompt` для инвалидации дисплея после внешних записей в `/dev/tty`.

- **Config** (`internal/config/`) — загрузка и валидация конфигурации из `~/.config/qx/config.yaml`. `Load()` — полная загрузка с валидацией LLM-полей и API-ключа. Включает `ThemeConfig` (настройки TUI-темы, конвертируемые в `ui.Theme`) и `ActionMenu bool` (показ меню действий после выбора команды).

- **Security guard** (`internal/guard/`) — обнаружение секретов в запросах и pipe-контексте, санитизация вывода.

- **History** (`internal/history/`) — персистентное хранение истории запросов в `~/.config/qx/history.json`. Хранит запрос, выбранную команду, pipe-контекст и timestamp. Ротация на 100 записей, атомарная запись.

### Key Libraries

- `spf13/cobra` + `spf13/viper` — CLI и конфигурация;
- `sashabaranov/go-openai` — OpenAI-compatible клиент;
- `charmbracelet/bubbletea` + `charmbracelet/bubbles` — TUI-фреймворк;
- `mattn/go-isatty` — TTY detection for stdin pipe mode and action menu display;
- `atotto/clipboard` — кроссплатформенная работа с буфером обмена;
- `golang.org/x/term` — raw-mode терминала для action menu.

## Configuration

Конфиг располагается в `~/.config/qx/config.yaml`:

```yaml
llm:
  base_url: "https://api.openai.com/v1"
  apikey: "${OPENAI_API_KEY}"
  model: "gpt-4o-mini"

theme:
  prompt: "> "
  pointer: "▌"
  selected_fg: "170"
  match_fg: "205"
  text_fg: "252"
  muted_fg: "241"
  border: "rounded"
  border_fg: "240"

action_menu: false  # show action menu in shell integration mode (default: false)
```

История хранится в `~/.config/qx/history.json` (JSON-массив записей, макс. 100, новейшие в конце).

## Plans

Design proposals and implementation plans are stored in `docs/plans/`.

### Structure

```text
docs/plans/
├── <date>-<short-description>.md   # active plans
└── completed/                       # implemented plans
```

### Naming Convention

```text
YYYYMMDD-short-description.md
```

Example: `20260220-revise-action-and-cleanup.md`

### Workflow

1. **Create** — add new file to `docs/plans/`
2. **Review** — discuss and refine the plan
3. **Implement** — write code according to the plan
4. **Close** — move to `docs/plans/completed/`

### Working with Plans

- Always read the relevant plan before implementing a feature
- Update plan status as work progresses
- Keep implementation aligned with the approved plan
- Document deviations or learnings in the plan

## References

- [fzf](https://github.com/junegunn/fzf) — образец TUI и shell-интеграции;
- [shell-gpt](https://github.com/TheR1D/shell_gpt) — похожий функционал на Python;
- [aichat](https://github.com/sigoden/aichat) — Shell Assistant режим.

<claude-mem-context>
# Recent Activity

<!-- This section is auto-generated by claude-mem. Edit content outside the tags. -->

### Jan 23, 2026

| ID | Time | T | Title | Read |
|----|------|---|-------|------|
| #177 | 9:25 AM | 🔵 | Project Overview: qx - LLM-powered Shell Command Generator | ~501 |
</claude-mem-context>
