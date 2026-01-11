package rtengo

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"math"
	"testing"
)

func TestLoadAndPreprocess(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode image: %v", err)
	}

	ctx := context.Background()
	data, err := LoadAndPreprocess(ctx, &buf, 2, 2)
	if err != nil {
		t.Fatalf("Failed to preprocess image: %v", err)
	}

	if len(data) != 3*2*2 {
		t.Fatalf("Expected %d values, got %d", 12, len(data))
	}
}

func TestLoadAndPreprocess_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode image: %v", err)
	}

	if _, err := LoadAndPreprocess(ctx, &buf, 1, 1); err == nil {
		t.Fatal("Expected error for canceled context")
	}
}

func TestLoadAndPreprocess_WithNormalization(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode image: %v", err)
	}

	ctx := context.Background()
	data, err := LoadAndPreprocess(ctx, &buf, 1, 1, WithNormalization([3]float32{0, 0, 0}, [3]float32{1, 1, 1}))
	if err != nil {
		t.Fatalf("Failed to preprocess image: %v", err)
	}
	if len(data) != 3 {
		t.Fatalf("Expected 3 values, got %d", len(data))
	}
	for i, v := range data {
		if math.Abs(float64(v-1.0)) > 1e-6 {
			t.Fatalf("Expected channel %d to be ~1.0, got %f", i, v)
		}
	}
}

func TestLoadAndPreprocessFile_Missing(t *testing.T) {
	ctx := context.Background()
	if _, err := LoadAndPreprocessFile(ctx, "no_such_file.png", 1, 1); err == nil {
		t.Fatal("Expected error for missing file")
	}
}
