# Synthetic Stock Analyst Agent (Go + RAG + Agent Loop)

## 🧠 Overview

This project implements a **simple but complete agentic AI system** in Go that performs synthetic stock analysis and outputs a **BUY / HOLD / SELL** decision.

The system demonstrates all core components of an agentic pipeline:

- RAG (Retrieval-Augmented Generation)
- Reasoning (LLM-based analysis)
- Tool calling (external computations)
- Reflection (self-correction loop)
- Evaluation (basic scoring)

The design prioritizes:

- Simplicity
- Determinism (via synthetic data)
- Minimal dependencies
- Go-first architecture

---

# 🎯 Goals

- Build a fully working agent with:
  - Context retrieval (RAG)
  - Multi-step reasoning
  - Tool usage
  - Reflection pass
- Keep everything in a **single Go service**
- Avoid overengineering (no microservices, no Python)

---

# 🧱 Project Structure

```
ai-agent/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── agent/
│   │   ├── flow.go          # main agent loop
│   │   ├── tools.go         # tool implementations
│   │   ├── reflection.go    # reflection logic
│   ├── rag/
│   │   ├── index.go         # indexing synthetic data
│   │   ├── retrieve.go      # vector search
│   │   ├── embed.go         # embeddings
│   ├── llm/
│   │   └── client.go        # LLM wrapper
│   ├── eval/
│   │   └── evaluator.go     # evaluation logic
│   └── config/
│       └── config.go
├── data/
│   ├── raw/                 # synthetic JSON data
│   └── processed/           # chunked data
├── scripts/
│   └── generate_data.go     # optional synthetic generator
├── go.mod
└── README.md
```

---

# ⚙️ Tech Stack

## Core

- Go (1.26)

## LLM

- Local ollama (llama3)

## Vector Database

- Qdrant (local via Docker)

## Embeddings

- Local Ollama embeddings (nomic-embed-text)

---

# 🧠 Agent Architecture

## High-Level Flow

```
User Input (Ticker)
    ↓
RAG Retrieval (Qdrant)
    ↓
LLM Reasoning (THINK)
    ↓
Tool Decision (DECIDE)
    ↓
Tool Execution (ACT)
    ↓
Updated Context
    ↓
Final Decision
    ↓
Reflection Pass
    ↓
Output
```

---

# 🧩 1. Data Preparation & Contextualization

## Data Types (Synthetic)

### 1. Company Profile

```json
{
  "ticker": "AAPL",
  "type": "profile",
  "content": "Apple designs consumer electronics...",
  "risks": "Dependence on iPhone sales"
}
```

### 2. Financial Data

```json
{
  "ticker": "AAPL",
  "type": "financial",
  "pe_ratio": 28,
  "revenue_growth": 0.05,
  "debt_to_equity": 1.5
}
```

### 3. News Headlines

```json
{
  "ticker": "AAPL",
  "type": "news",
  "headline": "Apple faces declining demand in China",
  "sentiment": "negative"
}
```

### 4. Ground Truth (for evaluation)

```json
{
  "ticker": "AAPL",
  "expected": "HOLD"
}
```

---

## Chunking Strategy

- Each fact = one document
- Examples:
  - "AAPL PE ratio is 28"
  - "News: Apple demand declining (negative)"

---

## Embedding Pipeline

```
raw data → chunk → embed → store in Qdrant
```

---

# 🔍 2. RAG Pipeline Design

## Retrieval Strategy

- Filter by ticker
- Top-K = 5–10 documents

## Query Example

```
Analyze AAPL stock and decide BUY, HOLD, or SELL
```

## Output of Retrieval

```
[]string (context documents)
```

---

# 🧠 3. Reasoning & Reflection

## Step 1 — Reasoning Prompt

```
You are a financial analyst.

Context:
{retrieved_docs}

Task:
Analyze the stock and decide BUY, HOLD, or SELL.

Return JSON:
{
  "decision": "...",
  "reasoning": "...",
  "confidence": 0.0-1.0
}
```

---

## Step 2 — Reflection Prompt

```
Review the previous analysis.

Check for:
- contradictions
- missing data
- weak reasoning

If needed, revise the decision.

Return final JSON.
```

---

## Reflection Logic

- Run second LLM pass
- Compare outputs
- Replace decision if improved

---

# 🛠️ 4. Tool-Calling Mechanisms

## Tool Interface

```go
type Tool interface {
    Name() string
    Execute(input string) (string, error)
}
```

---

## Tool 1: Sentiment Score

```go
func SentimentScore(headlines []string) float64
```

- Input: list of headlines
- Output: score (-1 to 1)

---

## Tool 2: Risk Detector

```go
func DetectRisk(text string) []string
```

- Extract risk keywords

---

## Tool 3: Ratio Analyzer

```go
func AnalyzePE(pe float64) string
```

- Output:
  - "undervalued"
  - "fair"
  - "overvalued"

---

## Tool Calling Strategy

### Option A (Simple)

- Always call tools before LLM

### Option B (Advanced)

- LLM decides:

```json
{
  "tool": "SentimentScore",
  "input": [...]
}
```

---

# 🔁 5. Agent Loop Implementation

## flow.go

Pseudo-code:

```go
func Run(query string) Output {
    docs := rag.Retrieve(query)

    toolResults := runTools(docs)

    prompt := buildPrompt(query, docs, toolResults)

    initial := llm.Generate(prompt)

    final := reflect(initial)

    return final
}
```

---

# 📊 6. Evaluation

## Metric 1 — Accuracy

```
accuracy = correct_predictions / total
```

---

## Metric 2 — Reasoning Quality

Check:

- mentions financial data
- references sentiment
- coherent explanation

---

## Metric 3 — Confidence Calibration

Compare:

- confidence vs correctness

---

## Evaluation Script

```go
func Evaluate(predicted, expected string) bool {
    return predicted == expected
}
```

---

# ⚙️ 7. Infrastructure Setup

## Qdrant

```bash
docker run -p 6333:6333 qdrant/qdrant
```

---

## Environment Variables

```
QDRANT_URL=http://localhost:6333
```

---

# 🧪 8. Testing Strategy

## Unit Tests

- tools
- retrieval

## Integration Tests

- full agent run

## Dataset

- 5–10 synthetic tickers

---

# 🚀 9. Development Phases

## Phase 1 — MVP

- basic RAG
- single LLM call
- static data

---

## Phase 2 — Agent

- add tools
- structured output

---

## Phase 3 — Reflection

- second LLM pass

---

## Phase 4 — Evaluation

- scoring system

---

# ⚠️ 10. Constraints & Simplifications

- No real-time APIs
- No microservices
- No Python
- No multi-agent system

---

# 🏁 Final Notes

This project is designed to:

- Demonstrate full agentic pipeline
- Stay simple and understandable
- Be fully implementable in Go

Focus on:

- clean structure
- deterministic data
- iterative improvements

Avoid:

- premature abstraction
- unnecessary frameworks
- distributed complexity

