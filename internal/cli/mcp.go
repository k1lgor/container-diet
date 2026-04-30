package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/k1lgor/container-diet/internal/config"
	mcpServer "github.com/k1lgor/container-diet/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpConfigPath string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP (Model Context Protocol) commands",
	Long:  `Commands for running container-diet as an MCP server for AI agents like Claude, Cursor, or Codex.`,
}

var mcpServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run container-diet as an MCP server",
	Long: `Run container-diet as an MCP (Model Context Protocol) server over stdin/stdout.

This starts an MCP server that AI agents (Claude Desktop, Cursor, Codex, etc.)
can launch as a child process and communicate with via standard I/O.

How it works:
  The MCP client spawns "container-diet mcp server" as a subprocess.
  The server reads JSON-RPC requests from stdin and writes responses to stdout.
  No network ports, no daemon, no registration — just a child process.

The server reads your API keys from ~/.config/container-diet/config.yaml
(the same config file used by the CLI). No need to set environment variables
in the MCP client config — just configure your provider once.

The server exposes these tools:
  - analyze_dockerfile: AI analysis of Dockerfiles with optional auto-fix
  - analyze_image: Layer breakdown + AI optimization advice for images
  - get_optimization_advice: General container advice from any context
  - get_image_summary: Quick image metrics without AI (free + fast)`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runMCPServer(); err != nil {
			os.Exit(1)
		}
	},
}

var mcpInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Show MCP client setup instructions",
	Long: `Display instructions for configuring various MCP clients to use container-diet.

The MCP server reads API keys from your config file — no need to put secrets
in the MCP client config. Just set up your provider once with init-config.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(aurora("🚀 MCP Server Setup"))
		fmt.Println()
		fmt.Println(neon("How it works:"))
		fmt.Println("  The MCP client (Claude, Cursor, etc.) launches container-diet as a")
		fmt.Println("  child process. It communicates over stdin/stdout using JSON-RPC.")
		fmt.Println("  No ports, no daemons, no registration — just a subprocess.")
		fmt.Println()
		fmt.Println(neon("Quick setup (3 steps):"))
		fmt.Println("  1. Install:   go install github.com/k1lgor/container-diet/cmd/cli@latest")
		fmt.Println("  2. Configure: container-diet init-config")
		fmt.Println("                → edit ~/.config/container-diet/config.yaml")
		fmt.Println("                → set default_provider + uncomment your provider + set api_key")
		fmt.Println("  3. Register:  add the JSON below to your MCP client config")
		fmt.Println()

		// ─── The one config that works everywhere ───
		fmt.Println(strings.Repeat("─", 64))
		fmt.Println(neon("MCP Client Config (Claude Desktop, Cursor, Claude Code, etc.)"))
		fmt.Println()
		fmt.Println("  Your API key lives in the config file, so the MCP client config")
		fmt.Println("  is minimal. Add this to your client's mcpServers section:")
		fmt.Println()
		fmt.Println(success("  {"))
		fmt.Println(success(`    "mcpServers": {`))
		fmt.Println(success(`      "container-diet": {`))
		fmt.Println(success(`        "command": "container-diet",`))
		fmt.Println(success(`        "args": ["mcp", "server"]`))
		fmt.Println(success(`      }`))
		fmt.Println(success(`    }`))
		fmt.Println(success("  }"))
		fmt.Println()
		fmt.Println("  That's it. The server reads ~/.config/container-diet/config.yaml")
		fmt.Println("  for the provider, model, and API key.")
		fmt.Println()

		// ─── Config file locations per client ───
		fmt.Println(strings.Repeat("─", 64))
		fmt.Println(neon("Where to put the JSON (per client)"))
		fmt.Println()
		fmt.Println("  Claude Desktop:")
		fmt.Println("    macOS:   ~/Library/Application Support/Claude/claude_desktop_config.json")
		fmt.Println("    Windows: %APPDATA%\\Claude\\claude_desktop_config.json")
		fmt.Println("    Linux:   ~/.config/Claude/claude_desktop_config.json")
		fmt.Println()
		fmt.Println("  Cursor:")
		fmt.Println("    Project: .cursor/mcp.json  (in your project root)")
		fmt.Println()
		fmt.Println("  Claude Code (CLI):")
		fmt.Println("    Global:  ~/.claude/claude_desktop_config.json")
		fmt.Println("    Project: .claude/mcp.json")
		fmt.Println()

		// ─── API Key Options ───
		fmt.Println(strings.Repeat("─", 64))
		fmt.Println(neon("API Key — pick one method"))
		fmt.Println()
		fmt.Println(success("  A) In config file (recommended) — set once, works everywhere:"))
		fmt.Println("       container-diet init-config")
		fmt.Println("       # Edit ~/.config/container-diet/config.yaml:")
		fmt.Println("       #")
		fmt.Println("       #   default_provider: \"openrouter\"")
		fmt.Println("       #   providers:")
		fmt.Println("       #     openrouter:")
		fmt.Println("       #       api_key: \"sk-or-v1-...\"")
		fmt.Println("       #       base_url: \"https://openrouter.ai/api/v1\"")
		fmt.Println("       #       default_model: \"openai/gpt-4o-mini\"")
		fmt.Println()
		fmt.Println(warn("  B) Env var reference in config (key stays in shell, not in file):"))
		fmt.Println("       # In config.yaml:")
		fmt.Println("       #   api_key: \"${OPENROUTER_API_KEY}\"")
		fmt.Println("       # In your shell profile (~/.zshrc, ~/.bashrc):")
		fmt.Println("       #   export OPENROUTER_API_KEY=sk-or-v1-...")
		fmt.Println()
		fmt.Println("  C) Env var in MCP client config (Claude Desktop only):")
		fmt.Println(`       Add "env": { "OPENROUTER_API_KEY": "sk-or-..." } to the JSON above`)
		fmt.Println()
		fmt.Println(neon("Supported providers:"))
		fmt.Println("  openai, anthropic, openrouter, groq, deepseek, mistral, ollama,")
		fmt.Println("  xai, perplexity, moonshot, huggingface, or any OpenAI-compatible API")
		fmt.Println()
		fmt.Println(warn("Make sure container-diet is in your PATH after installation."))
		fmt.Println()
	},
}

func init() {
	mcpServerCmd.Flags().StringVar(&mcpConfigPath, "config", "", "Path to configuration file")

	mcpCmd.AddCommand(mcpServerCmd)
	mcpCmd.AddCommand(mcpInitCmd)
	rootCmd.AddCommand(mcpCmd)
}

// runMCPServer is the testable core of the MCP server command.
func runMCPServer() error {
	cfg, err := config.LoadConfig(mcpConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Error loading config: %v\n", fail("✖"), err)
		return err
	}

	server, err := mcpServer.NewServer(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s Error creating MCP server: %v\n", fail("✖"), err)
		return err
	}

	if err := server.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "%s Server error: %v\n", fail("✖"), err)
		return err
	}
	return nil
}
