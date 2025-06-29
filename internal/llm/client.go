package llm

import (
	"context"
	"fmt"
	"io"
)

// Provider represents an LLM provider interface
type Provider interface {
	// Name returns the provider name
	Name() string

	// Chat sends a message and returns the response
	Chat(ctx context.Context, messages []Message) (*Response, error)

	// Stream sends a message and returns a streaming response
	Stream(ctx context.Context, messages []Message) (io.ReadCloser, error)

	// IsConfigured returns true if the provider is properly configured
	IsConfigured() bool
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// Response represents an LLM response
type Response struct {
	Content string         `json:"content"`
	Usage   *Usage         `json:"usage,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Client manages multiple LLM providers
type Client struct {
	providers       map[string]Provider
	defaultProvider string
}

// NewClient creates a new LLM client
func NewClient() *Client {
	return &Client{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers an LLM provider
func (c *Client) RegisterProvider(provider Provider) {
	c.providers[provider.Name()] = provider
}

// SetDefaultProvider sets the default provider
func (c *Client) SetDefaultProvider(name string) error {
	if _, exists := c.providers[name]; !exists {
		return fmt.Errorf("provider %s not found", name)
	}
	c.defaultProvider = name
	return nil
}

// GetProvider returns a provider by name
func (c *Client) GetProvider(name string) (Provider, error) {
	provider, exists := c.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return provider, nil
}

// GetDefaultProvider returns the default provider
func (c *Client) GetDefaultProvider() (Provider, error) {
	if c.defaultProvider == "" {
		return nil, fmt.Errorf("no default provider set")
	}
	return c.GetProvider(c.defaultProvider)
}

// Chat sends a message using the default provider
func (c *Client) Chat(ctx context.Context, messages []Message) (*Response, error) {
	provider, err := c.GetDefaultProvider()
	if err != nil {
		return nil, err
	}
	return provider.Chat(ctx, messages)
}

// ChatWithProvider sends a message using a specific provider
func (c *Client) ChatWithProvider(ctx context.Context, providerName string, messages []Message) (*Response, error) {
	provider, err := c.GetProvider(providerName)
	if err != nil {
		return nil, err
	}
	return provider.Chat(ctx, messages)
}

// ListProviders returns a list of available providers
func (c *Client) ListProviders() []string {
	var names []string
	for name := range c.providers {
		names = append(names, name)
	}
	return names
}
