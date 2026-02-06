package tools

import (
	"context"
	"encoding/json"
)

// Parameter describes a tool parameter
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "integer", "boolean", "array", "object"
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     any    `json:"default,omitempty"`
}

// Result represents the output of a tool execution
type Result struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"` // Structured data for further processing
}

// NewSuccessResult creates a successful result
func NewSuccessResult(output string) Result {
	return Result{
		Success: true,
		Output:  output,
	}
}

// NewErrorResult creates a failed result
func NewErrorResult(err string) Result {
	return Result{
		Success: false,
		Output:  "",
		Error:   err,
	}
}

// Tool defines the interface for all agent tools
type Tool interface {
	// Name returns the tool's unique identifier
	Name() string
	
	// Description returns a human-readable description of what the tool does
	Description() string
	
	// Parameters returns the list of parameters the tool accepts
	Parameters() []Parameter
	
	// Execute runs the tool with the provided arguments
	Execute(ctx context.Context, args map[string]any) (Result, error)
}

// ToolDefinition is used for LLM function calling schemas
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToDefinition converts a Tool to a ToolDefinition for LLM function calling
func ToDefinition(t Tool) ToolDefinition {
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for _, p := range t.Parameters() {
		properties[p.Name] = map[string]interface{}{
			"type":        p.Type,
			"description": p.Description,
		}
		if p.Required {
			required = append(required, p.Name)
		}
	}

	return ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": properties,
			"required":   required,
		},
	}
}

// ToJSON converts a ToolDefinition to JSON
func (td ToolDefinition) ToJSON() string {
	b, _ := json.Marshal(td)
	return string(b)
}
