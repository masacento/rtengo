package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	rten "github.com/masacento/rtengo"
)

func main() {
	var query, document string
	var threshold float32 = 0.5
	var simple bool
	var positionalArgs []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-q", "--query":
			if i+1 < len(args) {
				query = args[i+1]
				i++
			}
		case "-d", "--document":
			if i+1 < len(args) {
				document = args[i+1]
				i++
			}
		case "-t", "--threshold":
			if i+1 < len(args) {
				if val, err := strconv.ParseFloat(args[i+1], 32); err == nil {
					threshold = float32(val)
				}
				i++
			}
		case "-s", "--simple":
			simple = true
		default:
			if len(arg) > 0 && arg[0] != '-' {
				positionalArgs = append(positionalArgs, arg)
			}
		}
	}

	if len(positionalArgs) < 2 {
		fmt.Println("Usage: go run . <model.rten> <tokenizer.json> -q <query> -d <document> [-t <threshold>] [-s]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -q, --query      Query string")
		fmt.Println("  -d, --document   Document string to prune")
		fmt.Println("  -t, --threshold  Pruning threshold (0.0-1.0, default: 0.5)")
		fmt.Println("  -s, --simple     Simple output (pruned document only)")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  go run . model.rten tokenizer.json -q \"フランスの首都は？\" -d \"パリはフランスの首都です。東京は日本の首都です。\"")
		os.Exit(1)
	}

	modelPath := positionalArgs[0]
	tokenizerPath := positionalArgs[1]

	if query == "" || document == "" {
		fmt.Println("Error: Both -q (query) and -d (document) are required")
		os.Exit(1)
	}

	ctx := context.Background()

	if !simple {
		fmt.Println("Loading WASM module...")
	}
	start := time.Now()
	runtime, err := rten.NewRuntime(ctx)
	if err != nil {
		log.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close(ctx)
	if !simple {
		fmt.Printf("RTen runtime initialized (%s)\n", time.Since(start).Truncate(time.Millisecond))
	}

	if !simple {
		fmt.Println("Loading model...")
	}
	modelBytes, err := os.ReadFile(modelPath)
	if err != nil {
		log.Fatalf("Failed to read model file: %v", err)
	}

	model, err := runtime.LoadModel(ctx, modelBytes)
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}
	defer model.Close(ctx)
	if !simple {
		fmt.Println("Model loaded successfully")
	}

	if !simple {
		fmt.Println("Loading tokenizer...")
	}
	tk, err := rten.NewTokenizerFromFile(tokenizerPath)
	if err != nil {
		log.Fatalf("Failed to load tokenizer: %v", err)
	}
	if !simple {
		fmt.Println("Tokenizer loaded successfully")
	}

	pruner, err := rten.NewPruner(ctx, runtime, model, tk, rten.PrunerConfig{Threshold: threshold})
	if err != nil {
		log.Fatalf("Failed to create pruner: %v", err)
	}

	result, err := pruner.Prune(ctx, query, document)
	if err != nil {
		log.Fatalf("Failed to prune document: %v", err)
	}

	if simple {
		fmt.Println(result.PrunedDocument)
		return
	}

	fmt.Printf("\nQuery: %s\n", query)
	fmt.Printf("Document: %s\n", document)
	fmt.Printf("\nThreshold: %.2f\n", threshold)

	fmt.Println("\n--- Results ---")
	fmt.Printf("Ranking Score: %.6f\n", result.RankingScore)
	fmt.Printf("Ranking Probability (sigmoid): %.6f\n", result.Probability)

	fmt.Println("\nPruning Statistics:")
	fmt.Printf("  Kept tokens: %d / %d (%.1f%%)\n",
		result.KeptTokens, result.TotalTokens, result.KeptRatio*100)
	fmt.Printf("  Compression: %.1f%%\n", (1.0-result.KeptRatio)*100)

	fmt.Println("\nPruned Document:")
	fmt.Println(result.PrunedDocument)
}
