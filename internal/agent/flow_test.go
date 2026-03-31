package agent

import (
	"ai-agent/internal/model"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "Run error")
	require.NotEmpty(t, out.Final.Decision, "expected final decision")
	require.Equal(t, 2, llm.call, "expected reflection pass")
}

func TestRunnerRunWithQuery(t *testing.T) {
	llm := &fakeLLM{}
	r := Runner{
		retriever: fakeRetriever{},
		llm:       llm,
	}
	out, err := r.RunWithQuery(context.Background(), "AAPL", "What is the main risk for this name?")
	require.NoError(t, err, "RunWithQuery error")
	require.Equal(t, "What is the main risk for this name?", out.Query, "unexpected query")
	require.Equal(t, 2, llm.call, "expected reflection pass")
}
