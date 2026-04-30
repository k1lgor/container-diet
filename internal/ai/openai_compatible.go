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

// OpenAICompatibleProvider implements the Provider interface for OpenAI-compatible APIs
// This includes OpenAI, OpenRouter, Ollama, and any other OpenAI API-compatible service
type OpenAICompatibleProvider struct {
	config     config.Provider
	name       string
	httpClient *http.Client
	timeout    int
	maxTokens  int
}

// NewOpenAICompatibleProvider creates a new OpenAI-compatible provider
func NewOpenAICompatibleProvider(cfg config.Provider, name string) *OpenAICompatibleProvider {
	timeout := cfg.TimeoutSeconds
	if timeout == 0 {
		timeout = 90
	}

	return &OpenAICompatibleProvider{
		config:     cfg,
		name:       name,
		httpClient: &http.Client{Timeout: time.Duration(timeout) * time.Second},
		timeout:    timeout,
		maxTokens:  4096,
	}
}

// Name returns the provider name
func (p *OpenAICompatibleProvider) Name() string {
	return p.name
}

// DefaultModel returns the default model for this provider
func (p *OpenAICompatibleProvider) DefaultModel() string {
	if p.config.DefaultModel != "" {
		return p.config.DefaultModel
	}
	return "gpt-4o"
}

// SetMaxTokens sets the max tokens for generation
func (p *OpenAICompatibleProvider) SetMaxTokens(maxTokens int) {
	p.maxTokens = maxTokens
}

// AnalyzeReport sends the analysis to the AI and returns suggestions
func (p *OpenAICompatibleProvider) AnalyzeReport(analysis *analyzer.ImageAnalysis, dockerfileContent string, model string, autoFix bool) (*AnalysisResponse, error) {
	if model == "" {
		model = p.DefaultModel()
	}
	return p.callAPI(model, BuildPrompt(analysis, dockerfileContent, autoFix))
}

// AnalyzeAdvice provides general container optimization advice based on user context
func (p *OpenAICompatibleProvider) AnalyzeAdvice(contextStr string, dockerfileContent string, model string) (*AnalysisResponse, error) {
	if model == "" {
		model = p.DefaultModel()
	}
	return p.callAPI(model, BuildAdvicePrompt(contextStr, dockerfileContent))
}

// callAPI is the shared HTTP call logic for OpenAI-compatible APIs.
func (p *OpenAICompatibleProvider) callAPI(model string, prompt string) (*AnalysisResponse, error) {
	payload := map[string]interface{}{
		"model":      model,
		"max_tokens": p.maxTokens,
		"messages": []map[string]string{
			{"role": "system", "content": "You are the Container Dietician. You respond ONLY with JSON."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.timeout)*time.Second)
	defer cancel()

	baseURL := strings.TrimSuffix(p.config.BaseURL, "/")
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	// Add OpenRouter-specific headers if needed
	if strings.Contains(baseURL, "openrouter") {
		req.Header.Set("HTTP-Referer", "https://github.com/k1lgor/container-diet")
		req.Header.Set("X-Title", "Container Diet")
	}

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

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseOpenAIResponse(result)
}

// parseOpenAIResponse extracts the content from an OpenAI-compatible response
func parseOpenAIResponse(result map[string]interface{}) (*AnalysisResponse, error) {
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("invalid response format: no choices")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format: choice")
	}

	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		// Try "delta" for streaming responses
		message, ok = firstChoice["delta"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response format: message")
		}
	}

	content, ok := message["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format: content")
	}

	var analysisResp AnalysisResponse
	if err := json.Unmarshal([]byte(content), &analysisResp); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON content: %w", err)
	}

	return &analysisResp, nil
}
