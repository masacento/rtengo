package rtengo

import (
	"context"
	"testing"
)

func TestRuntime_CreateTensorFloat32(t *testing.T) {
	ctx := context.Background()

	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close(ctx)

	shape := []uint64{1, 2}
	data := []float32{1.5, -2.0}

	tensor, err := runtime.NewTensorFloat32(shape, data)
	if err != nil {
		t.Fatalf("Failed to create tensor: %v", err)
	}
	defer tensor.Close(ctx)

	if tensor.DataType() != DTypeFloat32 {
		t.Fatalf("Expected dtype float32, got %v", tensor.DataType())
	}

	if got := tensor.Shape(); len(got) != len(shape) {
		t.Fatalf("Expected shape len %d, got %d", len(shape), len(got))
	}

	values, err := tensor.Float32Data(ctx)
	if err != nil {
		t.Fatalf("Failed to read tensor data: %v", err)
	}
	if len(values) != len(data) {
		t.Fatalf("Expected %d values, got %d", len(data), len(values))
	}
}

func TestRuntime_CreateTensorMismatch(t *testing.T) {
	ctx := context.Background()

	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close(ctx)

	_, err = runtime.NewTensorFloat32([]uint64{2, 2}, []float32{1, 2, 3})
	if err == nil {
		t.Fatal("Expected error for mismatched data length")
	}
}

func TestRuntime_CreateTensorInt32(t *testing.T) {
	ctx := context.Background()

	runtime, err := NewRuntime(ctx)
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}
	defer runtime.Close(ctx)

	shape := []uint64{1, 3}
	data := []int32{1, -2, 3}

	tensor, err := runtime.NewTensorInt32(shape, data)
	if err != nil {
		t.Fatalf("Failed to create int32 tensor: %v", err)
	}
	defer tensor.Close(ctx)

	if tensor.DataType() != DTypeInt32 {
		t.Fatalf("Expected dtype int32, got %v", tensor.DataType())
	}

	if _, err := tensor.Float32Data(ctx); err == nil {
		t.Fatal("Expected error when reading float32 data from int32 tensor")
	}
}
