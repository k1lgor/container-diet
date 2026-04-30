<div align="center">
  <img src="assets/logo.png" alt="Container Diet Logo" width="400" />
</div>

# 🐳 Container Diet

**AI-powered Docker image optimization — because your containers could use a diet.**

Container Diet analyzes your Docker images and Dockerfiles, then serves up sassy, actionable advice backed by your choice of 12+ AI providers. Catch bloated layers, security holes, and missing best practices before they reach production. Works as a CLI, in CI/CD, or as an MCP tool inside your AI editor.

---

<p align="center">
  <a href="https://github.com/k1lgor/container-diet/releases"><img src="https://img.shields.io/github/v/release/k1lgor/container-diet?color=2496ED&label=latest" alt="Release"></a>
  <a href="https://github.com/k1lgor/container-diet/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-2496ED" alt="License"></a>
  <a href="https://pkg.go.dev/github.com/k1lgor/container-diet"><img src="https://img.shields.io/badge/go-reference-2496ED" alt="Go Reference"></a>
  <a href="https://www.producthunt.com/products/container-diet"><img src="https://img.shields.io/badge/Product_Hunt-featured-da552f" alt="Product Hunt"></a>
</p>

---

## ⚡ Quick Start

```bash
# 1. Install
go install github.com/k1lgor/container-diet/cmd/cli@latest

# 2. Configure your AI provider
container-diet init-config
# → edit ~/.config/container-diet/config.yaml
# → uncomment your provider, paste your API key

# 3. Analyze
container-diet analyze --dockerfile Dockerfile --auto-fix
```

That's it. You'll get a roast of your Dockerfile, actionable fixes, and a `Dockerfile.diet` with the optimizations applied.

---

## 🧠 AI Providers

Container Diet talks to **any** of these — bring the key you already have, or run locally for free.

| Provider             | Config Key    | API Key Env Var      |
| -------------------- | ------------- | -------------------- |
| OpenAI               | `openai`      | `OPENAI_API_KEY`     |
| Anthropic (native)   | `anthropic`   | `ANTHROPIC_API_KEY`  |
| OpenRouter           | `openrouter`  | `OPENROUTER_API_KEY` |
| Groq                 | `groq`        | `GROQ_API_KEY`       |
| DeepSeek             | `deepseek`    | `DEEPSEEK_API_KEY`   |
| Mistral              | `mistral`     | `MISTRAL_API_KEY`    |
| xAI (Grok)           | `xai`         | `XAI_API_KEY`        |
| Ollama (local, free) | `ollama`      | _none needed_        |
| Perplexity           | `perplexity`  | `PERPLEXITY_API_KEY` |
| Moonshot             | `moonshot`    | `MOONSHOT_API_KEY`   |
| Hugging Face         | `huggingface` | `HF_API_TOKEN`       |
| Custom endpoint      | `custom`      | `CUSTOM_API_KEY`     |

Switch providers with `--provider` or set `default_provider` in your config.

---

## 🔌 MCP Server — AI Agent Integration

Container Diet ships as an **MCP (Model Context Protocol) server** so AI agents can analyze containers directly from your editor.

**Supported clients:** Claude Desktop · Cursor · Claude Code · Codex · any MCP-compatible agent

**Add to your MCP client config:**

```json
{
  "mcpServers": {
    "container-diet": {
      "command": "container-diet",
      "args": ["mcp", "server"]
    }
  }
}
```

No API keys in the MCP config — the server reads them from `~/.config/container-diet/config.yaml`.

**Tools exposed to AI agents:**

| Tool                      | Description                                     |
| ------------------------- | ----------------------------------------------- |
| `analyze_dockerfile`      | AI analysis with optional auto-fix generation   |
| `analyze_image`           | Layer breakdown + AI optimization advice        |
| `get_optimization_advice` | General container advice from free-form context |
| `get_image_summary`       | Quick metrics without burning AI tokens         |

Full setup guide: `container-diet mcp init`

---

## 📖 Usage

### Analyze a Dockerfile

```bash
container-diet analyze --dockerfile Dockerfile
```

