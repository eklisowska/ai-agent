package rag

import (
	"context"
	"fmt"
	"strings"
)

type Retriever struct {
	embed      *Embedder
	qdrant     *QdrantClient
	collection string
	topK       int
}

func NewRetriever(embed *Embedder, qdrant *QdrantClient, collection string, topK int) *Retriever {
	return &Retriever{
		embed:      embed,
		qdrant:     qdrant,
		collection: collection,
		topK:       topK,
	}
}

func (r *Retriever) Retrieve(ctx context.Context, query, ticker string) ([]string, error) {
	vec, err := r.embed.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	docs, err := r.qdrant.Search(ctx, r.collection, strings.ToUpper(ticker), vec, r.topK)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return []string{}, nil
	}
	return docs, nil
}
