package llm

import "testing"

func TestParseAnalysis(t *testing.T) {
	raw := `{
	  "decision":"BUY",
	  "reasoning":"PE and news sentiment indicate upside.",
	  "confidence":0.76
	}`
	out, err := ParseAnalysis(raw)
	if err != nil {
		t.Fatalf("ParseAnalysis error: %v", err)
	}
	if out.Decision != "BUY" {
		t.Fatalf("unexpected decision %q", out.Decision)
	}
}

func TestParseAnalysisInvalid(t *testing.T) {
	raw := `{"decision":"MAYBE","reasoning":"x","confidence":0.2}`
	_, err := ParseAnalysis(raw)
	if err == nil {
		t.Fatalf("expected error for invalid decision")
	}
}
