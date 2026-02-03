# Add Status Badges to README

## Overview

Добавить status badges в README для визуализации состояния проекта:
- **Build** — статус CI (GitHub Actions);
- **Coverage** — процент покрытия тестами (Codecov);
- **Go Report** — оценка качества кода (goreportcard.com).

## Context

- Репозиторий: `evgfitil/qx`;
- CI workflow уже генерирует `coverage.out`;
- Файлы: `.github/workflows/ci.yml`, `README.md`.

## Development Approach

- **Testing approach**: не требуется (изменения только в CI и документации);
- изменения минимальны: один workflow и README.

## Implementation Steps

### Task 1: Update CI workflow for Codecov

- [x] добавить шаг upload coverage в `.github/workflows/ci.yml` после тестов;
- [x] использовать `codecov/codecov-action@v5`.

### Task 2: Add badges to README

- [x] добавить badge для GitHub Actions build status;
- [x] добавить badge для Codecov coverage;
- [x] добавить badge для Go Report Card;
- [x] разместить бейджи после заголовка `# qx`, перед описанием.

### Task 3: Verify

- [x] проверить синтаксис markdown;
- [x] убедиться что ссылки корректны.

## Technical Details

### Badges format

```markdown
[![Build](https://github.com/evgfitil/qx/actions/workflows/ci.yml/badge.svg)](https://github.com/evgfitil/qx/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/evgfitil/qx/branch/main/graph/badge.svg)](https://codecov.io/gh/evgfitil/qx)
[![Go Report Card](https://goreportcard.com/badge/github.com/evgfitil/qx)](https://goreportcard.com/report/github.com/evgfitil/qx)
```

### Codecov action

```yaml
- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v5
  with:
    files: coverage.out
```

## Post-Completion

**Manual verification:**
- после merge в main, проверить что Codecov получил данные;
- первый запуск Go Report Card может занять время — посетить goreportcard.com/report/github.com/evgfitil/qx для инициализации.
