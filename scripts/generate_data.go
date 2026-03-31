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

func main() {
	_ = write("data/raw/profiles.json", []profile{
		{"AAPL", "profile", "Apple sells premium devices and high-margin services across iPhone, Mac, wearables, and a growing subscription stack.", "Concentration in handset cycles, China demand, and ongoing app store regulation"},
		{"AMZN", "profile", "Amazon combines North American and international retail with AWS cloud, advertising, and logistics at global scale.", "Retail margin volatility, logistics competition, and cloud pricing pressure"},
		{"NVDA", "profile", "NVIDIA supplies accelerated computing for AI training, inference, data center, and gaming with CUDA ecosystem lock-in. Data center and AI software attach are the primary profit drivers; headline mixes can look noisy until you reconcile dated export chatter with current backlog and hyperscaler capex.", "Semiconductor cycle swings, export rules, and hyperscaler capex sensitivity"},
	})

	_ = write("data/raw/financials.json", []financial{
		{"AAPL", "financial", 26, 0.04, 1.2},
		{"AMZN", "financial", 42, 0.14, 0.4},
		{"NVDA", "financial", 40, 0.32, 0.25},
	})

	_ = write("data/raw/news.json", []news{
		{"AAPL", "news", "Apple faces declining demand in China", "negative", "2026-03-28"},
		{"AAPL", "news", "Apple services segment posts resilient growth", "positive", "2026-03-20"},
		{"AAPL", "news", "iPhone upgrade cycle shows mixed momentum", "neutral", "2026-03-12"},
		{"AAPL", "news", "Regulatory pressure on app store fees rises", "negative", "2026-03-05"},
		{"AAPL", "news", "Apple announces supply chain efficiency gains", "positive", "2026-02-22"},
		{"AAPL", "news", "Wearables and installed base drive recurring revenue upside", "positive", "2026-03-25"},

		{"AMZN", "news", "Amazon expands fulfillment automation and lowers unit costs", "positive", "2026-03-24"},
		{"AMZN", "news", "Retail profitability remains uneven across regions", "negative", "2026-03-17"},
		{"AMZN", "news", "AWS closes several multi-year enterprise AI contracts", "positive", "2026-03-11"},
		{"AMZN", "news", "Prime engagement trends remain stable quarter-over-quarter", "neutral", "2026-02-28"},
		{"AMZN", "news", "Third-party seller services growth improves marketplace economics", "positive", "2026-02-14"},
		{"AMZN", "news", "Advertising revenue mix shifts toward higher-margin streams", "positive", "2026-03-27"},

		{"NVDA", "news", "NVIDIA reports record data center revenue; management raises full-year AI infrastructure outlook", "positive", "2026-03-30"},
		{"NVDA", "news", "NVIDIA secures major AI datacenter orders", "positive", "2026-03-26"},
		{"NVDA", "news", "GPU supply constraints ease as packaging capacity expands", "positive", "2026-03-19"},
		{"NVDA", "news", "New enterprise software stack expands moat", "positive", "2026-03-10"},
		{"NVDA", "news", "Export restrictions create regional uncertainty", "negative", "2026-02-27"},
		{"NVDA", "news", "Hyperscaler capex outlook remains strong", "positive", "2026-02-13"},
		{"NVDA", "news", "Blackwell ramp meets early enterprise demand targets", "positive", "2026-03-29"},
		{"NVDA", "news", "NVDA updates guidance: AI backlog visibility improves versus prior-quarter uncertainty", "positive", "2026-03-31"},
		{"NVDA", "news", "Analysts note near-term headline mix can understate sequential data center strength", "neutral", "2026-03-27"},
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
