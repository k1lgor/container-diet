package ai

import (
	"strings"
	"testing"

	"github.com/k1lgor/container-diet/internal/analyzer"
	"github.com/k1lgor/container-diet/internal/config"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		contains string
	}{
		{"zero", 0, "bytes"},
		{"bytes", 512, "bytes"},
		{"kilobytes", 2048, "KB"},
		{"megabytes", 5 * 1024 * 1024, "MB"},
		{"gigabytes", 2 * 1024 * 1024 * 1024, "GB"},
		{"exact 1KB", 1024, "KB"},
		{"exact 1MB", 1024 * 1024, "MB"},
		{"exact 1GB", 1024 * 1024 * 1024, "GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("FormatBytes(%d) = %q, expected to contain %q", tt.bytes, result, tt.contains)
			}
		})
	}
}

func TestIsAnthropicProvider(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://api.anthropic.com", true},
		{"https://api.anthropic.com/v1", true},
		{"https://api.anthropic.com/", true},
		{"https://api.openai.com/v1", false},
		{"https://openrouter.ai/api/v1", false},
		{"http://localhost:11434/v1", false},
		{"", false},
		{"https://api.anthropic.com/v1/messages", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := isAnthropicURL(tt.url)
			if got != tt.expected {
				t.Errorf("isAnthropicURL(%q) = %v, want %v", tt.url, got, tt.expected)
			}
		})
	}

	// Verify explicit type field takes priority over URL
	t.Run("explicit type field", func(t *testing.T) {
		cfg := config.Provider{
			Type:    "anthropic",
			APIKey:  "key",
			BaseURL: "https://api.openai.com/v1", // non-Anthropic URL
		}
		p, err := NewProvider(cfg, "test", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := p.(*AnthropicProvider); !ok {
			t.Error("expected AnthropicProvider when type='anthropic' even with OpenAI URL")
		}
	})
}

func TestNewProvider(t *testing.T) {
	t.Run("creates anthropic provider", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "sk-ant-test",
			BaseURL: "https://api.anthropic.com/v1",
		}
		p, err := NewProvider(cfg, "anthropic", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name() != "anthropic" {
			t.Errorf("expected name 'anthropic', got %q", p.Name())
		}
	})

	t.Run("creates openai-compatible provider", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "sk-test",
			BaseURL: "https://api.openai.com/v1",
		}
		p, err := NewProvider(cfg, "openai", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name() != "openai" {
			t.Errorf("expected name 'openai', got %q", p.Name())
		}
	})

	t.Run("creates openrouter as openai-compatible", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "sk-or-test",
			BaseURL: "https://openrouter.ai/api/v1",
		}
		p, err := NewProvider(cfg, "openrouter", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name() != "openrouter" {
			t.Errorf("expected name 'openrouter', got %q", p.Name())
		}
	})

	t.Run("creates ollama as openai-compatible", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "ollama",
			BaseURL: "http://localhost:11434/v1",
		}
		p, err := NewProvider(cfg, "ollama", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name() != "ollama" {
			t.Errorf("expected name 'ollama', got %q", p.Name())
		}
	})

	t.Run("passes custom maxTokens", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "sk-test",
			BaseURL: "https://api.openai.com/v1",
		}
		p, err := NewProvider(cfg, "openai", 8192)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Verify by casting and checking the field
		if oai, ok := p.(*OpenAICompatibleProvider); ok {
			if oai.maxTokens != 8192 {
				t.Errorf("expected maxTokens 8192, got %d", oai.maxTokens)
			}
		}
	})
}

func TestBuildPrompt(t *testing.T) {
	t.Run("with image analysis, no autofix", func(t *testing.T) {
		analysis := &analyzer.ImageAnalysis{
			ImageName: "nginx:latest",
			TotalSize: 142000000,
			Layers: []analyzer.LayerAnalysis{
				{Size: 71000000, Command: "/bin/sh -c apt-get update"},
				{Size: 71000000, Command: "/bin/sh -c apt-get install -y curl"},
			},
		}

		prompt := BuildPrompt(analysis, "", false)

		for _, want := range []string{
			"Container Dietician",
			"nginx:latest",
			"IMAGE ANALYSIS",
			"apt-get update",
			"apt-get install",
			"advice",
		} {
			if !strings.Contains(prompt, want) {
				t.Errorf("prompt missing %q", want)
			}
		}
		// Should not contain fix output format when autofix=false
		if strings.Contains(prompt, `"fix"`) {
			t.Error("prompt should not contain fix format when autofix=false")
		}
	})

	t.Run("with dockerfile content, with autofix", func(t *testing.T) {
		dockerfile := `FROM ubuntu:22.04
RUN apt-get update && apt-get install -y curl
`
		prompt := BuildPrompt(nil, dockerfile, true)

		for _, want := range []string{
			"FROM ubuntu:22.04",
			"apt-get update",
			`"fix"`,
			"COMPLETE",
		} {
			if !strings.Contains(prompt, want) {
				t.Errorf("prompt missing %q", want)
			}
		}
	})

	t.Run("minimal prompt with no analysis or dockerfile", func(t *testing.T) {
		prompt := BuildPrompt(nil, "", false)

		if !strings.Contains(prompt, "Container Dietician") {
			t.Error("minimal prompt should still contain Container Dietician")
		}
		if strings.Contains(prompt, "IMAGE ANALYSIS") {
			t.Error("minimal prompt should not contain IMAGE ANALYSIS")
		}
		if strings.Contains(prompt, "DOCKERFILE CONTENT") {
			t.Error("minimal prompt should not contain DOCKERFILE CONTENT")
		}
	})

	t.Run("with both analysis and dockerfile", func(t *testing.T) {
		analysis := &analyzer.ImageAnalysis{
			ImageName: "myapp:latest",
			TotalSize: 500 * 1024 * 1024,
			Layers: []analyzer.LayerAnalysis{
				{Size: 500 * 1024 * 1024, Command: "COPY . /app"},
			},
		}
		dockerfile := "FROM node:18\nCOPY . /app\n"

		prompt := BuildPrompt(analysis, dockerfile, false)

		if !strings.Contains(prompt, "myapp:latest") {
			t.Error("prompt missing image name")
		}
		if !strings.Contains(prompt, "FROM node:18") {
			t.Error("prompt missing dockerfile content")
		}
	})
}

