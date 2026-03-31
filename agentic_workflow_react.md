# Agentic Workflow (ReAct)

## Objective

Demonstrate a self-reflective agentic system that combines retrieval, reasoning, action execution, and evaluation into one workflow that returns `BUY`, `HOLD`, or `SELL`.

## ReAct Loop in This Project

1. **Retrieve**: fetch ticker-scoped facts from vector search (RAG).
2. **Act**: compute deterministic signals (sentiment, risk, P/E assessment).
3. **Reason**: generate a structured decision with the LLM (pass 1).
4. **Reflect**: self-check and optionally revise with the LLM (pass 2).
5. **Evaluate**: score predictions against labeled ground truth.

## Flow

### 0) Bootstrap

**Files:** `cmd/server/main.go`, `internal/config/config.go`

- Load env config, initialize Qdrant + embedding + LLM clients, wire retriever/indexer/runner.
- Expose API routes: `/index`, `/analyze`, `/ask`, `/eval`, `/health`, `/ready`.

### 1) Prepare and Index Context

**Files:** `internal/rag/index.go`, `internal/rag/embed.go`, `internal/rag/qdrant.go`, `data/raw/*.json`

- Normalize raw synthetic data into fact documents.
- Embed each fact with Ollama embeddings.
- Upsert vectors and payloads into the Qdrant collection.

### 2) Run Analysis (`/analyze` or `/ask`)

**Files:** `internal/agent/flow.go`, `internal/rag/retrieve.go`, `internal/agent/tools.go`, `internal/llm/client.go`, `internal/agent/reflection.go`

- Retrieve top-k facts filtered by ticker.
- Run tools:
  - `SentimentScore(...)`
  - `DetectRisk(...)`
  - `AnalyzePE(...)`
- Build prompt and generate initial JSON output (`decision`, `reasoning`, `confidence`).
- Run reflection prompt and choose the final answer via policy (`ChooseFinal`).

### 3) Return Auditable Output

**Files:** `internal/agent/types.go`

- API returns full `RunResult` including retrieved docs, tool summary, initial result, reflected result, final selection, and reflection rationale.

### 4) Evaluate System (`/eval`)

**Files:** `cmd/server/main.go`, `internal/eval/evaluator.go`, `data/raw/ground_truth.json`

- Run end-to-end analysis for each labeled ticker.
- Use existing indexed data (no re-indexing inside `/eval`).
- Report:
  - accuracy
  - average reasoning quality
  - confidence calibration (`low` / `mid` / `high`)
  - per-ticker results

