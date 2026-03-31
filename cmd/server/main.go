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
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	qdrant := rag.NewQdrantClient(cfg.QdrantURL, cfg.HTTPTimeout)
	embedder := rag.NewEmbedder(cfg.OllamaURL, cfg.EmbedModel, cfg.HTTPTimeout)
	llmClient := llm.NewClient(cfg.OllamaURL, cfg.LLMModel, cfg.HTTPTimeout)
	retriever := rag.NewRetriever(embedder, qdrant, cfg.CollectionName, cfg.TopK)
	runner := agent.NewRunner(retriever, llmClient)
	indexer := rag.NewIndexer(qdrant, embedder)

	cmd := "analyze"
	if len(os.Args) > 1 {
		cmd = strings.ToLower(os.Args[1])
	}

	if cmd != "serve" {
		checkCtx, cancelCheck := context.WithTimeout(context.Background(), 2*cfg.HTTPTimeout)
		defer cancelCheck()
		if err := healthcheck(cfg, checkCtx); err != nil {
			return err
		}
	}

	switch cmd {
	case "serve":
		return runHTTPServer(cfg, runner, newDependencyMonitor(cfg), logger)
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

type dependencyMonitor struct {
	cfg       config.Config
	mu        sync.RWMutex
	lastErr   error
	lastCheck time.Time
}

func newDependencyMonitor(cfg config.Config) *dependencyMonitor {
	return &dependencyMonitor{cfg: cfg}
}

func (m *dependencyMonitor) run(ctx context.Context, interval time.Duration) {
	m.checkOnce(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkOnce(ctx)
		}
	}
}

func (m *dependencyMonitor) checkOnce(parent context.Context) {
	checkCtx, cancel := context.WithTimeout(parent, 2*m.cfg.HTTPTimeout)
	defer cancel()
	err := healthcheck(m.cfg, checkCtx)

	m.mu.Lock()
	m.lastErr = err
	m.lastCheck = time.Now().UTC()
	m.mu.Unlock()
}

func (m *dependencyMonitor) snapshot() (error, time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastErr, m.lastCheck
}

func runHTTPServer(cfg config.Config, runner *agent.Runner, monitor *dependencyMonitor, logger *slog.Logger) error {
	serverCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	checkInterval := cfg.HTTPTimeout
	if checkInterval < 5*time.Second {
		checkInterval = 5 * time.Second
	}
	go monitor.run(serverCtx, checkInterval)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(2 * cfg.HTTPTimeout))
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Duration("duration", time.Since(start)),
			)
		})
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"alive"}`))
	})
	router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		lastErr, checkedAt := monitor.snapshot()
		w.Header().Set("Content-Type", "application/json")
		if checkedAt.IsZero() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"starting","ready":false}`))
			return
		}
		if lastErr != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status":     "degraded",
				"ready":      false,
				"checked_at": checkedAt.Format(time.RFC3339),
				"error":      lastErr.Error(),
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":     "ok",
			"ready":      true,
			"checked_at": checkedAt.Format(time.RFC3339),
		})
	})
	router.Post("/ask", func(w http.ResponseWriter, r *http.Request) {
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
	logger.Info("server listening",
		slog.String("addr", cfg.ListenAddr),
		slog.String("routes", "POST /ask, GET /health, GET /ready"),
	)
	return http.ListenAndServe(cfg.ListenAddr, router)
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
	for i, t := range truth {
		fmt.Fprintf(os.Stderr, "evaluating %s (%d/%d)\n", t.Ticker, i+1, len(truth))
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
