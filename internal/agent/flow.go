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
	if len(docs) == 0 {
		noData := model.AnalysisOutput{
			Reasoning:  fmt.Sprintf("No context data found for ticker %s.", strings.ToUpper(ticker)),
			Confidence: 0,
		}
		return RunResult{
			Query:               query,
			Ticker:              strings.ToUpper(ticker),
			NoData:              true,
			NoDataReason:        noData.Reasoning,
			RetrievedDocs:       []string{},
			ToolSummary:         ToolSummary{},
			Initial:             noData,
			Reflected:           noData,
			Final:               noData,
			ReflectionReplaced:  false,
			ReflectionRationale: "analysis skipped: no retrieval context",
		}, nil
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
	final, policyApplied := applyPolicyGuardrail(final, tools)
	if policyApplied {
		reason = reason + "; policy guardrail applied"
	}
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
If the context is empty or does not contain ticker-specific facts, respond that there is no data for the ticker instead of inferring.
Prefer this policy unless context strongly contradicts it:
- composite_score >= 0.7 -> BUY
- composite_score <= -0.9 -> SELL
- otherwise -> HOLD

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
Keep decisions consistent with tool outputs and composite_score thresholds unless there is explicit contradictory evidence.

Return ONLY valid JSON:
{
  "decision": "BUY|HOLD|SELL",
  "reasoning": "short explanation",
  "confidence": 0.0
}`, string(initialJSON), docsText, string(toolJSON))
}

func applyPolicyGuardrail(in model.AnalysisOutput, tools ToolSummary) (model.AnalysisOutput, bool) {
	// Must match thresholds in buildReasoningPrompt (composite_score policy).
	// Sell threshold is stricter than buy so mixed/risk-heavy names stay in HOLD unless very bearish.
	const strongBuy = 0.7
	const strongSell = -0.9
	switch {
	case tools.CompositeScore >= strongBuy && in.Decision != model.DecisionBuy:
		in.Decision = model.DecisionBuy
		if in.Confidence < 0.65 {
			in.Confidence = 0.65
		}
		if !strings.Contains(strings.ToLower(in.Reasoning), "composite") {
			in.Reasoning += " Composite score and supporting signals indicate stronger upside bias."
		}
		return in, true
	case tools.CompositeScore <= strongSell && in.Decision != model.DecisionSell:
		in.Decision = model.DecisionSell
		if in.Confidence < 0.65 {
			in.Confidence = 0.65
		}
		if !strings.Contains(strings.ToLower(in.Reasoning), "composite") {
			in.Reasoning += " Composite score and risk/sentiment balance indicate stronger downside bias."
		}
		return in, true
	default:
		return in, false
	}
}
