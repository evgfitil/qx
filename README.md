# qx

[![Build](https://github.com/evgfitil/qx/actions/workflows/ci.yml/badge.svg)](https://github.com/evgfitil/qx/actions/workflows/ci.yml)
[![Coverage](https://codecov.io/gh/evgfitil/qx/branch/main/graph/badge.svg)](https://codecov.io/gh/evgfitil/qx)
[![Go Report Card](https://goreportcard.com/badge/github.com/evgfitil/qx)](https://goreportcard.com/report/github.com/evgfitil/qx)

Generate shell commands from natural language using LLM.

![qx demo](assets/demo.gif)

## Features

- Natural language to shell command conversion
- Multiple command variants with fuzzy selection
- Interactive TUI with real-time filtering
- Shell integration (Ctrl+G hotkey) for Bash, Zsh, and Fish with inline editing support
- Support for OpenAI-compatible APIs

## Installation

### Homebrew (macOS/Linux)

```bash
brew install evgfitil/tap/qx
```

### From releases

```bash
curl -sSL https://github.com/evgfitil/qx/releases/latest/download/qx_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv qx /usr/local/bin/
```

### From source

```bash
go install github.com/evgfitil/qx@latest
```

## Configuration

Create `~/.config/qx/config.yaml`:

```yaml
llm:
  base_url: "https://api.openai.com/v1"  # or any OpenAI-compatible API
  model: "gpt-4o-mini"
  count: 3  # how many commands to suggest
  apikey: "your-key-here"  # optional, can use env variable instead
```

**API Key**: Set via `OPENAI_API_KEY` environment variable or `llm.apikey` in config.
Environment variable takes precedence if both are set.

```bash
# Option 1: environment variable
export OPENAI_API_KEY="your-key-here"

# Option 2: set apikey in config.yaml (see above)
```

## Usage

### Shell integration (Recommended)

Add the appropriate line to your shell configuration:

```bash
# Bash (~/.bashrc)
eval "$(qx --shell-integration bash)"

# Zsh (~/.zshrc)
eval "$(qx --shell-integration zsh)"

# Fish (~/.config/fish/config.fish)
qx --shell-integration fish | source
```

Reload your shell config and press Ctrl+G to invoke qx.

**Inline editing**: Start typing a command, press Ctrl+G, and qx will use your
input as initial query. Add instructions to modify or extend the command.

**Prompt restoration**: Press Esc to cancel selection and restore your query
to the command line for editing.

**Error display**: If qx encounters an error (invalid configuration, API failure, etc.),
the error message is displayed in the terminal. Normal cancellation via Esc does not
produce error output.

### Direct mode

```bash
qx "find all go files modified today"
```

### Interactive mode

```bash
qx
# Type your query, press Enter, select command
```

### Pre-filled query

```bash
qx --query "git log"
# TUI opens with input field pre-filled
```

### Pipe mode

Pipe command output into qx to provide context for more precise generation:

```bash
ls -la | qx "delete files larger than 1GB"
docker ps | qx "stop all nginx containers"
git branch | qx "delete all merged branches"
```

Pipe mode also works with interactive TUI:

```bash
kubectl get pods | qx
# Type your query in the TUI with pod list as context
```

Stdin input is limited to 64KB. Content is checked for secrets before being sent to the LLM.

## License

MIT
