package tools

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestNewRegistry(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	if registry == nil {
		t.Fatal("Registry should not be nil")
	}

	if registry.Count() != 0 {
		t.Errorf("Expected 0 tools, got %d", registry.Count())
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	// Register a mock tool
	mockTool := &MockTool{name: "test_tool"}
	registry.Register(mockTool)

	if registry.Count() != 1 {
		t.Errorf("Expected 1 tool, got %d", registry.Count())
	}

	// Get the tool
	tool, ok := registry.Get("test_tool")
	if !ok {
		t.Fatal("Tool should exist")
	}

	if tool.Name() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", tool.Name())
	}
}

func TestRegistryExecute(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	mockTool := &MockTool{name: "echo", result: Result{Success: true, Output: "hello"}}
	registry.Register(mockTool)

	result, err := registry.Execute(context.Background(), "echo", map[string]any{"message": "hello"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Error("Result should be successful")
	}

	if result.Output != "hello" {
		t.Errorf("Expected output 'hello', got '%s'", result.Output)
	}
}

func TestRegistryExecuteNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	result, _ := registry.Execute(context.Background(), "nonexistent", nil)

	if result.Success {
		t.Error("Result should not be successful for non-existent tool")
	}

	if result.Error == "" {
		t.Error("Error message should not be empty")
	}
}

func TestGetDefinitions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	registry := NewRegistry(logger)

	registry.Register(&MockTool{name: "tool1"})
	registry.Register(&MockTool{name: "tool2"})

	definitions := registry.GetDefinitions()

	if len(definitions) != 2 {
		t.Errorf("Expected 2 definitions, got %d", len(definitions))
	}
}

// MockTool for testing
type MockTool struct {
	name   string
	result Result
}

func (m *MockTool) Name() string {
	return m.name
}

func (m *MockTool) Description() string {
	return "Mock tool for testing"
}

func (m *MockTool) Parameters() []Parameter {
	return []Parameter{
		{Name: "message", Type: "string", Description: "Test message", Required: true},
	}
}

func (m *MockTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	if m.result.Output != "" {
		return m.result, nil
	}
	return NewSuccessResult("mock result"), nil
}
