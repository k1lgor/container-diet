package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/k1lgor/container-diet/internal/analyzer"
	"github.com/k1lgor/container-diet/internal/config"
)

// AnthropicProvider implements the Provider interface for Anthropic's native API
type AnthropicProvider struct {
	config     config.Provider
	name       string
	httpClient *http.Client
	timeout    int
	maxTokens  int
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(cfg config.Provider, name string) *AnthropicProvider {
	timeout := cfg.TimeoutSeconds
	if timeout == 0 {
		timeout = 90
	}

	return &AnthropicProvider{
		config:     cfg,
		name:       name,
		httpClient: &http.Client{Timeout: time.Duration(timeout) * time.Second},
		timeout:    timeout,
		maxTokens:  4096,
	}
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return p.name
}

// DefaultModel returns the default model for this provider
func (p *AnthropicProvider) DefaultModel() string {
	if p.config.DefaultModel != "" {
		return p.config.DefaultModel
	}
	return "claude-sonnet-4-20250514"
}

// SetMaxTokens sets the max tokens for generation
func (p *AnthropicProvider) SetMaxTokens(maxTokens int) {
	p.maxTokens = maxTokens
}

// AnthropicRequest represents the request structure for Anthropic API
type AnthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []AnthropicMessage `json:"messages"`
}

// AnthropicMessage represents a message in the Anthropic API
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents the response structure from Anthropic API
type AnthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model string `json:"model"`
}

// AnalyzeReport sends the analysis to Anthropic and returns suggestions
func (p *AnthropicProvider) AnalyzeReport(analysis *analyzer.ImageAnalysis, dockerfileContent string, model string, autoFix bool) (*AnalysisResponse, error) {
	if model == "" {
		model = p.DefaultModel()
	}
	return p.callAPI(model, BuildPrompt(analysis, dockerfileContent, autoFix))
}

// AnalyzeAdvice provides general container optimization advice based on user context
func (p *AnthropicProvider) AnalyzeAdvice(contextStr string, dockerfileContent string, model string) (*AnalysisResponse, error) {
	if model == "" {
		model = p.DefaultModel()
	}
	return p.callAPI(model, BuildAdvicePrompt(contextStr, dockerfileContent))
}

// callAPI is the shared HTTP call logic for the Anthropic native API.
func (p *AnthropicProvider) callAPI(model string, prompt string) (*AnalysisResponse, error) {
	payload := AnthropicRequest{
		Model:     model,
		MaxTokens: p.maxTokens,
		System:    "You are the Container Dietician. You respond ONLY with JSON.",
		Messages: []AnthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.timeout)*time.Second)
	defer cancel()

	baseURL := strings.TrimSuffix(p.config.BaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", p.config.APIKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")

	// Enable request body rewinding for retries
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(requestBody)), nil
	}

	resp, err := doWithRetry(ctx, p.httpClient, req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("invalid response format: no content")
	}

	content := result.Content[0].Text

	var analysisResp AnalysisResponse
	if err := json.Unmarshal([]byte(content), &analysisResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON content: %w", err)
	}

	return &analysisResp, nil
}
