package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQdrantSearchTickerFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/collections/test/points/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		filter := payload["filter"].(map[string]any)
		must := filter["must"].([]any)
		entry := must[0].(map[string]any)
		match := entry["match"].(map[string]any)
		if match["value"] != "AAPL" {
			t.Fatalf("expected ticker filter AAPL, got %v", match["value"])
		}

		resp := map[string]any{
			"result": []map[string]any{
				{"payload": map[string]any{"text": "AAPL PE ratio is 28"}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewQdrantClient(server.URL, 2*time.Second)
	out, err := client.Search(context.Background(), "test", "aapl", []float64{0.1, 0.2}, 5)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(out))
	}
}
