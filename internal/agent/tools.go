package agent

import (
	"fmt"
	"regexp"
	"strings"
)

type Tool interface {
	Name() string
	Execute(input string) (string, error)
}

func SentimentScore(headlines []string) float64 {
	positive := []string{"growth", "beats", "surge", "expands", "strong", "record", "up"}
	negative := []string{"declining", "miss", "cut", "weak", "drops", "lawsuit", "down"}

	var score float64
	var tokens int
	for _, h := range headlines {
		low := strings.ToLower(h)
		for _, p := range positive {
			if strings.Contains(low, p) {
				score++
				tokens++
			}
		}
		for _, n := range negative {
			if strings.Contains(low, n) {
				score--
				tokens++
			}
		}
	}
	if tokens == 0 {
		return 0
	}
	normalized := score / float64(tokens)
	if normalized > 1 {
		return 1
	}
	if normalized < -1 {
		return -1
	}
	return normalized
}

func DetectRisk(text string) []string {
	keywords := []string{"regulation", "declining demand", "competition", "debt", "litigation", "margin pressure"}
	low := strings.ToLower(text)
	var risks []string
	for _, k := range keywords {
		if strings.Contains(low, k) {
			risks = append(risks, k)
		}
	}
	if len(risks) == 0 {
		if regexp.MustCompile(`\brisk\b`).MatchString(low) {
			risks = append(risks, "unspecified risk")
		}
	}
	return risks
}

func AnalyzePE(pe float64) string {
	switch {
	case pe < 15:
		return "undervalued"
	case pe <= 30:
		return "fair"
	default:
		return "overvalued"
	}
}

func RunToolSummary(retrievedDocs []string) ToolSummary {
	var headlines []string
	var merged string
	var pe float64 = 25

	for _, doc := range retrievedDocs {
		merged += doc + "\n"
		low := strings.ToLower(doc)
		if strings.Contains(low, "news for") {
			headlines = append(headlines, doc)
		}
		var ticker string
		_, _ = fmt.Sscanf(doc, "%s PE ratio is %f", &ticker, &pe)
	}

	return ToolSummary{
		SentimentScore: SentimentScore(headlines),
		Risks:          DetectRisk(merged),
		PEAssessment:   AnalyzePE(pe),
	}
}
