package rtengowasm

import (
	"context"
	"math"
	"testing"
)

func TestFloatTensorRoundTrip(t *testing.T) {
	rt := newTestRuntime(t)
	defer rt.Close(context.Background())

	shape := []uint64{2, 2}
	data := []float32{1.25, -2.5, 3.75, 4.5}

	id, err := rt.CreateFloatTensor(context.Background(), shape, data)
	if err != nil {
		t.Fatalf("CreateFloatTensor failed: %v", err)
	}
	defer rt.FreeTensor(context.Background(), id)

	gotShape, err := rt.GetTensorShape(context.Background(), id)
	if err != nil {
		t.Fatalf("GetTensorShape failed: %v", err)
	}
	if len(gotShape) != len(shape) {
		t.Fatalf("shape len mismatch: %v vs %v", gotShape, shape)
	}
	for i := range shape {
		if gotShape[i] != shape[i] {
			t.Fatalf("shape mismatch at %d: %d vs %d", i, gotShape[i], shape[i])
		}
	}

	got, err := rt.GetFloatData(context.Background(), id, uint64(len(data)))
	if err != nil {
		t.Fatalf("GetFloatData failed: %v", err)
	}
	if len(got) != len(data) {
		t.Fatalf("data len mismatch: %d vs %d", len(got), len(data))
	}
	for i := range data {
		if math.Abs(float64(got[i]-data[i])) > 1e-6 {
			t.Fatalf("data mismatch at %d: %f vs %f", i, got[i], data[i])
		}
	}
}

func TestIntTensorRoundTrip(t *testing.T) {
	rt := newTestRuntime(t)
	defer rt.Close(context.Background())

	shape := []uint64{3}
	data := []int32{1, -2, 3}

	id, err := rt.CreateIntTensor(context.Background(), shape, data)
	if err != nil {
		t.Fatalf("CreateIntTensor failed: %v", err)
	}
	defer rt.FreeTensor(context.Background(), id)

	gotShape, err := rt.GetTensorShape(context.Background(), id)
	if err != nil {
		t.Fatalf("GetTensorShape failed: %v", err)
	}
	if len(gotShape) != len(shape) {
		t.Fatalf("shape len mismatch: %v vs %v", gotShape, shape)
	}
	for i := range shape {
		if gotShape[i] != shape[i] {
			t.Fatalf("shape mismatch at %d: %d vs %d", i, gotShape[i], shape[i])
		}
	}

	got, err := rt.GetIntData(context.Background(), id, uint64(len(data)))
	if err != nil {
		t.Fatalf("GetIntData failed: %v", err)
	}
	if len(got) != len(data) {
		t.Fatalf("data len mismatch: %d vs %d", len(got), len(data))
	}
	for i := range data {
		if got[i] != data[i] {
			t.Fatalf("data mismatch at %d: %d vs %d", i, got[i], data[i])
		}
	}
}
