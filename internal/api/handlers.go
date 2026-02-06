package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/chiheb/sentinel-remediator/internal/domain"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// RemediateRequest is the request body for starting a remediation
type RemediateRequest struct {
	ScanResult *domain.ScanResult `json:"scan_result"`
}

// handleRemediate starts a new remediation job
func (s *Server) handleRemediate(w http.ResponseWriter, r *http.Request) {
	var req RemediateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body: %s", err)
		return
	}

	if req.ScanResult == nil {
		s.respondError(w, http.StatusBadRequest, "scan_result is required")
		return
	}

	if len(req.ScanResult.Vulnerabilities) == 0 {
		s.respondError(w, http.StatusBadRequest, "No vulnerabilities to remediate")
		return
	}

	// Create work directory for this job
	workDir := filepath.Join(s.cfg.WorkDir, fmt.Sprintf("job-%d", time.Now().UnixNano()))

	// Create job
	job := domain.NewRemediationJob(req.ScanResult, workDir)
	s.jobs.Store(job)

	s.logger.Info("Created remediation job",
		zap.String("job_id", job.ID),
		zap.Int("vulnerabilities", len(req.ScanResult.Vulnerabilities)),
	)

	// Start remediation in background
	go func() {
		ctx := context.Background()
		
		// Create event channel for streaming
		eventChan := make(chan domain.ThoughtStep, 100)
		s.jobs.SetEventChannel(job.ID, eventChan)
		s.agent.SetEventChannel(eventChan)
		defer close(eventChan)

		if err := s.agent.Run(ctx, job); err != nil {
			s.logger.Error("Remediation failed",
				zap.String("job_id", job.ID),
				zap.Error(err),
			)
		}
	}()

	s.respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id":  job.ID,
		"status":  job.Status,
		"message": "Remediation job started",
	})
}

// handleListJobs lists all remediation jobs
func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobs := s.jobs.List()
	
	// Convert to JSON-safe format
	result := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		result[i] = map[string]interface{}{
			"id":          job.ID,
			"status":      job.Status,
			"progress":    job.Progress,
			"fixed_count": job.FixedCount,
			"failed_count": job.FailedCount,
			"total_count": job.TotalCount,
			"created_at":  job.CreatedAt,
			"updated_at":  job.UpdatedAt,
			"pr_url":      job.PRUrl,
		}
	}

	s.respondJSON(w, http.StatusOK, result)
}

// handleGetJob returns a specific job's details
func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	
	job, ok := s.jobs.Get(jobID)
	if !ok {
		s.respondError(w, http.StatusNotFound, "Job not found: %s", jobID)
		return
	}

	s.respondJSON(w, http.StatusOK, job.ToJSON())
}

// handleStreamJob streams job events via Server-Sent Events
func (s *Server) handleStreamJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	
	job, ok := s.jobs.Get(jobID)
	if !ok {
		s.respondError(w, http.StatusNotFound, "Job not found: %s", jobID)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.respondError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	// Get or create event channel
	eventChan := s.jobs.GetEventChannel(jobID)
	if eventChan == nil {
		// Job might be completed, send current state and close
		s.sendSSE(w, "status", map[string]interface{}{
			"status":       job.Status,
			"progress":     job.Progress,
			"thought_trace": job.GetThoughtTrace(),
		})
		flusher.Flush()
		return
	}

	// Send initial state
	s.sendSSE(w, "init", map[string]interface{}{
		"job_id":   job.ID,
		"status":   job.Status,
		"progress": job.Progress,
	})
	flusher.Flush()

	// Stream events
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case thought, ok := <-eventChan:
			if !ok {
				// Channel closed, send final status
				s.sendSSE(w, "complete", map[string]interface{}{
					"status":      job.GetStatus(),
					"fixed_count": job.FixedCount,
					"failed_count": job.FailedCount,
					"pr_url":      job.PRUrl,
				})
				flusher.Flush()
				return
			}

			s.sendSSE(w, "thought", thought)
			flusher.Flush()
		}
	}
}

// handleCancelJob cancels a running job
func (s *Server) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	
	job, ok := s.jobs.Get(jobID)
	if !ok {
		s.respondError(w, http.StatusNotFound, "Job not found: %s", jobID)
		return
	}

	job.UpdateStatus(domain.JobStatusCancelled)

	s.respondJSON(w, http.StatusOK, map[string]string{
		"message": "Job cancelled",
		"job_id":  jobID,
	})
}

// handleSearchMemory searches for similar past fixes
func (s *Server) handleSearchMemory(w http.ResponseWriter, r *http.Request) {
	if s.memory == nil {
		s.respondError(w, http.StatusServiceUnavailable, "Memory store not available")
		return
	}

	query := r.URL.Query().Get("q")
	vulnType := r.URL.Query().Get("type")

	vuln := &domain.Vulnerability{
		Type:        domain.VulnType(vulnType),
		Description: query,
	}

	fixes, err := s.memory.RetrieveSimilar(r.Context(), vuln, 5)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to search memory: %s", err)
		return
	}

	s.respondJSON(w, http.StatusOK, fixes)
}

// Helper methods

func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) respondError(w http.ResponseWriter, status int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	s.logger.Error("API error", zap.String("error", msg), zap.Int("status", status))
	s.respondJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) sendSSE(w http.ResponseWriter, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(jsonData))
}
