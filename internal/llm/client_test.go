package llm

import (
	"context"
	"io"
	"testing"
)

// MockProvider implements the Provider interface for testing
type MockProvider struct {
	name       string
	configured bool
	chatResp   *Response
	chatErr    error
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Chat(ctx context.Context, messages []Message) (*Response, error) {
	if m.chatErr != nil {
		return nil, m.chatErr
	}
	return m.chatResp, nil
}

func (m *MockProvider) Stream(ctx context.Context, messages []Message) (io.ReadCloser, error) {
	// Simple mock implementation
	return nil, nil
}

func (m *MockProvider) IsConfigured() bool {
	return m.configured
}

func TestClient_RegisterProvider(t *testing.T) {
	client := NewClient()
	provider := &MockProvider{name: "test", configured: true}

	client.RegisterProvider(provider)

	providers := client.ListProviders()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(providers))
	}

	if providers[0] != "test" {
		t.Errorf("Expected provider name 'test', got '%s'", providers[0])
	}
}

func TestClient_SetDefaultProvider(t *testing.T) {
	client := NewClient()
	provider := &MockProvider{name: "test", configured: true}

	client.RegisterProvider(provider)

	err := client.SetDefaultProvider("test")
	if err != nil {
		t.Errorf("SetDefaultProvider() failed: %v", err)
	}

	// Test setting non-existent provider
	err = client.SetDefaultProvider("nonexistent")
	if err == nil {
		t.Error("Expected error when setting non-existent provider")
	}
}

func TestClient_GetProvider(t *testing.T) {
	client := NewClient()
	provider := &MockProvider{name: "test", configured: true}

	client.RegisterProvider(provider)

	retrievedProvider, err := client.GetProvider("test")
	if err != nil {
		t.Errorf("GetProvider() failed: %v", err)
	}

	if retrievedProvider.Name() != "test" {
		t.Errorf("Expected provider name 'test', got '%s'", retrievedProvider.Name())
	}

	// Test getting non-existent provider
	_, err = client.GetProvider("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent provider")
	}
}

func TestClient_Chat(t *testing.T) {
	client := NewClient()
	expectedResp := &Response{Content: "test response"}
	provider := &MockProvider{
		name:       "test",
		configured: true,
		chatResp:   expectedResp,
	}

	client.RegisterProvider(provider)
	client.SetDefaultProvider("test")

	ctx := context.Background()
	messages := []Message{{Role: "user", Content: "test message"}}

	resp, err := client.Chat(ctx, messages)
	if err != nil {
		t.Errorf("Chat() failed: %v", err)
	}

	if resp.Content != expectedResp.Content {
		t.Errorf("Expected response content '%s', got '%s'", expectedResp.Content, resp.Content)
	}
}

func TestClient_ChatWithProvider(t *testing.T) {
	client := NewClient()
	expectedResp := &Response{Content: "test response"}
	provider := &MockProvider{
		name:       "test",
		configured: true,
		chatResp:   expectedResp,
	}

	client.RegisterProvider(provider)

	ctx := context.Background()
	messages := []Message{{Role: "user", Content: "test message"}}

	resp, err := client.ChatWithProvider(ctx, "test", messages)
	if err != nil {
		t.Errorf("ChatWithProvider() failed: %v", err)
	}

	if resp.Content != expectedResp.Content {
		t.Errorf("Expected response content '%s', got '%s'", expectedResp.Content, resp.Content)
	}
}
