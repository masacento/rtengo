package rtengo

import (
	"context"
	"os"
	"testing"
)

// Regression test: repeated LoadModel/Close must not exhaust the 4GiB
// wasm linear memory (the staging buffer used to leak on every load).
func TestRepeatedLoadModel(t *testing.T) {
	path := os.Getenv("RTENGO_BENCH_MODEL")
	if path == "" {
		t.Skip("RTENGO_BENCH_MODEL is not set")
	}
	modelBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	rt, err := NewRuntime(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rt.Close(ctx)

	for i := 0; i < 10; i++ {
		model, err := rt.LoadModel(ctx, modelBytes)
		if err != nil {
			t.Fatalf("iteration %d: LoadModel failed: %v", i, err)
		}
		if err := model.Close(ctx); err != nil {
			t.Fatalf("iteration %d: Close failed: %v", i, err)
		}
		t.Logf("iteration %d ok", i)
	}
}
