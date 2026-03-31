package main

import (
	"ai-agent/internal/agent"
	"ai-agent/internal/config"
	"ai-agent/internal/eval"
	"ai-agent/internal/llm"
	"ai-agent/internal/rag"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	qdrant := rag.NewQdrantClient(cfg.QdrantURL, cfg.HTTPTimeout)
	embedder := rag.NewEmbedder(cfg.OllamaURL, cfg.EmbedModel, cfg.HTTPTimeout)
	llmClient := llm.NewClient(cfg.OllamaURL, cfg.LLMModel, cfg.HTTPTimeout)
	retriever := rag.NewRetriever(embedder, qdrant, cfg.CollectionName, cfg.TopK)
	runner := agent.NewRunner(retriever, llmClient)
	indexer := rag.NewIndexer(qdrant, embedder)

	checkCtx, cancelCheck := context.WithTimeout(context.Background(), 2*cfg.HTTPTimeout)
	defer cancelCheck()
	if err := healthcheck(cfg, checkCtx); err != nil {
		return err
	}

	cmd := "analyze"
	if len(os.Args) > 1 {
		cmd = strings.ToLower(os.Args[1])
	}

	switch cmd {
	case "serve":
		return runHTTPServer(cfg, runner)
	case "index":
		ctx, cancel := context.WithTimeout(context.Background(), 2*cfg.HTTPTimeout)
		defer cancel()
		return indexAll(ctx, cfg, indexer, qdrant, embedder)
	case "analyze":
		if len(os.Args) < 3 {
			return errors.New("usage: go run ./cmd/server analyze <ticker>")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*cfg.HTTPTimeout)
		defer cancel()
		result, err := runner.Run(ctx, os.Args[2])
		if err != nil {
			return err
		}
		return writeJSON(result)
	case "ask":
		if len(os.Args) < 4 {
			return errors.New(`usage: go run ./cmd/server ask <ticker> <question...>`)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*cfg.HTTPTimeout)
		defer cancel()
		ticker := os.Args[2]
		question := strings.Join(os.Args[3:], " ")
		result, err := runner.RunWithQuery(ctx, ticker, question)
		if err != nil {
			return err
		}
		return writeJSON(result)
	case "eval":
		ctx, cancel := context.WithTimeout(context.Background(), 2*cfg.HTTPTimeout)
		defer cancel()
		return runEval(ctx, cfg, indexer, qdrant, embedder, runner)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

type askRequest struct {
	Ticker   string `json:"ticker"`
	Question string `json:"question"`
}

func runHTTPServer(cfg config.Config, runner *agent.Runner) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req askRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		req.Ticker = strings.TrimSpace(req.Ticker)
		req.Question = strings.TrimSpace(req.Question)
		if req.Ticker == "" || req.Question == "" {
			http.Error(w, `body must include non-empty "ticker" and "question"`, http.StatusBadRequest)
			return
		}
		reqCtx, cancel := context.WithTimeout(r.Context(), 2*cfg.HTTPTimeout)
		defer cancel()
		result, err := runner.RunWithQuery(reqCtx, req.Ticker, req.Question)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	fmt.Fprintf(os.Stderr, "listening on %s (POST /ask, GET /health)\n", cfg.ListenAddr)
	return http.ListenAndServe(cfg.ListenAddr, mux)
}

func indexAll(ctx context.Context, cfg config.Config, indexer *rag.Indexer, qdrant *rag.QdrantClient, embedder *rag.Embedder) error {
	seed, err := embedder.Embed(ctx, "embedding size probe")
	if err != nil {
		return err
	}
	if err := qdrant.EnsureCollection(ctx, cfg.CollectionName, len(seed)); err != nil {
		return err
	}
	docs, _, err := rag.LoadRawData("data/raw")
	if err != nil {
		return err
	}
	if err := indexer.IndexDocuments(ctx, cfg.CollectionName, docs); err != nil {
		return err
	}
	fmt.Printf("indexed %d documents\n", len(docs))
	return nil
}

func runEval(ctx context.Context, cfg config.Config, indexer *rag.Indexer, qdrant *rag.QdrantClient, embedder *rag.Embedder, runner *agent.Runner) error {
	if err := indexAll(ctx, cfg, indexer, qdrant, embedder); err != nil {
		return err
	}

	_, truth, err := rag.LoadRawData("data/raw")
	if err != nil {
		return err
	}

	ev := eval.Evaluator{}
	items := make([]eval.ItemResult, 0, len(truth))
	for _, t := range truth {
		out, err := runner.Run(ctx, t.Ticker)
		if err != nil {
			return fmt.Errorf("ticker %s: %w", t.Ticker, err)
		}
		items = append(items, eval.MakeItem(t.Ticker, out, t.Expected, ev))
	}
	report := ev.BuildReport(items)
	return writeJSON(report)
}

func healthcheck(cfg config.Config, ctx context.Context) error {
	client := &http.Client{Timeout: cfg.HTTPTimeout}
	if err := probe(ctx, client, cfg.QdrantURL+"/collections"); err != nil {
		return fmt.Errorf("qdrant unavailable: %w", err)
	}
	if err := probe(ctx, client, cfg.OllamaURL+"/api/tags"); err != nil {
		return fmt.Errorf("ollama unavailable: %w", err)
	}
	return nil
}

func probe(ctx context.Context, client *http.Client, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %s", resp.Status)
	}
	return nil
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
