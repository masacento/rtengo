package rtengo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"unicode"
)

// RerankResult represents a single rerank result with score and passage.
type RerankResult struct {
	OriginalIndex int
	Score         float32
	Probability   float32
	Passage       string
}

// Reranker reranks passages by relevance.
type Reranker struct {
	rt    *Runtime
	model *Model
	tk    Tokenizer
}

// NewReranker creates a new Reranker.
func NewReranker(ctx context.Context, rt *Runtime, model *Model, tk Tokenizer) (*Reranker, error) {
	if rt == nil || model == nil || tk == nil {
		return nil, errors.New("runtime, model, and tokenizer are required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &Reranker{rt: rt, model: model, tk: tk}, nil
}

// Rerank reranks passages by relevance to a question.
func (r *Reranker) Rerank(ctx context.Context, question string, passages []string) ([]RerankResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(passages) == 0 {
		return nil, errors.New("no passages provided")
	}

	questionEnc, err := r.tk.Encode(question, true)
	if err != nil {
		return nil, fmt.Errorf("tokenization error for question: %w", err)
	}
	questionIDs := make([]int32, len(questionEnc.IDs))
	for i, id := range questionEnc.IDs {
		questionIDs[i] = int32(id)
	}

	allInputIDs := make([][]int32, len(passages))
	allAttentionMasks := make([][]int32, len(passages))
	maxSeqLen := 0

	for i, passage := range passages {
		passageEnc, err := r.tk.Encode(passage, true)
		if err != nil {
			return nil, fmt.Errorf("tokenization error for passage: %w", err)
		}

		passageIDs := make([]int32, len(passageEnc.IDs))
		for j, id := range passageEnc.IDs {
			passageIDs[j] = int32(id)
		}

		combined := make([]int32, 0, len(questionIDs)+len(passageIDs))
		combined = append(combined, questionIDs...)
		combined = append(combined, passageIDs...)

		mask := make([]int32, len(combined))
		for j := range mask {
			mask[j] = 1
		}

		allInputIDs[i] = combined
		allAttentionMasks[i] = mask
		if len(combined) > maxSeqLen {
			maxSeqLen = len(combined)
		}
	}

	batchSize := len(passages)
	inputIDsData := make([]int32, 0, batchSize*maxSeqLen)
	attentionMaskData := make([]int32, 0, batchSize*maxSeqLen)

	for i := range allInputIDs {
		ids := allInputIDs[i]
		mask := allAttentionMasks[i]
		seqLen := len(ids)

		inputIDsData = append(inputIDsData, ids...)
		attentionMaskData = append(attentionMaskData, mask...)
		for j := seqLen; j < maxSeqLen; j++ {
			inputIDsData = append(inputIDsData, 0)
			attentionMaskData = append(attentionMaskData, 0)
		}
	}

	inputIDsTensor, err := r.rt.createIntTensor(ctx, []uint64{uint64(batchSize), uint64(maxSeqLen)}, inputIDsData)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Close(ctx)

	attentionMaskTensor, err := r.rt.createIntTensor(ctx, []uint64{uint64(batchSize), uint64(maxSeqLen)}, attentionMaskData)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Close(ctx)

	outputs, err := r.model.Run(ctx, inputIDsTensor, attentionMaskTensor)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}
	if len(outputs) == 0 {
		return nil, errors.New("no outputs from model")
	}

	output := outputs[0]
	defer output.Close(ctx)

	logits, err := output.Float32Data(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get output data: %w", err)
	}

	results := make([]RerankResult, batchSize)
	for i := 0; i < batchSize; i++ {
		score := logits[i]
		results[i] = RerankResult{
			OriginalIndex: i,
			Score:         score,
			Probability:   sigmoid(score),
			Passage:       passages[i],
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// SplitIntoSentences splits context into sentences for reranking.
func SplitIntoSentences(context string) []string {
	var sentences []string
	var current []rune

	runes := []rune(context)
	i := 0

	for i < len(runes) {
		ch := runes[i]
		current = append(current, ch)

		if ch == '.' || ch == '!' || ch == '?' {
			if i+1 >= len(runes) || unicode.IsSpace(runes[i+1]) {
				sentence := trimString(string(current))
				if len(sentence) > 0 {
					sentences = append(sentences, sentence)
				}
				current = current[:0]
				for i+1 < len(runes) && unicode.IsSpace(runes[i+1]) {
					i++
				}
			}
		}
		i++
	}

	remaining := trimString(string(current))
	if len(remaining) > 0 {
		sentences = append(sentences, remaining)
	}

	return sentences
}

func trimString(s string) string {
	runes := []rune(s)
	start := 0
	end := len(runes)

	for start < end && unicode.IsSpace(runes[start]) {
		start++
	}
	for end > start && unicode.IsSpace(runes[end-1]) {
		end--
	}

	return string(runes[start:end])
}

func sigmoid(x float32) float32 {
	return float32(1.0 / (1.0 + math.Exp(float64(-x))))
}
