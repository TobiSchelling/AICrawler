package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var DefaultConfigYAML []byte

type Config struct {
	Sources       Sources       `yaml:"sources"`
	Keywords      []string      `yaml:"keywords"`
	Summarization Summarization `yaml:"summarization"`
	Output        Output        `yaml:"output"`
	Server        Server        `yaml:"server"`
	Logging       Logging       `yaml:"logging"`
}

type Sources struct {
	Feeds []Feed     `yaml:"feeds"`
	APIs  APIsConfig `yaml:"apis"`
}

type Feed struct {
	URL  string `yaml:"url"`
	Name string `yaml:"name"`
}

type APIsConfig struct {
	NewsAPI NewsAPIConfig `yaml:"newsapi"`
}

type NewsAPIConfig struct {
	Enabled   bool   `yaml:"enabled"`
	APIKeyEnv string `yaml:"api_key_env"`
	Query     string `yaml:"query"`
}

type Summarization struct {
	Provider       string `yaml:"provider"`
	Model          string `yaml:"model"`
	OllamaURL      string `yaml:"ollama_url"`
	EmbeddingModel string `yaml:"embedding_model"`
	OpenAIModel    string `yaml:"openai_model"`
	APIKeyEnv      string `yaml:"api_key_env"`
	MaxTokens      int    `yaml:"max_tokens"`
}

type Output struct {
	DataDir string `yaml:"data_dir"`
}

type Server struct {
	Port int `yaml:"port"`
}

type Logging struct {
	Level string `yaml:"level"`
}

// ConfigDir returns the XDG config directory for aicrawler.
func ConfigDir() string {
	return filepath.Join(homeDir(), ".config", "aicrawler")
}

// DataDir returns the XDG data directory for aicrawler.
func DataDir() string {
	return filepath.Join(homeDir(), ".local", "share", "aicrawler")
}

// ResolveConfigPath finds the config file following priority:
// explicit path > ~/.config/aicrawler/config.yaml > ./config.yaml
func ResolveConfigPath(explicit string) (string, error) {
	if explicit != "" {
		if _, err := os.Stat(explicit); err != nil {
			return "", fmt.Errorf("config file not found: %s", explicit)
		}
		return explicit, nil
	}

	xdgConfig := filepath.Join(ConfigDir(), "config.yaml")
	if _, err := os.Stat(xdgConfig); err == nil {
		return xdgConfig, nil
	}

	cwdConfig := "config.yaml"
	if _, err := os.Stat(cwdConfig); err == nil {
		return cwdConfig, nil
	}

	return "", fmt.Errorf(
		"no config file found; searched:\n  %s\n  ./config.yaml\n\nRun 'aicrawler init' to create a default config",
		xdgConfig,
	)
}

// Load reads and parses a config YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	return parse(data)
}

// parse parses YAML bytes into a Config, applying defaults.
func parse(data []byte) (*Config, error) {
	cfg := &Config{
		Sources: Sources{
			APIs: APIsConfig{
				NewsAPI: NewsAPIConfig{
					Enabled:   true,
					APIKeyEnv: "NEWSAPI_KEY",
					Query:     "artificial intelligence software development",
				},
			},
		},
		Summarization: Summarization{
			Provider:       "ollama",
			Model:          "qwen2.5:7b",
			OllamaURL:      "http://localhost:11434",
			EmbeddingModel: "nomic-embed-text",
			OpenAIModel:    "gpt-4o-mini",
			APIKeyEnv:      "OPENAI_API_KEY",
			MaxTokens:      512,
		},
		Server: Server{Port: 8000},
		Logging: Logging{Level: "INFO"},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// GetDataDir returns the effective data directory from config or XDG default.
func (c *Config) GetDataDir() string {
	if c.Output.DataDir != "" {
		return c.Output.DataDir
	}
	return DataDir()
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
