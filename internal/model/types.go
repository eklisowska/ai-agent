package model

type Decision string

const (
	DecisionBuy  Decision = "BUY"
	DecisionHold Decision = "HOLD"
	DecisionSell Decision = "SELL"
)

type AnalysisOutput struct {
	Decision   Decision `json:"decision"`
	Reasoning  string   `json:"reasoning"`
	Confidence float64  `json:"confidence"`
}
