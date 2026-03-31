package agent

import "ai-agent/internal/model"

func ChooseFinal(initial, reflected model.AnalysisOutput) (model.AnalysisOutput, bool, string) {
	if reflected.Decision == "" {
		return initial, false, "reflection output invalid"
	}

	improvedConfidence := reflected.Confidence >= initial.Confidence+0.1
	nonDegradedConfidence := reflected.Confidence >= initial.Confidence
	changedDecision := reflected.Decision != initial.Decision
	longerReasoning := len(reflected.Reasoning) >= len(initial.Reasoning)+20

	// Reflective revisions should be evidence-led:
	// either materially higher confidence, or a better-supported decision change.
	if improvedConfidence || (changedDecision && nonDegradedConfidence && longerReasoning) {
		return reflected, true, "reflection improved evidence/consistency"
	}
	return initial, false, "initial decision retained"
}
