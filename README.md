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
- Shell integration (Ctrl+G hotkey) with inline editing support
- Pipe/stdin support for context-aware command generation
- Command description mode to explain existing commands
- Support for OpenAI-compatible APIs

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap evgfitil/tap
brew install qx
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

```bash
# Bash: add to ~/.bashrc
eval "$(qx --shell-integration bash)"

# Zsh: add to ~/.zshrc
eval "$(qx --shell-integration zsh)"

# Fish: add to ~/.config/fish/config.fish
qx --shell-integration fish | source

# Then reload your shell config
source ~/.bashrc  # or ~/.zshrc, or restart Fish

# Now press Ctrl+G in terminal to invoke qx
```

**Inline editing**: Start typing a command, press Ctrl+G, and qx will use your
input as initial query. Add instructions to modify or extend the command.

**Prompt restoration**: Press Esc to cancel selection and restore your query
to the command line for editing.

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

### Pipe/stdin support

qx can read context from stdin to generate more relevant commands:

```bash
# Use file content as context
cat error.log | qx "find the cause"

# Use command output as context
docker ps | qx "stop the nginx container"

# Use git diff as context
git diff | qx "create a commit message"
```

### Command description mode

Explain what a command does instead of generating new ones:

```bash
qx -d "find . -name '*.go' -exec grep TODO {} +"
# Outputs explanation of the command

qx --describe "awk '{print \$1}' file.txt"
# Explains what awk is doing
```

## License

MIT
