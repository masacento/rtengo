package rtengo

import (
	"context"
	"io"

	rtenwasm "github.com/masacento/rtengo/internal/rtenwasm"
)

// Runtime holds the WebAssembly runtime and module.
type Runtime struct {
	rten *rtenwasm.Runtime
}

// RuntimeOption configures runtime creation.
type RuntimeOption func(*RuntimeConfig)

// RuntimeConfig holds options for runtime creation.
type RuntimeConfig struct {
	Stdout io.Writer
	Stderr io.Writer
}

// WithStdout sets the runtime stdout.
func WithStdout(w io.Writer) RuntimeOption {
	return func(cfg *RuntimeConfig) {
		cfg.Stdout = w
	}
}

// WithStderr sets the runtime stderr.
func WithStderr(w io.Writer) RuntimeOption {
	return func(cfg *RuntimeConfig) {
		cfg.Stderr = w
	}
}

// NewRuntime creates a new RTen runtime from a WASM module.
func NewRuntime(ctx context.Context, opts ...RuntimeOption) (*Runtime, error) {
	cfg := RuntimeConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	rt, err := rtenwasm.NewRuntime(ctx, rtenwasm.Config{
		Stdout: cfg.Stdout,
		Stderr: cfg.Stderr,
	})
	if err != nil {
		return nil, err
	}

	return &Runtime{
		rten: rt,
	}, nil
}

// Close releases resources associated with the runtime.
func (r *Runtime) Close(ctx context.Context) error {
	return r.rten.Close(ctx)
}
