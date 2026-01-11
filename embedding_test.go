package rtengo

import (
	"context"
	"testing"
)

type dummyTokenizer struct{}

func (d dummyTokenizer) Encode(text string, addSpecial bool) (*Encoding, error) {
	return &Encoding{Tokens: []string{text}, IDs: []int{1}, AttentionMask: []int{1}}, nil
}

func (d dummyTokenizer) Decode(ids []int) string {
	return "decoded"
}

func TestNewEmbedder_RequiresDeps(t *testing.T) {
	ctx := context.Background()
	_, err := NewEmbedder(ctx, nil, nil, nil)
	if err == nil {
		t.Fatal("Expected error when dependencies are nil")
	}
}

func TestNewEmbedder_OptionsApplied(t *testing.T) {
	ctx := context.Background()
	embedder, err := NewEmbedder(ctx, &Runtime{}, &Model{}, dummyTokenizer{},
		WithEmbeddingNormalize(true),
		WithEmbeddingPooling(PoolingMean),
	)
	if err != nil {
		t.Fatalf("Expected embedder to be created: %v", err)
	}
	if !embedder.cfg.Normalize {
		t.Fatal("Expected normalize option to be applied")
	}
	if embedder.cfg.Pooling != PoolingMean {
		t.Fatalf("Expected pooling %v, got %v", PoolingMean, embedder.cfg.Pooling)
	}
}
