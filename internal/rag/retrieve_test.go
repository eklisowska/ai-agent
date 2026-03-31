package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQdrantSearchTickerFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/collections/test/points/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err, "decode request")
		filter := payload["filter"].(map[string]any)
		must := filter["must"].([]any)
		entry := must[0].(map[string]any)
		match := entry["match"].(map[string]any)
		require.Equal(t, "AAPL", match["value"], "expected ticker filter")

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
	require.NoError(t, err, "Search returned error")
	require.Len(t, out, 1, "expected one document")
}
