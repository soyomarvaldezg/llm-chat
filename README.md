# LLM Chat CLI

A powerful, fast, and beautiful terminal-based chat application for interacting with multiple Large Language Model providers.

```
╔═══════════════════════════════════════╗
║     🤖 LLM Chat CLI v0.1.0        ║
╚═══════════════════════════════════════╝
```

## ✨ Features

### Core Features

- 🚀 **Multiple LLM Providers** - Ollama, Together AI, Groq, SambaNova, Google Gemini
- 💬 **Interactive Chat Mode** - Full conversational context
- 🔧 **Shell Mode** - Pipe input for scripting and automation
- ⚡ **Streaming Responses** - Real-time output as models think
- 🔄 **Model Switching** - Switch models on the fly
- 📝 **Multiple Output Formats** - text, json, markdown, raw

### Advanced Features

- 📊 **Prompt Assessment** - AI-powered analysis of your prompts (8 criteria)
- ✨ **Auto-Improvement** - Let AI help you write better prompts
- 📚 **Prompt Engineering Guide** - Built-in best practices
- 💾 **Persistent History** - All conversations saved and searchable
- 🔍 **Search History** - Find past conversations
- 📤 **Export Conversations** - Markdown, JSON, or plain text
- 📈 **Usage Statistics** - Track your conversations

### Polish

- 🎨 **Beautiful UI** - Colors, emojis, clean layout
- 🎯 **Smart Commands** - Intuitive `/` commands
- ⚙️ **Highly Configurable** - CLI flags for everything
- 🔐 **Privacy-Focused** - Local-first, optional history saving

---

## 🚀 Quick Start

### Installation

#### Option 1: Build from source (recommended)

```bash
# Clone the repository
git clone https://github.com/soyomarvaldezg/llm-chat.git
cd llm-chat

# Build
make build

# Or install to your PATH
make install
```

#### Option 2: Manual build

```bash
git clone https://github.com/soyomarvaldezg/llm-chat.git
cd llm-chat
go build -o llm-chat ./cmd/llm-chat
```

### Setup

Configure at least one provider:

```bash
# For Ollama (local, free)
export OLLAMA_URL=http://localhost:11434

# For Groq (fast, free tier)
export GROQ_API_KEY=your-api-key

# For Together AI
export TOGETHER_API_KEY=your-api-key

# For SambaNova
export SAMBA_API_KEY=your-api-key

# For Google Gemini
export GEMINI_API_KEY=your-api-key
```

### Basic Usage

```bash
# Start interactive chat
llm-chat

# Use specific provider
llm-chat -p groq

# Shell mode with piped input
cat myfile.py | llm-chat -s "explain this code"

# With assessment enabled
llm-chat --assess
```

---

## 📖 Usage Examples

### Interactive Mode

```bash
# Basic chat
llm-chat

# With verbose metrics
llm-chat --verbose

# With prompt assessment
llm-chat --assess --auto-improve

# Using specific model
llm-chat --model qwen2.5-coder:7b-instruct-q4_K_M
```

### Shell Mode (Piping)

```bash
# Explain code
cat main.go | llm-chat -s "explain this code"

# Generate commit message
git diff | llm-chat -s "write a concise commit message"

# Analyze data
cat data.csv | llm-chat -s "summarize this data"

# Debug errors
cat error.log | llm-chat -s "what's causing this error?"

# Code review
cat pr.diff | llm-chat -s "review this code for bugs and improvements"

# Generate documentation
cat *.go | llm-chat -s "create API documentation" -f markdown > docs.md

# Multiple providers
echo "Hello world" | llm-chat -p groq -s "translate to Spanish"
echo "Hola mundo" | llm-chat -p ollama -s "translate to French"
```

### Advanced Examples

```bash
# Output formats
cat code.py | llm-chat -s "find bugs" -f json
cat code.py | llm-chat -s "document this" -f markdown > docs.md

# Temperature control
llm-chat --temperature 0.3  # More focused/deterministic
llm-chat --temperature 0.9  # More creative/diverse

# Token limits
llm-chat --max-tokens 500  # Short responses
llm-chat --max-tokens 4000 # Long responses

# Disable history for sensitive data
llm-chat --no-history

# Chain commands
cat file.txt | llm-chat -s "summarize" | llm-chat -s "translate to French"

# Verbose with metrics
cat large-file.txt | llm-chat -s "analyze" --verbose
```

---

## 🎯 Interactive Commands

Once in interactive mode, use these commands:

### Basic Commands

- `/help` - Show all commands
- `/clear` - Clear the screen
- `/reset` - Start fresh conversation
- `/exit` or `/quit` - Exit

### Provider & Model Management

- `/providers` - List all available providers
- `/models` - List models for current provider
- `/switch` - Switch to different model

### History Management

- `/history` - Show current conversation
- `/saved` - Show recent saved conversations
- `/search` - Search through history
- `/export` - Export current conversation
- `/stats` - Show usage statistics

### Prompt Engineering

- `/assess` - Toggle prompt assessment
- `/guide` - Show prompt engineering guide
- `/improve <prompt>` - Get AI help improving a prompt

---

## ⚙️ Configuration

### Environment Variables

#### Ollama

```bash
export OLLAMA_URL=http://localhost:11434        # Default
export OLLAMA_MODEL=llama3:8b-instruct-q4_K_M   # Your model
```

