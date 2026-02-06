package tools

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Registry holds all available tools and provides lookup functionality
type Registry struct {
	mu     sync.RWMutex
	tools  map[string]Tool
	logger *zap.Logger
}

// NewRegistry creates a new tool registry
func NewRegistry(logger *zap.Logger) *Registry {
	return &Registry{
		tools:  make(map[string]Tool),
		logger: logger,
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
	r.logger.Debug("Registered tool", zap.String("name", tool.Name()))
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// Execute runs a tool by name with the provided arguments
func (r *Registry) Execute(ctx context.Context, name string, args map[string]any) (Result, error) {
	tool, ok := r.Get(name)
	if !ok {
		return NewErrorResult(fmt.Sprintf("tool not found: %s", name)), nil
	}

	r.logger.Info("Executing tool",
		zap.String("name", name),
		zap.Any("args", args),
	)

	result, err := tool.Execute(ctx, args)
	if err != nil {
		r.logger.Error("Tool execution failed",
			zap.String("name", name),
			zap.Error(err),
		)
		return NewErrorResult(err.Error()), err
	}

	r.logger.Info("Tool execution completed",
		zap.String("name", name),
		zap.Bool("success", result.Success),
	)

	return result, nil
}

// List returns all registered tool names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GetDefinitions returns tool definitions for LLM function calling
func (r *Registry) GetDefinitions() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definitions := make([]ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		definitions = append(definitions, ToDefinition(tool))
	}
	return definitions
}

// Count returns the number of registered tools
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}
