# rtengo

Lightweight Go bindings for the RTen WebAssembly runtime. It runs RTen WASM on top of `wazero` and provides
high-level APIs for text embedding, reranking, pruning, and image preprocessing.

## Features

- Pure Go WASM runtime via `wazero` (no CGO)
- RTen WASM embedded by default
- Simple Runtime / Model / Tensor API
- SentencePiece tokenizer support
- High-level task APIs (Embedding / Rerank / Prune)

## Requirements

- Go 1.24+
- RTen `.rten` model file
- SentencePiece `tokenizer.model`

## Install

```bash
go get github.com/masacento/rtengo
```

## Quick Start

```go
package main

import (
  "context"
  "log"
  "os"

  rten "github.com/masacento/rtengo"
)

func main() {
  ctx := context.Background()

  // Runtime
  rt, err := rten.NewRuntime(ctx)
  if err != nil {
    log.Fatal(err)
  }
  defer rt.Close(ctx)

  // Model
  modelBytes, err := os.ReadFile("model.rten")
  if err != nil {
    log.Fatal(err)
  }
  model, err := rt.LoadModel(ctx, modelBytes)
  if err != nil {
    log.Fatal(err)
  }
  defer model.Close(ctx)

  // Tokenizer
  tk, err := rten.NewSentencePiece("tokenizer.model")
  if err != nil {
    log.Fatal(err)
  }

  // Embedding
  embedder, err := rten.NewEmbedder(ctx, rt, model, tk)
  if err != nil {
    log.Fatal(err)
  }
  vec, err := embedder.Embed(ctx, "Hello")
  if err != nil {
    log.Fatal(err)
  }
  _ = vec
}
```

## Core API

### Runtime / Model / Tensor

```go
rt, _ := rten.NewRuntime(ctx)
model, _ := rt.LoadModel(ctx, modelBytes)
outputs, _ := model.Run(ctx, inputTensor)

shape := outputs[0].Shape()
vec, _ := outputs[0].Float32Data(ctx)
```

### Embedding

```go
embedder, _ := rten.NewEmbedder(ctx, rt, model, tk,
  rten.WithEmbeddingNormalize(true),
  rten.WithEmbeddingPooling(rten.PoolingMean),
)
vec, _ := embedder.Embed(ctx, "text")
```

### Rerank

```go
reranker, _ := rten.NewReranker(ctx, rt, model, tk)
results, _ := reranker.Rerank(ctx, "question", passages)
```

### Prune (OpenProvence, etc.)

```go
pruner, _ := rten.NewPruner(ctx, rt, model, tk, rten.PrunerConfig{Threshold: 0.5})
result, _ := pruner.Prune(ctx, "query", "document")
```

### Image Preprocessing

```go
data, _ := rten.LoadAndPreprocessFile(ctx, "cat.jpg", 224, 224)
inputTensor, _ := rt.NewTensorFloat32([]uint64{1, 3, 224, 224}, data)
```

## Examples

- `examples/embedding`: text embedding
- `examples/rerank`: passage reranking
- `examples/openprovence`: document pruning
- `examples/mobilenet`: image classification

## License

MIT
