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
	client openai.Client
	model  string
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
	return &OpenAIProvider{client: client, model: model}
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
