package llm

import (
	"context"
	"fmt"
	"strings"
)

// Summarizer provides functionality to summarize conversation history
// when the context window is exceeded.
type Summarizer struct {
	config SummarizerConfig
}

// SummarizerConfig contains configuration for the summarizer
type SummarizerConfig struct {
	// SummarizeAtPercent is the percentage of context length at which to trigger summarization (default 70%)
	SummarizeAtPercent int
	// KeepConversations is the number of recent conversation exchanges to keep intact (default 1)
	KeepConversations int
}

// NewSummarizer creates a new conversation summarizer with the given configuration
func NewSummarizer(config SummarizerConfig) *Summarizer {
	// Set defaults if not provided
	if config.SummarizeAtPercent <= 0 {
		config.SummarizeAtPercent = 70
	}
	if config.KeepConversations <= 0 {
		config.KeepConversations = 1
	}
	
	return &Summarizer{config: config}
}

// SummarizeConversation condenses a long conversation while preserving
// the system prompt, goal, and recent context based on configuration.
func (s *Summarizer) SummarizeConversation(ctx context.Context, provider Provider, messages []Message) ([]Message, error) {
	// Calculate how many recent messages to keep based on conversation count
	// Each conversation typically consists of 2 messages (user + assistant)
	keepRecent := s.config.KeepConversations * 2
	
	if len(messages) <= keepRecent+2 { // system + goal + recent messages
		return messages, nil
	}

	var systemMsg, goalMsg Message
	var recentMsgs []Message
	var middleMsgs []Message

	// Extract system message (first message)
	if len(messages) > 0 && messages[0].Role == "system" {
		systemMsg = messages[0]
		messages = messages[1:]
	}

	// Extract goal message (likely the first user message)
	if len(messages) > 0 && messages[0].Role == "user" {
		goalMsg = messages[0]
		messages = messages[1:]
	}

	// Extract recent messages
	if len(messages) > keepRecent {
		recentMsgs = messages[len(messages)-keepRecent:]
		middleMsgs = messages[:len(messages)-keepRecent]
	} else {
		recentMsgs = messages
	}

	// If there are no middle messages to summarize, return as is
	if len(middleMsgs) == 0 {
		result := []Message{}
		if systemMsg.Role != "" {
			result = append(result, systemMsg)
		}
		if goalMsg.Role != "" {
			result = append(result, goalMsg)
		}
		result = append(result, recentMsgs...)
		return result, nil
	}

	// Create summarization prompt
	summaryPrompt := s.buildSummaryPrompt(middleMsgs)
	
	// Request summary from the LLM
	summaryMessages := []Message{
		{Role: "system", Content: "You are an expert at summarizing technical conversations. Provide concise, accurate summaries that preserve key information."},
		{Role: "user", Content: summaryPrompt},
	}

	response, err := provider.Chat(ctx, summaryMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	// Build the condensed conversation
	result := []Message{}
	if systemMsg.Role != "" {
		result = append(result, systemMsg)
	}
	if goalMsg.Role != "" {
		result = append(result, goalMsg)
	}
	
	// Add the summary as an assistant message
	result = append(result, Message{
		Role:    "assistant",
		Content: fmt.Sprintf("**Previous conversation summary:**\n%s", response.Content),
	})
	
	// Add recent messages
	result = append(result, recentMsgs...)

	return result, nil
}

// buildSummaryPrompt creates a prompt for summarizing the middle conversation
func (s *Summarizer) buildSummaryPrompt(messages []Message) string {
	var conversation strings.Builder
	conversation.WriteString("Please summarize the following technical conversation concisely. ")
	conversation.WriteString("Focus on:\n")
	conversation.WriteString("1. Commands executed and their purposes\n")
	conversation.WriteString("2. Key findings or issues discovered\n")
	conversation.WriteString("3. Important progress made toward the goal\n")
	conversation.WriteString("4. Any errors or challenges encountered\n\n")
	conversation.WriteString("Conversation to summarize:\n\n")

	for i, msg := range messages {
		switch msg.Role {
		case "user":
			conversation.WriteString(fmt.Sprintf("Human: %s\n\n", msg.Content))
		case "assistant":
			conversation.WriteString(fmt.Sprintf("Assistant: %s\n\n", msg.Content))
		case "tool":
			conversation.WriteString(fmt.Sprintf("Tool output: %s\n\n", msg.Content))
		}
		
		// Prevent the summary prompt itself from being too long
		if i >= 20 {
			conversation.WriteString("... (additional messages truncated for brevity)\n\n")
			break
		}
	}

	conversation.WriteString("Please provide a clear, concise summary in 2-3 paragraphs.")
	return conversation.String()
}

// IsContextWindowError checks if an error indicates that the context window was exceeded
func IsContextWindowError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	
	// Check for common context window error patterns
	contextErrorPatterns := []string{
		"context_length_exceeded",
		"context window",
		"context length",
		"too many tokens",
		"token limit",
		"maximum context",
		"exceeds the limit",
		"context size",
		"request too large",
		"prompt too long",
	}
	
	for _, pattern := range contextErrorPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	
	return false
}

// ShouldSummarize checks if the conversation should be summarized based on token usage
func (s *Summarizer) ShouldSummarize(provider Provider, messages []Message) bool {
	contextLength := provider.GetContextLength()
	currentTokens := provider.EstimateTokens(messages)
	
	threshold := (contextLength * s.config.SummarizeAtPercent) / 100
	return currentTokens >= threshold
}