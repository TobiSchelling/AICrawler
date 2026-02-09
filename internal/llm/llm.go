package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Provider is the interface for LLM providers.
type Provider interface {
	Generate(ctx context.Context, prompt string, maxTokens int) (string, error)
	IsConfigured() bool
}

// Embedder is the interface for generating embeddings.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

// OllamaProvider is a local Ollama LLM provider.
type OllamaProvider struct {
	Model   string
	BaseURL string
	client  *http.Client
}

// NewOllamaProvider creates a new Ollama provider.
func NewOllamaProvider(model, baseURL string) *OllamaProvider {
	return &OllamaProvider{
		Model:   model,
		BaseURL: baseURL,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// IsConfigured checks if Ollama is running and the model is available.
func (o *OllamaProvider) IsConfigured() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", o.BaseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false
	}

	modelBase := strings.SplitN(o.Model, ":", 2)[0]
	for _, m := range result.Models {
		if strings.Contains(m.Name, modelBase) {
			return true
		}
	}
	log.Printf("Ollama model %q not found", o.Model)
	return false
}

// Generate sends a prompt to Ollama and returns the response.
func (o *OllamaProvider) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	body := map[string]any{
		"model": o.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"stream": false,
		"options": map[string]any{
			"num_predict":  maxTokens,
			"temperature": 0.3,
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.BaseURL+"/api/chat", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return result.Message.Content, nil
}

// OllamaEmbedder generates embeddings via the Ollama API.
type OllamaEmbedder struct {
	Model   string
	BaseURL string
	client  *http.Client
}

// NewOllamaEmbedder creates a new Ollama embedder.
func NewOllamaEmbedder(model, baseURL string) *OllamaEmbedder {
	return &OllamaEmbedder{
		Model:   model,
		BaseURL: baseURL,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// Embed generates embeddings for the given texts.
func (e *OllamaEmbedder) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	body := map[string]any{
		"model": e.Model,
		"input": texts,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.BaseURL+"/api/embed", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding embeddings: %w", err)
	}

	return result.Embeddings, nil
}

// OpenAIProvider is an OpenAI API provider.
type OpenAIProvider struct {
	Model  string
	APIKey string
	client *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(model, apiKeyEnv string) *OpenAIProvider {
	return &OpenAIProvider{
		Model:  model,
		APIKey: os.Getenv(apiKeyEnv),
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

// IsConfigured checks if the API key is set.
func (o *OpenAIProvider) IsConfigured() bool {
	return o.APIKey != ""
}

// Generate sends a prompt to OpenAI and returns the response.
func (o *OpenAIProvider) Generate(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if o.APIKey == "" {
		return "", fmt.Errorf("OpenAI API key not configured")
	}

	body := map[string]any{
		"model": o.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens":  maxTokens,
		"temperature": 0.3,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.APIKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return result.Choices[0].Message.Content, nil
}

// CreateProvider creates an LLM provider based on configuration.
func CreateProvider(provider, model, ollamaURL, openaiModel, apiKeyEnv string) Provider {
	if strings.ToLower(provider) == "ollama" {
		p := NewOllamaProvider(model, ollamaURL)
		if p.IsConfigured() {
			log.Printf("Using Ollama with model: %s", model)
			return p
		}
		log.Println("Ollama not available, trying OpenAI fallback...")
	}

	p := NewOpenAIProvider(openaiModel, apiKeyEnv)
	if p.IsConfigured() {
		log.Printf("Using OpenAI with model: %s", openaiModel)
		return p
	}

	log.Println("No LLM provider available. Check Ollama is running or set OPENAI_API_KEY.")
	return nil
}
