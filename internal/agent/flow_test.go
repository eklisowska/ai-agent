package agent

import (
	"ai-agent/internal/model"
	"context"
	"testing"
)

type fakeRetriever struct{}

func (fakeRetriever) Retrieve(context.Context, string, string) ([]string, error) {
	return []string{
		"AAPL PE ratio is 28",
		"News for AAPL: Apple faces declining demand in China (sentiment: negative)",
	}, nil
}

type fakeLLM struct {
	call int
}

func (f *fakeLLM) Generate(_ context.Context, _ string) (model.AnalysisOutput, string, error) {
	f.call++
	if f.call == 1 {
		return model.AnalysisOutput{
			Decision:   model.DecisionHold,
			Reasoning:  "PE is fair and sentiment is mixed.",
			Confidence: 0.60,
		}, "", nil
	}
	return model.AnalysisOutput{
		Decision:   model.DecisionHold,
		Reasoning:  "PE is fair, sentiment is negative, so hold with caution.",
		Confidence: 0.75,
	}, "", nil
}

func TestRunnerRun(t *testing.T) {
	llm := &fakeLLM{}
	r := Runner{
		retriever: fakeRetriever{},
		llm:       llm,
	}
	out, err := r.Run(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if out.Final.Decision == "" {
		t.Fatalf("expected final decision")
	}
	if llm.call != 2 {
		t.Fatalf("expected reflection pass, calls=%d", llm.call)
	}
}

func TestRunnerRunWithQuery(t *testing.T) {
	llm := &fakeLLM{}
	r := Runner{
		retriever: fakeRetriever{},
		llm:       llm,
	}
	out, err := r.RunWithQuery(context.Background(), "AAPL", "What is the main risk for this name?")
	if err != nil {
		t.Fatalf("RunWithQuery error: %v", err)
	}
	if out.Query != "What is the main risk for this name?" {
		t.Fatalf("unexpected query: %q", out.Query)
	}
	if llm.call != 2 {
		t.Fatalf("expected reflection pass, calls=%d", llm.call)
	}
}
