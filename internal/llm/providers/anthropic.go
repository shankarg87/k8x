package providers

import (
	"context"
	"fmt"
	"io"

	"k8x/internal/llm"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider implements the llm.Provider interface for Claude
type AnthropicProvider struct {
	client  *anthropic.Client
	model   string
	apiKey  string
	baseURL string
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey, baseURL, model string) *AnthropicProvider {
	// Default to Claude Sonnet 4 if no model is specified
	if model == "" {
		model = string(anthropic.ModelClaudeSonnet4_0)
	}
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	var client *anthropic.Client
	if apiKey != "" {
		opts := []option.RequestOption{option.WithAPIKey(apiKey)}
		if baseURL != "https://api.anthropic.com" {
			opts = append(opts, option.WithBaseURL(baseURL))
		}
		c := anthropic.NewClient(opts...)
		client = &c
	}

	return &AnthropicProvider{
		client:  client,
		model:   model,
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// IsConfigured returns true if the provider has been initialized
func (p *AnthropicProvider) IsConfigured() bool {
	return p.apiKey != "" && p.client != nil
}

// Chat sends messages to Claude and returns the response
func (p *AnthropicProvider) Chat(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("Anthropic provider not configured: missing API key or client")
	}

	// Convert messages
	anthroMsgs := make([]anthropic.MessageParam, 0, len(messages))
	var systemMsg string
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemMsg = msg.Content
		case "user":
			anthroMsgs = append(anthroMsgs, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case "assistant":
			anthroMsgs = append(anthroMsgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content)))
		}
	}

	// Build and send
	req := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 4096,
		Messages:  anthroMsgs,
	}
	if systemMsg != "" {
		req.System = []anthropic.TextBlockParam{{Text: systemMsg}}
	}

	resp, err := p.client.Messages.New(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Extract text
	var content string
	if len(resp.Content) > 0 {
		if block := resp.Content[0].AsText(); block.Text != "" {
			content = block.Text
		}
	}

	return &llm.Response{
		Content: content,
		Usage: &llm.Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
	}, nil
}

// Stream is not yet supported for Anthropic
func (p *AnthropicProvider) Stream(ctx context.Context, messages []llm.Message) (io.ReadCloser, error) {
	return nil, fmt.Errorf("streaming not yet implemented for Anthropic provider")
}

// ChatWithTools sends a message with tool support and returns the response
func (p *AnthropicProvider) ChatWithTools(ctx context.Context, messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	// For now, Anthropic tools support is simplified - we'll use regular chat
	// and handle tool calls in a basic way
	return p.Chat(ctx, messages)
}
