package eval

import (
	"ai-agent/internal/agent"
	"strings"
)

type Evaluator struct{}

type ItemResult struct {
	Ticker             string  `json:"ticker"`
	Predicted          string  `json:"predicted"`
	Expected           string  `json:"expected"`
	Correct            bool    `json:"correct"`
	ReasoningQuality   float64 `json:"reasoning_quality"`
	Confidence         float64 `json:"confidence"`
	ConfidenceBucket   string  `json:"confidence_bucket"`
	ReasoningChecklist []bool  `json:"reasoning_checklist"`
}

type Report struct {
	Total               int                `json:"total"`
	Correct             int                `json:"correct"`
	Accuracy            float64            `json:"accuracy"`
	AverageReasoning    float64            `json:"avg_reasoning_quality"`
	CalibrationByBucket map[string]float64 `json:"calibration_by_bucket"`
	PerTicker           []ItemResult       `json:"per_ticker"`
}

func (Evaluator) Evaluate(predicted, expected string) bool {
	return strings.EqualFold(strings.TrimSpace(predicted), strings.TrimSpace(expected))
}

func (Evaluator) ReasoningQuality(reason string) (float64, []bool) {
	low := strings.ToLower(reason)
	financial := strings.Contains(low, "pe") || strings.Contains(low, "revenue") || strings.Contains(low, "debt")
	sentiment := strings.Contains(low, "sentiment") || strings.Contains(low, "news")
	coherent := len(strings.Fields(reason)) >= 8
	checks := []bool{financial, sentiment, coherent}
	var passed float64
	for _, ok := range checks {
		if ok {
			passed++
		}
	}
	return passed / float64(len(checks)), checks
}

func (Evaluator) BuildReport(results []ItemResult) Report {
	report := Report{
		Total:               len(results),
		CalibrationByBucket: map[string]float64{"low": 0, "mid": 0, "high": 0},
		PerTicker:           results,
	}
	if len(results) == 0 {
		return report
	}

	countByBucket := map[string]int{"low": 0, "mid": 0, "high": 0}
	correctByBucket := map[string]int{"low": 0, "mid": 0, "high": 0}

	var rq float64
	for _, r := range results {
		if r.Correct {
			report.Correct++
			correctByBucket[r.ConfidenceBucket]++
		}
		countByBucket[r.ConfidenceBucket]++
		rq += r.ReasoningQuality
	}

	report.Accuracy = float64(report.Correct) / float64(report.Total)
	report.AverageReasoning = rq / float64(report.Total)

	for bucket, total := range countByBucket {
		if total == 0 {
			report.CalibrationByBucket[bucket] = 0
			continue
		}
		report.CalibrationByBucket[bucket] = float64(correctByBucket[bucket]) / float64(total)
	}
	return report
}

func ConfidenceBucket(v float64) string {
	switch {
	case v < 0.4:
		return "low"
	case v < 0.7:
		return "mid"
	default:
		return "high"
	}
}

func MakeItem(ticker string, run agent.RunResult, expected string, ev Evaluator) ItemResult {
	correct := ev.Evaluate(string(run.Final.Decision), expected)
	rq, checks := ev.ReasoningQuality(run.Final.Reasoning)
	return ItemResult{
		Ticker:             ticker,
		Predicted:          string(run.Final.Decision),
		Expected:           expected,
		Correct:            correct,
		ReasoningQuality:   rq,
		Confidence:         run.Final.Confidence,
		ConfidenceBucket:   ConfidenceBucket(run.Final.Confidence),
		ReasoningChecklist: checks,
	}
}
