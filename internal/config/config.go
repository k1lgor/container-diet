package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the full configuration for container-diet
type Config struct {
	DefaultProvider string              `yaml:"default_provider"`
	Providers       map[string]Provider `yaml:"providers"`
	Analysis        AnalysisConfig      `yaml:"analysis"`
	UI              UIConfig            `yaml:"ui"`
}

// Provider represents a single AI provider configuration
type Provider struct {
	Type           string `yaml:"type"` // "anthropic" for Anthropic native API, empty/anything else for OpenAI-compatible
	APIKey         string `yaml:"api_key"`
	BaseURL        string `yaml:"base_url"`
	DefaultModel   string `yaml:"default_model"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

// AnalysisConfig holds analysis-related settings
type AnalysisConfig struct {
	MaxTokens int `yaml:"max_tokens"`
}

// UIConfig holds UI preference settings
type UIConfig struct {
	Theme      string `yaml:"theme"`
	ShowEmojis bool   `yaml:"show_emojis"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultProvider: "",
		Providers:       map[string]Provider{},
		Analysis: AnalysisConfig{
			MaxTokens: 4096,
		},
		UI: UIConfig{
			Theme:      "docker",
			ShowEmojis: true,
		},
	}
}

// GetProvider returns the configuration for a specific provider
func (c *Config) GetProvider(name string) (Provider, error) {
	if provider, ok := c.Providers[name]; ok {
		provider.APIKey = expandEnvVars(provider.APIKey)
		return provider, nil
	}
	return Provider{}, fmt.Errorf("provider '%s' not found in config", name)
}

// LoadConfig loads configuration from file or returns default if not found.
// Config hierarchy (later overrides earlier):
//  1. Global config: ~/.config/container-diet/config.yaml
//  2. Local config: ./.container-diet/config.yaml (project-specific)
//  3. Explicit path (if provided, overrides everything)
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// If explicit path provided, load only that
	if path != "" {
		return loadConfigFromPath(config, path)
	}

	// Try global config first
	if globalPath := findGlobalConfigPath(); globalPath != "" {
		if merged, err := loadConfigFromPath(config, globalPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load global config '%s': %v\n", globalPath, err)
		} else {
			config = merged
		}
	}

	// Try local config (overrides global)
	if localPath := findLocalConfigPath(); localPath != "" {
		if merged, err := loadConfigFromPath(config, localPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load local config '%s': %v\n", localPath, err)
		} else {
			config = merged
		}
	}

	return config, nil
}

// loadConfigFromPath loads config from a specific path, merging with existing config
func loadConfigFromPath(baseConfig *Config, path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return baseConfig, nil
		}
		return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
	}

	config := baseConfig
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file '%s': %w", path, err)
	}

	return config, nil
}

// findGlobalConfigPath searches for global config in home directory
func findGlobalConfigPath() string {
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		path := filepath.Join(xdgConfig, "container-diet", "config.yaml")
		if fileExists(path) {
			return path
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	path := filepath.Join(homeDir, ".config", "container-diet", "config.yaml")
	if fileExists(path) {
		return path
	}

	return ""
}

// findLocalConfigPath searches for local config in current directory
func findLocalConfigPath() string {
	localPath := filepath.Join(".container-diet", "config.yaml")
	if fileExists(localPath) {
		return localPath
	}
	return ""
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// envVarPattern matches ${VAR} and $VAR syntax for environment variable expansion.
var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// expandEnvVars expands environment variables in the format ${VAR} or $VAR
// Uses a single regex pass to prevent double-expansion of substituted values.
func expandEnvVars(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		var varName string
		if strings.HasPrefix(match, "${") {
			// ${VAR} syntax
			varName = match[2 : len(match)-1]
		} else {
			// $VAR syntax
			varName = match[1:]
		}
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match
	})
}

