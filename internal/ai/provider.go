package ai

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/k1lgor/container-diet/internal/analyzer"
	"github.com/k1lgor/container-diet/internal/config"
)

// AnalysisResponse represents the response from any AI provider
type AnalysisResponse struct {
	Advice string `json:"advice"`
	Fix    string `json:"fix,omitempty"`
}

// Provider defines the interface for AI providers
type Provider interface {
	// AnalyzeReport analyzes the image and Dockerfile, returns AI-generated advice
	AnalyzeReport(analysis *analyzer.ImageAnalysis, dockerfileContent string, model string, autoFix bool) (*AnalysisResponse, error)

	// AnalyzeAdvice provides general container optimization advice based on user context
	AnalyzeAdvice(contextStr string, dockerfileContent string, model string) (*AnalysisResponse, error)

	// Name returns the provider name
	Name() string

	// DefaultModel returns the default model for this provider
	DefaultModel() string

	// SetMaxTokens sets the max tokens for generation
	SetMaxTokens(maxTokens int)
}

// NewProvider creates a provider instance based on configuration
func NewProvider(providerConfig config.Provider, providerName string, maxTokens int) (Provider, error) {
	// Determine provider type: explicit type field takes priority, fall back to URL detection
	isAnthropic := providerConfig.Type == "anthropic" || isAnthropicURL(providerConfig.BaseURL)

	if isAnthropic {
		p := NewAnthropicProvider(providerConfig, providerName)
		if maxTokens > 0 {
			p.SetMaxTokens(maxTokens)
		}
		return p, nil
	}

	// All others use OpenAI-compatible API (OpenAI, OpenRouter, Ollama, etc.)
	p := NewOpenAICompatibleProvider(providerConfig, providerName)
	if maxTokens > 0 {
		p.SetMaxTokens(maxTokens)
	}
	return p, nil
}

// isAnthropicURL checks if the URL is Anthropic's native API (legacy detection)
func isAnthropicURL(baseURL string) bool {
	return baseURL == "https://api.anthropic.com" ||
		baseURL == "https://api.anthropic.com/v1" ||
		baseURL == "https://api.anthropic.com/"
}

// BuildPrompt creates the analysis prompt used by all providers
func BuildPrompt(analysis *analyzer.ImageAnalysis, dockerfileContent string, autoFix bool) string {
	var prompt string

	prompt += `You are the "Container Dietician", a ruthless but helpful AI expert dedicated to slimming down bloated Docker images.
Your goal is to roast the user's configuration slightly while providing critical optimization advice. Make it entertaining but useful.

TONE:
- Sassy, professional, and futuristic.
- Use emojis freely (e.g., 🐳, 🗑️, ⚡, 📉).
- Refer to large layers as "fat" or "bloat".

FORMATTING RULES for "advice":
- Start each warning with "⚠ WARNING: "
- Start each suggestion with "✓ SUGGESTION: "
- Keep the advice actionable and technical.

Focus on:
1. Large layers that can be optimized.
2. Unnecessary packages or build tools.
3. Security risks (e.g., running as root, secrets).
4. Multi-stage build opportunities.
5. Package manager caching (e.g., apt cache, pip cache, npm cache).
`

	if autoFix {
		prompt += `
OUTPUT FORMAT:
You MUST return a JSON object with the following keys:
- "advice": Your sassy roast and technical suggestions (string).
- "fix": The COMPLETE, optimized Dockerfile content based on your suggestions (string).
`
	} else {
		prompt += `
OUTPUT FORMAT:
Return a JSON object with the following key:
- "advice": Your sassy roast and technical suggestions (string).
`
	}

	if analysis != nil {
		prompt += `
IMAGE ANALYSIS:
Image Name: ` + analysis.ImageName + `
Total Size: ` + FormatBytes(analysis.TotalSize) + `
Layers:
`
		for i, l := range analysis.Layers {
			prompt += fmt.Sprintf("%d. Size: %s, Command: %s\n", i+1, FormatBytes(l.Size), l.Command)
		}
	}

	if dockerfileContent != "" {
		prompt += `
DOCKERFILE CONTENT:
` + dockerfileContent + `
`
	}

	return prompt
}

// BuildAdvicePrompt creates a general optimization advice prompt with user context.
// Used by the MCP get_optimization_advice tool for open-ended container questions.
func BuildAdvicePrompt(contextStr string, dockerfileContent string) string {
	prompt := `You are the "Container Dietician", a ruthless but helpful AI expert dedicated to slimming down bloated Docker images.
Your goal is to provide critical optimization advice for the user's specific container scenario. Make it entertaining but useful.

TONE:
- Sassy, professional, and futuristic.
- Use emojis freely (e.g., 🐳, 🗑️, ⚡, 📉).
- Refer to large layers as "fat" or "bloat".

FORMATTING RULES for "advice":
- Start each warning with "⚠ WARNING: "
- Start each suggestion with "✓ SUGGESTION: "
- Keep the advice actionable and technical.

Focus on:
1. Base image selection (alpine vs slim vs full).
2. Multi-stage build opportunities.
3. Package manager caching (e.g., apt cache, pip cache, npm cache).
4. Security best practices (non-root users, secrets).
5. Layer ordering and optimization.
6. Build context and .dockerignore usage.

OUTPUT FORMAT:
Return a JSON object with the following key:
- "advice": Your sassy roast and technical suggestions (string).

USER CONTEXT:
` + contextStr + "\n"

	if dockerfileContent != "" {
		prompt += "\nDOCKERFILE CONTENT:\n" + dockerfileContent + "\n"
	}

	return prompt
}

// FormatBytes converts bytes to human-readable format
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// isRetryable checks if an HTTP status code warrants a retry.
func isRetryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= 500
}

// doWithRetry executes an HTTP request with exponential backoff on retryable errors.
// Retries up to maxRetries times on 429 and 5xx responses.
func doWithRetry(ctx context.Context, client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	var lastErr error
	backoff := 1 * time.Second

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
		}

		// Each retry needs a fresh request body since it may have been consumed
		if attempt > 0 && req.Body != nil {
			// The body is a bytes.Buffer, which can be reset via req.GetBody
			if req.GetBody != nil {
				newBody, err := req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to rewind request body: %w", err)
				}
				req.Body = newBody
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if isRetryable(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("API request failed with status %d (attempt %d/%d)", resp.StatusCode, attempt+1, maxRetries+1)
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}
