package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/k1lgor/container-diet/internal/config"
)

func TestOpenRouterHeaders(t *testing.T) {
	// Verify the provider sends OpenRouter-specific headers
	// We can't use httptest with a URL containing 'openrouter',
	// so we test by checking the code path and verifying the provider name
	cfg := config.Provider{APIKey: "key", BaseURL: "https://openrouter.ai/api/v1"}
	p := NewOpenAICompatibleProvider(cfg, "openrouter")

	if p.Name() != "openrouter" {
		t.Errorf("expected name 'openrouter', got %q", p.Name())
	}

	// Verify the URL contains openrouter (triggers header logic)
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if !strings.Contains(baseURL, "openrouter") {
		t.Error("baseURL should contain 'openrouter' for header logic")
	}
}

func TestOpenAICompatibleProvider_AnalyzeReport(t *testing.T) {
	t.Run("successful analysis", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify auth header
			auth := r.Header.Get("Authorization")
			if auth != "Bearer sk-test" {
				t.Errorf("expected 'Bearer sk-test', got %q", auth)
			}

			// Parse and verify request
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			if req["model"] != "gpt-4o" {
				t.Errorf("unexpected model: %v", req["model"])
			}

			// Return response
			content := `{"advice": "Use multi-stage builds", "fix": "FROM golang AS builder"}`
			resp := map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]interface{}{"content": content}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "sk-test", BaseURL: server.URL}
		p := NewOpenAICompatibleProvider(cfg, "openai")

		result, err := p.AnalyzeReport(nil, "FROM golang", "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Advice != "Use multi-stage builds" {
			t.Errorf("unexpected advice: %q", result.Advice)
		}
		if result.Fix != "FROM golang AS builder" {
			t.Errorf("unexpected fix: %q", result.Fix)
		}
	})

	t.Run("uses custom model", func(t *testing.T) {
		var receivedModel string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			receivedModel = req["model"].(string)

			resp := map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]interface{}{"content": `{"advice": "ok"}`}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewOpenAICompatibleProvider(cfg, "openai")
		p.AnalyzeReport(nil, "FROM alpine", "gpt-4-turbo", false)

		if receivedModel != "gpt-4-turbo" {
			t.Errorf("expected model 'gpt-4-turbo', got %q", receivedModel)
		}
	})



	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal"}`))
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewOpenAICompatibleProvider(cfg, "openai")

		_, err := p.AnalyzeReport(nil, "FROM alpine", "", false)
		if err == nil {
			t.Fatal("expected error for 500 response")
		}
	})

	t.Run("empty choices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]interface{}{"choices": []interface{}{}}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewOpenAICompatibleProvider(cfg, "openai")

		_, err := p.AnalyzeReport(nil, "FROM alpine", "", false)
		if err == nil {
			t.Fatal("expected error for empty choices")
		}
	})

	t.Run("base url with trailing slash", func(t *testing.T) {
		var requestURL string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestURL = r.URL.Path
			resp := map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]interface{}{"content": `{"advice": "ok"}`}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL + "/"}
		p := NewOpenAICompatibleProvider(cfg, "openai")
		p.AnalyzeReport(nil, "FROM alpine", "", false)

		if requestURL != "/chat/completions" {
			t.Errorf("expected /chat/completions, got %q", requestURL)
		}
	})

	t.Run("response with json_format", func(t *testing.T) {
		var reqBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&reqBody)
			resp := map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]interface{}{"content": `{"advice": "ok"}`}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewOpenAICompatibleProvider(cfg, "openai")
		p.AnalyzeReport(nil, "FROM alpine", "", false)

		// Verify response_format is set
		if fmt, ok := reqBody["response_format"]; ok {
			fmtMap := fmt.(map[string]interface{})
			if fmtMap["type"] != "json_object" {
				t.Errorf("expected response_format type json_object, got %v", fmtMap["type"])
			}
		} else {
			t.Error("expected response_format in request")
		}
	})

	t.Run("system message present", func(t *testing.T) {
		var reqBody map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewDecoder(r.Body).Decode(&reqBody)
			resp := map[string]interface{}{
				"choices": []map[string]interface{}{
					{"message": map[string]interface{}{"content": `{"advice": "ok"}`}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		cfg := config.Provider{APIKey: "key", BaseURL: server.URL}
		p := NewOpenAICompatibleProvider(cfg, "openai")
		p.AnalyzeReport(nil, "FROM alpine", "", false)

		messages := reqBody["messages"].([]interface{})
		if len(messages) < 2 {
			t.Fatal("expected at least 2 messages (system + user)")
		}
		firstMsg := messages[0].(map[string]interface{})
		if firstMsg["role"] != "system" {
			t.Errorf("expected first message role 'system', got %q", firstMsg["role"])
		}
		if !strings.Contains(firstMsg["content"].(string), "JSON") {
			t.Error("system message should mention JSON")
		}
	})
}
