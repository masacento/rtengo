package rtengo

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// PoolingType defines how to pool token embeddings.
type PoolingType int

const (
	PoolingMean PoolingType = iota + 1
)

// EmbedderConfig configures embedding behavior.
type EmbedderConfig struct {
	Normalize bool
	Pooling   PoolingType
}

// EmbedderOption configures an Embedder.
type EmbedderOption func(*EmbedderConfig)

// WithEmbeddingNormalize enables or disables L2 normalization.
func WithEmbeddingNormalize(normalize bool) EmbedderOption {
	return func(cfg *EmbedderConfig) {
		cfg.Normalize = normalize
	}
}

// WithEmbeddingPooling sets the pooling strategy.
func WithEmbeddingPooling(pooling PoolingType) EmbedderOption {
	return func(cfg *EmbedderConfig) {
		cfg.Pooling = pooling
	}
}

// Embedder produces embeddings from text.
type Embedder struct {
	rt    *Runtime
	model *Model
	tk    Tokenizer
	cfg   EmbedderConfig
}

// NewEmbedder creates a new Embedder.
func NewEmbedder(ctx context.Context, rt *Runtime, model *Model, tk Tokenizer, opts ...EmbedderOption) (*Embedder, error) {
	if rt == nil || model == nil || tk == nil {
		return nil, errors.New("runtime, model, and tokenizer are required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	cfg := EmbedderConfig{Pooling: PoolingMean}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Pooling == 0 {
		cfg.Pooling = PoolingMean
	}

	return &Embedder{rt: rt, model: model, tk: tk, cfg: cfg}, nil
}

// Embed generates an embedding for a single text.
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	vectors, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) != 1 {
		return nil, errors.New("unexpected embedding count")
	}
	return vectors[0], nil
}

// EmbedBatch generates embeddings for a batch of texts.
func (e *Embedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(texts) == 0 {
		return nil, errors.New("no texts provided")
	}

	encodings := make([]*Encoding, len(texts))
	maxSeqLen := 0
	for i, text := range texts {
		enc, err := e.tk.Encode(text, true)
		if err != nil {
			return nil, fmt.Errorf("tokenization error: %w", err)
		}
		encodings[i] = enc
		if len(enc.IDs) > maxSeqLen {
			maxSeqLen = len(enc.IDs)
		}
	}

	batchSize := len(texts)
	inputIdsData := make([]int32, 0, batchSize*maxSeqLen)
	attentionMaskData := make([]int32, 0, batchSize*maxSeqLen)

	for _, enc := range encodings {
		seqLen := len(enc.IDs)
		for _, id := range enc.IDs {
			inputIdsData = append(inputIdsData, int32(id))
		}
		for _, m := range enc.AttentionMask {
			attentionMaskData = append(attentionMaskData, int32(m))
		}
		for j := seqLen; j < maxSeqLen; j++ {
			inputIdsData = append(inputIdsData, 0)
			attentionMaskData = append(attentionMaskData, 0)
		}
	}

	inputIdsTensor, err := e.rt.createIntTensor(ctx, []uint64{uint64(batchSize), uint64(maxSeqLen)}, inputIdsData)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIdsTensor.Close(ctx)

	attentionMaskTensor, err := e.rt.createIntTensor(ctx, []uint64{uint64(batchSize), uint64(maxSeqLen)}, attentionMaskData)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Close(ctx)

	outputs, err := e.model.Run(ctx, inputIdsTensor, attentionMaskTensor)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}
	if len(outputs) == 0 {
		return nil, errors.New("no outputs from model")
	}

	output := outputs[0]
	defer output.Close(ctx)

	shape := output.Shape()
	data, err := output.Float32Data(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get output data: %w", err)
	}

	var embeddings [][]float32
	switch len(shape) {
	case 2:
		hidden := int(shape[1])
		embeddings = make([][]float32, batchSize)
		for i := 0; i < batchSize; i++ {
			start := i * hidden
			end := start + hidden
			vec := append([]float32(nil), data[start:end]...)
			embeddings[i] = vec
		}
	case 3:
		hidden := int(shape[2])
		if e.cfg.Pooling != PoolingMean {
			return nil, errors.New("unsupported pooling type")
		}

		embeddings = make([][]float32, batchSize)
		for i := 0; i < batchSize; i++ {
			vec := make([]float32, hidden)
			sumCount := 0
			for token := 0; token < maxSeqLen; token++ {
				if attentionMaskData[i*maxSeqLen+token] == 0 {
					continue
				}
				sumCount++
				base := (i*maxSeqLen*hidden + token*hidden)
				for d := 0; d < hidden; d++ {
					vec[d] += data[base+d]
				}
			}
			if sumCount == 0 {
				sumCount = 1
			}
			for d := range vec {
				vec[d] /= float32(sumCount)
			}
			embeddings[i] = vec
		}
	default:
		return nil, fmt.Errorf("unexpected embedding output shape: %v", shape)
	}

	if e.cfg.Normalize {
		for i := range embeddings {
			l2norm := float32(0)
			for _, v := range embeddings[i] {
				l2norm += v * v
			}
			l2norm = float32(math.Sqrt(float64(l2norm)))
			if l2norm == 0 {
				continue
			}
			for j := range embeddings[i] {
				embeddings[i][j] /= l2norm
			}
		}
	}

	return embeddings, nil
}
