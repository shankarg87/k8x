package providers

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/shankgan/k8x/internal/llm"
)

// OpenAIProvider implements the LLM Provider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	if model == "" {
		model = openai.GPT4
	}

	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)

	return &OpenAIProvider{
		client: client,
		model:  model,
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// IsConfigured returns true if the provider is properly configured
func (p *OpenAIProvider) IsConfigured() bool {
	return p.client != nil
}

// Chat sends a message and returns the response
func (p *OpenAIProvider) Chat(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("OpenAI provider not configured")
	}

	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	request := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: openaiMessages,
	}

	resp, err := p.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	return &llm.Response{
		Content: resp.Choices[0].Message.Content,
		Usage: &llm.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// Stream sends a message and returns a streaming response
func (p *OpenAIProvider) Stream(ctx context.Context, messages []llm.Message) (io.ReadCloser, error) {
	// TODO: Implement streaming support with OpenAI SDK
	return nil, fmt.Errorf("streaming not yet implemented for OpenAI provider")
}
