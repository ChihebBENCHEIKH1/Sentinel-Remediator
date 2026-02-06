package domain

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current state of a remediation job
type JobStatus string

const (
	JobStatusPending    JobStatus = "PENDING"
	JobStatusReasoning  JobStatus = "REASONING"
	JobStatusApplying   JobStatus = "APPLYING_FIX"
	JobStatusBuilding   JobStatus = "BUILDING"
	JobStatusTesting    JobStatus = "TESTING"
	JobStatusPushing    JobStatus = "PUSHING"
	JobStatusSuccess    JobStatus = "SUCCESS"
	JobStatusFailed     JobStatus = "FAILED"
	JobStatusCancelled  JobStatus = "CANCELLED"
)

// ThoughtStep represents a single step in the agent's reasoning trace
type ThoughtStep struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // "thought", "action", "observation", "error"
	Content     string    `json:"content"`
	ToolName    string    `json:"tool_name,omitempty"`
	ToolArgs    string    `json:"tool_args,omitempty"`
	ToolResult  string    `json:"tool_result,omitempty"`
	Iteration   int       `json:"iteration"`
}

// FixAttempt represents a single attempt to fix a vulnerability
type FixAttempt struct {
	VulnerabilityID string    `json:"vulnerability_id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time,omitempty"`
	Success         bool      `json:"success"`
	FilesModified   []string  `json:"files_modified"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	Iterations      int       `json:"iterations"`
}

// RemediationJob represents a complete remediation workflow
type RemediationJob struct {
	mu sync.RWMutex

	ID              string          `json:"id"`
	ScanResult      *ScanResult     `json:"scan_result"`
	Status          JobStatus       `json:"status"`
	Progress        float64         `json:"progress"` // 0.0 to 1.0
	ThoughtTrace    []ThoughtStep   `json:"thought_trace"`
	FixAttempts     []FixAttempt    `json:"fix_attempts"`
	PRUrl           string          `json:"pr_url,omitempty"`
	BranchName      string          `json:"branch_name,omitempty"`
	WorkDir         string          `json:"work_dir"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	ErrorMessage    string          `json:"error_message,omitempty"`
	FixedCount      int             `json:"fixed_count"`
	FailedCount     int             `json:"failed_count"`
	TotalCount      int             `json:"total_count"`
}

// NewRemediationJob creates a new remediation job
func NewRemediationJob(scanResult *ScanResult, workDir string) *RemediationJob {
	now := time.Now()
	return &RemediationJob{
		ID:           uuid.New().String(),
		ScanResult:   scanResult,
		Status:       JobStatusPending,
		Progress:     0.0,
		ThoughtTrace: make([]ThoughtStep, 0),
		FixAttempts:  make([]FixAttempt, 0),
		WorkDir:      workDir,
		CreatedAt:    now,
		UpdatedAt:    now,
		TotalCount:   len(scanResult.Vulnerabilities),
	}
}

// AddThought appends a thought step to the trace (thread-safe)
func (j *RemediationJob) AddThought(thought ThoughtStep) {
	j.mu.Lock()
	defer j.mu.Unlock()
	thought.Timestamp = time.Now()
	j.ThoughtTrace = append(j.ThoughtTrace, thought)
	j.UpdatedAt = time.Now()
}

// UpdateStatus updates the job status (thread-safe)
func (j *RemediationJob) UpdateStatus(status JobStatus) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Status = status
	j.UpdatedAt = time.Now()
	
	if status == JobStatusSuccess || status == JobStatusFailed || status == JobStatusCancelled {
		now := time.Now()
		j.CompletedAt = &now
	}
}

// SetError sets an error message and marks the job as failed
func (j *RemediationJob) SetError(message string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.ErrorMessage = message
	j.Status = JobStatusFailed
	j.UpdatedAt = time.Now()
	now := time.Now()
	j.CompletedAt = &now
}

// GetStatus returns the current status (thread-safe)
func (j *RemediationJob) GetStatus() JobStatus {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.Status
}

// GetThoughtTrace returns a copy of the thought trace (thread-safe)
func (j *RemediationJob) GetThoughtTrace() []ThoughtStep {
	j.mu.RLock()
	defer j.mu.RUnlock()
	trace := make([]ThoughtStep, len(j.ThoughtTrace))
	copy(trace, j.ThoughtTrace)
	return trace
}

// UpdateProgress updates the job progress (thread-safe)
func (j *RemediationJob) UpdateProgress(progress float64) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.Progress = progress
	j.UpdatedAt = time.Now()
}

// RecordFixAttempt records a fix attempt result
func (j *RemediationJob) RecordFixAttempt(attempt FixAttempt) {
	j.mu.Lock()
	defer j.mu.Unlock()
	attempt.EndTime = time.Now()
	j.FixAttempts = append(j.FixAttempts, attempt)
	if attempt.Success {
		j.FixedCount++
	} else {
		j.FailedCount++
	}
	j.UpdatedAt = time.Now()
}

// ToJSON returns a JSON-safe copy of the job
func (j *RemediationJob) ToJSON() *RemediationJob {
	j.mu.RLock()
	defer j.mu.RUnlock()
	
	// Create a shallow copy
	copy := *j
	copy.ThoughtTrace = make([]ThoughtStep, len(j.ThoughtTrace))
	copy.FixAttempts = make([]FixAttempt, len(j.FixAttempts))
	for i, t := range j.ThoughtTrace {
		copy.ThoughtTrace[i] = t
	}
	for i, f := range j.FixAttempts {
		copy.FixAttempts[i] = f
	}
	return &copy
}
