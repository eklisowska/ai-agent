# AI Stock Analyst Agent

This is a **Go backend API** that demonstrates an **agentic stock analyst** on synthetic data. Given a ticker that exists in the indexed dataset (the bundled data covers `**AAPL`**, `**AMZN`**, and `**NVDA**` only), it returns `BUY`, `HOLD`, or `SELL` using:

- Retrieval-augmented context from financial, news, and profile data
- Tool-calling for sentiment, risk, and valuation signals
- A reflection pass that can revise the first answer

## How it is built


| Area             | Choice                                              |
| ---------------- | --------------------------------------------------- |
| Language         | Go 1.26                                             |
| HTTP             | `chi` — `POST /index`, `POST /analyze`, `POST /ask` |
| LLM & embeddings | Ollama (`llama3`, `nomic-embed-text`)               |
| Vector store     | Qdrant                                              |


**Data:** JSON under `data/raw/` — `profiles.json`, `financials.json`, and `news.json` — are normalized into fact documents, embedded, and indexed in Qdrant.

**Diagrams:** See `[architecture.mmd](./architecture.mmd)` for the high-level flow (Mermaid). Preview it in VS Code with a Mermaid extension, in GitHub if your viewer supports `.mmd`, or paste into [mermaid.live](https://mermaid.live).

**Narrative walkthrough:** `[agentic_workflow_react.md](./agentic_workflow_react.md)` describes the ReAct-style loop (retrieve → act → reason → reflect).

## Configuration (environment)

Loaded at startup (`internal/config`). Optional `.env` in the working directory, or set `ENV_FILE` to a specific path.


| Variable            | Role                                              | Default                  |
| ------------------- | ------------------------------------------------- | ------------------------ |
| `QDRANT_URL`        | Qdrant HTTP API                                   | `http://localhost:6333`  |
| `OLLAMA_URL`        | Ollama API                                        | `http://localhost:11434` |
| `LLM_MODEL`         | Chat model name                                   | `llama3`                 |
| `EMBED_MODEL`       | Embedding model name                              | `nomic-embed-text`       |
| `QDRANT_COLLECTION` | Collection name                                   | `stock_facts`            |
| `TOP_K`             | Retrieved chunks per query                        | `8`                      |
| `HTTP_TIMEOUT`      | Client timeouts (Go duration, e.g. `30s`, `600s`) | `30s`                    |
| `LISTEN_ADDR`       | Bind address                                      | `:8080`                  |


`docker-compose.yml` sets these for the `agent` service (Ollama is expected on the host, reachable via `host.docker.internal`).

## Repository layout


| Path                       | Purpose                                                   |
| -------------------------- | --------------------------------------------------------- |
| `cmd/server/`              | HTTP server entrypoint                                    |
| `internal/agent/`          | Run loop, tools, reflection, `RunResult` types            |
| `internal/rag/`            | Load raw JSON, embed, Qdrant index & retrieve             |
| `internal/llm/`            | Ollama chat + JSON analysis parsing                       |
| `internal/model/`          | Structured analysis output (`decision`, `reasoning`, …)   |
| `internal/config/`         | Environment-based configuration                           |
| `data/raw/`                | Source JSON for indexing                                  |
| `scripts/generate_data.go` | Optional helper to regenerate synthetic `data/raw/*.json` |


## Quick start (Docker-first)

1. Start Ollama on your host:

```bash
ollama serve
```

1. Pull models:

```bash
ollama pull llama3
ollama pull nomic-embed-text
```

1. Start Qdrant and the API (Ollama stays on the host; see `extra_hosts` in `docker-compose.yml`):

```bash
docker compose up -d --build
```

1. Check liveness:

```bash
curl -s http://localhost:8080/health
```

1. Index documents (required before analysis):

```bash
curl -s -X POST http://localhost:8080/index
```

1. Call the API:

- `POST /analyze` — default recommendation; JSON body: `{ "ticker": "AAPL" }` (use a ticker present in `data/raw/`).
- `POST /ask` — question-driven analysis; JSON body: `{ "ticker": "AAPL", "question": "…" }`.
 Example — `/analyze`:

```bash
curl -s -X POST http://localhost:8080/analyze \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AAPL"}'
```

   Examples — `/ask`:

```bash
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AAPL","question":"Is valuation stretched vs peers?"}'
```

```bash
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AMZN","question":"What is the biggest downside risk over the next 2 quarters?"}'
```

```bash
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"NVDA","question":"Revenue growth is strong but sentiment is mixed. Should I still rate it BUY?"}'
```

```bash
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AMZN","question":"Give a conservative recommendation focused on risk control."}'
```

## Requirements coverage (agentic components)

1. **Data preparation & contextualization** — Raw JSON (`profiles`, `financials`, `news`) is turned into normalized fact documents and ticker-scoped text for retrieval.
2. **RAG pipeline** — Facts are embedded with Ollama, stored in Qdrant; retrieval uses query embedding, ticker filter, and top-k search.
3. **Reasoning & reflection** — The agent produces structured output, then a second reflection pass may revise the answer (`ChooseFinal`).
4. **Tool-calling** — Deterministic helpers compute sentiment score, risk keywords, and P/E band; results are injected into prompts.

