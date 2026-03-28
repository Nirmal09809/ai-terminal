package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Provider ProviderConfig `yaml:"provider" json:"provider"`
	Tools    ToolsConfig    `yaml:"tools" json:"tools"`
	Session  SessionConfig  `yaml:"session" json:"session"`
	Context  ContextConfig  `yaml:"context" json:"context"`
	MCP      MCPConfig      `yaml:"mcp" json:"mcp"`
	UI       UIConfig       `yaml:"ui" json:"ui"`
	Security SecurityConfig `yaml:"security" json:"security"`
	Log      LogConfig      `yaml:"log" json:"log"`
}

type ProviderConfig struct {
	Default string           `yaml:"default" json:"default"`
	Model   string           `yaml:"model" json:"model"`
	Gemini  GeminiConfig    `yaml:"gemini" json:"gemini"`
	OpenAI  OpenAIConfig    `yaml:"openai" json:"openai"`
	Anthropic AnthropicConfig `yaml:"anthropic" json:"anthropic"`
	Ollama  OllamaConfig    `yaml:"ollama" json:"ollama"`
	Vertex  VertexConfig    `yaml:"vertex" json:"vertex"`
}

type GeminiConfig struct {
	APIKey      string  `yaml:"api_key" json:"api_key"`
	Model       string  `yaml:"model" json:"model"`
	Temperature float64 `yaml:"temperature" json:"temperature"`
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
	Grounding   bool    `yaml:"grounding" json:"grounding"`
}

type OpenAIConfig struct {
	APIKey       string `yaml:"api_key" json:"api_key"`
	BaseURL      string `yaml:"base_url" json:"base_url"`
	Model        string `yaml:"model" json:"model"`
	Temperature  float64 `yaml:"temperature" json:"temperature"`
	MaxTokens    int    `yaml:"max_tokens" json:"max_tokens"`
	Organization string `yaml:"organization" json:"organization"`
}

type AnthropicConfig struct {
	APIKey      string  `yaml:"api_key" json:"api_key"`
	BaseURL     string  `yaml:"base_url" json:"base_url"`
	Model       string  `yaml:"model" json:"model"`
	Temperature float64 `yaml:"temperature" json:"temperature"`
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
}

type OllamaConfig struct {
	URL     string `yaml:"url" json:"url"`
	Model   string `yaml:"model" json:"model"`
	Timeout int    `yaml:"timeout" json:"timeout"`
}

type VertexConfig struct {
	ProjectID   string `yaml:"project_id" json:"project_id"`
	Location    string `yaml:"location" json:"location"`
	Model       string `yaml:"model" json:"model"`
	Credentials string `yaml:"credentials" json:"credentials"`
}

type ToolsConfig struct {
	Shell ShellConfig `yaml:"shell" json:"shell"`
	Web   WebConfig   `yaml:"web" json:"web"`
	File  FileConfig  `yaml:"file" json:"file"`
}

type ShellConfig struct {
	Enabled         bool     `yaml:"enabled" json:"enabled"`
	AllowedCommands []string `yaml:"allowed_commands" json:"allowed_commands"`
	BlockedCommands []string `yaml:"blocked_commands" json:"blocked_commands"`
	Timeout         int      `yaml:"timeout" json:"timeout"`
	MaxOutputSize   int      `yaml:"max_output_size" json:"max_output_size"`
}

type WebConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`
	UserAgent  string `yaml:"user_agent" json:"user_agent"`
	Timeout    int    `yaml:"timeout" json:"timeout"`
	MaxRetries int    `yaml:"max_retries" json:"max_retries"`
}

type FileConfig struct {
	AllowedDirs []string `yaml:"allowed_dirs" json:"allowed_dirs"`
	BlockedDirs []string `yaml:"blocked_dirs" json:"blocked_dirs"`
	MaxFileSize int64    `yaml:"max_file_size" json:"max_file_size"`
}

type SessionConfig struct {
	StorageDir  string `yaml:"storage_dir" json:"storage_dir"`
	MaxHistory  int    `yaml:"max_history" json:"max_history"`
	AutoSave    bool   `yaml:"auto_save" json:"auto_save"`
	SaveInterval int   `yaml:"save_interval" json:"save_interval"`
}

type ContextConfig struct {
	MaxTokens     int    `yaml:"max_tokens" json:"max_tokens"`
	Strategy      string `yaml:"strategy" json:"strategy"`
	SlidingWindow int    `yaml:"sliding_window" json:"sliding_window"`
}

type MCPConfig struct {
	Enabled bool       `yaml:"enabled" json:"enabled"`
	Servers []MCPServer `yaml:"servers" json:"servers"`
}

type MCPServer struct {
	Name    string            `yaml:"name" json:"name"`
	Command string            `yaml:"command" json:"command"`
	Args    []string          `yaml:"args" json:"args"`
	Env     map[string]string `yaml:"env" json:"env"`
}

type UIConfig struct {
	Theme      string `yaml:"theme" json:"theme"`
	Streaming  bool   `yaml:"streaming" json:"streaming"`
	Color256   bool   `yaml:"color_256" json:"color_256"`
	ShowTyping bool   `yaml:"show_typing" json:"show_typing"`
	MaxWidth   int    `yaml:"max_width" json:"max_width"`
}

type SecurityConfig struct {
	Sandbox   bool `yaml:"sandbox" json:"sandbox"`
	RateLimit int  `yaml:"rate_limit" json:"rate_limit"`
	AuditLog  bool `yaml:"audit_log" json:"audit_log"`
}

type LogConfig struct {
	Level  string `yaml:"level" json:"level"`
	File   string `yaml:"file" json:"file"`
	Format string `yaml:"format" json:"format"`
}

func Default() *Config {
	return &Config{
		Provider: ProviderConfig{
			Default: "gemini",
			Model:   "gemini-2.0-flash",
			Gemini: GeminiConfig{
				Model:       "gemini-2.0-flash",
				Temperature: 0.7,
				MaxTokens:   8192,
				Grounding:   true,
			},
			OpenAI: OpenAIConfig{
				Model:       "gpt-4o",
				Temperature: 0.7,
				MaxTokens:   8192,
			},
			Anthropic: AnthropicConfig{
				Model:       "claude-3-5-sonnet-20241022",
				Temperature: 0.7,
				MaxTokens:   8192,
			},
			Ollama: OllamaConfig{
				URL:     "http://localhost:11434",
				Model:   "llama3",
				Timeout: 120,
			},
		},
		Tools: ToolsConfig{
			Shell: ShellConfig{
				Enabled:         true,
				Timeout:         30,
				MaxOutputSize:   1024 * 1024,
			},
			Web: WebConfig{
				Enabled:    true,
				UserAgent: "ai-terminal/1.0",
				Timeout:    30,
				MaxRetries: 3,
			},
			File: FileConfig{
				MaxFileSize: 10 * 1024 * 1024,
			},
		},
		Session: SessionConfig{
			StorageDir:  "~/.ai-terminal/sessions",
			MaxHistory:  100,
			AutoSave:    true,
			SaveInterval: 60,
		},
		Context: ContextConfig{
			MaxTokens:     128000,
			Strategy:      "sliding",
			SlidingWindow: 10,
		},
		MCP: MCPConfig{
			Enabled: true,
			Servers: []MCPServer{},
		},
		UI: UIConfig{
			Theme:      "dark",
			Streaming:  true,
			Color256:   true,
			ShowTyping: true,
			MaxWidth:   80,
		},
		Security: SecurityConfig{
			Sandbox:   true,
			RateLimit: 60,
			AuditLog:  true,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func Load(path string) (*Config, error) {
	godotenv.Load()

	cfg := Default()

	if path == "" {
		path = GetConfigPath()
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	viper.SetDefault("provider.default", "gemini")
	viper.SetDefault("provider.gemini.model", "gemini-2.0-flash")
	viper.SetDefault("provider.openai.model", "gpt-4o")
	viper.SetDefault("provider.anthropic.model", "claude-3-5-sonnet-20241022")
	viper.SetDefault("provider.ollama.model", "llama3")

	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create config directory: %w", err)
			}
			if err := Save(path, cfg); err != nil {
				return nil, fmt.Errorf("failed to save default config: %w", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cfg.Provider.Gemini.APIKey == "" {
		cfg.Provider.Gemini.APIKey = os.Getenv("GEMINI_API_KEY")
	}
	if cfg.Provider.OpenAI.APIKey == "" {
		cfg.Provider.OpenAI.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.Provider.Anthropic.APIKey == "" {
		cfg.Provider.Anthropic.APIKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	return cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-terminal", "config.yaml")
}
