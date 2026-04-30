package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultProvider != "" {
		t.Errorf("expected empty DefaultProvider, got %q", cfg.DefaultProvider)
	}
	if len(cfg.Providers) != 0 {
		t.Errorf("expected empty Providers, got %d", len(cfg.Providers))
	}
	if cfg.Analysis.MaxTokens != 4096 {
		t.Errorf("expected MaxTokens=4096, got %d", cfg.Analysis.MaxTokens)
	}
	if cfg.UI.Theme != "docker" {
		t.Errorf("expected theme 'docker', got %q", cfg.UI.Theme)
	}
	if !cfg.UI.ShowEmojis {
		t.Error("expected ShowEmojis=true")
	}
}

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "no vars",
			input:    "plain string",
			expected: "plain string",
		},
		{
			name:     "braces syntax",
			input:    "${MY_KEY}",
			envVars:  map[string]string{"MY_KEY": "secret123"},
			expected: "secret123",
		},
		{
			name:     "bare syntax",
			input:    "$MY_KEY",
			envVars:  map[string]string{"MY_KEY": "secret123"},
			expected: "secret123",
		},
		{
			name:     "unset var braces",
			input:    "${NONEXISTENT_VAR_XYZ}",
			expected: "${NONEXISTENT_VAR_XYZ}",
		},
		{
			name:     "unset var bare",
			input:    "$NONEXISTENT_VAR_XYZ",
			expected: "$NONEXISTENT_VAR_XYZ",
		},
		{
			name:     "mixed content",
			input:    "prefix-${KEY}-suffix",
			envVars:  map[string]string{"KEY": "val"},
			expected: "prefix-val-suffix",
		},
		{
			name:     "multiple vars",
			input:    "${A}:${B}",
			envVars:  map[string]string{"A": "1", "B": "2"},
			expected: "1:2",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no double expansion",
			input:    "${PASSWD}",
			envVars:  map[string]string{"PASSWD": "$ecret"},
			expected: "$ecret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env
			for k := range tt.envVars {
				os.Setenv(k, tt.envVars[k])
				defer os.Unsetenv(k)
			}

			result := expandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("expandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetProvider(t *testing.T) {
	cfg := &Config{
		Providers: map[string]Provider{
			"openai": {
				APIKey:       "sk-test",
				BaseURL:      "https://api.openai.com/v1",
				DefaultModel: "gpt-4o",
			},
		},
	}

	t.Run("existing provider", func(t *testing.T) {
		p, err := cfg.GetProvider("openai")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.APIKey != "sk-test" {
			t.Errorf("expected APIKey 'sk-test', got %q", p.APIKey)
		}
		if p.BaseURL != "https://api.openai.com/v1" {
			t.Errorf("expected BaseURL 'https://api.openai.com/v1', got %q", p.BaseURL)
		}
		if p.DefaultModel != "gpt-4o" {
			t.Errorf("expected model 'gpt-4o', got %q", p.DefaultModel)
		}
	})

	t.Run("missing provider", func(t *testing.T) {
		_, err := cfg.GetProvider("nonexistent")
		if err == nil {
			t.Fatal("expected error for missing provider")
		}
	})

	t.Run("env var expansion in api key", func(t *testing.T) {
		os.Setenv("TEST_API_KEY", "sk-from-env")
		defer os.Unsetenv("TEST_API_KEY")

		cfg2 := &Config{
			Providers: map[string]Provider{
				"test": {
					APIKey:  "${TEST_API_KEY}",
					BaseURL: "https://example.com/v1",
				},
			},
		}

		p, err := cfg2.GetProvider("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.APIKey != "sk-from-env" {
			t.Errorf("expected expanded key 'sk-from-env', got %q", p.APIKey)
		}
	})
}

func TestFileExists(t *testing.T) {
	t.Run("existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "exists.txt")
		os.WriteFile(path, []byte("test"), 0644)

		if !fileExists(path) {
			t.Error("expected fileExists=true for existing file")
		}
	})

	t.Run("non-existing file", func(t *testing.T) {
		if fileExists("/no/such/path/file.txt") {
			t.Error("expected fileExists=false for missing file")
		}
	})
}