// SaveExampleConfig creates an example configuration file
func SaveExampleConfig(path string) error {
	example := `# Container Diet Configuration
#
# IMPORTANT: This is the MAIN configuration file. You MUST configure your providers here.
# The application will NOT work without proper configuration.
#
# Config file locations (searched in order, later overrides earlier):
#   1. Global: ~/.config/container-diet/config.yaml
#   2. Local:  ./.container-diet/config.yaml (project-specific, overrides global)
#
# Setup Instructions:
#   1. Uncomment and configure the provider you want to use
#   2. Set default_provider to the provider name you configured
#   3. Set the api_key (use environment variables with ${VAR_NAME} syntax)
#   4. Set the default_model to your preferred model
#   5. For MCP: The server will read this same config file
#
# Security best practices:
#   - Use environment variable references: api_key: "${OPENAI_API_KEY}"
#   - Add .container-diet/ to .gitignore for local configs
#   - Never commit API keys to version control

# Set this to the provider you want to use by default
# Must match one of the provider names below
default_provider: ""  # Example: "openai", "anthropic", "openrouter", etc.

providers:
  # === OpenAI ===
  # Uncomment and configure if using OpenAI
  # openai:
  #   api_key: "${OPENAI_API_KEY}"
  #   base_url: "https://api.openai.com/v1"
  #   default_model: ""  # Example: "gpt-4o", "gpt-4", "gpt-5"
  #   timeout_seconds: 90

  # === Anthropic (Native API) ===
  # Uncomment and configure if using Anthropic directly
  # anthropic:
  #   type: "anthropic"  # Required for Anthropic's native API (non-OpenAI-compatible)
  #   api_key: "${ANTHROPIC_API_KEY}"
  #   base_url: "https://api.anthropic.com/v1"
  #   default_model: ""  # Example: "claude-sonnet-4.6", "claude-opus-4.6"
  #   timeout_seconds: 90

  # === OpenRouter ===
  # Access models from multiple providers through one API
  # openrouter:
  #   api_key: "${OPENROUTER_API_KEY}"
  #   base_url: "https://openrouter.ai/api/v1"
  #   default_model: ""  # Example: "anthropic/claude-sonnet-4.6", "openai/gpt-5"
  #   timeout_seconds: 90

  # === Ollama (Local LLM) ===
  # Run models locally - no API key needed for local instance
  # ollama:
  #   api_key: "ollama"  # Can be any non-empty string for local Ollama
  #   base_url: "http://localhost:11434/v1"
  #   default_model: ""  # Example: "llama3.1", "mistral", "codellama"
  #   timeout_seconds: 120

  # === Groq ===
  # Fast inference API
  # groq:
  #   api_key: "${GROQ_API_KEY}"
  #   base_url: "https://api.groq.com/openai/v1"
  #   default_model: ""  # Example: "llama-3.1-70b-versatile", "mixtral-8x7b-32768"
  #   timeout_seconds: 90

  # === Perplexity ===
  # AI search and conversational API
  # perplexity:
  #   api_key: "${PERPLEXITY_API_KEY}"
  #   base_url: "https://api.perplexity.ai"
  #   default_model: ""  # Example: "llama-3.1-sonar-large-128k-online"
  #   timeout_seconds: 90

  # === Mistral AI ===
  # Native Mistral API
  # mistral:
  #   api_key: "${MISTRAL_API_KEY}"
  #   base_url: "https://api.mistral.ai/v1"
  #   default_model: ""  # Example: "mistral-large-latest", "codestral-latest"
  #   timeout_seconds: 90

  # === DeepSeek ===
  # DeepSeek's models
  # deepseek:
  #   api_key: "${DEEPSEEK_API_KEY}"
  #   base_url: "https://api.deepseek.com/v1"
  #   default_model: ""  # Example: "deepseek-chat", "deepseek-coder"
  #   timeout_seconds: 90

  # === xAI (Grok) ===
  # Elon Musk's xAI with Grok models
  # xai:
  #   api_key: "${XAI_API_KEY}"
  #   base_url: "https://api.x.ai/v1"
  #   default_model: ""  # Example: "grok-beta", "grok-vision-beta"
  #   timeout_seconds: 90

  # === Moonshot AI ===
  # Chinese AI company (月之暗面)
  # moonshot:
  #   api_key: "${MOONSHOT_API_KEY}"
  #   base_url: "https://api.moonshot.cn/v1"
  #   default_model: ""  # Example: "moonshot-v1-8k", "moonshot-v1-128k"
  #   timeout_seconds: 90

  # === Hugging Face Inference API ===
  # Access thousands of models
  # huggingface:
  #   api_key: "${HF_API_TOKEN}"
  #   base_url: "https://api-inference.huggingface.co/models"
  #   default_model: ""  # Example: "meta-llama/Meta-Llama-3-70B-Instruct"
  #   timeout_seconds: 120

  # === Custom/OpenAI-Compatible ===
  # Any OpenAI-compatible API endpoint
  # custom:
  #   api_key: "${CUSTOM_API_KEY}"
  #   base_url: "https://your-api-endpoint.com/v1"
  #   default_model: ""  # Your model name
  #   timeout_seconds: 90

analysis:
  max_tokens: 4096

ui:
  theme: "docker"  # Options: docker, minimal
  show_emojis: true
`

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(example), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Tighten directory permissions
	dirPerm := os.FileMode(0700)
	os.Chmod(dir, dirPerm)

	return nil
}
