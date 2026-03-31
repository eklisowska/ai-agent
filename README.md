# Synthetic Stock Analyst Agent

Agentic AI demo in Go that analyzes a stock ticker and returns `BUY`, `HOLD`, or `SELL` using:
- Retrieval-augmented context from synthetic financial/news/profile data
- Lightweight tool-calling for sentiment/risk/valuation signals
- A reflection pass that can revise the first answer
- Simple evaluation against labeled ground truth

## How It Was Built

- **Language/runtime:** Go 1.26
- **HTTP API:** `chi` router (`/index`, `/analyze`, `/ask`, `/eval`, `/health`, `/ready`)
- **LLM + embeddings:** Ollama (`llama3`, `nomic-embed-text`)
- **Vector store:** Qdrant
- **Data source:** `data/raw/*.json` synthetic dataset

See `architecture.mmd` for architecture and run flow.

## Quick Start (Docker-first)

1) Pull Ollama models on your host:

```bash
ollama pull llama3
ollama pull nomic-embed-text
```

2) Start the stack (Qdrant + API container; Ollama remains on host):

```bash
docker compose up -d --build
```

3) Check health/readiness:

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/ready
```

4) Index data and run:

```bash
curl -s -X POST http://localhost:8080/index
curl -s -X POST http://localhost:8080/analyze \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AAPL"}'
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AAPL","question":"Is valuation stretched vs peers?"}'
curl -s -X POST http://localhost:8080/eval
```

Stop services:

```bash
docker compose down
```

## Requirements Coverage (Agentic Components)

1. **Data Preparation & Contextualization**  
   Raw JSON files (`profiles`, `financials`, `news`) are transformed into normalized fact documents and ticker-scoped context.

2. **RAG Pipeline Design**  
   Facts are embedded with Ollama embeddings and indexed in Qdrant; retrieval uses query embedding + ticker filter + top-k search.

3. **Reasoning & Reflection**  
   The agent first generates a structured analysis, then runs a second reflection prompt to self-check and optionally replace the initial answer.

4. **Tool-Calling Mechanisms**  
   Deterministic tools compute sentiment score, risk keyword detection, and P/E valuation class; outputs are injected into prompts.

5. **Evaluation**  
   `/eval` runs end-to-end over labeled tickers and reports accuracy, reasoning-quality score, and confidence calibration buckets.