#### Together AI

```bash
export TOGETHER_API_KEY=your-api-key
export TOGETHER_MODEL=llama-70b-free            # Options: llama-70b, llama-70b-free, deepseek, qwen-72b
```

#### Groq

```bash
export GROQ_API_KEY=your-api-key
export GROQ_MODEL=llama-70b                     # Options: llama-70b, llama-8b, mixtral, gemma-7b
```

#### SambaNova

```bash
export SAMBA_API_KEY=your-api-key
export SAMBA_MODEL=llama-70b                    # Options: llama-70b, llama-8b, qwen-72b
```

#### Google Gemini

```bash
export GEMINI_API_KEY=your-api-key
export GEMINI_MODEL=flash-lite                  # Options: flash, flash-lite, pro
```

### CLI Flags

```bash
-p, --provider string       LLM provider (default "ollama")
-v, --verbose              Show detailed metrics
-t, --temperature float    Temperature 0.0-2.0 (default 0.7)
-m, --max-tokens int       Maximum tokens (default 4000)
    --model string         Specific model to use
-s, --shell string         Shell mode with prompt
-f, --format string        Output format: text, json, markdown, raw
-a, --assess              Enable prompt assessment
    --auto-improve        Auto-offer prompt improvements
    --no-history          Don't save conversation
-h, --help                Show help
```

---

## 📊 Prompt Assessment

The built-in prompt assessment analyzes your prompts on 8 criteria:

1. **Clarity** - How clear and understandable
2. **Specificity** - How specific and detailed
3. **Context** - Background information provided
4. **Structure** - Organization and formatting
5. **Constraints** - Limitations specified
6. **Output Format** - Desired format specified
7. **Role/Persona** - Role definition
8. **Examples** - Examples provided

### Usage

```bash
# Enable assessment
llm-chat --assess

# With auto-improvement
llm-chat --assess --auto-improve

# In chat, toggle with
/assess

# Improve a specific prompt
/improve explain python decorators
```

---

## 💾 History Management

All conversations are automatically saved to `~/.llm-chat/history.json`

### Commands

```bash
/saved        # View recent conversations
/search       # Search through history
/export       # Export current chat
/stats        # View statistics
```

### Export Formats

```bash
/export
# Choose: markdown, json, or txt
```

### Disable History

```bash
llm-chat --no-history  # For sensitive conversations
```

---

## 🏗️ Project Structure

```
llm-chat/
├── cmd/
│   └── llm-chat/
│       └── main.go              # Entry point
├── internal/
│   ├── providers/               # LLM provider implementations
│   │   ├── provider.go         # Interface
│   │   ├── ollama.go
│   │   ├── together.go
│   │   ├── groq.go
│   │   ├── samba.go
│   │   └── gemini.go
│   ├── registry/               # Provider registry
│   │   └── registry.go
│   ├── chat/                   # Chat session logic
│   │   ├── session.go
│   │   └── shell.go
│   ├── assessment/             # Prompt assessment
│   │   ├── analyzer.go
│   │   └── improver.go
│   ├── history/                # History management
│   │   └── manager.go
│   ├── ui/                     # Terminal UI
│   │   ├── display.go
│   │   └── markdown.go
│   └── config/                 # Configuration
│       └── config.go
├── pkg/
│   └── models/                 # Shared data models
│       └── message.go
├── Makefile
├── go.mod
└── README.md
```

---

## 🤝 Contributing

Contributions are welcome! To add a new provider:

1. Create `internal/providers/yourprovider.go`
2. Implement the `Provider` interface
3. Register in `cmd/llm-chat/main.go`
4. Add environment variables to README

---

## 📝 Tips & Tricks

### Writing Better Prompts

Use `/guide` to see the full prompt engineering guide. Quick tips:

1. **Be specific** - "Explain Python decorators" → "Explain Python decorators with 3 code examples"
2. **Add context** - Include why you need it
3. **Set constraints** - "Keep under 200 words"
4. **Specify format** - "Format as a numbered list"
5. **Define role** - "As an expert Python developer..."
6. **Use examples** - Show what you want

### Workflow Examples

#### Code Review Workflow

```bash
git diff | llm-chat -s "review for bugs, security, performance" > review.md
```

#### Documentation Generation

```bash
cat src/*.go | llm-chat -s "generate API docs" -f markdown > API.md
```

#### Data Analysis

```bash
cat data.csv | llm-chat -s "analyze and visualize key trends" --verbose
```

#### Learning Assistant

```bash
llm-chat --assess --auto-improve  # Interactive learning with feedback
```

---

## 🐛 Troubleshooting

### Provider not available

```bash
# Check your setup
llm-chat

# You'll see which providers are configured
# Set the required environment variables
```

### Ollama connection issues

```bash
# Check Ollama is running
curl http://localhost:11434

# Start Ollama
ollama serve
```

### History file location

```bash
~/.llm-chat/history.json
```

---

## 📄 License

MIT License - See LICENSE file for details

---

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Uses [go-openai](https://github.com/sashabaranov/go-openai) for OpenAI-compatible APIs
- Integrates with [Ollama](https://ollama.ai) for local models
- Colors by [fatih/color](https://github.com/fatih/color)

---

## ⭐ Star History

If you find this useful, consider giving it a star on GitHub!

---

**Happy chatting!** 🤖✨
