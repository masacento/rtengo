package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	rten "github.com/masacento/rtengo"
)

func main() {
	var question, contextText string
	var positionalArgs []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-q" || arg == "--q" {
			if i+1 < len(args) {
				question = args[i+1]
				i++
			}
		} else if arg == "-c" || arg == "--c" {
			if i+1 < len(args) {
				contextText = args[i+1]
				i++
			}
		} else if len(arg) > 0 && arg[0] != '-' {
			positionalArgs = append(positionalArgs, arg)
		}
	}

	if len(positionalArgs) < 2 {
		fmt.Println("Usage: go run . <model.rten> <tokenizer.model> -q <question> -c <context>")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  go run . model.rten tokenizer.model -q \"What's your favorite Japanese food?\" -c \"Work deadlines piled up today.\"")
		os.Exit(1)
	}

	modelPath := positionalArgs[0]
	tokenizerPath := positionalArgs[1]

	if question == "" || contextText == "" {
		fmt.Println("Error: Both -q (question) and -c (context) are required for rerank mode")
		os.Exit(1)
	}

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

	reranker, err := rten.NewReranker(ctx, runtime, model, tk)
	if err != nil {
		log.Fatalf("Failed to create reranker: %v", err)
	}

	fmt.Printf("\nQuestion: %s\n", question)
	fmt.Printf("\nContext:\n%s\n", contextText)

	passages := rten.SplitIntoSentences(contextText)
	fmt.Printf("\nFound %d passage(s) to rank:\n", len(passages))
	for i, passage := range passages {
		fmt.Printf("  [%d]: %s\n", i+1, passage)
	}

	fmt.Println("\nReranking passages...")
	results, err := reranker.Rerank(ctx, question, passages)
	if err != nil {
		log.Fatalf("Failed to rerank passages: %v", err)
	}

	fmt.Println("\nRanked results (by relevance):")
	for rank, result := range results {
		fmt.Printf("\n  Rank %d (original position %d):\n", rank+1, result.OriginalIndex+1)
		fmt.Printf("    Passage: %s\n", result.Passage)
		fmt.Printf("    Score: %.6f\n", result.Score)
		fmt.Printf("    Probability (sigmoid): %.6f\n", result.Probability)
	}
}
