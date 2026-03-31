package main

import (
	"encoding/json"
	"os"
)

type profile struct {
	Ticker  string `json:"ticker"`
	Type    string `json:"type"`
	Content string `json:"content"`
	Risks   string `json:"risks"`
}

type financial struct {
	Ticker        string  `json:"ticker"`
	Type          string  `json:"type"`
	PERatio       float64 `json:"pe_ratio"`
	RevenueGrowth float64 `json:"revenue_growth"`
	DebtToEquity  float64 `json:"debt_to_equity"`
}

type news struct {
	Ticker    string `json:"ticker"`
	Type      string `json:"type"`
	Headline  string `json:"headline"`
	Sentiment string `json:"sentiment"`
	Date      string `json:"date"`
}

type groundTruth struct {
	Ticker   string `json:"ticker"`
	Expected string `json:"expected"`
}

func main() {
	_ = write("data/raw/profiles.json", []profile{
		{"AAPL", "profile", "Apple designs consumer electronics and software ecosystems.", "Dependence on iPhone sales and China demand"},
		{"AMZN", "profile", "Amazon runs e-commerce, cloud, and logistics businesses.", "Execution variability across retail and logistics operations"},
		{"NVDA", "profile", "NVIDIA builds GPUs used in AI and gaming workloads.", "Cyclic demand in semiconductors"},
	})

	_ = write("data/raw/financials.json", []financial{
		{"AAPL", "financial", 28, 0.05, 1.5},
		{"AMZN", "financial", 45, 0.12, 0.45},
		{"NVDA", "financial", 38, 0.35, 0.3},
	})

	_ = write("data/raw/news.json", []news{
		{"AAPL", "news", "Apple faces declining demand in China", "negative", "2026-03-26"},
		{"AAPL", "news", "Apple services segment posts resilient growth", "positive", "2026-03-19"},
		{"AAPL", "news", "iPhone upgrade cycle shows mixed momentum", "neutral", "2026-03-12"},
		{"AAPL", "news", "Regulatory pressure on app store fees rises", "negative", "2026-03-04"},
		{"AAPL", "news", "Apple announces supply chain efficiency gains", "positive", "2026-02-21"},

		{"AMZN", "news", "Amazon expands fulfillment automation and lowers unit costs", "positive", "2026-03-22"},
		{"AMZN", "news", "Retail profitability remains uneven across regions", "negative", "2026-03-16"},
		{"AMZN", "news", "AWS closes several multi-year enterprise AI contracts", "positive", "2026-03-10"},
		{"AMZN", "news", "Prime engagement trends remain stable quarter-over-quarter", "neutral", "2026-02-27"},
		{"AMZN", "news", "Third-party seller services growth improves marketplace economics", "positive", "2026-02-13"},

		{"NVDA", "news", "NVIDIA secures major AI datacenter orders", "positive", "2026-03-24"},
		{"NVDA", "news", "GPU supply constraints ease as packaging capacity expands", "positive", "2026-03-18"},
		{"NVDA", "news", "New enterprise software stack expands moat", "positive", "2026-03-09"},
		{"NVDA", "news", "Export restrictions create regional uncertainty", "negative", "2026-02-26"},
		{"NVDA", "news", "Hyperscaler capex outlook remains strong", "positive", "2026-02-12"},

	})

	_ = write("data/raw/ground_truth.json", []groundTruth{
		{"AAPL", "HOLD"},
		{"AMZN", "BUY"},
		{"NVDA", "BUY"},
	})
}

func write(path string, payload any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}
