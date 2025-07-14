package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"k8x/internal/llm"

	"google.golang.org/genai"
)

// GoogleProvider implements llm.Provider via Google's GenAI SDK.
type GoogleProvider struct {
	client *genai.Client
	model  string
}

func NewGoogleProvider(apiKey, applicationCredentials, model string) (*GoogleProvider, error) {
	ctx := context.Background()

	var backend genai.Backend

	// Determine which backend to use based on authentication method
	if apiKey != "" {
		// Use Gemini API with API key
		if err := os.Setenv("GEMINI_API_KEY", apiKey); err != nil {
			return nil, fmt.Errorf("failed to set GEMINI_API_KEY: %w", err)
		}
		backend = genai.BackendGeminiAPI
	} else if applicationCredentials != "" {
		// Use Vertex AI with service account credentials
		if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", applicationCredentials); err != nil {
			return nil, fmt.Errorf("failed to set GOOGLE_APPLICATION_CREDENTIALS: %w", err)
		}
		backend = genai.BackendVertexAI
	} else {
		// Try to use existing environment variables
		if os.Getenv("GEMINI_API_KEY") != "" {
			backend = genai.BackendGeminiAPI
		} else if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") != "" {
			backend = genai.BackendVertexAI
		} else {
			return nil, fmt.Errorf("no authentication method provided: either set apiKey for Gemini API or applicationCredentials for Vertex AI")
		}
	}

	cc := &genai.ClientConfig{
		Backend: backend,
	}
	client, err := genai.NewClient(ctx, cc)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	if model == "" {
		model = "gemini-2.5-flash"
	}

	return &GoogleProvider{client: client, model: model}, nil
}

// Name returns the provider identifier.
func (p *GoogleProvider) Name() string {
	return "google"
}

// IsConfigured returns true if the underlying GenAI client exists.
func (p *GoogleProvider) IsConfigured() bool {
	return p.client != nil
}

// Chat sends a simple chat (no tools) to the model.
func (p *GoogleProvider) Chat(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("google provider not configured")
	}

	// Build one content sequence from all messages
	parts := make([]*genai.Part, 0, len(messages))
	for _, msg := range messages {
		parts = append(parts, genai.NewPartFromText(msg.Content))
	}
	resp, err := p.client.Models.GenerateContent(
		ctx,
		p.model,
		[]*genai.Content{{Parts: parts}},
		nil, // no special config
	)
	if err != nil {
		return nil, err
	}

	content := resp.Text()

	return &llm.Response{
		Content: content,
	}, nil
}

// ChatWithTools sends chat with tool and function-calling support.
func (p *GoogleProvider) ChatWithTools(
	ctx context.Context,
	messages []llm.Message,
	tools []llm.Tool,
) (*llm.Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("google provider not configured")
	}

	// 1) Build a mapping from tool call ID to function name
	toolCallIDToName := make(map[string]string)
	for _, msg := range messages {
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			for _, toolCall := range msg.ToolCalls {
				toolCallIDToName[toolCall.ID] = toolCall.Function.Name
			}
		}
	}

	// 2) Convert messages → genai.Parts
	parts := make([]*genai.Part, 0, len(messages))
	for _, msg := range messages {
		switch msg.Role {
		case "system", "user", "assistant":
			parts = append(parts, genai.NewPartFromText(msg.Content))
		case "tool":
			// If it's a prior tool response, inject it as function response
			// Google GenAI expects the function name, not the tool call ID
			functionName := toolCallIDToName[msg.ToolCallID]
			if functionName == "" {
				// Fallback: skip this tool message if we can't find the function name
				continue
			}
			parts = append(parts, genai.NewPartFromFunctionResponse(
				functionName,
				map[string]any{"data": msg.Content},
			))
		}
	}

	// 3) Build GenerateContentConfig with tools
	var genaiTools []*genai.Tool
	for _, tool := range tools {
		// For Google GenAI, we use the ParametersJsonSchema approach
		// Convert our tool schema to a JSON schema format
		properties := make(map[string]interface{})
		for name, prop := range tool.Function.Parameters.Properties {
			schema := map[string]interface{}{
				"type":        prop.Type,
				"description": prop.Description,
			}
			if len(prop.Enum) > 0 {
				schema["enum"] = prop.Enum
			}
			properties[name] = schema
		}

		// Create the JSON schema for the function parameters
		jsonSchema := map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   tool.Function.Parameters.Required,
		}

		funcDecl := &genai.FunctionDeclaration{
			Name:                 tool.Function.Name,
			Description:          tool.Function.Description,
			ParametersJsonSchema: jsonSchema,
		}

		genaiTool := &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{funcDecl},
		}
		genaiTools = append(genaiTools, genaiTool)
	}

	mode := genai.FunctionCallingConfigModeAuto
	fcConfig := &genai.FunctionCallingConfig{
		Mode: mode,
	}
	// Only set AllowedFunctionNames if mode is ANY
	if mode == genai.FunctionCallingConfigModeAny {
		names := make([]string, len(tools))
		for i, t := range tools {
			names[i] = t.Function.Name
		}
		fcConfig.AllowedFunctionNames = names
	}
	config := &genai.GenerateContentConfig{
		Tools: genaiTools,
		ToolConfig: &genai.ToolConfig{
			FunctionCallingConfig: fcConfig,
		},
	}

	// 4) Call the model
	resp, err := p.client.Models.GenerateContent(
		ctx,
		p.model,
		[]*genai.Content{{Parts: parts}},
		config,
	)
	if err != nil {
		return nil, err
	}

	// 5) Extract the first text response and any function calls
	out := &llm.Response{}

	if resp != nil && len(resp.Candidates) > 0 && resp.Candidates[0] != nil &&
		resp.Candidates[0].Content != nil && len(resp.Candidates[0].Content.Parts) > 0 && resp.Candidates[0].Content.Parts[0] != nil {
		// Extract text from the first part
		if part := resp.Candidates[0].Content.Parts[0]; part.Text != "" {
			out.Content = part.Text
		}
	}
	for _, fc := range resp.FunctionCalls() {
		// Convert arguments to JSON string for consistency with other providers
		argsJSON, err := json.Marshal(fc.Args)
		if err != nil {
			// Fallback to string representation if JSON marshaling fails
			argsJSON = []byte(fmt.Sprintf("%v", fc.Args))
		}

		toolCall := llm.ToolCall{
			ID:   fc.ID,
			Type: "function",
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      fc.Name,
				Arguments: string(argsJSON),
			},
		}
		out.ToolCalls = append(out.ToolCalls, toolCall)
	}

	return out, nil
}

// Stream is not yet supported for GenAI.
func (p *GoogleProvider) Stream(ctx context.Context, messages []llm.Message) (io.ReadCloser, error) {
	return nil, fmt.Errorf("streaming not yet implemented for Google GenAI provider")
}
