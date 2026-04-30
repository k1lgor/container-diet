package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/k1lgor/container-diet/internal/config"
)

func TestAnthropicProvider_AnalyzeReport(t *testing.T) {
	t.Run("successful analysis", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request headers
			if r.Header.Get("X-API-Key") != "sk-ant-test" {
				t.Errorf("expected X-API-Key header, got %q", r.Header.Get("X-API-Key"))
			}
			if r.Header.Get("Anthropic-Version") != "2023-06-01" {
				t.Errorf("expected Anthropic-Version header")
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json")
			}

			// Parse request
			var req AnthropicRequest
			json.NewDecoder(r.Body).Decode(&req)

			if req.MaxTokens != 4096 {
				t.Errorf("expected MaxTokens 4096, got %d", req.MaxTokens)
			}
			if req.Model != "claude-sonnet-4-20250514" {
				t.Errorf("unexpected model: %s", req.Model)
			}

			// Return mock response
			content := `{"advice": "Use alpine base image 🐳", "fix": "FROM alpine"}`
			resp := AnthropicResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{{Type: "text", Text: content}},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{
			APIKey:  "sk-ant-test",
			BaseURL: server.URL,
		}
		p := NewAnthropicProvider(cfg, "anthropic")

		result, err := p.AnalyzeReport(nil, "FROM ubuntu", "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Advice != "Use alpine base image 🐳" {
			t.Errorf("unexpected advice: %q", result.Advice)
		}
		if result.Fix != "FROM alpine" {
			t.Errorf("unexpected fix: %q", result.Fix)
		}
	})

	t.Run("uses custom model", func(t *testing.T) {
		var receivedModel string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req AnthropicRequest
			json.NewDecoder(r.Body).Decode(&req)
			receivedModel = req.Model

			resp := AnthropicResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{{Type: "text", Text: `{"advice": "ok"}`}},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewAnthropicProvider(cfg, "anthropic")
		p.AnalyzeReport(nil, "FROM alpine", "claude-opus-4", false)

		if receivedModel != "claude-opus-4" {
			t.Errorf("expected model 'claude-opus-4', got %q", receivedModel)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "invalid api key"}`))
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "bad", BaseURL: server.URL}
		p := NewAnthropicProvider(cfg, "anthropic")

		_, err := p.AnalyzeReport(nil, "FROM alpine", "", false)
		if err == nil {
			t.Fatal("expected error for 401 response")
		}
	})

	t.Run("empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := AnthropicResponse{Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{}}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewAnthropicProvider(cfg, "anthropic")

		_, err := p.AnalyzeReport(nil, "FROM alpine", "", false)
		if err == nil {
			t.Fatal("expected error for empty content")
		}
	})

	t.Run("invalid json in content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := AnthropicResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{{Type: "text", Text: "not json"}},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewAnthropicProvider(cfg, "anthropic")

		_, err := p.AnalyzeReport(nil, "FROM alpine", "", false)
		if err == nil {
			t.Fatal("expected error for invalid JSON content")
		}
	})

	t.Run("advice only (no fix)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := AnthropicResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{{Type: "text", Text: `{"advice": "slim it down"}`}},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewAnthropicProvider(cfg, "anthropic")

		result, err := p.AnalyzeReport(nil, "FROM ubuntu", "", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Advice != "slim it down" {
			t.Errorf("unexpected advice: %q", result.Advice)
		}
		if result.Fix != "" {
			t.Errorf("expected empty fix, got %q", result.Fix)
		}
	})

	t.Run("base url with trailing slash", func(t *testing.T) {
		var requestURL string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURL = r.URL.Path
			resp := AnthropicResponse{
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{{Type: "text", Text: `{"advice": "ok"}`}},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL + "/"}
		p := NewAnthropicProvider(cfg, "anthropic")
		p.AnalyzeReport(nil, "FROM alpine", "", false)

		if requestURL != "/messages" {
			t.Errorf("expected request to /messages, got %q", requestURL)
		}
	})
}
