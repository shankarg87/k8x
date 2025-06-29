package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/shankgan/k8x/internal/llm"
)

// AnthropicProvider implements the LLM Provider interface for Anthropic Claude
type AnthropicProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, baseURL, model string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-3-sonnet-20240229"
	}

	return &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// IsConfigured returns true if the provider is properly configured
func (p *AnthropicProvider) IsConfigured() bool {
	return p.apiKey != ""
}

// Chat sends a message and returns the response
func (p *AnthropicProvider) Chat(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Anthropic provider not configured: missing API key")
	}

	// Convert messages to Anthropic format
	anthropicMessages := make([]map[string]string, 0, len(messages))
	var systemMessage string

	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessage = msg.Content
		} else {
			anthropicMessages = append(anthropicMessages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	requestBody := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 4096,
		"messages":   anthropicMessages,
	}

	if systemMessage != "" {
		requestBody["system"] = systemMessage
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no response from Anthropic")
	}

	var content strings.Builder
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			content.WriteString(c.Text)
		}
	}

	return &llm.Response{
		Content: strings.TrimSpace(content.String()),
		Usage: &llm.Usage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
	}, nil
}

// Stream sends a message and returns a streaming response
func (p *AnthropicProvider) Stream(ctx context.Context, messages []llm.Message) (io.ReadCloser, error) {
	// TODO: Implement streaming support
	return nil, fmt.Errorf("streaming not yet implemented for Anthropic provider")
}
