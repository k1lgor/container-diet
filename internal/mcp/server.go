package mcp

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/k1lgor/container-diet/internal/ai"
	"github.com/k1lgor/container-diet/internal/analyzer"
	"github.com/k1lgor/container-diet/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const mcpVersion = "0.4.0"

// Server represents the MCP server instance
type Server struct {
	config   *config.Config
	provider ai.Provider
}

// knownBaseURLs maps common provider names to their default API base URLs.
var knownBaseURLs = map[string]string{
	"openai":      "https://api.openai.com/v1",
	"anthropic":   "https://api.anthropic.com/v1",
	"openrouter":  "https://openrouter.ai/api/v1",
	"groq":        "https://api.groq.com/openai/v1",
	"deepseek":    "https://api.deepseek.com/v1",
	"mistral":     "https://api.mistral.ai/v1",
	"xai":         "https://api.x.ai/v1",
	"perplexity":  "https://api.perplexity.ai",
	"moonshot":    "https://api.moonshot.cn/v1",
	"huggingface": "https://api-inference.huggingface.co/models",
	"ollama":      "http://localhost:11434/v1",
}

// NewServer creates a new MCP server instance
func NewServer(cfg *config.Config) (*Server, error) {
	providerCfg, err := cfg.GetProvider(cfg.DefaultProvider)
	if err != nil {
		// Try env var fallback: check for {PROVIDER}_API_KEY
		envKey := strings.ToUpper(cfg.DefaultProvider) + "_API_KEY"
		if apiKey := os.Getenv(envKey); apiKey != "" {
			baseURL := knownBaseURLs[cfg.DefaultProvider]
			if baseURL == "" && cfg.DefaultProvider != "ollama" {
				return nil, fmt.Errorf("unknown provider '%s' — please configure it in your config file (container-diet init-config) with a base_url", cfg.DefaultProvider)
			}
			providerType := ""
			if cfg.DefaultProvider == "anthropic" {
				providerType = "anthropic"
			}
			providerCfg = config.Provider{
				Type:           providerType,
				APIKey:         apiKey,
				BaseURL:        baseURL,
				TimeoutSeconds: 90,
			}
		} else {
			return nil, fmt.Errorf("failed to get provider config: %w", err)
		}
	}

	if providerCfg.APIKey == "" {
		envKey := strings.ToUpper(cfg.DefaultProvider) + "_API_KEY"
		return nil, fmt.Errorf("no API key found for provider '%s'\n\nTo set up API keys for MCP:\n1. Create config: container-diet init-config\n2. Set env var: export %s=your-key-here\n3. For Claude Desktop, add env var to mcpServers config\n\nSee: container-diet mcp init", cfg.DefaultProvider, envKey)
	}

	provider, err := ai.NewProvider(providerCfg, cfg.DefaultProvider, cfg.Analysis.MaxTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	return &Server{
		config:   cfg,
		provider: provider,
	}, nil
}

// Start starts the MCP server
func (s *Server) Start() error {
	mcpServer := server.NewMCPServer(
		"container-diet",
		mcpVersion,
		server.WithToolCapabilities(false),
	)

	s.registerTools(mcpServer)

	// Handle signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ServeStdio(mcpServer)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (s *Server) registerTools(mcpServer *server.MCPServer) {
	// Tool 1: analyze_dockerfile — AI analysis of a Dockerfile
	tool := mcp.NewTool("analyze_dockerfile",
		mcp.WithDescription("Analyze a Dockerfile and get AI-powered optimization suggestions. Returns sassy, actionable advice with warnings and suggestions. Use auto_fix=true to receive an optimized Dockerfile in the response."),
		mcp.WithString("dockerfile_content",
			mcp.Required(),
			mcp.Description("The complete Dockerfile content to analyze"),
		),
		mcp.WithBoolean("auto_fix",
			mcp.Description("Whether to generate an optimized version of the Dockerfile (default: false)"),
		),
		mcp.WithString("provider",
			mcp.Description("AI provider to use (defaults to configured provider)"),
		),
		mcp.WithString("model",
			mcp.Description("AI model to use (defaults to provider's default)"),
		),
	)
	mcpServer.AddTool(tool, s.handleAnalyzeDockerfile)

	// Tool 2: analyze_image — AI analysis of a Docker image
	tool = mcp.NewTool("analyze_image",
		mcp.WithDescription("Analyze a Docker image's layers, size, and configuration, then get AI-powered optimization advice. Pulls from local daemon by default; use remote=true for registries. Returns layer breakdown with sizes and AI suggestions."),
		mcp.WithString("image_name",
			mcp.Required(),
			mcp.Description("The Docker image name to analyze (e.g., 'nginx:latest', 'python:3.11-slim')"),
		),
		mcp.WithBoolean("remote",
			mcp.Description("Pull image from remote registry if not found locally (default: false)"),
		),
		mcp.WithBoolean("pull_missing",
			mcp.Description("Auto-pull missing images from remote (default: false)"),
		),
		mcp.WithBoolean("auto_fix",
			mcp.Description("Generate an optimized Dockerfile based on the image analysis (default: false)"),
		),
		mcp.WithString("provider",
			mcp.Description("AI provider to use (defaults to configured provider)"),
		),
		mcp.WithString("model",
			mcp.Description("AI model to use (defaults to provider's default)"),
		),
	)
	mcpServer.AddTool(tool, s.handleAnalyzeImage)

	// Tool 3: get_optimization_advice — General container advice (context-aware)
	tool = mcp.NewTool("get_optimization_advice",
		mcp.WithDescription("Get AI-powered container optimization advice for any scenario. Describe your tech stack, constraints, or goals and receive tailored suggestions for base images, multi-stage builds, caching strategies, and security hardening."),
		mcp.WithString("context",
			mcp.Required(),
			mcp.Description("Describe what you want to optimize (e.g., 'Python Flask app with large dependencies, need to reduce image from 1.2GB', 'Go microservice that needs fast builds')"),
		),
		mcp.WithString("dockerfile_content",
			mcp.Description("Optional: your current Dockerfile for context-aware advice"),
		),
		mcp.WithString("provider",
			mcp.Description("AI provider to use (defaults to configured provider)"),
		),
		mcp.WithString("model",
			mcp.Description("AI model to use (defaults to provider's default)"),
		),
	)
	mcpServer.AddTool(tool, s.handleGetOptimizationAdvice)

	// Tool 4: get_image_summary — Quick metrics without AI
	tool = mcp.NewTool("get_image_summary",
		mcp.WithDescription("Get a quick summary of a Docker image's layers and sizes without AI analysis. Fast and free — returns total size, layer count, and per-layer size breakdown."),
		mcp.WithString("image_name",
			mcp.Required(),
			mcp.Description("The Docker image name (e.g., 'nginx:latest')"),
		),
		mcp.WithBoolean("remote",
			mcp.Description("Pull image from remote registry if not found locally (default: false)"),
		),
		mcp.WithBoolean("pull_missing",
			mcp.Description("Auto-pull missing images from remote (default: false)"),
		),
	)
	mcpServer.AddTool(tool, s.handleGetImageSummary)
}

func (s *Server) getProvider(args map[string]any) (ai.Provider, string, error) {
	provider := s.provider
	model := ""

	if providerName, ok := args["provider"].(string); ok && providerName != "" {
		if providerName != s.config.DefaultProvider {
			providerCfg, err := s.config.GetProvider(providerName)
			if err != nil {
				return nil, "", fmt.Errorf("provider '%s' not found: %w", providerName, err)
			}
			newProvider, err := ai.NewProvider(providerCfg, providerName, s.config.Analysis.MaxTokens)
			if err != nil {
				return nil, "", fmt.Errorf("failed to create provider: %w", err)
			}
			provider = newProvider
		}
	}

	if modelArg, ok := args["model"].(string); ok && modelArg != "" {
		model = modelArg
	} else {
		model = provider.DefaultModel()
	}

	return provider, model, nil
}

func (s *Server) handleAnalyzeDockerfile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	dockerfileContent, ok := args["dockerfile_content"].(string)
	if !ok || dockerfileContent == "" {
		return mcp.NewToolResultError("dockerfile_content is required"), nil
	}

	autoFix := false
	if val, ok := args["auto_fix"].(bool); ok {
		autoFix = val
	}

	provider, model, err := s.getProvider(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := provider.AnalyzeReport(nil, dockerfileContent, model, autoFix)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("AI analysis failed: %v", err)), nil
	}

	response := result.Advice

	if autoFix && result.Fix != "" {
		response += fmt.Sprintf("\n\n--- Optimized Dockerfile ---\n\n```dockerfile\n%s\n```", result.Fix)
	}

	return mcp.NewToolResultText(response), nil
}

