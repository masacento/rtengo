package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	rten "github.com/masacento/rtengo"
)

// softmax applies softmax to convert logits to probabilities.
func softmax(logits []float32) []float32 {
	max := logits[0]
	for _, v := range logits {
		if v > max {
			max = v
		}
	}

	probs := make([]float32, len(logits))
	var sum float32
	for i, v := range logits {
		probs[i] = float32(exp64(float64(v - max)))
		sum += probs[i]
	}

	for i := range probs {
		probs[i] /= sum
	}

	return probs
}

// exp64 computes e^x using Taylor series (simple implementation).
func exp64(x float64) float64 {
	if x > 88 {
		return 1e38
	}
	if x < -88 {
		return 0
	}
	result := 1.0
	term := 1.0
	for i := 1; i < 100; i++ {
		term *= x / float64(i)
		result += term
		if term < 1e-15 && term > -1e-15 {
			break
		}
	}
	return result
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . <image.jpg>")
		fmt.Println("Example: go run . cat.jpg")
		os.Exit(1)
	}
	imagePath := os.Args[1]

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
	modelBytes, err := os.ReadFile("mobilenet.rten")
	if err != nil {
		log.Fatalf("Failed to read model file: %v", err)
	}

	model, err := runtime.LoadModel(ctx, modelBytes)
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}
	defer model.Close(ctx)
	fmt.Println("Model loaded successfully")

	inputCount, err := model.GetInputCount(ctx)
	if err != nil {
		log.Fatalf("Failed to get input count: %v", err)
	}
	fmt.Printf("Model has %d input(s)\n", inputCount)

	outputCount, err := model.GetOutputCount(ctx)
	if err != nil {
		log.Fatalf("Failed to get output count: %v", err)
	}
	fmt.Printf("Model has %d output(s)\n", outputCount)

	for i := uint32(0); i < inputCount; i++ {
		dims, err := model.GetInputDims(ctx, i)
		if err != nil {
			log.Printf("Failed to get input %d dimensions: %v", i, err)
			continue
		}
		fmt.Printf("Input %d shape: %v\n", i, dims)
	}

	const inputWidth, inputHeight = 224, 224
	fmt.Printf("\nLoading image: %s\n", imagePath)
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatalf("Failed to open image: %v", err)
	}
	defer file.Close()

	inputData, err := rten.LoadAndPreprocess(ctx, file, inputWidth, inputHeight)
	if err != nil {
		log.Fatalf("Failed to load image: %v", err)
	}

	inputShape := []uint64{1, 3, inputHeight, inputWidth}
	fmt.Println("Creating input tensor...")
	inputTensor, err := runtime.NewTensorFloat32(inputShape, inputData)
	if err != nil {
		log.Fatalf("Failed to create input tensor: %v", err)
	}
	defer inputTensor.Close(ctx)
	fmt.Printf("Input tensor created with shape: %v\n", inputTensor.Shape())

	fmt.Println("\nRunning inference...")
	outputs, err := model.Run(ctx, inputTensor)
	if err != nil {
		log.Fatalf("Inference failed: %v", err)
	}
	fmt.Printf("Inference completed, got %d output(s)\n", len(outputs))

	for i, output := range outputs {
		defer output.Close(ctx)

		fmt.Printf("\nOutput %d:\n", i)
		fmt.Printf("  Shape: %v\n", output.Shape())

		if output.DataType() == rten.DTypeFloat32 {
			data, err := output.Float32Data(ctx)
			if err != nil {
				log.Printf("Failed to get output data: %v", err)
				continue
			}

			probs := softmax(data)

			if len(probs) > 1 {
				fmt.Println("\nTop 5 predictions:")
				fmt.Println("--------------------------------------------------")
				topIndices := findTopN(probs, 5)
				for j, idx := range topIndices {
					fmt.Printf("  %d. Class %4d: %.4f (%5.1f%%)\n",
						j+1, idx, probs[idx], probs[idx]*100)
				}
			} else {
				fmt.Printf("  Value: %v\n", probs)
			}
		}
	}

	fmt.Println("\nDone!")
}

// findTopN returns indices of the top N values in descending order.
func findTopN(values []float32, n int) []int {
	type indexValue struct {
		index int
		value float32
	}

	pairs := make([]indexValue, len(values))
	for i, v := range values {
		pairs[i] = indexValue{index: i, value: v}
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	result := make([]int, 0, n)
	for i := 0; i < n && i < len(pairs); i++ {
		result = append(result, pairs[i].index)
	}

	return result
}
