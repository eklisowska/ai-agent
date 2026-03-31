package llm

import (
	"ai-agent/internal/model"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAnalysis(t *testing.T) {
	raw := `{
	  "decision":"BUY",
	  "reasoning":"PE and news sentiment indicate upside.",
	  "confidence":0.76
	}`
	out, err := ParseAnalysis(raw)
	require.NoError(t, err, "ParseAnalysis error")
	require.Equal(t, model.DecisionBuy, out.Decision, "unexpected decision")
}

func TestParseAnalysisInvalid(t *testing.T) {
	raw := `{"decision":"MAYBE","reasoning":"x","confidence":0.2}`
	_, err := ParseAnalysis(raw)
	require.Error(t, err, "expected error for invalid decision")
}