func TestSetMaxTokens(t *testing.T) {
	cfg := config.Provider{
		APIKey:  "sk-ant-test",
		BaseURL: "https://api.anthropic.com/v1",
	}
	p := NewAnthropicProvider(cfg, "anthropic")

	// Default is 4096
	if p.maxTokens != 4096 {
		t.Errorf("expected default maxTokens 4096, got %d", p.maxTokens)
	}

	p.SetMaxTokens(2048)
	if p.maxTokens != 2048 {
		t.Errorf("expected maxTokens 2048, got %d", p.maxTokens)
	}
}

func TestSetMaxTokensOpenAI(t *testing.T) {
	cfg := config.Provider{
		APIKey:  "sk-test",
		BaseURL: "https://api.openai.com/v1",
	}
	p := NewOpenAICompatibleProvider(cfg, "openai")

	p.SetMaxTokens(1024)
	if p.maxTokens != 1024 {
		t.Errorf("expected maxTokens 1024, got %d", p.maxTokens)
	}
}

func TestAnthropicDefaultModel(t *testing.T) {
	t.Run("uses config model", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:       "sk-ant-test",
			BaseURL:      "https://api.anthropic.com/v1",
			DefaultModel: "claude-opus-4-20250514",
		}
		p := NewAnthropicProvider(cfg, "anthropic")
		if p.DefaultModel() != "claude-opus-4-20250514" {
			t.Errorf("expected config model, got %q", p.DefaultModel())
		}
	})

	t.Run("falls back to hardcoded default", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "sk-ant-test",
			BaseURL: "https://api.anthropic.com/v1",
		}
		p := NewAnthropicProvider(cfg, "anthropic")
		if p.DefaultModel() == "" {
			t.Error("expected non-empty default model")
		}
	})

	t.Run("uses custom timeout", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:         "sk-ant-test",
			BaseURL:        "https://api.anthropic.com/v1",
			TimeoutSeconds: 120,
		}
		p := NewAnthropicProvider(cfg, "anthropic")
		if p.httpClient.Timeout.Seconds() != 120 {
			t.Errorf("expected timeout 120s, got %v", p.httpClient.Timeout)
		}
	})

	t.Run("uses default timeout when zero", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "sk-ant-test",
			BaseURL: "https://api.anthropic.com/v1",
		}
		p := NewAnthropicProvider(cfg, "anthropic")
		if p.httpClient.Timeout.Seconds() != 90 {
			t.Errorf("expected default timeout 90s, got %v", p.httpClient.Timeout)
		}
	})
}

func TestOpenAICompatibleDefaultModel(t *testing.T) {
	t.Run("uses config model", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:       "sk-test",
			BaseURL:      "https://api.openai.com/v1",
			DefaultModel: "gpt-4-turbo",
		}
		p := NewOpenAICompatibleProvider(cfg, "openai")
		if p.DefaultModel() != "gpt-4-turbo" {
			t.Errorf("expected config model, got %q", p.DefaultModel())
		}
	})

	t.Run("falls back to gpt-4o", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "sk-test",
			BaseURL: "https://api.openai.com/v1",
		}
		p := NewOpenAICompatibleProvider(cfg, "openai")
		if p.DefaultModel() != "gpt-4o" {
			t.Errorf("expected 'gpt-4o', got %q", p.DefaultModel())
		}
	})

	t.Run("uses custom timeout", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:         "sk-test",
			BaseURL:        "https://api.openai.com/v1",
			TimeoutSeconds: 60,
		}
		p := NewOpenAICompatibleProvider(cfg, "openai")
		if p.httpClient.Timeout.Seconds() != 60 {
			t.Errorf("expected timeout 60s, got %v", p.httpClient.Timeout)
		}
	})

	t.Run("uses default timeout when zero", func(t *testing.T) {
		cfg := config.Provider{
			APIKey:  "sk-test",
			BaseURL: "https://api.openai.com/v1",
		}
		p := NewOpenAICompatibleProvider(cfg, "openai")
		if p.httpClient.Timeout.Seconds() != 90 {
			t.Errorf("expected default timeout 90s, got %v", p.httpClient.Timeout)
		}
	})
}

