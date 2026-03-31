package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSentimentScore(t *testing.T) {
	score := SentimentScore([]string{
		"Cloud revenue growth beats expectations",
		"Company faces declining demand",
	})
	require.NotZero(t, score, "expected non-zero sentiment score")
}

func TestAnalyzePE(t *testing.T) {
	require.Equal(t, "undervalued", AnalyzePE(10), "AnalyzePE undervalued")
	require.Equal(t, "fair", AnalyzePE(20), "AnalyzePE fair")
	require.Equal(t, "overvalued", AnalyzePE(40), "AnalyzePE overvalued")
}

func TestDetectRisk(t *testing.T) {
	risks := DetectRisk("Regulation risk and debt pressure are rising.")
	require.NotEmpty(t, risks, "expected at least one risk")
}
