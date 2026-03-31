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

func TestDetectRiskIgnoresDebtToEquityPhrase(t *testing.T) {
	risks := DetectRisk("AAPL debt-to-equity ratio is 1.50")
	require.Empty(t, risks, "metric line should not count as narrative debt risk")
}

// Regression: composite_score policy in prompts/guardrails should match labeled eval tickers.
func TestRunToolSummaryEvalTickers(t *testing.T) {
	aapl := []string{
		"AAPL PE ratio is 28.00",
		"AAPL revenue growth is 5.00%",
		"AAPL debt-to-equity ratio is 1.50",
		"News for AAPL: Apple faces declining demand in China (sentiment: negative, date: 2026-03-26)",
		"News for AAPL: Apple services segment posts resilient growth (sentiment: positive, date: 2026-03-19)",
	}
	amzn := []string{
		"AMZN PE ratio is 45.00",
		"AMZN revenue growth is 12.00%",
		"AMZN debt-to-equity ratio is 0.45",
		"News for AMZN: Amazon expands fulfillment automation and lowers unit costs (sentiment: positive, date: 2026-03-22)",
		"News for AMZN: AWS closes several multi-year enterprise AI contracts (sentiment: positive, date: 2026-03-10)",
	}
	nvda := []string{
		"NVDA PE ratio is 38.00",
		"NVDA revenue growth is 35.00%",
		"NVDA debt-to-equity ratio is 0.30",
		"News for NVDA: NVIDIA secures major AI datacenter orders (sentiment: positive, date: 2026-03-24)",
		"News for NVDA: GPU supply constraints ease as packaging capacity expands (sentiment: positive, date: 2026-03-18)",
	}
	a := RunToolSummary(aapl)
	require.Greater(t, a.CompositeScore, -0.7, "AAPL expected HOLD band")
	require.Less(t, a.CompositeScore, 0.7)

	m := RunToolSummary(amzn)
	require.GreaterOrEqual(t, m.CompositeScore, 0.7, "AMZN expected BUY threshold")

	n := RunToolSummary(nvda)
	require.GreaterOrEqual(t, n.CompositeScore, 0.7, "NVDA expected BUY threshold")
}
