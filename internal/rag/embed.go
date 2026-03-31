package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Embedder struct {
	baseURL string
	model   string
	http    *http.Client
}

func NewEmbedder(baseURL, model string, timeout time.Duration) *Embedder {
	return &Embedder{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		http:    &http.Client{Timeout: timeout},
	}
}

type embeddingsRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embeddingsResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (e *Embedder) Embed(ctx context.Context, text string) ([]float64, error) {
	reqBody := embeddingsRequest{
		Model:  e.model,
		Prompt: text,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("ollama embeddings failed: %s (%s)", resp.Status, string(data))
	}
	var parsed embeddingsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if len(parsed.Embedding) == 0 {
		return nil, fmt.Errorf("empty embedding")
	}
	return parsed.Embedding, nil
}