func TestParseOpenAIResponse(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		content := `{"advice": "Use a smaller base image", "fix": "FROM alpine"}`
		result := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": content,
					},
				},
			},
		}

		resp, err := parseOpenAIResponse(result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Advice != "Use a smaller base image" {
			t.Errorf("expected advice 'Use a smaller base image', got %q", resp.Advice)
		}
		if resp.Fix != "FROM alpine" {
			t.Errorf("expected fix 'FROM alpine', got %q", resp.Fix)
		}
	})

	t.Run("valid response advice only", func(t *testing.T) {
		content := `{"advice": "Multi-stage builds are great"}`
		result := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": content,
					},
				},
			},
		}

		resp, err := parseOpenAIResponse(result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Advice != "Multi-stage builds are great" {
			t.Errorf("unexpected advice: %q", resp.Advice)
		}
		if resp.Fix != "" {
			t.Errorf("expected empty fix, got %q", resp.Fix)
		}
	})

	t.Run("delta (streaming) response", func(t *testing.T) {
		content := `{"advice": "Stream advice"}`
		result := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"delta": map[string]interface{}{
						"content": content,
					},
				},
			},
		}

		resp, err := parseOpenAIResponse(result)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Advice != "Stream advice" {
			t.Errorf("unexpected advice: %q", resp.Advice)
		}
	})

	t.Run("no choices", func(t *testing.T) {
		result := map[string]interface{}{
			"choices": []interface{}{},
		}
		_, err := parseOpenAIResponse(result)
		if err == nil {
			t.Error("expected error for empty choices")
		}
	})

	t.Run("missing choices key", func(t *testing.T) {
		result := map[string]interface{}{}
		_, err := parseOpenAIResponse(result)
		if err == nil {
			t.Error("expected error for missing choices")
		}
	})

	t.Run("invalid choice format", func(t *testing.T) {
		result := map[string]interface{}{
			"choices": []interface{}{"not a map"},
		}
		_, err := parseOpenAIResponse(result)
		if err == nil {
			t.Error("expected error for invalid choice format")
		}
	})

	t.Run("missing message", func(t *testing.T) {
		result := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{},
			},
		}
		_, err := parseOpenAIResponse(result)
		if err == nil {
			t.Error("expected error for missing message")
		}
	})

	t.Run("missing content", func(t *testing.T) {
		result := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{},
				},
			},
		}
		_, err := parseOpenAIResponse(result)
		if err == nil {
			t.Error("expected error for missing content")
		}
	})

	t.Run("invalid json in content", func(t *testing.T) {
		result := map[string]interface{}{
			"choices": []interface{}{
				map[string]interface{}{
					"message": map[string]interface{}{
						"content": "not valid json",
					},
				},
			},
		}
		_, err := parseOpenAIResponse(result)
		if err == nil {
			t.Error("expected error for invalid JSON content")
		}
	})
}

func TestBuildAdvicePrompt(t *testing.T) {
	t.Run("includes user context", func(t *testing.T) {
		prompt := BuildAdvicePrompt("Python Flask app with 1.2GB image using TensorFlow", "")

		for _, want := range []string{
			"Container Dietician",
			"Python Flask app with 1.2GB image using TensorFlow",
			"USER CONTEXT",
			"OUTPUT FORMAT",
			`"advice"`,
		} {
			if !strings.Contains(prompt, want) {
				t.Errorf("prompt missing %q", want)
			}
		}

		// Should NOT contain autofix or image analysis sections
		if strings.Contains(prompt, `"fix"`) {
			t.Error("advice prompt should not contain fix output format")
		}
		if strings.Contains(prompt, "IMAGE ANALYSIS") {
			t.Error("advice prompt should not contain IMAGE ANALYSIS")
		}
	})

	t.Run("includes dockerfile when provided", func(t *testing.T) {
		prompt := BuildAdvicePrompt("Go microservice", "FROM golang:1.21\nCOPY . /app")

		if !strings.Contains(prompt, "Go microservice") {
			t.Error("prompt missing context")
		}
		if !strings.Contains(prompt, "FROM golang:1.21") {
			t.Error("prompt missing dockerfile content")
		}
		if !strings.Contains(prompt, "DOCKERFILE CONTENT") {
			t.Error("prompt missing DOCKERFILE CONTENT header")
		}
	})

	t.Run("no dockerfile section when empty", func(t *testing.T) {
		prompt := BuildAdvicePrompt("Optimize my Node app", "")

		if strings.Contains(prompt, "DOCKERFILE CONTENT") {
			t.Error("prompt should not contain DOCKERFILE CONTENT when empty")
		}
	})
}
