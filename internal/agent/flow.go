package agent

import (
	"ai-agent/internal/llm"
	"ai-agent/internal/model"
	"ai-agent/internal/rag"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type LLMClient interface {
	Generate(ctx context.Context, prompt string) (model.AnalysisOutput, string, error)
}

type Retriever interface {
	Retrieve(ctx context.Context, query, ticker string) ([]string, error)
}

type Runner struct {
	retriever Retriever
	llm       LLMClient
}

func NewRunner(retriever *rag.Retriever, llmClient *llm.Client) *Runner {
	return &Runner{
		retriever: retriever,
		llm:       llmClient,
	}
}

func (r *Runner) Run(ctx context.Context, ticker string) (RunResult, error) {
	query := fmt.Sprintf("Analyze %s stock and decide BUY, HOLD, or SELL", strings.ToUpper(ticker))
	return r.RunWithQuery(ctx, ticker, query)
}

// RunWithQuery runs retrieval and analysis using the given question for embedding and prompts.
// ticker scopes vector search to that symbol's facts.
func (r *Runner) RunWithQuery(ctx context.Context, ticker, query string) (RunResult, error) {
	ticker = strings.TrimSpace(ticker)
	query = strings.TrimSpace(query)
	if ticker == "" {
		return RunResult{}, fmt.Errorf("ticker is required")
	}
	if query == "" {
		return RunResult{}, fmt.Errorf("question is required")
	}

	docs, err := r.retriever.Retrieve(ctx, query, ticker)
	if err != nil {
		return RunResult{}, err
	}

	tools := RunToolSummary(docs)
	initialPrompt := buildReasoningPrompt(query, docs, tools)
	initial, _, err := r.llm.Generate(ctx, initialPrompt)
	if err != nil {
		return RunResult{}, err
	}

	reflectionPrompt := buildReflectionPrompt(initial, docs, tools)
	reflected, _, err := r.llm.Generate(ctx, reflectionPrompt)
	if err != nil {
		reflected = initial
	}

	final, replaced, reason := ChooseFinal(initial, reflected)
	return RunResult{
		Query:               query,
		Ticker:              strings.ToUpper(ticker),
		RetrievedDocs:       docs,
		ToolSummary:         tools,
		Initial:             initial,
		Reflected:           reflected,
		Final:               final,
		ReflectionReplaced:  replaced,
		ReflectionRationale: reason,
	}, nil
}

func buildReasoningPrompt(query string, docs []string, tools ToolSummary) string {
	docsText := strings.Join(docs, "\n- ")
	toolJSON, _ := json.MarshalIndent(tools, "", "  ")
	return fmt.Sprintf(`You are a financial analyst.

Query:
%s

Context:
- %s

Tool outputs:
%s

Task:
Use the context to address the query above, then give a BUY, HOLD, or SELL recommendation.

Return ONLY valid JSON:
{
  "decision": "BUY|HOLD|SELL",
  "reasoning": "short explanation",
  "confidence": 0.0
}`, query, docsText, string(toolJSON))
}

func buildReflectionPrompt(initial model.AnalysisOutput, docs []string, tools ToolSummary) string {
	initialJSON, _ := json.MarshalIndent(initial, "", "  ")
	docsText := strings.Join(docs, "\n- ")
	toolJSON, _ := json.MarshalIndent(tools, "", "  ")
	return fmt.Sprintf(`Review the previous stock analysis.

Previous output:
%s

Context:
- %s

Tool outputs:
%s

Check for contradictions, missing data, and weak reasoning relative to the query.
If needed, revise the decision.

Return ONLY valid JSON:
{
  "decision": "BUY|HOLD|SELL",
  "reasoning": "short explanation",
  "confidence": 0.0
}`, string(initialJSON), docsText, string(toolJSON))
}
