# Agentic workflow (ReAct)

## Objective

Demonstrate a self-reflective agentic system that combines **retrieval**, deterministic **tools (Act)**, LLM **reasoning**, and **reflection** into one workflow that returns `BUY`, `HOLD`, or `SELL`.

There is **no separate batch-evaluation HTTP route**; quality checks live in unit tests and manual review of `RunResult` JSON.

## ReAct loop in this project

1. **Retrieve** — Ticker-scoped facts from vector search (Qdrant).
2. **Act** — Deterministic signals: sentiment score, risk keywords, P/E band (`internal/agent/tools.go`).
3. **Reason** — First LLM pass: structured JSON (`decision`, `reasoning`, `confidence`).
4. **Reflect** — Second LLM pass + policy (`ChooseFinal` in `internal/agent/reflection.go`) to accept or revise.

## Flow

### 0) Bootstrap

**Files:** `cmd/server/main.go`, `internal/config/config.go`

- Load configuration from the environment (optional `.env` or `ENV_FILE`).
- Construct Qdrant client, embedder, LLM client, retriever, indexer, and `agent.Runner`.
- Register routes: `POST /index`, `POST /analyze`, `POST /ask`, `GET /health`, `GET /ready`.

### 1) Prepare and index context

**Files:** `internal/rag/index.go`, `internal/rag/embed.go`, `internal/rag/qdrant.go`, `data/raw/*.json`

- Read `profiles.json`, `financials.json`, `news.json` and build `FactDocument` records.
- Embed each fact and upsert into the configured Qdrant collection (`POST /index`).

### 2) Run analysis

**Files:** `internal/agent/flow.go`, `internal/rag/retrieve.go`, `internal/agent/tools.go`, `internal/llm/client.go`, `internal/agent/reflection.go`

- **`/analyze`** — Default query text for embedding/prompts (ticker-only flow).
- **`/ask`** — User question drives embedding and prompt framing.
- Retrieve top-k chunks filtered by ticker; run tools; build prompts; parse LLM JSON; run reflection; apply guardrails where applicable.

### 3) Return auditable output

**Files:** `internal/agent/types.go`

- Response is `RunResult`: retrieved snippets, tool summary, initial and reflected model outputs, final choice, and reflection rationale.

## Diagram

See [`architecture.mmd`](./architecture.mmd) for a compact Mermaid flowchart (index path vs ReAct path).
