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

type QdrantClient struct {
	baseURL string
	http    *http.Client
}

func NewQdrantClient(baseURL string, timeout time.Duration) *QdrantClient {
	return &QdrantClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: timeout},
	}
}

func (q *QdrantClient) EnsureCollection(ctx context.Context, collection string, vectorSize int) error {
	body := map[string]any{
		"vectors": map[string]any{
			"size":     vectorSize,
			"distance": "Cosine",
		},
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/collections/%s", q.baseURL, collection), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 200/201 OK, 409 means collection already exists – treat both as success.
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		return nil
	}

	if resp.StatusCode >= 300 {
		d, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("ensure collection failed: %s (%s)", resp.Status, string(d))
	}
	return nil
}

func (q *QdrantClient) UpsertPoint(ctx context.Context, collection string, doc FactDocument, vector []float64) error {
	body := map[string]any{
		"points": []map[string]any{
			{
				"id":     doc.ID,
				"vector": vector,
				"payload": map[string]any{
					"ticker": doc.Ticker,
					"type":   doc.Type,
					"text":   doc.Text,
				},
			},
		},
	}
	return q.put(ctx, fmt.Sprintf("%s/collections/%s/points?wait=true", q.baseURL, collection), body)
}

type searchResult struct {
	Result []struct {
		Payload map[string]any `json:"payload"`
		Score   float64        `json:"score"`
	} `json:"result"`
}

func (q *QdrantClient) Search(ctx context.Context, collection, ticker string, vector []float64, limit int) ([]string, error) {
	body := map[string]any{
		"vector": vector,
		"limit":  limit,
		"filter": map[string]any{
			"must": []map[string]any{
				{
					"key": "ticker",
					"match": map[string]any{
						"value": strings.ToUpper(ticker),
					},
				},
			},
		},
		"with_payload": true,
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/collections/%s/points/search", q.baseURL, collection), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		d, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("search failed: %s (%s)", resp.Status, string(d))
	}
	var parsed searchResult
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(parsed.Result))
	for _, item := range parsed.Result {
		text, ok := item.Payload["text"].(string)
		if !ok {
			continue
		}
		out = append(out, text)
	}
	return out, nil
}

func (q *QdrantClient) post(ctx context.Context, url string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		d, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("qdrant post failed: %s (%s)", resp.Status, string(d))
	}
	return nil
}

func (q *QdrantClient) put(ctx context.Context, url string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := q.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		d, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("qdrant put failed: %s (%s)", resp.Status, string(d))
	}
	return nil
}
