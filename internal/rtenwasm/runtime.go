package rtengowasm

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Config controls runtime creation.
type Config struct {
	Stdout io.Writer
	Stderr io.Writer
}

// Runtime holds the WebAssembly runtime and module.
type Runtime struct {
	rt     wazero.Runtime
	mod    api.Module
	memory api.Memory
}

// NewRuntime creates a new RTen runtime from a WASM module.
func NewRuntime(ctx context.Context, cfg Config) (*Runtime, error) {
	wasmBytes := EmbeddedWASM()
	if len(wasmBytes) == 0 {
		return nil, errors.New("WASM module is empty")
	}

	rt := wazero.NewRuntime(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	compiled, err := rt.CompileModule(ctx, wasmBytes)
	if err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	config := wazero.NewModuleConfig().
		WithStdout(cfg.Stdout).
		WithStderr(cfg.Stderr).
		WithName("rten")

	mod, err := rt.InstantiateModule(ctx, compiled, config)
	if err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	memory := mod.Memory()
	if memory == nil {
		_ = rt.Close(ctx)
		return nil, errors.New("WASM module has no memory export")
	}

	r := &Runtime{
		rt:     rt,
		mod:    mod,
		memory: memory,
	}

	initFunc := r.mod.ExportedFunction("rten_init")
	if initFunc == nil {
		_ = rt.Close(ctx)
		return nil, errors.New("rten_init function not found")
	}

	results, err := initFunc.Call(ctx)
	if err != nil {
		_ = rt.Close(ctx)
		return nil, fmt.Errorf("rten_init failed: %w", err)
	}
	if results[0] == 0 {
		_ = rt.Close(ctx)
		return nil, errors.New("rten_init returned failure")
	}

	return r, nil
}

// Close releases resources associated with the runtime.
func (r *Runtime) Close(ctx context.Context) error {
	return r.rt.Close(ctx)
}

// Call invokes an exported function by name.
func (r *Runtime) Call(ctx context.Context, name string, params ...uint64) ([]uint64, error) {
	fn := r.mod.ExportedFunction(name)
	if fn == nil {
		return nil, fmt.Errorf("%s function not found", name)
	}

	return fn.Call(ctx, params...)
}

// Allocate allocates memory in the WASM module.
func (r *Runtime) Allocate(ctx context.Context, size uint32) (uint32, error) {
	results, err := r.Call(ctx, "allocate", uint64(size))
	if err != nil {
		return 0, fmt.Errorf("allocate failed: %w", err)
	}

	return uint32(results[0]), nil
}

// Deallocate frees memory in the WASM module.
func (r *Runtime) Deallocate(ctx context.Context, ptr, size uint32) error {
	_, err := r.Call(ctx, "deallocate", uint64(ptr), uint64(size))
	return err
}

// GetError retrieves the last error message from RTen.
func (r *Runtime) GetError(ctx context.Context) string {
	results, err := r.Call(ctx, "rten_get_error_len")
	if err != nil || results[0] == 0 {
		return ""
	}

	errLen := uint32(results[0])
	errPtr, err := r.Allocate(ctx, errLen)
	if err != nil {
		return ""
	}
	defer r.Deallocate(ctx, errPtr, errLen)

	results, err = r.Call(ctx, "rten_get_error", uint64(errPtr), uint64(errLen))
	if err != nil || results[0] == 0 {
		return ""
	}

	errBytes, ok := r.MemoryRead(errPtr, errLen)
	if !ok {
		return ""
	}

	return string(errBytes)
}
