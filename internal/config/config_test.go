package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDefaultConfig(t *testing.T) {
	cfg, err := parse(DefaultConfigYAML)
	if err != nil {
		t.Fatalf("failed to parse default config: %v", err)
	}

	if len(cfg.Sources.Feeds) == 0 {
		t.Error("expected feeds to be populated")
	}

	if cfg.Summarization.Provider != "ollama" {
		t.Errorf("expected provider 'ollama', got %q", cfg.Summarization.Provider)
	}

	if cfg.Summarization.Model != "qwen2.5:7b" {
		t.Errorf("expected model 'qwen2.5:7b', got %q", cfg.Summarization.Model)
	}

	if cfg.Server.Port != 8000 {
		t.Errorf("expected port 8000, got %d", cfg.Server.Port)
	}
}

func TestParseMinimalConfig(t *testing.T) {
	data := []byte(`
summarization:
  provider: openai
  model: gpt-4o
server:
  port: 9000
`)
	cfg, err := parse(data)
	if err != nil {
		t.Fatalf("failed to parse minimal config: %v", err)
	}

	if cfg.Summarization.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %q", cfg.Summarization.Provider)
	}
	if cfg.Server.Port != 9000 {
		t.Errorf("expected port 9000, got %d", cfg.Server.Port)
	}
	// Defaults should still be set for unspecified fields
	if cfg.Summarization.OllamaURL != "http://localhost:11434" {
		t.Errorf("expected default ollama_url, got %q", cfg.Summarization.OllamaURL)
	}
}

func TestLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, DefaultConfigYAML, 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if len(cfg.Sources.Feeds) == 0 {
		t.Error("expected feeds to be populated from file")
	}
}

func TestGetDataDir(t *testing.T) {
	cfg := &Config{}
	defaultDir := cfg.GetDataDir()
	if defaultDir == "" {
		t.Error("expected non-empty default data dir")
	}

	cfg.Output.DataDir = "/custom/path"
	if cfg.GetDataDir() != "/custom/path" {
		t.Errorf("expected '/custom/path', got %q", cfg.GetDataDir())
	}
}
