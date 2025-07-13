package providers

import (
	"context"
	"fmt"
	"io"

	"k8x/internal/llm"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
)

// OpenAIProvider implements the llm.Provider interface for OpenAI
type OpenAIProvider struct {
	client        openai.Client
	model         string
	contextLength int
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	// Default to o3-mini if no model is specified
	if model == "" {
		model = "o3-mini"
	}

	opts := []option.RequestOption{}
	if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := openai.NewClient(opts...)
	
	// Get default context length for the model
	contextLength := getDefaultContextLengthForOpenAI(model)
	
	return &OpenAIProvider{
		client:        client,
		model:         model,
		contextLength: contextLength,
	}
}

// getDefaultContextLengthForOpenAI returns default context lengths for OpenAI models
func getDefaultContextLengthForOpenAI(model string) int {
	defaults := map[string]int{
		"gpt-4":           8192,
		"gpt-4-turbo":     128000,
		"gpt-4o":          128000,
		"gpt-4o-mini":     128000,
		"o1":              200000,
		"o1-mini":         128000,
		"o3-mini":         128000,
		"gpt-3.5-turbo":   16385,
	}
	
	if contextLength, exists := defaults[model]; exists {
		return contextLength
	}
	
	// Default fallback
	return 8192
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// IsConfigured returns true if the provider has a valid model
func (p *OpenAIProvider) IsConfigured() bool {
	return p.model != ""
}

// Chat sends a message and returns the response
func (p *OpenAIProvider) Chat(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("OpenAI provider not configured")
	}

	// Convert messages
	openaiMsgs := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "user":
			openaiMsgs[i] = openai.UserMessage(msg.Content)
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				toolCalls := make([]openai.ChatCompletionMessageToolCallParam, len(msg.ToolCalls))
				for j, tc := range msg.ToolCalls {
					toolCalls[j] = openai.ChatCompletionMessageToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				}
				// Create assistant message with tool calls manually
				openaiMsgs[i] = openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						Content: openai.ChatCompletionAssistantMessageParamContentUnion{
							OfString: openai.String(msg.Content),
						},
						ToolCalls: toolCalls,
					},
				}
			} else {
				openaiMsgs[i] = openai.AssistantMessage(msg.Content)
			}
		case "system":
			openaiMsgs[i] = openai.SystemMessage(msg.Content)
		case "tool":
			// Tool response message - use the exact tool call ID as provided
			openaiMsgs[i] = openai.ToolMessage(msg.Content, msg.ToolCallID)
		}
	}

	resp, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(p.model),
		Messages: openaiMsgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	response := &llm.Response{
		Content: resp.Choices[0].Message.Content,
		Usage: &llm.Usage{
			PromptTokens:     int(resp.Usage.PromptTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		},
	}

	// Handle tool calls if present
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		toolCalls := make([]llm.ToolCall, len(resp.Choices[0].Message.ToolCalls))
		for i, tc := range resp.Choices[0].Message.ToolCalls {
			toolCalls[i] = llm.ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
		response.ToolCalls = toolCalls
	}

	return response, nil
}

// ChatWithTools sends a message with tool support and returns the response
func (p *OpenAIProvider) ChatWithTools(ctx context.Context, messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("OpenAI provider not configured")
	}

	// Convert messages
	openaiMsgs := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "user":
			openaiMsgs[i] = openai.UserMessage(msg.Content)
		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// Assistant message with tool calls
				toolCalls := make([]openai.ChatCompletionMessageToolCallParam, len(msg.ToolCalls))
				for j, tc := range msg.ToolCalls {
					toolCalls[j] = openai.ChatCompletionMessageToolCallParam{
						ID: tc.ID,
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Function.Name,
							Arguments: tc.Function.Arguments,
						},
					}
				}
				// Create assistant message with tool calls manually
				openaiMsgs[i] = openai.ChatCompletionMessageParamUnion{
					OfAssistant: &openai.ChatCompletionAssistantMessageParam{
						Content: openai.ChatCompletionAssistantMessageParamContentUnion{
							OfString: openai.String(msg.Content),
						},
						ToolCalls: toolCalls,
					},
				}
			} else {
				openaiMsgs[i] = openai.AssistantMessage(msg.Content)
			}
		case "system":
			openaiMsgs[i] = openai.SystemMessage(msg.Content)
		case "tool":
			// Tool response message - use the exact tool call ID as provided
			openaiMsgs[i] = openai.ToolMessage(msg.Content, msg.ToolCallID)
		}
	}

	// Convert tools
	openaiTools := make([]openai.ChatCompletionToolParam, len(tools))
	for i, tool := range tools {
		properties := make(map[string]interface{})
		for name, prop := range tool.Function.Parameters.Properties {
			propMap := map[string]interface{}{
				"type":        prop.Type,
				"description": prop.Description,
			}
			if len(prop.Enum) > 0 {
				propMap["enum"] = prop.Enum
			}
			properties[name] = propMap
		}

		openaiTools[i] = openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Function.Name,
				Description: openai.String(tool.Function.Description),
				Parameters: openai.FunctionParameters{
					"type":       "object",
					"properties": properties,
					"required":   tool.Function.Parameters.Required,
				},
			},
		}
	}

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(p.model),
		Messages: openaiMsgs,
		Tools:    openaiTools,
	}

	resp, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	response := &llm.Response{
		Content: resp.Choices[0].Message.Content,
		Usage: &llm.Usage{
			PromptTokens:     int(resp.Usage.PromptTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		},
	}

	// Handle tool calls if present
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		toolCalls := make([]llm.ToolCall, len(resp.Choices[0].Message.ToolCalls))
		for i, tc := range resp.Choices[0].Message.ToolCalls {
			toolCalls[i] = llm.ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
		response.ToolCalls = toolCalls
	}

	return response, nil
}

// Stream returns an error until streaming is implemented
func (p *OpenAIProvider) Stream(ctx context.Context, messages []llm.Message) (io.ReadCloser, error) {
	return nil, fmt.Errorf("streaming not yet implemented for OpenAI provider")
}

// EstimateTokens provides a rough estimation of token count for messages
func (p *OpenAIProvider) EstimateTokens(messages []llm.Message) int {
	totalTokens := 0
	
	for _, msg := range messages {
		// Rough estimation: 1 token ≈ 4 characters for English text
		// Add some overhead for message structure
		contentTokens := len(msg.Content) / 4
		totalTokens += contentTokens + 10 // +10 for message overhead
		
		// Add tokens for tool calls if present
		if len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				totalTokens += len(tc.Function.Name)/4 + len(tc.Function.Arguments)/4 + 20
			}
		}
	}
	
	// Add some overhead for the overall request structure
	return totalTokens + 50
}

// GetContextLength returns the context window size for this provider
func (p *OpenAIProvider) GetContextLength() int {
	return p.contextLength
}

// SetContextLength allows overriding the context length from configuration
func (p *OpenAIProvider) SetContextLength(length int) {
	if length > 0 {
		p.contextLength = length
	}
}
