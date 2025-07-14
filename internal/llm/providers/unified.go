package providers

import (
	"context"
	"fmt"
	"io"

	"k8x/internal/llm"
)

// Credentials holds API credentials and selected provider configuration.
type Credentials struct {
	// SelectedProvider indicates which LLM provider to use ("openai", "anthropic", or "google").
	SelectedProvider string `yaml:"selected_provider,omitempty"`

	OpenAI struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"openai"`

	Anthropic struct {
		APIKey string `yaml:"api_key"`
	} `yaml:"anthropic"`

	Google struct {
		APIKey                 string `yaml:"api_key"`
		ApplicationCredentials string `yaml:"application_credentials"`
	} `yaml:"google"`
}

// UnifiedProvider wraps a concrete llm.Provider (OpenAI or Anthropic) behind one interface.
type UnifiedProvider struct {
	provider llm.Provider
}

// NewUnifiedProvider instantiates a UnifiedProvider based on creds.SelectedProvider.
// It returns an error if the provider is unsupported or not configured.
func NewUnifiedProvider(creds Credentials) (*UnifiedProvider, error) {
	var provider llm.Provider
	var err error

	switch creds.SelectedProvider {
	case "openai":
		provider = NewOpenAIProvider(creds.OpenAI.APIKey, "", "")
	case "anthropic":
		provider = NewAnthropicProvider(creds.Anthropic.APIKey, "", "")
	case "google":
		provider, err = NewGoogleProvider(creds.Google.APIKey, creds.Google.ApplicationCredentials, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create Google provider: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %s", creds.SelectedProvider)
	}

	if !provider.IsConfigured() {
		return nil, fmt.Errorf("%s provider not configured", provider.Name())
	}

	return &UnifiedProvider{provider: provider}, nil
}

// Name returns the active provider's name.
func (u *UnifiedProvider) Name() string {
	return u.provider.Name()
}

// IsConfigured returns true if the underlying provider is properly configured.
func (u *UnifiedProvider) IsConfigured() bool {
	return u.provider.IsConfigured()
}

// Chat sends messages to the selected LLM provider and returns its response.
func (u *UnifiedProvider) Chat(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	if !u.IsConfigured() {
		return nil, fmt.Errorf("%s provider not configured", u.Name())
	}
	return u.provider.Chat(ctx, messages)
}

// Stream starts a streaming chat session with the selected provider.
func (u *UnifiedProvider) Stream(ctx context.Context, messages []llm.Message) (io.ReadCloser, error) {
	if !u.IsConfigured() {
		return nil, fmt.Errorf("%s provider not configured", u.Name())
	}
	return u.provider.Stream(ctx, messages)
}

// ChatWithTools sends messages with tool support to the selected LLM provider and returns its response.
func (u *UnifiedProvider) ChatWithTools(ctx context.Context, messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	if !u.IsConfigured() {
		return nil, fmt.Errorf("%s provider not configured", u.Name())
	}

	// Check if the provider supports tools
	switch p := u.provider.(type) {
	case *OpenAIProvider:
		return p.ChatWithTools(ctx, messages, tools)
	case *AnthropicProvider:
		return p.ChatWithTools(ctx, messages, tools)
	case *GoogleProvider:
		return p.ChatWithTools(ctx, messages, tools)
	default:
		// Fallback to regular chat if tools not supported
		return u.provider.Chat(ctx, messages)
	}
}
