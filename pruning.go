package rtengo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
)

// PruneResult represents the result of document pruning.
type PruneResult struct {
	PrunedDocument string
	KeptRatio      float32
	KeptTokens     int
	TotalTokens    int
	RankingScore   float32
	Probability    float32
}

// PrunerConfig configures pruning behavior.
type PrunerConfig struct {
	Threshold float32
}

// Pruner prunes documents based on a query.
type Pruner struct {
	rt    *Runtime
	model *Model
	tk    Tokenizer
	cfg   PrunerConfig
}

// NewPruner creates a new Pruner.
func NewPruner(ctx context.Context, rt *Runtime, model *Model, tk Tokenizer, cfg PrunerConfig) (*Pruner, error) {
	if rt == nil || model == nil || tk == nil {
		return nil, errors.New("runtime, model, and tokenizer are required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return &Pruner{rt: rt, model: model, tk: tk, cfg: cfg}, nil
}

// Prune processes a query and document through OpenProvence model for pruning and ranking.
func (p *Pruner) Prune(ctx context.Context, query, document string) (*PruneResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	queryEnc, err := p.tk.Encode(query, true)
	if err != nil {
		return nil, fmt.Errorf("query tokenization error: %w", err)
	}
	queryIDs := make([]int32, len(queryEnc.IDs))
	for i, id := range queryEnc.IDs {
		queryIDs[i] = int32(id)
	}

	documentEnc, err := p.tk.Encode(document, true)
	if err != nil {
		return nil, fmt.Errorf("document tokenization error: %w", err)
	}
	documentIDs := make([]int32, len(documentEnc.IDs))
	for i, id := range documentEnc.IDs {
		documentIDs[i] = int32(id)
	}

	combinedIDs := make([]int32, 0, len(queryIDs)+len(documentIDs))
	combinedIDs = append(combinedIDs, queryIDs...)
	combinedIDs = append(combinedIDs, documentIDs...)

	seqLen := len(combinedIDs)
	attentionMask := make([]int32, seqLen)
	for i := range attentionMask {
		attentionMask[i] = 1
	}

	queryLen := len(queryIDs)

	inputIDsTensor, err := p.rt.createIntTensor(ctx, []uint64{1, uint64(seqLen)}, combinedIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Close(ctx)

	attentionMaskTensor, err := p.rt.createIntTensor(ctx, []uint64{1, uint64(seqLen)}, attentionMask)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Close(ctx)

	outputs, err := p.model.Run(ctx, inputIDsTensor, attentionMaskTensor)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}
	if len(outputs) < 2 {
		return nil, fmt.Errorf("expected 2 outputs (ranking_logits, pruning_logits), got %d", len(outputs))
	}

	rankingOutput := outputs[0]
	defer rankingOutput.Close(ctx)
	rankingData, err := rankingOutput.Float32Data(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ranking logits: %w", err)
	}
	rankingScore := rankingData[0]
	probability := sigmoid(rankingScore)

	pruningOutput := outputs[1]
	defer pruningOutput.Close(ctx)
	pruningData, err := pruningOutput.Float32Data(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pruning logits: %w", err)
	}

	keepProbs := softmaxLastDim(pruningData, 2)

	totalDocumentTokens := 0
	keptDocumentTokens := 0
	keptTokenIndices := []int{}

	for i := 0; i < len(documentEnc.IDs); i++ {
		if i == 0 || i == len(documentEnc.IDs)-1 {
			continue
		}

		totalDocumentTokens++
		combinedPos := queryLen + i
		if combinedPos >= seqLen {
			continue
		}

		keepProb := keepProbs[combinedPos*2+1]
		if keepProb >= p.cfg.Threshold {
			keptDocumentTokens++
			keptTokenIndices = append(keptTokenIndices, i)
		}
	}

	keptIDs := make([]int, 0, len(keptTokenIndices))
	for _, idx := range keptTokenIndices {
		keptIDs = append(keptIDs, documentEnc.IDs[idx])
	}
	prunedDocument := p.tk.Decode(keptIDs)

	keptRatio := float32(0)
	if totalDocumentTokens > 0 {
		keptRatio = float32(keptDocumentTokens) / float32(totalDocumentTokens)
	}

	return &PruneResult{
		PrunedDocument: prunedDocument,
		KeptRatio:      keptRatio,
		KeptTokens:     keptDocumentTokens,
		TotalTokens:    totalDocumentTokens,
		RankingScore:   rankingScore,
		Probability:    probability,
	}, nil
}

// PruneBatch processes multiple documents with the same query.
func (p *Pruner) PruneBatch(ctx context.Context, query string, documents []string) ([]*PruneResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, errors.New("no documents provided")
	}

	queryEnc, err := p.tk.Encode(query, true)
	if err != nil {
		return nil, fmt.Errorf("query tokenization error: %w", err)
	}
	queryIDs := make([]int32, len(queryEnc.IDs))
	for i, id := range queryEnc.IDs {
		queryIDs[i] = int32(id)
	}
	queryLen := len(queryIDs)

	type docInfo struct {
		enc         *Encoding
		combinedIDs []int32
	}

	docInfos := make([]docInfo, len(documents))
	maxSeqLen := 0
	for i, doc := range documents {
		enc, err := p.tk.Encode(doc, true)
		if err != nil {
			return nil, fmt.Errorf("document tokenization error: %w", err)
		}

		docIDs := make([]int32, len(enc.IDs))
		for j, id := range enc.IDs {
			docIDs[j] = int32(id)
		}

		combinedIDs := make([]int32, 0, len(queryIDs)+len(docIDs))
		combinedIDs = append(combinedIDs, queryIDs...)
		combinedIDs = append(combinedIDs, docIDs...)

		docInfos[i] = docInfo{enc: enc, combinedIDs: combinedIDs}
		if len(combinedIDs) > maxSeqLen {
			maxSeqLen = len(combinedIDs)
		}
	}

	batchSize := len(documents)
	inputIDsData := make([]int32, 0, batchSize*maxSeqLen)
	attentionMaskData := make([]int32, 0, batchSize*maxSeqLen)

	for _, info := range docInfos {
		seqLen := len(info.combinedIDs)
		inputIDsData = append(inputIDsData, info.combinedIDs...)
		for j := seqLen; j < maxSeqLen; j++ {
			inputIDsData = append(inputIDsData, 0)
		}

		for j := 0; j < seqLen; j++ {
			attentionMaskData = append(attentionMaskData, 1)
		}
		for j := seqLen; j < maxSeqLen; j++ {
			attentionMaskData = append(attentionMaskData, 0)
		}
	}

	inputIDsTensor, err := p.rt.createIntTensor(ctx, []uint64{uint64(batchSize), uint64(maxSeqLen)}, inputIDsData)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Close(ctx)

	attentionMaskTensor, err := p.rt.createIntTensor(ctx, []uint64{uint64(batchSize), uint64(maxSeqLen)}, attentionMaskData)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Close(ctx)

	outputs, err := p.model.Run(ctx, inputIDsTensor, attentionMaskTensor)
	if err != nil {
		return nil, fmt.Errorf("inference failed: %w", err)
	}
	if len(outputs) < 2 {
		return nil, fmt.Errorf("expected 2 outputs, got %d", len(outputs))
	}

	rankingOutput := outputs[0]
	defer rankingOutput.Close(ctx)
	rankingData, err := rankingOutput.Float32Data(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get ranking logits: %w", err)
	}

	pruningOutput := outputs[1]
	defer pruningOutput.Close(ctx)
	pruningData, err := pruningOutput.Float32Data(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pruning logits: %w", err)
	}

	results := make([]*PruneResult, batchSize)
	for i := 0; i < batchSize; i++ {
		info := docInfos[i]
		seqLen := len(info.combinedIDs)

		batchPruningStart := i * maxSeqLen * 2
		batchPruningData := pruningData[batchPruningStart : batchPruningStart+seqLen*2]
		keepProbs := softmaxLastDim(batchPruningData, 2)

		totalDocumentTokens := 0
		keptDocumentTokens := 0
		keptIDs := []int{}

		for j := 0; j < len(info.enc.IDs); j++ {
			if j == 0 || j == len(info.enc.IDs)-1 {
				continue
			}

			totalDocumentTokens++
			combinedPos := queryLen + j
			if combinedPos >= seqLen {
				continue
			}

			keepProb := keepProbs[combinedPos*2+1]
			if keepProb >= p.cfg.Threshold {
				keptDocumentTokens++
				keptIDs = append(keptIDs, info.enc.IDs[j])
			}
		}

		prunedDocument := p.tk.Decode(keptIDs)
		keptRatio := float32(0)
		if totalDocumentTokens > 0 {
			keptRatio = float32(keptDocumentTokens) / float32(totalDocumentTokens)
		}

		rankingScore := rankingData[i]
		results[i] = &PruneResult{
			PrunedDocument: prunedDocument,
			KeptRatio:      keptRatio,
			KeptTokens:     keptDocumentTokens,
			TotalTokens:    totalDocumentTokens,
			RankingScore:   rankingScore,
			Probability:    sigmoid(rankingScore),
		}
	}

	return results, nil
}

// softmaxLastDim applies softmax to the last dimension of a flat array.
func softmaxLastDim(logits []float32, numClasses int) []float32 {
	seqLen := len(logits) / numClasses
	probs := make([]float32, len(logits))

	for i := 0; i < seqLen; i++ {
		start := i * numClasses
		end := start + numClasses

		maxVal := float32(math.Inf(-1))
		for j := start; j < end; j++ {
			if logits[j] > maxVal {
				maxVal = logits[j]
			}
		}

		sum := float32(0)
		for j := start; j < end; j++ {
			probs[j] = float32(math.Exp(float64(logits[j] - maxVal)))
			sum += probs[j]
		}

		for j := start; j < end; j++ {
			probs[j] /= sum
		}
	}

	return probs
}

// mergeRanges merges overlapping or adjacent ranges.
func mergeRanges(ranges [][2]int) [][2]int {
	if len(ranges) == 0 {
		return nil
	}

	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i][0] == ranges[j][0] {
			return ranges[i][1] < ranges[j][1]
		}
		return ranges[i][0] < ranges[j][0]
	})

	merged := [][2]int{ranges[0]}
	for _, r := range ranges[1:] {
		last := &merged[len(merged)-1]
		if r[0] > last[1] {
			merged = append(merged, r)
		} else if r[1] > last[1] {
			last[1] = r[1]
		}
	}

	return merged
}
