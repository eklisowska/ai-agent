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

1) Start Qdrant:

```bash
docker run -d --name qdrant -p 6333:6333 qdrant/qdrant
```

2) Pull Ollama models:

```bash
ollama pull llama3
ollama pull nomic-embed-text
```

3) Configure environment with a `.env` file:

From the project root:

```bash
cp .env.example .env
# edit .env with your URLs, models, and optional overrides (see Environment variables)
```

The binary loads `.env` from the **current working directory** when you run `go run` or `./agent`. To use a different path, set the `ENV_FILE` variable for that process (IDE run config, systemd, container, etc.).

4) Index synthetic data (from the project root, with `.env` loaded):

```bash
ENV_FILE=.env go run ./cmd/server index
```

5) Analyze a ticker:

```bash
ENV_FILE=.env go run ./cmd/server analyze AAPL
```

6) Ask a custom question (same pipeline; your text drives retrieval and prompts):

```bash
ENV_FILE=.env go run ./cmd/server ask AAPL What are the main risks given the current narrative?
```

7) Run evaluation:

```bash
ENV_FILE=.env go run ./cmd/server eval
```

8) Optional — HTTP API (`POST /ask`, `GET /health`). Set `LISTEN_ADDR` in `.env` if you need something other than the default `:8080`, then:

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

| Variable | Default | Description |
|----------|---------|-------------|
| `QDRANT_URL` | `http://localhost:6333` | Qdrant HTTP API |
| `OLLAMA_URL` | `http://localhost:11434` | Ollama API |
| `LLM_MODEL` | `llama3` | Chat model |
| `EMBED_MODEL` | `nomic-embed-text` | Embedding model |
| `QDRANT_COLLECTION` | `stock_facts` | Qdrant collection name |
| `TOP_K` | `8` | Retrieval hits |
| `HTTP_TIMEOUT` | `30s` | Per-request HTTP timeout (Go duration). For larger models like `llama3`, `60s`–`120s` is often more practical. |
| `LISTEN_ADDR` | `:8080` | Bind address for `serve` |
| `ENV_FILE` | _(unset)_ | If set, load this path instead of `.env` (file must exist) |

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

Run Qdrant and the agent with a single command (agent runs `analyze AAPL` by default, as configured in `docker-compose.yml`):

```bash
docker compose up
```

Or detached:

```bash
docker compose up -d
```

Other commands using the same compose setup:

```bash
docker compose run --rm agent index
docker compose run --rm agent analyze AAPL
docker compose run --rm agent ask AAPL "What is the sentiment picture?"
docker compose run --rm agent eval
```

HTTP server from the agent container: put `LISTEN_ADDR=0.0.0.0:8080` in your `.env` (needed so the port is reachable from the host), mount the file, map the port:

```bash
docker compose run --rm -p 8080:8080 \
  -v "$PWD/.env:/app/.env:ro" -e ENV_FILE=/app/.env \
  agent serve
```

Notes:
- `docker-compose.yml` sets `OLLAMA_URL=http://host.docker.internal:11434` for the agent container.
- On Linux, `extra_hosts: host.docker.internal:host-gateway` is included so containers can reach host Ollama.

## Commands

- `index`: embeds and upserts all facts from `data/raw/`.
- `analyze <TICKER>`: runs retrieval, tools, reasoning, reflection (fixed prompt).
- `ask <TICKER> <question...>`: same as analyze, but the question is your natural-language query (retrieval + prompts use it).
- `serve`: HTTP server; `POST /ask` with JSON `{"ticker":"...","question":"..."}`, `GET /health`.
- `eval`: indexes data and evaluates all tickers against `ground_truth.json`.
