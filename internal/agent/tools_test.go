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

// Regression: composite_score policy in prompts/guardrails for AAPL, AMZN, NVDA tool summaries.
func TestRunToolSummaryDatasetTickers(t *testing.T) {
	aapl := []string{
		"AAPL PE ratio is 26.00",
		"AAPL revenue growth is 4.00%",
		"AAPL debt-to-equity ratio is 1.20",
		"News for AAPL: Apple faces declining demand in China (sentiment: negative, date: 2026-03-28)",
		"News for AAPL: Apple services segment posts resilient growth (sentiment: positive, date: 2026-03-20)",
	}
	amzn := []string{
		"AMZN PE ratio is 42.00",
		"AMZN revenue growth is 14.00%",
		"AMZN debt-to-equity ratio is 0.40",
		"News for AMZN: Amazon expands fulfillment automation and lowers unit costs (sentiment: positive, date: 2026-03-24)",
		"News for AMZN: AWS closes several multi-year enterprise AI contracts (sentiment: positive, date: 2026-03-11)",
	}
	nvda := []string{
		"NVDA PE ratio is 40.00",
		"NVDA revenue growth is 32.00%",
		"NVDA debt-to-equity ratio is 0.25",
		"News for NVDA: NVIDIA secures major AI datacenter orders (sentiment: positive, date: 2026-03-26)",
		"News for NVDA: GPU supply constraints ease as packaging capacity expands (sentiment: positive, date: 2026-03-19)",
	}
	a := RunToolSummary(aapl)
	require.Greater(t, a.CompositeScore, -0.9, "AAPL expected HOLD band (not strong sell)")
	require.Less(t, a.CompositeScore, 0.7)

	m := RunToolSummary(amzn)
	require.Greater(t, m.CompositeScore, -0.9, "AMZN expected HOLD band")
	require.Less(t, m.CompositeScore, 0.7)

	n := RunToolSummary(nvda)
	require.GreaterOrEqual(t, n.CompositeScore, 0.7, "NVDA expected BUY threshold")
}

// Regression: financial facts must not depend on chunk order (last-wins used to drop revenue).
func TestRunToolSummaryFinancialsAnyOrder(t *testing.T) {
	// Full AMZN facts in retrieval order that used to leave revenue unset when news chunks came last.
	docs := []string{
		"AMZN profile: Amazon combines North American and international retail with AWS cloud, advertising, and logistics at global scale.",
		"AMZN key risk: Retail margin volatility, logistics competition, and cloud pricing pressure",
		"News for AMZN: Amazon expands fulfillment automation and lowers unit costs (sentiment: positive, date: 2026-03-24)",
		"AMZN PE ratio is 42.00",
		"News for AMZN: Retail profitability remains uneven across regions (sentiment: negative, date: 2026-03-17)",
		"AMZN revenue growth is 14.00%",
		"News for AMZN: AWS closes several multi-year enterprise AI contracts (sentiment: positive, date: 2026-03-11)",
		"AMZN debt-to-equity ratio is 0.40",
		"News for AMZN: Prime engagement trends remain stable quarter-over-quarter (sentiment: neutral, date: 2026-02-28)",
		"News for AMZN: Third-party seller services growth improves marketplace economics (sentiment: positive, date: 2026-02-14)",
	}
	a := RunToolSummary(docs)
	b := RunToolSummary(reverseCopy(docs))
	require.InDelta(t, a.CompositeScore, b.CompositeScore, 1e-9, "composite must not depend on doc order")
	require.Greater(t, a.CompositeScore, -0.9, "AMZN full facts stay out of strong sell, got %f", a.CompositeScore)
	require.Less(t, a.CompositeScore, 0.7, "AMZN full facts stay below BUY guardrail, got %f", a.CompositeScore)
}

func reverseCopy(s []string) []string {
	out := make([]string, len(s))
	for i := range s {
		out[i] = s[len(s)-1-i]
	}
	return out
}
