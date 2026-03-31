package rag

type FactDocument struct {
	ID     string
	Ticker string
	Type   string
	Text   string
}

type RawProfile struct {
	Ticker  string `json:"ticker"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Risks   string `json:"risks"`
}

type RawFinancial struct {
	Ticker        string  `json:"ticker"`
	Type          string  `json:"type"`
	PERatio       float64 `json:"pe_ratio"`
	RevenueGrowth float64 `json:"revenue_growth"`
	DebtToEquity  float64 `json:"debt_to_equity"`
}

type RawNews struct {
	Ticker    string `json:"ticker"`
	Type      string `json:"type"`
	Headline  string `json:"headline"`
	Sentiment string `json:"sentiment"`
	Date      string `json:"date"`
}
