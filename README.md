# qx

Generate shell commands from natural language using LLM.

![qx demo](assets/demo.gif)

## Features

- Natural language to shell command conversion
- Multiple command variants with fuzzy selection
- Interactive TUI with real-time filtering
- Shell integration (Ctrl+G hotkey) with inline editing support
- Support for OpenAI-compatible APIs

## Installation

### From releases (recommended)

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
  count: 5  # how many commands to suggest
```

Set your API key:

```bash
export OPENAI_API_KEY="your-key-here"
```

## Usage

### Shell integration (Recommended)

```bash
# Add to ~/.bashrc or ~/.zshrc
eval "$(qx --shell-integration zsh)"

# Then reload your shell config
source ~/.zshrc  # or ~/.bashrc

# Now press Ctrl+G in terminal to invoke qx
```

**Inline editing**: Start typing a command, press Ctrl+G, and qx will use your
input as initial query. Add instructions to modify or extend the command.

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

## License

MIT