func (s *Server) handleAnalyzeImage(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	imageName, ok := args["image_name"].(string)
	if !ok || imageName == "" {
		return mcp.NewToolResultError("image_name is required"), nil
	}

	remote := false
	if val, ok := args["remote"].(bool); ok {
		remote = val
	}

	pullMissing := false
	if val, ok := args["pull_missing"].(bool); ok {
		pullMissing = val
	}

	autoFix := false
	if val, ok := args["auto_fix"].(bool); ok {
		autoFix = val
	}

	analysis, err := analyzer.AnalyzeImage(imageName, remote, pullMissing)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze image: %v", err)), nil
	}

	provider, model, err := s.getProvider(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := provider.AnalyzeReport(analysis, "", model, autoFix)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("AI analysis failed: %v", err)), nil
	}

	summary := formatImageSummary(analysis)
	response := summary + "\n\n--- AI Analysis ---\n\n" + result.Advice

	if autoFix && result.Fix != "" {
		response += fmt.Sprintf("\n\n--- Optimized Dockerfile ---\n\n```dockerfile\n%s\n```", result.Fix)
	}

	return mcp.NewToolResultText(response), nil
}

func (s *Server) handleGetOptimizationAdvice(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	contextStr, ok := args["context"].(string)
	if !ok || contextStr == "" {
		return mcp.NewToolResultError("context is required"), nil
	}

	dockerfileContent := ""
	if val, ok := args["dockerfile_content"].(string); ok {
		dockerfileContent = val
	}

	provider, model, err := s.getProvider(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := provider.AnalyzeAdvice(contextStr, dockerfileContent, model)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("AI analysis failed: %v", err)), nil
	}

	return mcp.NewToolResultText(result.Advice), nil
}

