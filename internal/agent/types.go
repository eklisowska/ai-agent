package agent

import "ai-agent/internal/model"

type RunResult struct {
	Query               string               `json:"query"`
	Ticker              string               `json:"ticker"`
	NoData              bool                 `json:"no_data"`
	NoDataReason        string               `json:"no_data_reason,omitempty"`
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
	RiskCount      int      `json:"risk_count"`
	PEAssessment   string   `json:"pe_assessment"`
	CompositeScore float64  `json:"composite_score"`
}
