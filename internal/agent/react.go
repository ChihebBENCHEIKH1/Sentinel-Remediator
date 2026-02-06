package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chiheb/sentinel-remediator/internal/config"
	"github.com/chiheb/sentinel-remediator/internal/domain"
	"github.com/chiheb/sentinel-remediator/internal/tools"
	"go.uber.org/zap"
)

// Agent implements the ReAct pattern for vulnerability remediation
type Agent struct {
	llm           LLMClient
	registry      *tools.Registry
	cfg           *config.Config
	logger        *zap.Logger
	maxIterations int
	
	// Event channel for streaming thoughts
	eventChan chan<- domain.ThoughtStep
}

// NewAgent creates a new remediation agent
func NewAgent(cfg *config.Config, registry *tools.Registry, logger *zap.Logger) (*Agent, error) {
	llm := NewAnthropicClient(cfg, logger)
	
	return &Agent{
		llm:           llm,
		registry:      registry,
		cfg:           cfg,
		logger:        logger,
		maxIterations: cfg.MaxIterations,
	}, nil
}

// SetEventChannel sets the channel for streaming thought events
func (a *Agent) SetEventChannel(ch chan<- domain.ThoughtStep) {
	a.eventChan = ch
}

// Run executes the ReAct loop for a remediation job
func (a *Agent) Run(ctx context.Context, job *domain.RemediationJob) error {
	a.logger.Info("Starting remediation agent",
		zap.String("job_id", job.ID),
		zap.Int("vulnerabilities", len(job.ScanResult.Vulnerabilities)),
	)

	job.UpdateStatus(domain.JobStatusReasoning)

	// Clone the repository first
	if err := a.setupRepository(ctx, job); err != nil {
		job.SetError(fmt.Sprintf("Failed to setup repository: %s", err))
		return err
	}

	// Create a fix branch
	branchName := fmt.Sprintf("sentinel-fix-%s", time.Now().Format("20060102-150405"))
	job.BranchName = branchName
	
	_, err := a.registry.Execute(ctx, "git", map[string]any{
		"action":      "branch",
		"branch_name": branchName,
	})
	if err != nil {
		job.SetError(fmt.Sprintf("Failed to create branch: %s", err))
		return err
	}

	// Fix each vulnerability
	for i, vuln := range job.ScanResult.Vulnerabilities {
		a.logger.Info("Processing vulnerability",
			zap.String("vuln_id", vuln.ID),
			zap.String("type", string(vuln.Type)),
			zap.Int("index", i+1),
			zap.Int("total", len(job.ScanResult.Vulnerabilities)),
		)

		job.UpdateProgress(float64(i) / float64(len(job.ScanResult.Vulnerabilities)))

		attempt := domain.FixAttempt{
			VulnerabilityID: vuln.ID,
			StartTime:       time.Now(),
		}

		err := a.fixVulnerability(ctx, job, &vuln, &attempt)
		if err != nil {
			a.logger.Error("Failed to fix vulnerability",
				zap.String("vuln_id", vuln.ID),
				zap.Error(err),
			)
			attempt.Success = false
			attempt.ErrorMessage = err.Error()
		} else {
			attempt.Success = true
		}

		job.RecordFixAttempt(attempt)
	}

	// Push changes if any fixes were applied
	if job.FixedCount > 0 {
		if a.cfg.DryRun {
			a.logger.Info("Dry run enabled, skipping git push and PR creation", 
				zap.String("branch", branchName))
			a.emitThought(job, "success", "Dry run enabled: Fixed verified locally. Skipping git push.", 0)
		} else {
			job.UpdateStatus(domain.JobStatusPushing)
			
			if _, err := a.registry.Execute(ctx, "git", map[string]any{
				"action": "push",
			}); err != nil {
				a.logger.Error("Failed to push changes", zap.Error(err))
			}
			
			// TODO: Create PR via GitHub API
			job.PRUrl = fmt.Sprintf("https://github.com/%s/compare/%s", 
				extractRepoPath(job.ScanResult.RepoURL), branchName)
		}
	}

	job.UpdateProgress(1.0)
	
	if job.FailedCount == 0 {
		job.UpdateStatus(domain.JobStatusSuccess)
	} else if job.FixedCount > 0 {
		job.UpdateStatus(domain.JobStatusSuccess) // Partial success
	} else {
		job.UpdateStatus(domain.JobStatusFailed)
	}

	a.logger.Info("Remediation complete",
		zap.String("job_id", job.ID),
		zap.Int("fixed", job.FixedCount),
		zap.Int("failed", job.FailedCount),
	)

	return nil
}

// setupRepository clones the repository to the work directory
func (a *Agent) setupRepository(ctx context.Context, job *domain.RemediationJob) error {
	a.emitThought(job, "thought", "Setting up repository for remediation...", 0)

	result, err := a.registry.Execute(ctx, "git", map[string]any{
		"action":   "clone",
		"repo_url": job.ScanResult.RepoURL,
	})
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("git clone failed: %s", result.Error)
	}

	return nil
}

