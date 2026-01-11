package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	rten "github.com/masacento/rtengo"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 3 {
		fmt.Println("Usage: go run . <model.rten> <tokenizer.model> <text>")
		fmt.Println("Example: go run . model.rten tokenizer.model \"text\"")
		os.Exit(1)
	}

	modelPath := args[0]
	tokenizerPath := args[1]
	text := args[2]

	ctx := context.Background()

	fmt.Println("Loading WASM module...")
	start := time.Now()
	runtime, err := rten.NewRuntime(ctx)
	if err != nil {
		log.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close(ctx)
	fmt.Printf("RTen runtime initialized (%s)\n", time.Since(start).Truncate(time.Millisecond))

	fmt.Println("Loading model...")
	modelBytes, err := os.ReadFile(modelPath)
	if err != nil {
		log.Fatalf("Failed to read model file: %v", err)
	}

	model, err := runtime.LoadModel(ctx, modelBytes)
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}
	defer model.Close(ctx)
	fmt.Println("Model loaded successfully")

	fmt.Println("Loading tokenizer...")
	tk, err := rten.NewSentencePiece(tokenizerPath)
	if err != nil {
		log.Fatalf("Failed to load tokenizer: %v", err)
	}
	fmt.Println("Tokenizer loaded successfully")

	embedder, err := rten.NewEmbedder(ctx, runtime, model, tk)
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
	}

	fmt.Printf("Generating embedding for: \"%s\"\n", text)
	embedding, err := embedder.Embed(ctx, text)
	if err != nil {
		log.Fatalf("Failed to generate embedding: %v", err)
	}

	fmt.Printf("\nEmbedding dimension: %d\n", len(embedding))
	fmt.Println("\nEmbedding vector (first 10 values):")
	for i := 0; i < 10 && i < len(embedding); i++ {
		fmt.Printf("  [%3d]: %.6f\n", i, embedding[i])
	}
	fmt.Println("  ...")

	var sumSq float64
	for _, v := range embedding {
		sumSq += float64(v) * float64(v)
	}
	norm := math.Sqrt(sumSq)
	fmt.Printf("\nL2 norm: %.6f\n", norm)
}
