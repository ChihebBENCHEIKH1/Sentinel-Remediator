package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/chiheb/sentinel-remediator/internal/config"
	"github.com/chiheb/sentinel-remediator/internal/tools"
	"go.uber.org/zap"
)

// LLMClient abstracts LLM interactions
type LLMClient interface {
	// Chat sends a message and returns the response
	Chat(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error)
}

// Message represents a conversation message
type Message struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

// ToolCall represents a requested tool invocation
type ToolCall struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Response represents the LLM response
type Response struct {
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	StopReason string     `json:"stop_reason"`
}

// AnthropicClient implements LLMClient using Claude via HTTP
type AnthropicClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewAnthropicClient creates a new Anthropic/Claude client
func NewAnthropicClient(cfg *config.Config, logger *zap.Logger) *AnthropicClient {
	return &AnthropicClient{
		apiKey:     cfg.AnthropicAPIKey,
		model:      cfg.LLMModel,
		httpClient: &http.Client{},
		logger:     logger,
	}
}

// Anthropic API types
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type anthropicResponse struct {
	ID         string               `json:"id"`
	Type       string               `json:"type"`
	Role       string               `json:"role"`
	Content    []anthropicContent   `json:"content"`
	StopReason string               `json:"stop_reason"`
}

type anthropicContent struct {
	Type  string            `json:"type"`
	Text  string            `json:"text,omitempty"`
	ID    string            `json:"id,omitempty"`
	Name  string            `json:"name,omitempty"`
	Input map[string]any    `json:"input,omitempty"`
}

func (c *AnthropicClient) Chat(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error) {
	// Convert messages to Anthropic format
	anthropicMsgs := make([]anthropicMessage, 0, len(messages))
	systemPrompt := ""

	for _, msg := range messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
			continue
		}
		anthropicMsgs = append(anthropicMsgs, anthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Convert tool definitions to Anthropic format
	anthropicTools := make([]anthropicTool, 0, len(toolDefs))
	for _, td := range toolDefs {
		anthropicTools = append(anthropicTools, anthropicTool{
			Name:        td.Name,
			Description: td.Description,
			InputSchema: td.Parameters,
		})
	}

	// Build request
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: 4096,
		System:    systemPrompt,
		Messages:  anthropicMsgs,
	}

	if len(anthropicTools) > 0 {
		reqBody.Tools = anthropicTools
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	c.logger.Debug("Sending request to Claude",
		zap.String("model", c.model),
		zap.Int("messages", len(anthropicMsgs)),
		zap.Int("tools", len(anthropicTools)),
	)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Make the API call
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to our response format
	response := &Response{
		StopReason: anthropicResp.StopReason,
		ToolCalls:  make([]ToolCall, 0),
	}

	for _, content := range anthropicResp.Content {
		switch content.Type {
		case "text":
			response.Content = content.Text
		case "tool_use":
			response.ToolCalls = append(response.ToolCalls, ToolCall{
				ID:        content.ID,
				Name:      content.Name,
				Arguments: content.Input,
			})
		}
	}

	c.logger.Debug("Received response from Claude",
		zap.String("stop_reason", response.StopReason),
		zap.Int("tool_calls", len(response.ToolCalls)),
		zap.Int("content_length", len(response.Content)),
	)

	return response, nil
}
