package agent

import "ai-agent/internal/model"

type RunResult struct {
	Query               string               `json:"query"`
	Ticker              string               `json:"ticker"`
	RetrievedDocs       []string             `json:"retrieved_docs"`
	ToolSummary         ToolSummary          `json:"tool_summary"`
	Initial             model.AnalysisOutput `json:"initial"`
	Reflected           model.AnalysisOutput `json:"reflected"`
	Final               model.AnalysisOutput `json:"final"`
	ReflectionReplaced  bool                 `json:"reflection_replaced"`
	ReflectionRationale string               `json:"reflection_rationale"`
}

type ToolSummary struct {
	SentimentScore float64  `json:"sentiment_score"`
	Risks          []string `json:"risks"`
	PEAssessment   string   `json:"pe_assessment"`
}
