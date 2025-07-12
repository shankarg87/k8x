package llm

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestIsContextWindowError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "context_length_exceeded",
			err:      fmt.Errorf("context_length_exceeded: Request too large"),
			expected: true,
		},
		{
			name:     "token limit exceeded",
			err:      fmt.Errorf("too many tokens in request"),
			expected: true,
		},
		{
			name:     "context window error",
			err:      fmt.Errorf("context window size exceeded"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      fmt.Errorf("network connection failed"),
			expected: false,
		},
		{
			name:     "case insensitive match",
			err:      fmt.Errorf("CONTEXT_LENGTH_EXCEEDED"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsContextWindowError(tt.err)
			if result != tt.expected {
				t.Errorf("IsContextWindowError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSummarizeConversation(t *testing.T) {
	summarizer := NewSummarizer(SummarizerConfig{
		SummarizeAtPercent: 70,
		KeepConversations:  1,
	})
	
	// Mock provider that returns a simple summary
	mockProvider := &MockProvider{
		name:       "mock",
		configured: true,
		chatResp: &Response{
			Content: "This is a test summary of the conversation.",
		},
		contextLength: 4096,
	}
	
	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant"},
		{Role: "user", Content: "Help me with Kubernetes"},
		{Role: "assistant", Content: "I'll help you with Kubernetes tasks"},
		{Role: "user", Content: "List all pods"},
		{Role: "assistant", Content: "I'll list the pods for you"},
		{Role: "tool", Content: "pod1, pod2, pod3"},
		{Role: "assistant", Content: "Here are your pods"},
		{Role: "user", Content: "What's the status of pod1?"},
	}
	
	summarized, err := summarizer.SummarizeConversation(context.Background(), mockProvider, messages)
	if err != nil {
		t.Fatalf("SummarizeConversation failed: %v", err)
	}
	
	// Should have: system + goal + summary + 2 recent messages (1 conversation = 2 messages)
	expectedLength := 5 // system, goal, summary, 2 recent
	if len(summarized) != expectedLength {
		t.Errorf("Expected %d messages, got %d", expectedLength, len(summarized))
	}
	
	// Check that system message is preserved
	if summarized[0].Role != "system" {
		t.Error("System message should be first")
	}
	
	// Check that goal is preserved
	if summarized[1].Role != "user" || summarized[1].Content != "Help me with Kubernetes" {
		t.Error("Goal message should be preserved as second message")
	}
	
	// Check that summary is included
	if summarized[2].Role != "assistant" || !strings.Contains(summarized[2].Content, "Previous conversation summary") {
		t.Error("Summary should be included as third message")
	}
}

func TestSummarizeConversationShortHistory(t *testing.T) {
	summarizer := NewSummarizer(SummarizerConfig{
		SummarizeAtPercent: 70,
		KeepConversations:  1,
	})
	mockProvider := &MockProvider{
		name:          "mock",
		configured:    true,
		contextLength: 4096,
	}
	
	// Short conversation that doesn't need summarization
	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant"},
		{Role: "user", Content: "Help me"},
		{Role: "assistant", Content: "Sure, I'll help"},
	}
	
	summarized, err := summarizer.SummarizeConversation(context.Background(), mockProvider, messages)
	if err != nil {
		t.Fatalf("SummarizeConversation failed: %v", err)
	}
	
	// Should return original messages unchanged
	if len(summarized) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(summarized))
	}
	
	for i, msg := range messages {
		if summarized[i].Content != msg.Content {
			t.Errorf("Message %d content mismatch", i)
		}
	}
}

func TestShouldSummarize(t *testing.T) {
	summarizer := NewSummarizer(SummarizerConfig{
		SummarizeAtPercent: 70,
		KeepConversations:  1,
	})
	
	mockProvider := &MockProvider{
		name:          "mock",
		configured:    true,
		contextLength: 1000, // Small context for easy testing
	}
	
	// Small message that shouldn't trigger summarization
	smallMessages := []Message{
		{Role: "user", Content: "Hi"}, // ~1 token
	}
	
	// Large message that should trigger summarization (>70% of 1000 = 700 tokens)
	// Each "word " is 5 chars, so 800 * 5 = 4000 chars, which is 1000+ tokens
	largeMessages := []Message{
		{Role: "user", Content: strings.Repeat("word ", 800)}, // ~1000+ tokens
	}
	
	if summarizer.ShouldSummarize(mockProvider, smallMessages) {
		t.Error("Should not summarize small messages")
	}
	
	if !summarizer.ShouldSummarize(mockProvider, largeMessages) {
		t.Error("Should summarize large messages")
	}
}

