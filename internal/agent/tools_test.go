package agent

import "testing"

func TestSentimentScore(t *testing.T) {
	score := SentimentScore([]string{
		"Cloud revenue growth beats expectations",
		"Company faces declining demand",
	})
	if score == 0 {
		t.Fatalf("expected non-zero sentiment score")
	}
}

func TestAnalyzePE(t *testing.T) {
	if got := AnalyzePE(10); got != "undervalued" {
		t.Fatalf("AnalyzePE undervalued = %q", got)
	}
	if got := AnalyzePE(20); got != "fair" {
		t.Fatalf("AnalyzePE fair = %q", got)
	}
	if got := AnalyzePE(40); got != "overvalued" {
		t.Fatalf("AnalyzePE overvalued = %q", got)
	}
}

func TestDetectRisk(t *testing.T) {
	risks := DetectRisk("Regulation risk and debt pressure are rising.")
	if len(risks) == 0 {
		t.Fatalf("expected at least one risk")
	}
}