func (s *Server) handleGetImageSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	imageName, ok := args["image_name"].(string)
	if !ok || imageName == "" {
		return mcp.NewToolResultError("image_name is required"), nil
	}

	remote := false
	if val, ok := args["remote"].(bool); ok {
		remote = val
	}

	pullMissing := false
	if val, ok := args["pull_missing"].(bool); ok {
		pullMissing = val
	}

	analysis, err := analyzer.AnalyzeImage(imageName, remote, pullMissing)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze image: %v", err)), nil
	}

	summary := formatImageSummary(analysis)

	return mcp.NewToolResultText(summary), nil
}

func formatImageSummary(analysis *analyzer.ImageAnalysis) string {
	summary := fmt.Sprintf("📦 Image: %s\n", analysis.ImageName)
	summary += fmt.Sprintf("📊 Total Size: %s\n", ai.FormatBytes(analysis.TotalSize))
	summary += fmt.Sprintf("🍰 Layers: %d\n\n", len(analysis.Layers))
	summary += "--- Layer Breakdown ---\n"

	for i, layer := range analysis.Layers {
		cmd := layer.Command
		if len(cmd) > 60 {
			cmd = cmd[:57] + "..."
		}
		summary += fmt.Sprintf("%d. %s - %s\n", i+1, ai.FormatBytes(layer.Size), cmd)
	}

	return summary
}
