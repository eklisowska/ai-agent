package agent

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

var (
	newsDateRe = regexp.MustCompile(`date:\s*(\d{4}-\d{2}-\d{2})`)
	riskWordRe = regexp.MustCompile(`\brisk\b`)
)

type Tool interface {
	Name() string
	Execute(input string) (string, error)
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func sentimentFromNewsLine(line string) float64 {
	low := strings.ToLower(line)
	switch {
	case strings.Contains(low, "sentiment: positive"):
		return 1
	case strings.Contains(low, "sentiment: negative"):
		return -1
	default:
		return 0
	}
}

func newsRecencyWeight(line string) float64 {
	m := newsDateRe.FindStringSubmatch(strings.ToLower(line))
	if len(m) != 2 {
		return 1
	}
	d, err := time.Parse("2006-01-02", m[1])
	if err != nil {
		return 1
	}
	days := time.Since(d).Hours() / 24
	if days < 0 {
		days = 0
	}
	// Exponential decay; very recent headlines weigh more.
	return math.Exp(-days / 120.0)
}

func SentimentScore(newsDocs []string) float64 {
	if len(newsDocs) == 0 {
		return 0
	}
	var weightedSum float64
	var weightTotal float64
	for _, line := range newsDocs {
		w := newsRecencyWeight(line)
		weightedSum += sentimentFromNewsLine(line) * w
		weightTotal += w
	}
	if weightTotal == 0 {
		return 0
	}
	return clamp(weightedSum/weightTotal, -1, 1)
}

func DetectRisk(text string) []string {
	keywords := []string{"regulation", "declining demand", "competition", "debt", "litigation", "margin pressure"}
	low := strings.ToLower(text)
	// Avoid matching "debt" inside the financial metric phrase "debt-to-equity".
	lowForDebt := strings.ReplaceAll(strings.ReplaceAll(low, "debt-to-equity", ""), "debt to equity", "")
	var risks []string
	for _, k := range keywords {
		src := low
		if k == "debt" {
			src = lowForDebt
		}
		if strings.Contains(src, k) {
			risks = append(risks, k)
		}
	}
	if len(risks) == 0 {
		if riskWordRe.MatchString(low) {
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
	var revenueGrowth float64
	var debtToEquity float64

	for _, doc := range retrievedDocs {
		merged += doc + "\n"
		low := strings.ToLower(doc)
		if strings.Contains(low, "news for") {
			headlines = append(headlines, doc)
		}
		var ticker string
		_, _ = fmt.Sscanf(doc, "%s PE ratio is %f", &ticker, &pe)
		_, _ = fmt.Sscanf(doc, "%s revenue growth is %f%%", &ticker, &revenueGrowth)
		_, _ = fmt.Sscanf(doc, "%s debt-to-equity ratio is %f", &ticker, &debtToEquity)
	}
	risks := DetectRisk(merged)
	sentiment := SentimentScore(headlines)
	peAssessment := AnalyzePE(pe)

	composite := 0.0
	switch peAssessment {
	case "undervalued":
		composite += 0.5
	case "overvalued":
		// High revenue growth partially offsets a rich multiple (growth names).
		if revenueGrowth >= 10 {
			composite -= 0.22
		} else {
			composite -= 0.4
		}
	}
	switch {
	case revenueGrowth >= 15:
		composite += 0.8
	case revenueGrowth >= 10:
		// Strong double-digit growth (e.g. AMZN) should compete with rich multiples.
		composite += 0.63
	case revenueGrowth >= 8:
		composite += 0.4
	case revenueGrowth < 0:
		composite -= 0.8
	}
	switch {
	case debtToEquity > 1.5:
		composite -= 0.6
	case debtToEquity > 1.0:
		composite -= 0.3
	case debtToEquity < 0.5 && debtToEquity > 0:
		composite += 0.2
	}
	composite += sentiment * 0.8
	composite -= math.Min(float64(len(risks))*0.2, 1.0)
	composite = clamp(composite, -2, 2)

	return ToolSummary{
		SentimentScore: sentiment,
		Risks:          risks,
		RiskCount:      len(risks),
		PEAssessment:   peAssessment,
		CompositeScore: composite,
	}
}