### Analyze + Auto-Fix

```bash
container-diet analyze --dockerfile Dockerfile --auto-fix
# Writes Dockerfile.diet with optimizations applied
```

### Analyze a Docker Image

```bash
# Local daemon
container-diet analyze nginx:latest

# Pull from registry
container-diet analyze python:3.12-slim --remote

# Auto-pull if missing locally
container-diet analyze busybox --pull-missing
```

### JSON Output (CI/CD)

```bash
container-diet analyze --dockerfile Dockerfile --format json
```

```json
{
  "advice": "⚠ WARNING: Root user detected...\n✓ SUGGESTION: Use non-root user...",
  "fix": "FROM nginx:alpine\nUSER nginx\n..."
}
```

### Choose Provider + Model

```bash
container-diet analyze --dockerfile Dockerfile --provider anthropic --model claude-sonnet-4-6
```

---

## 🎮 Demo

Running the "Nightmare Monolith" Dockerfile through Container Diet:

```text
📄 Reading Dockerfile: samples/Dockerfile.nightmare...

🤖 [AI ANALYSIS]
🚢 Asking the Container Dietician for insights using openrouter (openai/gpt-4o-mini)

================================================================

Oh, honey, what do we have here? A "Monolith Monster" Dockerfile that's
about to sink your ship with its fatty layers and spicy security risks! 🚀

---

⚠ WARNING: Version Drift Alert!
Using `ubuntu:latest` means you're playing roulette with your build
environment. 🎰

✓ SUGGESTION: Pin your base image to a specific version like
`ubuntu:22.04` to keep things predictable.

---

⚠ WARNING: Apt-get Avalanche!
Installing everything but the kitchen sink? This is the definition of
bloat, my dear. 🐘

✓ SUGGESTION: Ditch `openjdk-11-jdk`, `build-essential`, `cmake`, and
`gdb` unless you truly need them.

---

⚠ WARNING: 777 permissions? Root SSH? Root login permitted?
You might as well hand over the keys to the universe. 🔑

✓ SUGGESTION: Use non-root user, restrictive permissions. And ask
yourself — do you really need SSH in a container?

---

Oh, darling, trim that bloated ship before it swallows a sun! 🐳✨
```

---

## ⚙️ Configuration

Config file locations (searched in order, later overrides earlier):

| Priority    | Path                                   | Scope       |
| ----------- | -------------------------------------- | ----------- |
| 1 (highest) | `--config <path>` flag                 | Per-command |
| 2           | `.container-diet/config.yaml`          | Per-project |
| 3 (lowest)  | `~/.config/container-diet/config.yaml` | Global      |

**Minimal config example:**

```yaml
default_provider: "openrouter"

providers:
  openrouter:
    api_key: "${OPENROUTER_API_KEY}" # or paste the key directly
    base_url: "https://openrouter.ai/api/v1"
    default_model: "openai/gpt-4o-mini"
    timeout_seconds: 90

analysis:
  max_tokens: 4096

ui:
  theme: "docker"
  show_emojis: true
```

Generate the full example with all providers: `container-diet init-config`

Pro-tip: use `--local` for project-specific config, `--force` to overwrite.

---

## 🏗️ Architecture

```
cmd/cli/main.go               # Entry point
internal/
├── ai/                       # Provider interface + implementations
│   ├── provider.go           # Interface, prompt builders, retry logic
│   ├── anthropic.go          # Anthropic native API
│   └── openai_compatible.go  # OpenAI, OpenRouter, Ollama, Groq, etc.
├── analyzer/                 # Docker image layer inspection
├── cli/                      # Cobra commands (analyze, init-config, mcp)
├── config/                   # YAML config loading with env var expansion
├── mcp/                      # MCP server (stdio transport, 4 tools)
├── samples/                  # Test Dockerfiles (light → nightmare)
└── web/                      # Landing page (React + Three.js)
```

---

## 📄 License

[MIT](LICENSE) © 2026 Container Diet

---

<p align="center">
  <sub>Built with 🐳 and a touch of sass.</sub>
</p>
