package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"k8x/internal/llm"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider implements the llm.Provider interface for Claude
// It ensures no trailing whitespace on assistant messages to satisfy API requirements.
type AnthropicProvider struct {
	client  *anthropic.Client
	model   string
	apiKey  string
	baseURL string
}

// NewAnthropicProvider creates a new Anthropic provider with defaults.
func NewAnthropicProvider(apiKey, baseURL, model string) *AnthropicProvider {
	if model == "" {
		model = string(anthropic.ModelClaudeSonnet4_0)
	}
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "https://api.anthropic.com" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	// NewClient returns an anthropic.Client (value); take its address to match *anthropic.Client
	clientVal := anthropic.NewClient(opts...)

	return &AnthropicProvider{
		client:  &clientVal,
		model:   model,
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// IsConfigured returns true if the provider has the API key and client
func (p *AnthropicProvider) IsConfigured() bool {
	return p.apiKey != "" && p.client != nil
}

// Chat sends messages to Claude and returns the response
func (p *AnthropicProvider) Chat(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("anthropic provider not configured: missing API key or client")
	}

	var anthroMsgs []anthropic.MessageParam
	var systemPrompt string
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			anthroMsgs = append(anthroMsgs, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case "assistant":
			// Always trim trailing whitespace on assistant content
			content := strings.TrimRight(msg.Content, " \t\n\r")
			anthroMsgs = append(anthroMsgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(content)))
		}
	}

	req := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 4096,
		Messages:  anthroMsgs,
	}
	if systemPrompt != "" {
		req.System = []anthropic.TextBlockParam{{Text: systemPrompt}}
	}

	resp, err := p.client.Messages.New(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Extract content text
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

// ChatWithTools sends a message with tool support and returns the response
func (p *AnthropicProvider) ChatWithTools(ctx context.Context, messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("anthropic provider not configured: missing API key or client")
	}

	var anthroMsgs []anthropic.MessageParam
	var systemBlocks []anthropic.TextBlockParam
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemBlocks = append(systemBlocks, anthropic.TextBlockParam{Text: msg.Content})
		case "user":
			anthroMsgs = append(anthroMsgs, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		case "assistant":
			// Always trim trailing whitespace on assistant content
			content := strings.TrimRight(msg.Content, " \t\n\r")

			// If the assistant message has tool calls, create a message with both text and tool use blocks
			if len(msg.ToolCalls) > 0 {
				var blocks []anthropic.ContentBlockParamUnion

				// Add text content if present
				if content != "" {
					blocks = append(blocks, anthropic.NewTextBlock(content))
				}

				// Add tool use blocks
				for _, toolCall := range msg.ToolCalls {
					var input interface{}
					if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &input); err != nil {
						// If JSON unmarshal fails, create a simple input
						input = map[string]interface{}{"arguments": toolCall.Function.Arguments}
					}
					toolUse := anthropic.NewToolUseBlock(toolCall.ID, input, toolCall.Function.Name)
					blocks = append(blocks, toolUse)
				}

				anthroMsgs = append(anthroMsgs, anthropic.MessageParam{
					Role:    "assistant",
					Content: blocks,
				})
			} else {
				// Regular assistant message without tool calls
				anthroMsgs = append(anthroMsgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(content)))
			}
		case "tool":
			// Tool results are sent as user messages with tool result blocks
			toolResult := anthropic.NewToolResultBlock(msg.ToolCallID, msg.Content, false)
			anthroMsgs = append(anthroMsgs, anthropic.NewUserMessage(toolResult))
		}
	}

	// Build tool definitions
	var toolParams []anthropic.ToolParam
	for _, tl := range tools {
		props := make(map[string]interface{})
		for name, pinfo := range tl.Function.Parameters.Properties {
			schema := map[string]interface{}{"type": pinfo.Type}
			if pinfo.Description != "" {
				schema["description"] = pinfo.Description
			}
			if len(pinfo.Enum) > 0 {
				schema["enum"] = pinfo.Enum
			}
			props[name] = schema
		}
		inputSchema := anthropic.ToolInputSchemaParam{
			Properties: props,
			Required:   tl.Function.Parameters.Required,
			Type:       "object",
		}
		toolParams = append(toolParams, anthropic.ToolParam{
			Name:        tl.Function.Name,
			Description: anthropic.String(tl.Function.Description),
			InputSchema: inputSchema,
		})
	}

	// Wrap as union params
	var unions []anthropic.ToolUnionParam
	for i := range toolParams {
		unions = append(unions, anthropic.ToolUnionParam{OfTool: &toolParams[i]})
	}

	req := anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model),
		MaxTokens: 4096,
		Messages:  anthroMsgs,
	}
	if len(systemBlocks) > 0 {
		req.System = systemBlocks
	}
	if len(unions) > 0 {
		req.Tools = unions
	}

	resp, err := p.client.Messages.New(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create message with tools: %w", err)
	}

	// Build response
	var content string
	if len(resp.Content) > 0 {
		if block := resp.Content[0].AsText(); block.Text != "" {
			content = block.Text
		}
	}
	r := &llm.Response{
		Content: content,
		Usage: &llm.Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
	}

	// Extract tool calls
	for _, blk := range resp.Content {
		switch tblk := blk.AsAny().(type) {
		case anthropic.ToolUseBlock:
			inputJSON, _ := json.Marshal(tblk.Input)
			r.ToolCalls = append(r.ToolCalls, llm.ToolCall{
				ID:   tblk.ID,
				Type: "function",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{Name: tblk.Name, Arguments: string(inputJSON)},
			})
		}
	}

	return r, nil
}

// Stream is not yet supported for Anthropic
func (p *AnthropicProvider) Stream(ctx context.Context, messages []llm.Message) (io.ReadCloser, error) {
	return nil, fmt.Errorf("streaming not yet implemented for Anthropic provider")
}
