package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/k1lgor/container-diet/internal/analyzer"
)

type AIClient interface {
	AnalyzeReport(analysis *analyzer.ImageAnalysis, dockerfileContent string, model string) (string, error)
}

type OpenAIClient struct {
	APIKey     string
	HTTPClient *http.Client
}

func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	return &OpenAIClient{
		APIKey: strings.TrimSpace(apiKey),
		HTTPClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}, nil
}

func (c *OpenAIClient) AnalyzeReport(analysis *analyzer.ImageAnalysis, dockerfileContent string, model string) (string, error) {
	if model == "" {
		model = "gpt-4o"
	}

	// Prepare the prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString(`
You are the "Container Dietician", a ruthless but helpful AI expert dedicated to slimming down bloated Docker images.
Your goal is to roast the user's configuration slightly while providing critical optimization advice. Make it entertaining but useful.

TONE:
- Sassy, professional, and futuristic.
- Use emojis freely (e.g., 🐳, 🗑️, ⚡, 📉).
- Refer to large layers as "fat" or "bloat".

FORMATTING RULES:
- Start each warning with "⚠ WARNING: "
- Start each suggestion with "✓ SUGGESTION: "
- Keep the advice actionable and technical.

Focus on:
1. Large layers that can be optimized.
2. Unnecessary packages or build tools.
3. Security risks (e.g., running as root, secrets).
4. Multi-stage build opportunities.
5. Package manager caching (e.g., apt cache, pip cache, npm cache).
`)

	if analysis != nil {
		promptBuilder.WriteString(fmt.Sprintf(`
IMAGE ANALYSIS:
Image Name: %s
Total Size: %d bytes
Layers:
`, analysis.ImageName, analysis.TotalSize))
		for i, l := range analysis.Layers {
			promptBuilder.WriteString(fmt.Sprintf("%d. Size: %d bytes, Command: %s\n", i+1, l.Size, l.Command))
		}
	}

	if dockerfileContent != "" {
		promptBuilder.WriteString(fmt.Sprintf(`
DOCKERFILE CONTENT:
%s
`, dockerfileContent))
	}
	prompt := promptBuilder.String()

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are the Container Dietician, a sassy, strict, and futuristic optimization expert. You roast users for bloated images but provide helpful advice."},
			{"role": "user", "content": prompt},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("invalid response format")
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: choice")
	}
	message, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid response format: message")
	}
	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format: content")
	}

	return content, nil
}