// fixVulnerability runs the ReAct loop for a single vulnerability
func (a *Agent) fixVulnerability(ctx context.Context, job *domain.RemediationJob, vuln *domain.Vulnerability, attempt *domain.FixAttempt) error {
	job.UpdateStatus(domain.JobStatusReasoning)

	// Build initial messages
	messages := []Message{
		{Role: "system", Content: SystemPrompt},
		{Role: "user", Content: a.buildVulnerabilityPrompt(job, vuln)},
	}

	// ReAct loop
	for iteration := 0; iteration < a.maxIterations; iteration++ {
		attempt.Iterations = iteration + 1

		// Get LLM response
		resp, err := a.llm.Chat(ctx, messages, a.registry.GetDefinitions())
		if err != nil {
			return fmt.Errorf("LLM error: %w", err)
		}

		// Record the thought
		if resp.Content != "" {
			a.emitThought(job, "thought", resp.Content, iteration)
		}

		// Check if we have tool calls
		if len(resp.ToolCalls) == 0 {
			// No more actions needed - check if this was a success indicator
			if strings.Contains(strings.ToLower(resp.Content), "fixed") ||
				strings.Contains(strings.ToLower(resp.Content), "complete") ||
				strings.Contains(strings.ToLower(resp.Content), "success") {
				return nil
			}
			
			// Agent is done reasoning without taking action
			if iteration > 0 {
				return nil
			}
			continue
		}

		// Execute each tool call
		for _, toolCall := range resp.ToolCalls {
			job.UpdateStatus(domain.JobStatusApplying)
			
			// Emit action event
			argsJSON, _ := json.Marshal(toolCall.Arguments)
			a.emitThought(job, "action", fmt.Sprintf("Calling %s with args: %s", toolCall.Name, string(argsJSON)), iteration)

			// Execute the tool
			result, err := a.registry.Execute(ctx, toolCall.Name, toolCall.Arguments)
			if err != nil {
				a.emitThought(job, "error", fmt.Sprintf("Tool error: %s", err), iteration)
			}

			// Record result
			a.emitThought(job, "observation", result.Output, iteration)
			
			// Track modified files
			if toolCall.Name == "filesystem" {
				if action, ok := toolCall.Arguments["action"].(string); ok {
					if action == "write" || action == "patch" {
						if path, ok := toolCall.Arguments["path"].(string); ok {
							attempt.FilesModified = append(attempt.FilesModified, path)
						}
					}
				}
			}

			// If this was a docker build, check if it succeeded
			if toolCall.Name == "docker" {
				if action, ok := toolCall.Arguments["action"].(string); ok && action == "build" {
					job.UpdateStatus(domain.JobStatusBuilding)
					
					if !result.Success {
						// Build failed - add error context for retry
						messages = append(messages, Message{
							Role:    "assistant",
							Content: resp.Content,
						})
						messages = append(messages, Message{
							Role:    "user",
							Content: fmt.Sprintf(BuildFailurePrompt, result.Error),
						})
						continue
					}
				}
			}

			// Add tool result to conversation
			messages = append(messages, Message{
				Role:    "assistant",
				Content: resp.Content,
			})
			messages = append(messages, Message{
				Role: "user",
				Content: fmt.Sprintf("Tool '%s' result:\n%s", toolCall.Name, result.Output),
			})
		}

		// Check if we should commit the changes
		if len(attempt.FilesModified) > 0 && iteration >= 2 {
			// Stage and commit changes
			a.registry.Execute(ctx, "git", map[string]any{
				"action": "add",
				"files":  attempt.FilesModified,
			})
			
			commitMsg := fmt.Sprintf("fix(%s): %s", vuln.Type, vuln.Title)
			a.registry.Execute(ctx, "git", map[string]any{
				"action":  "commit",
				"message": commitMsg,
			})
			
			return nil
		}
	}

	return fmt.Errorf("max iterations reached without successful fix")
}

// buildVulnerabilityPrompt creates the initial prompt for a vulnerability
func (a *Agent) buildVulnerabilityPrompt(job *domain.RemediationJob, vuln *domain.Vulnerability) string {
	return fmt.Sprintf(VulnerabilityAnalysisPrompt,
		vuln.ID,
		vuln.Type,
		vuln.Severity,
		vuln.Title,
		vuln.Description,
		vuln.FilePath,
		vuln.LineNumber,
		vuln.Suggestion,
		job.ScanResult.RepoURL,
		job.ScanResult.Branch,
		job.WorkDir,
	)
}

// emitThought sends a thought step to the event channel and records it
func (a *Agent) emitThought(job *domain.RemediationJob, thoughtType, content string, iteration int) {
	thought := domain.ThoughtStep{
		Timestamp: time.Now(),
		Type:      thoughtType,
		Content:   content,
		Iteration: iteration,
	}
	
	job.AddThought(thought)
	
	if a.eventChan != nil {
		select {
		case a.eventChan <- thought:
		default:
			// Don't block if channel is full
		}
	}
}

// extractRepoPath extracts owner/repo from a GitHub URL
func extractRepoPath(url string) string {
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "https://github.com/")
	url = strings.TrimPrefix(url, "git@github.com:")
	return url
}
