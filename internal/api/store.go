package api

import (
	"sync"

	"github.com/chiheb/sentinel-remediator/internal/domain"
)

// JobStore provides thread-safe storage for remediation jobs
type JobStore struct {
	mu            sync.RWMutex
	jobs          map[string]*domain.RemediationJob
	eventChannels map[string]chan domain.ThoughtStep
}

// NewJobStore creates a new job store
func NewJobStore() *JobStore {
	return &JobStore{
		jobs:          make(map[string]*domain.RemediationJob),
		eventChannels: make(map[string]chan domain.ThoughtStep),
	}
}

// Store adds a job to the store
func (s *JobStore) Store(job *domain.RemediationJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[job.ID] = job
}

// Get retrieves a job by ID
func (s *JobStore) Get(id string) (*domain.RemediationJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	return job, ok
}

// List returns all jobs
func (s *JobStore) List() []*domain.RemediationJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	jobs := make([]*domain.RemediationJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// Delete removes a job from the store
func (s *JobStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.jobs, id)
	delete(s.eventChannels, id)
}

// SetEventChannel sets the event channel for a job
func (s *JobStore) SetEventChannel(id string, ch chan domain.ThoughtStep) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.eventChannels[id] = ch
}

// GetEventChannel gets the event channel for a job
func (s *JobStore) GetEventChannel(id string) chan domain.ThoughtStep {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.eventChannels[id]
}

// Count returns the number of jobs
func (s *JobStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobs)
}
