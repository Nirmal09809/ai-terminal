# AI Terminal

Enterprise-level AI terminal agent built with Go. A powerful CLI tool that brings the power of LLMs directly into your terminal.

## Features

- **Multiple LLM Providers**: Support for Google Gemini, OpenAI GPT, Anthropic Claude, and Ollama (local)
- **25+ Built-in Tools**: Shell execution, file operations, web search, code analysis, and more
- **Plan Mode**: Analyze complex tasks before execution
- **Session Persistence**: Save and resume conversations
- **MCP Support**: Model Context Protocol integration
- **Enterprise-Grade Security**: Sandboxed execution, rate limiting, audit logging
- **Beautiful TUI**: Modern terminal UI built with Bubble Tea
- **Streaming Responses**: Real-time AI output
- **Cross-Platform**: Works on Linux, macOS, Windows, and Termux (Android)

## Installation

### Quick Install

```bash
go install github.com/ai-terminal/ai-terminal@latest
```

### Build from Source

```bash
git clone https://github.com/ai-terminal/ai-terminal.git
cd ai-terminal
make build
./bin/ai-terminal
```

### Termux (Android)

```bash
pkg update
pkg install golang git
git clone https://github.com/ai-terminal/ai-terminal.git
cd ai-terminal
make build
./bin/ai-terminal
```

## Configuration

On first run, a default config is created at `~/.ai-terminal/config.yaml`

### Provider Setup

#### Gemini (Recommended - Free Tier Available)

```yaml
provider: gemini

providers:
  gemini:
    api_key: "YOUR_GEMINI_API_KEY"
    model: "gemini-2.0-flash"
```

Get your API key from: https://aistudio.google.com/app/apikey

#### OpenAI

```yaml
provider: openai

providers:
  openai:
    api_key: "YOUR_OPENAI_API_KEY"
    model: "gpt-4o"
```

#### Anthropic Claude

```yaml
provider: anthropic

providers:
  anthropic:
    api_key: "YOUR_ANTHROPIC_API_KEY"
    model: "claude-3-5-sonnet-20241022"
```

#### Ollama (Local - Free)

```yaml
provider: ollama

providers:
  ollama:
    url: "http://localhost:11434"
    model: "llama3"
```

Install Ollama: https://ollama.com

## Usage

### Interactive Mode

```bash
ai-terminal
```

### Non-Interactive Mode

```bash
ai-terminal ask "Explain what Go is"
```

### Commands

| Command | Description |
|---------|-------------|
| `ai-terminal` | Start interactive mode |
| `ai-terminal ask <prompt>` | Ask without interactive mode |
| `ai-terminal config` | Create default config |
| `ai-terminal completion bash` | Generate bash completion |
| `ai-terminal version` | Show version |

## Tools

| Tool | Description |
|------|-------------|
| `shell` | Execute shell commands |
| `read` | Read file contents |
| `write` | Write content to files |
| `edit` | Make targeted edits |
| `glob` | Find files by pattern |
| `grep` | Search in files |
| `search` | Web search |
| `fetch` | Fetch URL content |
| `ask_user` | Ask questions |
| `todo` | Manage todo list |
| `memory` | Persistent storage |
| `analyze` | Code analysis |
| `explain` | Explain concepts |
| `review` | Code review |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Enter | Send message |
| Esc | Cancel |
| Ctrl+C | Quit |

## Architecture

```
┌─────────────────────────────────────┐
│           TUI Layer                 │
│       (Bubble Tea UI)               │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│         Agent Engine                │
│    (Tool calling loop)              │
└──────────────┬──────────────────────┘
               │
    ┌──────────┴──────────┐
    ▼                     ▼
┌─────────┐          ┌──────────┐
│ Provider│          │  Tools   │
│  Layer  │          │ Executor │
└─────────┘          └──────────┘
```

## Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Run in development
make dev

# Format code
make fmt
```

## License

MIT License - See LICENSE file

## Contributing

Contributions are welcome! Please open an issue or submit a PR.

## Credits

Built with inspiration from:
- [OpenCode](https://github.com/anomalyco/opencode)
- [Gemini CLI](https://github.com/google-gemini/gemini-cli)
