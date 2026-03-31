package rag

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Indexer struct {
	qdrant *QdrantClient
	embed  *Embedder
}

func NewIndexer(qdrant *QdrantClient, embed *Embedder) *Indexer {
	return &Indexer{qdrant: qdrant, embed: embed}
}

func LoadRawData(rawDir string) ([]FactDocument, []GroundTruth, error) {
	var docs []FactDocument

	var profiles []RawProfile
	if err := readJSON(filepath.Join(rawDir, "profiles.json"), &profiles); err != nil {
		return nil, nil, err
	}
	for _, p := range profiles {
		docs = append(docs, FactDocument{
			ID:     makeFactID(p.Ticker, "profile", p.Content),
			Ticker: p.Ticker,
			Type:   "profile",
			Text:   fmt.Sprintf("%s profile: %s", p.Ticker, p.Content),
		})
		if strings.TrimSpace(p.Risks) != "" {
			docs = append(docs, FactDocument{
				ID:     makeFactID(p.Ticker, "risk", p.Risks),
				Ticker: p.Ticker,
				Type:   "risk",
				Text:   fmt.Sprintf("%s key risk: %s", p.Ticker, p.Risks),
			})
		}
	}

	var financials []RawFinancial
	if err := readJSON(filepath.Join(rawDir, "financials.json"), &financials); err != nil {
		return nil, nil, err
	}
	for _, f := range financials {
		docs = append(docs,
			FactDocument{
				ID:     makeFactID(f.Ticker, "financial", fmt.Sprintf("pe:%f", f.PERatio)),
				Ticker: f.Ticker,
				Type:   "financial",
				Text:   fmt.Sprintf("%s PE ratio is %.2f", f.Ticker, f.PERatio),
			},
			FactDocument{
				ID:     makeFactID(f.Ticker, "financial", fmt.Sprintf("rev:%f", f.RevenueGrowth)),
				Ticker: f.Ticker,
				Type:   "financial",
				Text:   fmt.Sprintf("%s revenue growth is %.2f%%", f.Ticker, f.RevenueGrowth*100),
			},
			FactDocument{
				ID:     makeFactID(f.Ticker, "financial", fmt.Sprintf("dte:%f", f.DebtToEquity)),
				Ticker: f.Ticker,
				Type:   "financial",
				Text:   fmt.Sprintf("%s debt-to-equity ratio is %.2f", f.Ticker, f.DebtToEquity),
			},
		)
	}

	var news []RawNews
	if err := readJSON(filepath.Join(rawDir, "news.json"), &news); err != nil {
		return nil, nil, err
	}
	for _, n := range news {
		datePart := ""
		if strings.TrimSpace(n.Date) != "" {
			datePart = fmt.Sprintf(", date: %s", n.Date)
		}
		docs = append(docs, FactDocument{
			ID:     makeFactID(n.Ticker, "news", n.Headline),
			Ticker: n.Ticker,
			Type:   "news",
			Text:   fmt.Sprintf("News for %s: %s (sentiment: %s%s)", n.Ticker, n.Headline, strings.ToLower(n.Sentiment), datePart),
		})
	}

	var truth []GroundTruth
	if err := readJSON(filepath.Join(rawDir, "ground_truth.json"), &truth); err != nil {
		return nil, nil, err
	}

	return docs, truth, nil
}

func (i *Indexer) IndexDocuments(ctx context.Context, collection string, docs []FactDocument) error {
	for _, doc := range docs {
		vector, err := i.embed.Embed(ctx, doc.Text)
		if err != nil {
			return fmt.Errorf("embed %s: %w", doc.ID, err)
		}
		if err := i.qdrant.UpsertPoint(ctx, collection, doc, vector); err != nil {
			return fmt.Errorf("upsert %s: %w", doc.ID, err)
		}
	}
	return nil
}

func makeFactID(ticker, kind, content string) string {
	// Qdrant in this config only accepts unsigned integers or UUIDs as point IDs.
	// Build a deterministic UUID-like ID from a SHA-1 hash of the content.
	h := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(content))))
	hexStr := hex.EncodeToString(h[:16]) // 32 hex chars

	// Format as 8-4-4-4-12 (UUID-style) using the first 32 chars.
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32],
	)
}

func readJSON(path string, out any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
