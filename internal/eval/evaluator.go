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
	Total               int                       `json:"total"`
	Correct             int                       `json:"correct"`
	Accuracy            float64                   `json:"accuracy"`
	AverageReasoning    float64                   `json:"avg_reasoning_quality"`
	CalibrationByBucket map[string]float64        `json:"calibration_by_bucket"`
	ConfusionMatrix     map[string]map[string]int `json:"confusion_matrix"`
	PerClass            map[string]ClassMetrics   `json:"per_class"`
	DirectionalRate     float64                   `json:"directional_rate"`
	PerTicker           []ItemResult              `json:"per_ticker"`
}

type ClassMetrics struct {
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1        float64 `json:"f1"`
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
		ConfusionMatrix: map[string]map[string]int{
			"BUY":  {"BUY": 0, "HOLD": 0, "SELL": 0},
			"HOLD": {"BUY": 0, "HOLD": 0, "SELL": 0},
			"SELL": {"BUY": 0, "HOLD": 0, "SELL": 0},
		},
		PerClass: map[string]ClassMetrics{
			"BUY":  {},
			"HOLD": {},
			"SELL": {},
		},
		PerTicker: results,
	}
	if len(results) == 0 {
		return report
	}

	countByBucket := map[string]int{"low": 0, "mid": 0, "high": 0}
	correctByBucket := map[string]int{"low": 0, "mid": 0, "high": 0}
	classes := []string{"BUY", "HOLD", "SELL"}
	tp := map[string]int{"BUY": 0, "HOLD": 0, "SELL": 0}
	fp := map[string]int{"BUY": 0, "HOLD": 0, "SELL": 0}
	fn := map[string]int{"BUY": 0, "HOLD": 0, "SELL": 0}
	var directional int

	var rq float64
	for _, r := range results {
		if r.Correct {
			report.Correct++
			correctByBucket[r.ConfidenceBucket]++
		}
		if _, ok := report.ConfusionMatrix[r.Expected]; ok {
			if _, ok := report.ConfusionMatrix[r.Expected][r.Predicted]; ok {
				report.ConfusionMatrix[r.Expected][r.Predicted]++
			}
		}
		if r.Predicted == "BUY" || r.Predicted == "SELL" {
			directional++
		}
		countByBucket[r.ConfidenceBucket]++
		rq += r.ReasoningQuality
	}

	report.Accuracy = float64(report.Correct) / float64(report.Total)
	report.AverageReasoning = rq / float64(report.Total)
	report.DirectionalRate = float64(directional) / float64(report.Total)

	for bucket, total := range countByBucket {
		if total == 0 {
			report.CalibrationByBucket[bucket] = 0
			continue
		}
		report.CalibrationByBucket[bucket] = float64(correctByBucket[bucket]) / float64(total)
	}
	for _, c := range classes {
		tp[c] = report.ConfusionMatrix[c][c]
	}
	for _, c := range classes {
		for _, actual := range classes {
			if actual != c {
				fp[c] += report.ConfusionMatrix[actual][c]
			}
		}
		for _, predicted := range classes {
			if predicted != c {
				fn[c] += report.ConfusionMatrix[c][predicted]
			}
		}
		precision := ratio(tp[c], tp[c]+fp[c])
		recall := ratio(tp[c], tp[c]+fn[c])
		f1 := 0.0
		if precision+recall > 0 {
			f1 = 2 * precision * recall / (precision + recall)
		}
		report.PerClass[c] = ClassMetrics{
			Precision: precision,
			Recall:    recall,
			F1:        f1,
		}
	}
	return report
}

func ratio(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return float64(num) / float64(den)
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
