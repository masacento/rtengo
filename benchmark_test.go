package rtengo

import (
	"context"
	"os"
	"testing"
)

// Benchmarks for the current (wazero-based) implementation.
//
// These benchmarks require a real model and tokenizer, provided via
// environment variables:
//
//	RTENGO_BENCH_MODEL      path to the model file (e.g. model.onnx / model.rten)
//	RTENGO_BENCH_TOKENIZER  path to tokenizer.json
//
// Example:
//
//	RTENGO_BENCH_MODEL=/path/to/model.onnx \
//	RTENGO_BENCH_TOKENIZER=/path/to/tokenizer.json \
//	go test -run '^$' -bench . -benchmem -count 5

func benchModelPath(b *testing.B) []byte {
	b.Helper()
	path := os.Getenv("RTENGO_BENCH_MODEL")
	if path == "" {
		b.Skip("RTENGO_BENCH_MODEL is not set")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("failed to read model file: %v", err)
	}
	return data
}

func benchTokenizer(b *testing.B) Tokenizer {
	b.Helper()
	path := os.Getenv("RTENGO_BENCH_TOKENIZER")
	if path == "" {
		b.Skip("RTENGO_BENCH_TOKENIZER is not set")
	}
	tk, err := NewTokenizerFromFile(path)
	if err != nil {
		b.Fatalf("failed to load tokenizer: %v", err)
	}
	return tk
}

// BenchmarkRuntimeInit measures WASM module compile + instantiate + rten_init.
func BenchmarkRuntimeInit(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt, err := NewRuntime(ctx)
		if err != nil {
			b.Fatalf("NewRuntime failed: %v", err)
		}
		if err := rt.Close(ctx); err != nil {
			b.Fatalf("Close failed: %v", err)
		}
	}
}

// BenchmarkModelLoad measures copying the model into WASM memory and
// running rten_load_model.
func BenchmarkModelLoad(b *testing.B) {
	ctx := context.Background()
	modelBytes := benchModelPath(b)

	rt, err := NewRuntime(ctx)
	if err != nil {
		b.Fatalf("NewRuntime failed: %v", err)
	}
	defer rt.Close(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model, err := rt.LoadModel(ctx, modelBytes)
		if err != nil {
			b.Fatalf("LoadModel failed: %v", err)
		}
		if err := model.Close(ctx); err != nil {
			b.Fatalf("model Close failed: %v", err)
		}
	}
}

// BenchmarkEmbed measures a single end-to-end embedding inference.
func BenchmarkEmbed(b *testing.B) {
	ctx := context.Background()
	modelBytes := benchModelPath(b)
	tk := benchTokenizer(b)

	rt, err := NewRuntime(ctx)
	if err != nil {
		b.Fatalf("NewRuntime failed: %v", err)
	}
	defer rt.Close(ctx)

	model, err := rt.LoadModel(ctx, modelBytes)
	if err != nil {
		b.Fatalf("LoadModel failed: %v", err)
	}
	defer model.Close(ctx)

	embedder, err := NewEmbedder(ctx, rt, model, tk, WithEmbeddingNormalize(true))
	if err != nil {
		b.Fatalf("NewEmbedder failed: %v", err)
	}

	texts := map[string]string{
		"Short": "hello",
		"Long":  "これは埋め込みモデルのベンチマークのための、ある程度の長さを持つ日本語のテキストです。モデルの推論時間を計測するために使用します。",
	}

	for name, text := range texts {
		b.Run(name, func(b *testing.B) {
			// Sanity check once before timing.
			vec, err := embedder.Embed(ctx, text)
			if err != nil {
				b.Fatalf("Embed failed: %v", err)
			}
			if len(vec) == 0 {
				b.Fatal("empty embedding")
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := embedder.Embed(ctx, text); err != nil {
					b.Fatalf("Embed failed: %v", err)
				}
			}
		})
	}
}
