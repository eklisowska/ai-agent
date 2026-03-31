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
}

type groundTruth struct {
	Ticker   string `json:"ticker"`
	Expected string `json:"expected"`
}

func main() {
	_ = write("data/raw/profiles.json", []profile{
		{"AAPL", "profile", "Apple designs consumer electronics and software ecosystems.", "Dependence on iPhone sales and China demand"},
		{"MSFT", "profile", "Microsoft offers cloud, software, and enterprise platforms.", "Competition in cloud pricing"},
		{"TSLA", "profile", "Tesla manufactures EVs and energy products.", "Execution risk in production ramp"},
		{"AMZN", "profile", "Amazon runs e-commerce, cloud, and logistics businesses.", "Margin pressure from retail costs"},
		{"NVDA", "profile", "NVIDIA builds GPUs used in AI and gaming workloads.", "Cyclic demand in semiconductors"},
	})

	_ = write("data/raw/financials.json", []financial{
		{"AAPL", "financial", 28, 0.05, 1.5},
		{"MSFT", "financial", 33, 0.11, 0.4},
		{"TSLA", "financial", 52, 0.16, 0.2},
		{"AMZN", "financial", 45, 0.10, 0.7},
		{"NVDA", "financial", 40, 0.28, 0.3},
	})

	_ = write("data/raw/news.json", []news{
		{"AAPL", "news", "Apple faces declining demand in China", "negative"},
		{"MSFT", "news", "Microsoft cloud revenue growth beats expectations", "positive"},
		{"TSLA", "news", "Tesla cuts prices as EV competition increases", "negative"},
		{"AMZN", "news", "Amazon expands logistics automation to reduce costs", "positive"},
		{"NVDA", "news", "NVIDIA secures major AI datacenter orders", "positive"},
	})

	_ = write("data/raw/ground_truth.json", []groundTruth{
		{"AAPL", "HOLD"},
		{"MSFT", "BUY"},
		{"TSLA", "SELL"},
		{"AMZN", "HOLD"},
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
