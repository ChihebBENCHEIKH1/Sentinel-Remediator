package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/chiheb/sentinel-remediator/internal/config"
	"go.uber.org/zap"
)

// Embedder creates vector embeddings from text
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// AnthropicEmbedder uses Anthropic's embedding model (via a proxy or alternative)
// Note: Anthropic doesn't have a native embedding API, so we use OpenAI's
type OpenAIEmbedder struct {
	apiKey string
	model  string
	logger *zap.Logger
}

// NewEmbedder creates a new embedder
func NewEmbedder(cfg *config.Config, logger *zap.Logger) Embedder {
	// Use OpenAI for embeddings (more widely available)
	return &OpenAIEmbedder{
		apiKey: cfg.OpenAIAPIKey,
		model:  "text-embedding-3-small",
		logger: logger,
	}
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if e.apiKey == "" {
		// Return a dummy embedding for development
		e.logger.Warn("No OpenAI API key configured, using dummy embedding")
		return make([]float32, 1536), nil
	}

	reqBody := embeddingRequest{
		Model: e.model,
		Input: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding API returned status %d", resp.StatusCode)
	}

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embResp.Data[0].Embedding, nil
}
