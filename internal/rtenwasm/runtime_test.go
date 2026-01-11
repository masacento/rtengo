package rtengowasm

import (
	"bytes"
	"context"
	"testing"
)

func TestRuntimeCallMissingFunction(t *testing.T) {
	rt := newTestRuntime(t)
	defer rt.Close(context.Background())

	_, err := rt.Call(context.Background(), "no_such_function")
	if err == nil {
		t.Fatal("expected error for missing function")
	}
}

func TestMemoryReadWrite(t *testing.T) {
	rt := newTestRuntime(t)
	defer rt.Close(context.Background())

	payload := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	ptr, err := rt.Allocate(context.Background(), uint32(len(payload)))
	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}
	defer rt.Deallocate(context.Background(), ptr, uint32(len(payload)))

	if ok := rt.MemoryWrite(ptr, payload); !ok {
		t.Fatal("memory write failed")
	}

	read, ok := rt.MemoryRead(ptr, uint32(len(payload)))
	if !ok {
		t.Fatal("memory read failed")
	}

	if !bytes.Equal(payload, read) {
		t.Fatalf("memory mismatch: %v vs %v", payload, read)
	}
}

func TestGetErrorEmpty(t *testing.T) {
	rt := newTestRuntime(t)
	defer rt.Close(context.Background())

	if msg := rt.GetError(context.Background()); msg != "" {
		t.Fatalf("expected empty error, got %q", msg)
	}
}

func newTestRuntime(t *testing.T) *Runtime {
	t.Helper()

	rt, err := NewRuntime(context.Background(), Config{})
	if err != nil {
		t.Fatalf("failed to create runtime: %v", err)
	}
	return rt
}
