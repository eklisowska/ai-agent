package agent

import "ai-agent/internal/model"

func ChooseFinal(initial, reflected model.AnalysisOutput) (model.AnalysisOutput, bool, string) {
	if reflected.Decision == "" {
		return initial, false, "reflection output invalid"
	}

	improvedConfidence := reflected.Confidence >= initial.Confidence+0.1
	changedDecision := reflected.Decision != initial.Decision
	longerReasoning := len(reflected.Reasoning) > len(initial.Reasoning)

	if improvedConfidence || (changedDecision && longerReasoning) {
		return reflected, true, "reflection improved confidence/consistency"
	}
	return initial, false, "initial decision retained"
}
