package agent

import (
	"context"
	"testing"

	"github.com/chiheb/sentinel-remediator/internal/config"
	"github.com/chiheb/sentinel-remediator/internal/domain"
	"github.com/chiheb/sentinel-remediator/internal/tools"
	"go.uber.org/zap"
)

func TestNewAgent(t *testing.T) {
	cfg := &config.Config{
		AnthropicAPIKey: "test-key",
		LLMModel:        "claude-3-5-sonnet-20241022",
		MaxIterations:   10,
		WorkDir:         "/tmp/test",
	}
	logger, _ := zap.NewDevelopment()
	registry := tools.NewRegistry(logger)

	agent, err := NewAgent(cfg, registry, logger)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	if agent == nil {
		t.Fatal("Agent should not be nil")
	}

	if agent.maxIterations != 10 {
		t.Errorf("Expected maxIterations=10, got %d", agent.maxIterations)
	}
}

func TestBuildVulnerabilityPrompt(t *testing.T) {
	cfg := &config.Config{
		AnthropicAPIKey: "test-key",
		LLMModel:        "claude-3-5-sonnet-20241022",
		MaxIterations:   10,
	}
	logger, _ := zap.NewDevelopment()
	registry := tools.NewRegistry(logger)
	agent, _ := NewAgent(cfg, registry, logger)

	job := &domain.RemediationJob{
		ScanResult: &domain.ScanResult{
			RepoURL: "https://github.com/test/repo",
			Branch:  "main",
		},
		WorkDir: "/tmp/test",
	}

	vuln := &domain.Vulnerability{
		ID:          "VULN-001",
		Type:        domain.VulnTypeRunAsRoot,
		Severity:    domain.SeverityHigh,
		Title:       "Container runs as root",
		Description: "Test description",
		FilePath:    "Dockerfile",
		LineNumber:  1,
		Suggestion:  "Add USER directive",
	}

	prompt := agent.buildVulnerabilityPrompt(job, vuln)

	if prompt == "" {
		t.Fatal("Prompt should not be empty")
	}

	// Check that key information is in the prompt
	if !contains(prompt, "VULN-001") {
		t.Error("Prompt should contain vulnerability ID")
	}
	if !contains(prompt, "RUN_AS_ROOT") {
		t.Error("Prompt should contain vulnerability type")
	}
	if !contains(prompt, "Dockerfile") {
		t.Error("Prompt should contain file path")
	}
}

func TestExtractRepoPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com/owner/repo", "owner/repo"},
		{"https://github.com/owner/repo.git", "owner/repo"},
		{"git@github.com:owner/repo.git", "owner/repo"},
	}

	for _, tt := range tests {
		result := extractRepoPath(tt.input)
		if result != tt.expected {
			t.Errorf("extractRepoPath(%s) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// MockLLMClient for testing
type MockLLMClient struct {
	responses []*Response
	callCount int
}

func (m *MockLLMClient) Chat(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error) {
	if m.callCount < len(m.responses) {
		resp := m.responses[m.callCount]
		m.callCount++
		return resp, nil
	}
	return &Response{Content: "Done", StopReason: "end_turn"}, nil
}
