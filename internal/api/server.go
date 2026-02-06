package api

import (
	"net/http"
	"time"

	"github.com/chiheb/sentinel-remediator/internal/agent"
	"github.com/chiheb/sentinel-remediator/internal/config"
	"github.com/chiheb/sentinel-remediator/internal/memory"
	"github.com/chiheb/sentinel-remediator/internal/tools"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// Server holds the HTTP server and dependencies
type Server struct {
	router   *chi.Mux
	cfg      *config.Config
	logger   *zap.Logger
	agent    *agent.Agent
	registry *tools.Registry
	memory   *memory.FixMemory
	jobs     *JobStore
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	// Initialize tool registry
	registry := tools.NewRegistry(logger)
	registry.Register(tools.NewGitTool(cfg.WorkDir, cfg.GitHubToken, logger))
	registry.Register(tools.NewFilesystemTool(cfg.WorkDir, logger))
	registry.Register(tools.NewDockerTool(cfg.WorkDir, logger))

	// Initialize memory (optional - may fail if Qdrant not available)
	fixMemory, err := memory.NewFixMemory(cfg, logger)
	if err != nil {
		logger.Warn("Failed to initialize memory store", zap.Error(err))
		// Continue without memory
	}

	// Initialize agent
	remedAgent, err := agent.NewAgent(cfg, registry, logger)
	if err != nil {
		return nil, err
	}

	// Initialize job store
	jobs := NewJobStore()

	s := &Server{
		cfg:      cfg,
		logger:   logger,
		agent:    remedAgent,
		registry: registry,
		memory:   fixMemory,
		jobs:     jobs,
	}

	s.setupRouter()

	return s, nil
}

// Router returns the HTTP router
func (s *Server) Router() http.Handler {
	return s.router
}

// setupRouter configures all routes
func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS for dashboard
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", s.handleHealth)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Remediation endpoints
		r.Post("/remediate", s.handleRemediate)
		r.Get("/jobs", s.handleListJobs)
		r.Get("/jobs/{jobID}", s.handleGetJob)
		r.Get("/jobs/{jobID}/stream", s.handleStreamJob)
		r.Delete("/jobs/{jobID}", s.handleCancelJob)

		// Memory endpoints
		r.Get("/memory/search", s.handleSearchMemory)
	})

	s.router = r
}

// handleHealth returns service health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"sentinel-remediator"}`))
}
