package domain

import (
	"testing"
	"time"
)

func TestNewVulnerability(t *testing.T) {
	vuln := NewVulnerability(VulnTypeRunAsRoot, SeverityHigh, "Test Title", "Test Description")

	if vuln.ID == "" {
		t.Error("ID should not be empty")
	}

	if vuln.Type != VulnTypeRunAsRoot {
		t.Errorf("Expected type RUN_AS_ROOT, got %s", vuln.Type)
	}

	if vuln.Severity != SeverityHigh {
		t.Errorf("Expected severity HIGH, got %s", vuln.Severity)
	}
}

func TestVulnerabilityIsCritical(t *testing.T) {
	tests := []struct {
		severity Severity
		expected bool
	}{
		{SeverityCritical, true},
		{SeverityHigh, true},
		{SeverityMedium, false},
		{SeverityLow, false},
		{SeverityInfo, false},
	}

	for _, tt := range tests {
		vuln := &Vulnerability{Severity: tt.severity}
		if vuln.IsCritical() != tt.expected {
			t.Errorf("IsCritical() for %s = %v, expected %v", tt.severity, vuln.IsCritical(), tt.expected)
		}
	}
}

func TestNewRemediationJob(t *testing.T) {
	scanResult := &ScanResult{
		ScanID:    "scan-001",
		ImageName: "test",
		Vulnerabilities: []Vulnerability{
			{ID: "v1"},
			{ID: "v2"},
		},
	}

	job := NewRemediationJob(scanResult, "/tmp/test")

	if job.ID == "" {
		t.Error("Job ID should not be empty")
	}

	if job.Status != JobStatusPending {
		t.Errorf("Expected status PENDING, got %s", job.Status)
	}

	if job.TotalCount != 2 {
		t.Errorf("Expected TotalCount=2, got %d", job.TotalCount)
	}
}

func TestRemediationJobAddThought(t *testing.T) {
	job := &RemediationJob{
		ThoughtTrace: make([]ThoughtStep, 0),
	}

	job.AddThought(ThoughtStep{
		Type:    "thought",
		Content: "Test thought",
	})

	if len(job.ThoughtTrace) != 1 {
		t.Errorf("Expected 1 thought, got %d", len(job.ThoughtTrace))
	}

	if job.ThoughtTrace[0].Content != "Test thought" {
		t.Errorf("Expected content 'Test thought', got '%s'", job.ThoughtTrace[0].Content)
	}

	// Timestamp should be set
	if job.ThoughtTrace[0].Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestRemediationJobUpdateStatus(t *testing.T) {
	job := &RemediationJob{
		Status: JobStatusPending,
	}

	job.UpdateStatus(JobStatusReasoning)

	if job.Status != JobStatusReasoning {
		t.Errorf("Expected status REASONING, got %s", job.Status)
	}

	// Test completion status
	job.UpdateStatus(JobStatusSuccess)

	if job.CompletedAt == nil {
		t.Error("CompletedAt should be set for SUCCESS status")
	}
}

func TestRemediationJobRecordFixAttempt(t *testing.T) {
	job := &RemediationJob{
		FixAttempts: make([]FixAttempt, 0),
	}

	job.RecordFixAttempt(FixAttempt{
		VulnerabilityID: "v1",
		Success:         true,
		StartTime:       time.Now(),
	})

	if job.FixedCount != 1 {
		t.Errorf("Expected FixedCount=1, got %d", job.FixedCount)
	}

	job.RecordFixAttempt(FixAttempt{
		VulnerabilityID: "v2",
		Success:         false,
		ErrorMessage:    "Build failed",
	})

	if job.FailedCount != 1 {
		t.Errorf("Expected FailedCount=1, got %d", job.FailedCount)
	}
}

func TestRemediationJobThreadSafety(t *testing.T) {
	job := NewRemediationJob(&ScanResult{}, "/tmp/test")

	// Run concurrent operations
	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func(n int) {
			job.AddThought(ThoughtStep{
				Type:    "thought",
				Content: "Concurrent thought",
			})
			job.UpdateProgress(float64(n) / 100)
			job.GetStatus()
			job.GetThoughtTrace()
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	// Should not panic and should have recorded thoughts
	if len(job.GetThoughtTrace()) != 100 {
		t.Errorf("Expected 100 thoughts, got %d", len(job.GetThoughtTrace()))
	}
}
