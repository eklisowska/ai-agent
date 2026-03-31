# Synthetic Stock Analyst Agent (Go)

Agentic Go app that demonstrates an agentic pipeline:

- RAG over synthetic stock facts in Qdrant
- Tool-assisted reasoning
- Reflection pass
- Evaluation on synthetic ground truth

## Requirements

- Go 1.26.x (latest stable line)
- Docker (for Qdrant)
- Ollama with models:
  - `llama3`
  - `nomic-embed-text`

## Quick Start

1. Start Qdrant:

```bash
docker run -d --name qdrant -p 6333:6333 qdrant/qdrant
```

1. Pull Ollama models:

```bash
ollama pull llama3
ollama pull nomic-embed-text
```

1. Configure environment with a `.env` file:

From the project root:

```bash
cp .env.example .env
# edit .env with your URLs, models, and optional overrides (see Environment variables)
```

The binary loads `.env` from the **current working directory** when you run `go run` or `./agent`. 

1. Index synthetic data (run from the project root):

```bash
go run ./cmd/server index
```

1. Analyze a ticker:

```bash
go run ./cmd/server analyze AAPL
```

1. Ask a custom question (same pipeline; your text drives retrieval and prompts):

```bash
go run ./cmd/server ask AAPL What are the main risks given the current narrative?
```

1. Run evaluation:

```bash
go run ./cmd/server eval
```

1. Optional — HTTP API (`POST /ask`, `GET /health`, `GET /ready`). Set `LISTEN_ADDR` in `.env` if you need something other than the default `:8080`, then:

```bash
go run ./cmd/server serve
```

```bash
curl -s http://localhost:8080/health
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AAPL","question":"Is valuation stretched vs peers?"}'
```

## Environment variables


| Variable            | Default                  | Description                                                                                                    |
| ------------------- | ------------------------ | -------------------------------------------------------------------------------------------------------------- |
| `QDRANT_URL`        | `http://localhost:6333`  | Qdrant HTTP API                                                                                                |
| `OLLAMA_URL`        | `http://localhost:11434` | Ollama API                                                                                                     |
| `LLM_MODEL`         | `llama3`                 | Chat model                                                                                                     |
| `EMBED_MODEL`       | `nomic-embed-text`       | Embedding model                                                                                                |
| `QDRANT_COLLECTION` | `stock_facts`            | Qdrant collection name                                                                                         |
| `TOP_K`             | `8`                      | Retrieval hits                                                                                                 |
| `HTTP_TIMEOUT`      | `30s`                    | Per-request HTTP timeout (Go duration). For larger models like `llama3`, `120s`–`600s` is often more practical (especially for `eval`). |
| `LISTEN_ADDR`       | `:8080`                  | Bind address for `serve`                                                                                       |
| `ENV_FILE`          | *(unset)*                | If set, load this path instead of `.env` (file must exist)                                                     |


## Tests

```bash
go test ./...
```

## Docker (Qdrant + Agent, host Ollama)

This setup runs only Qdrant and the Go agent in containers. Ollama stays local on your host machine.

Prerequisites:

- Ollama running on host (`http://localhost:11434`)
- Models pulled locally:
  - `ollama pull llama3`
  - `ollama pull nomic-embed-text`

Run the API stack (Qdrant + long-running `serve`) with:

```bash
docker compose up
```

Or detached:

```bash
docker compose up -d
```

Check liveness/readiness:

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/ready
```

Run one-off jobs without replacing the API container:

```bash
docker compose run --rm --no-deps agent-index
docker compose run --rm --no-deps agent-eval
docker compose run --rm --no-deps agent analyze AAPL
docker compose run --rm --no-deps agent ask AAPL "What is the sentiment picture?"
```

API behavior:

- `GET /health` is liveness-only (`alive`).
- `GET /ready` checks Qdrant and Ollama reachability and returns `503` while dependencies are unavailable.

Notes:

- `docker-compose.yml` sets `OLLAMA_URL=http://host.docker.internal:11434` for the agent container.
- On Linux, `extra_hosts: host.docker.internal:host-gateway` is included so containers can reach host Ollama.
- If Ollama is not running on host, `agent` stays alive but `GET /ready` reports not ready.

## Commands

- `index`: embeds and upserts all facts from `data/raw/`.
- `analyze <TICKER>`: runs retrieval, tools, reasoning, reflection (fixed prompt).
- `ask <TICKER> <question...>`: same as analyze, but the question is your natural-language query (retrieval + prompts use it).
- `serve`: HTTP server; `POST /ask` with JSON `{"ticker":"...","question":"..."}`, `GET /health`, `GET /ready`.
- `eval`: indexes data and evaluates all tickers against `ground_truth.json`.