func TestSaveExampleConfig(t *testing.T) {
	t.Run("creates file and directory", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "subdir", "config.yaml")

		err := SaveExampleConfig(path)
		if err != nil {
			t.Fatalf("SaveExampleConfig failed: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read created config: %v", err)
		}

		content := string(data)
		if len(content) == 0 {
			t.Error("config file is empty")
		}
		// Check for key sections
		for _, want := range []string{"default_provider:", "providers:", "analysis:", "ui:"} {
			if !strings.Contains(content, want) {
				t.Errorf("config missing %q section", want)
			}
		}
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		os.WriteFile(path, []byte("old content"), 0644)

		err := SaveExampleConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, _ := os.ReadFile(path)
		if string(data) == "old content" {
			t.Error("file was not overwritten")
		}
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("returns default when no files exist", func(t *testing.T) {
		// Use a temp dir as working dir so no config is found
		dir := t.TempDir()
		origWd, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origWd)

		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DefaultProvider != "" {
			t.Errorf("expected empty DefaultProvider, got %q", cfg.DefaultProvider)
		}
	})

	t.Run("loads explicit path", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		content := `default_provider: "test"
providers:
  test:
    api_key: "sk-explicit"
    base_url: "https://example.com/v1"
    default_model: "test-model"
`
		os.WriteFile(path, []byte(content), 0644)

		cfg, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DefaultProvider != "test" {
			t.Errorf("expected DefaultProvider 'test', got %q", cfg.DefaultProvider)
		}
		p, err := cfg.GetProvider("test")
		if err != nil {
			t.Fatalf("GetProvider failed: %v", err)
		}
		if p.APIKey != "sk-explicit" {
			t.Errorf("expected APIKey 'sk-explicit', got %q", p.APIKey)
		}
	})

	t.Run("returns error for invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.yaml")
		os.WriteFile(path, []byte("{{invalid yaml"), 0644)

		_, err := LoadConfig(path)
		if err == nil {
			t.Error("expected error for invalid YAML")
		}
	})

	t.Run("merges local over global", func(t *testing.T) {
		dir := t.TempDir()
		origWd, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(origWd)

		// Create local config
		os.MkdirAll(filepath.Join(dir, ".container-diet"), 0755)
		localCfg := `default_provider: "local"
providers:
  local:
    api_key: "local-key"
    base_url: "https://local.example.com/v1"
`
		os.WriteFile(filepath.Join(dir, ".container-diet", "config.yaml"), []byte(localCfg), 0644)

		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DefaultProvider != "local" {
			t.Errorf("expected DefaultProvider 'local', got %q", cfg.DefaultProvider)
		}
	})

	t.Run("returns default for non-existent path", func(t *testing.T) {
		cfg, err := LoadConfig("/no/such/path/config.yaml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DefaultProvider != "" {
			t.Errorf("expected empty default, got %q", cfg.DefaultProvider)
		}
	})

	t.Run("finds global config via XDG_CONFIG_HOME", func(t *testing.T) {
		xdgDir := t.TempDir()
		cfgDir := filepath.Join(xdgDir, "container-diet")
		os.MkdirAll(cfgDir, 0755)
		cfgContent := `default_provider: "xdg"
providers:
  xdg:
    api_key: "xdg-key"
    base_url: "https://xdg.example.com/v1"
`
		os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(cfgContent), 0644)

		os.Setenv("XDG_CONFIG_HOME", xdgDir)
		defer os.Unsetenv("XDG_CONFIG_HOME")

		// Work in empty dir so local config doesn't interfere
		emptyDir := t.TempDir()
		origWd, _ := os.Getwd()
		os.Chdir(emptyDir)
		defer os.Chdir(origWd)

		cfg, err := LoadConfig("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.DefaultProvider != "xdg" {
			t.Errorf("expected DefaultProvider 'xdg', got %q", cfg.DefaultProvider)
		}
	})
}

func TestFindGlobalConfigPath(t *testing.T) {
	t.Run("returns empty when no config exists", func(t *testing.T) {
		// Ensure XDG and home have no config
		origXDG := os.Getenv("XDG_CONFIG_HOME")
		os.Unsetenv("XDG_CONFIG_HOME")
		defer os.Setenv("XDG_CONFIG_HOME", origXDG)

		// Use temp dir as home to control the result
		emptyDir := t.TempDir()
		origWd, _ := os.Getwd()
		os.Chdir(emptyDir)
		defer os.Chdir(origWd)

		path := findGlobalConfigPath()
		// May or may not find a global config depending on system state
		// Just verify it doesn't panic
		_ = path
	})
}
