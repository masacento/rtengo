package rtengo

import "testing"

func TestSoftmaxLastDim(t *testing.T) {
	logits := []float32{0, 0, 0, 0}
	probs := softmaxLastDim(logits, 2)
	if len(probs) != 4 {
		t.Fatalf("Expected 4 probs, got %d", len(probs))
	}
	if probs[0] < 0.49 || probs[0] > 0.51 {
		t.Fatalf("Expected prob ~0.5, got %f", probs[0])
	}
}

func TestSoftmaxLastDim_PairsSumToOne(t *testing.T) {
	logits := []float32{1, 2, 3, 4}
	probs := softmaxLastDim(logits, 2)
	if len(probs) != 4 {
		t.Fatalf("Expected 4 probs, got %d", len(probs))
	}
	for i := 0; i < 2; i++ {
		sum := probs[i*2] + probs[i*2+1]
		if sum < 0.999 || sum > 1.001 {
			t.Fatalf("Expected pair sum ~1.0, got %f", sum)
		}
		if probs[i*2+1] <= probs[i*2] {
			t.Fatalf("Expected second prob > first in pair %d", i)
		}
	}
}

func TestMergeRanges(t *testing.T) {
	ranges := [][2]int{{1, 3}, {2, 5}, {10, 11}}
	merged := mergeRanges(ranges)
	if len(merged) != 2 {
		t.Fatalf("Expected 2 ranges, got %d", len(merged))
	}
	if merged[0][0] != 1 || merged[0][1] != 5 {
		t.Fatalf("Unexpected merge result: %+v", merged[0])
	}
}

func TestMergeRanges_Empty(t *testing.T) {
	if merged := mergeRanges(nil); merged != nil {
		t.Fatalf("Expected nil for empty ranges, got %v", merged)
	}
}
