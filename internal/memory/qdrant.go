package memory

import (
	"context"
	"sync"

	"github.com/chiheb/sentinel-remediator/internal/config"
	"github.com/chiheb/sentinel-remediator/internal/domain"
	"go.uber.org/zap"
)

// FixMemory stores and retrieves successful vulnerability fixes
// This is a simplified in-memory implementation. For production,
// integrate with Qdrant or another vector database.
type FixMemory struct {
	mu     sync.RWMutex
	fixes  []StoredFix
	logger *zap.Logger
}

// StoredFix represents a successful fix stored in memory
type StoredFix struct {
	ID                string          `json:"id"`
	VulnerabilityType domain.VulnType `json:"vulnerability_type"`
	Description       string          `json:"description"`
	FilePath          string          `json:"file_path"`
	OriginalContent   string          `json:"original_content"`
	FixedContent      string          `json:"fixed_content"`
	CommitMessage     string          `json:"commit_message"`
	Score             float32         `json:"score,omitempty"`
}

// NewFixMemory creates a new fix memory store
func NewFixMemory(cfg *config.Config, logger *zap.Logger) (*FixMemory, error) {
	logger.Info("Initialized in-memory fix storage (Qdrant integration available for production)")
	
	return &FixMemory{
		fixes:  make([]StoredFix, 0),
		logger: logger,
	}, nil
}

// Close is a no-op for the in-memory implementation
func (fm *FixMemory) Close() error {
	return nil
}

// StoreFix saves a successful fix to memory
func (fm *FixMemory) StoreFix(ctx context.Context, fix *StoredFix) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.fixes = append(fm.fixes, *fix)
	fm.logger.Info("Stored fix in memory",
		zap.String("id", fix.ID),
		zap.String("type", string(fix.VulnerabilityType)),
	)

	return nil
}

// RetrieveSimilar finds similar past fixes for a vulnerability
// In this simplified version, we just filter by vulnerability type
func (fm *FixMemory) RetrieveSimilar(ctx context.Context, vuln *domain.Vulnerability, limit int) ([]StoredFix, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	var results []StoredFix
	for _, fix := range fm.fixes {
		if fix.VulnerabilityType == vuln.Type {
			results = append(results, fix)
			if len(results) >= limit {
				break
			}
		}
	}

	fm.logger.Debug("Retrieved similar fixes",
		zap.String("vuln_type", string(vuln.Type)),
		zap.Int("count", len(results)),
	)

	return results, nil
}
