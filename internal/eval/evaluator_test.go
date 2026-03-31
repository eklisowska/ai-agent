package eval

import "testing"

func TestBuildReport(t *testing.T) {
	ev := Evaluator{}
	results := []ItemResult{
		{Ticker: "AAPL", Correct: true, ReasoningQuality: 1, ConfidenceBucket: "high"},
		{Ticker: "TSLA", Correct: false, ReasoningQuality: 0.66, ConfidenceBucket: "mid"},
	}
	report := ev.BuildReport(results)
	if report.Total != 2 {
		t.Fatalf("total got %d", report.Total)
	}
	if report.Accuracy != 0.5 {
		t.Fatalf("accuracy got %f", report.Accuracy)
	}
}
