package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSentimentScore(t *testing.T) {
	score := SentimentScore([]string{
		"News for MSFT: Cloud revenue growth beats expectations (sentiment: positive, date: 2026-03-10)",
		"News for AAPL: Company faces declining demand (sentiment: negative, date: 2026-03-11)",
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
