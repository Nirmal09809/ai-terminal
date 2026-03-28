# AI Terminal

Enterprise-level AI terminal agent built with Go. A powerful CLI tool that brings the power of LLMs directly into your terminal.

## Features

- **Multiple LLM Providers**: Support for Google Gemini, OpenAI GPT, Anthropic Claude, and Ollama (local)
- **60+ Built-in Tools**: Shell, File, Git, Docker, Network, Dev, Web, System, AI/ML tools
- **Interactive TUI**: Modern terminal UI built with Bubble Tea
- **Session Persistence**: Save and resume conversations
- **Tool Calling Loop**: ReAct pattern for AI function calling
- **Streaming Responses**: Real-time AI output
- **Pure Go**: Lightweight, fast, cross-platform
- **Termux Ready**: ARM64 binary for Android

## Installation

### Termux (Android)

```bash
pkg update && pkg install golang git
git clone https://github.com/Nirmal09809/ai-terminal.git
cd ai-terminal
go build -o bin/ai-terminal ./cmd/ai-terminal
./bin/ai-terminal config create
export GEMINI_API_KEY="your-api-key-here"
./bin/ai-terminal ask "hello"
```

### Linux/macOS

```bash
git clone https://github.com/Nirmal09809/ai-terminal.git
cd ai-terminal
go build -o bin/ai-terminal ./cmd/ai-terminal
./bin/ai-terminal config create
export GEMINI_API_KEY="your-api-key-here"
./bin/ai-terminal ask "hello"
```

### Windows

```bash
git clone https://github.com/Nirmal09809/ai-terminal.git
cd ai-terminal
go build -o bin/ai-terminal.exe ./cmd/ai-terminal
```

## Configuration

On first run, a default config is created at `~/.ai-terminal/config.yaml`

### Provider Setup

#### Gemini (Recommended - Free Tier Available)

```yaml
provider:
  default: gemini
  gemini:
    api_key: "YOUR_GEMINI_API_KEY"
```

Get your API key from: https://aistudio.google.com/app/apikey

Or set environment variable:
```bash
export GEMINI_API_KEY="your-key-here"
```

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

## Tools (60+)

### File Tools (9)
read, write, edit, delete, mkdir, cp, mv, chmod, stat

### Search Tools (5)
glob, grep, find, locate, ripgrep

### Shell Tools (4)
shell, bash, sudo, exec

### Git Tools (7)
git, git_status, git_log, git_diff, git_commit, git_push, git_pull

### Docker Tools (5)
docker, docker_ps, docker_images, docker_run, docker_logs

### Web Tools (4)
search, fetch, scrape, curl

### System Tools (8)
system, cpu, memory, disk, process, top, uptime, whoami

### Network Tools (6)
ping, nslookup, netstat, wget, ssh, scp

### Dev Tools (8)
npm, pip, go, cargo, build, test, lint, format

### AI/ML Tools (6)
analyze, explain, review, refactor, test_gen, doc_gen

### Utility Tools (7)
todo, note, calc, hash, encode, decode, date

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
