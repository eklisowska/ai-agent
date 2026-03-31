# Synthetic Stock Analyst Agent

This is a **Go backend API** that demonstrates an **agentic stock analyst** on synthetic data. Given a ticker (like `AAPL`), it returns `BUY`, `HOLD`, or `SELL` using:

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

1. Pull Ollama models on your host:

```bash
ollama pull llama3
ollama pull nomic-embed-text
```

2. Start the stack (Qdrant + API container; Ollama remains on host):

```bash
docker compose up -d --build
```

3. Check health:

```bash
curl -s http://localhost:8080/health
```

4. Index data

```bash
curl -s -X POST http://localhost:8080/index
```

5. Make calls to API:
   - Use `POST /analyze` for a default quick recommendation (input: `ticker`).
   - Use `POST /ask` for question-driven analysis (input: `ticker` + `question`).
   - Shortcut: if you have a specific question, use `/ask`; otherwise use `/analyze`.

   `/analyze` example:

```bash
curl -s -X POST http://localhost:8080/analyze \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AAPL"}'
```

`/ask` examples (specific, question-driven):

```bash
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"AAPL","question":"Is valuation stretched vs peers?"}'
```

Question about downside risk:

```bash
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"TSLA","question":"What is the biggest downside risk over the next 2 quarters?"}'
```

Question about conflicting signals:

```bash
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"NVDA","question":"Revenue growth is strong but sentiment is mixed. Should I still rate it BUY?"}'
```

Question requesting a conservative view:

```bash
curl -s -X POST http://localhost:8080/ask \
  -H 'Content-Type: application/json' \
  -d '{"ticker":"MSFT","question":"Give a conservative recommendation focused on risk control."}'
```

6. Run end-to-end evaluation over labeled tickers and report accuracy, reasoning-quality score, and confidence calibration buckets.

```bash
curl -s -X POST http://localhost:8080/eval
```

## ReAct Workflow Documentation

For a concise end-to-end breakdown of the agent loop, see:

- `[agentic_worklow_react.md](./agentic_worklow_react.md)`

That document explains how this project implements a cohesive ReAct-style system across:

- retrieval (RAG indexing + search),
- action execution (deterministic tools),
- reasoning (LLM initial pass),
- self-reflection (LLM revision pass),
- and evaluation (accuracy, reasoning quality, confidence calibration).

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

