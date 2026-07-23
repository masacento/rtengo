package rtengowasm

import (
	"context"
	"fmt"
	"io"

	"github.com/masacento/rtengo/internal/genwasm"
	"github.com/masacento/rtengo/internal/genwasm/base"
)

// Config controls runtime creation.
type Config struct {
	Stdout io.Writer
	Stderr io.Writer
}

// Runtime holds the AOT-compiled RTen module.
type Runtime struct {
	mod *base.Module
}

// NewRuntime creates a new RTen runtime from the wasm2go-generated module.
func NewRuntime(ctx context.Context, cfg Config) (*Runtime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	wasi := base.DefaultWASI()
	// Match the previous wazero behavior: guest stdout/stderr are discarded
	// unless a writer is explicitly configured.
	if cfg.Stdout != nil {
		wasi.SetStdout(cfg.Stdout)
	} else {
		wasi.SetStdout(io.Discard)
	}
	if cfg.Stderr != nil {
		wasi.SetStderr(cfg.Stderr)
	} else {
		wasi.SetStderr(io.Discard)
	}

	r := &Runtime{
		mod: genwasm.NewWithWASI(wasi),
	}

	result, err := r.call0(genwasm.RtenInit)
	if err != nil {
		return nil, fmt.Errorf("rten_init failed: %w", err)
	}
	if result == 0 {
		return nil, fmt.Errorf("rten_init returned failure")
	}

	return r, nil
}

// Close releases resources associated with the runtime.
// The generated module is plain Go memory, so there is nothing to release;
// resources are left to the garbage collector.
func (r *Runtime) Close(ctx context.Context) error {
	return nil
}

// call0 invokes a generated export with no arguments, converting wasm traps
// (panics) into errors.
func (r *Runtime) call0(fn func(*base.Module) int32) (result uint64, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("wasm trap: %v", rec)
		}
	}()
	return uint64(uint32(fn(r.mod))), nil
}

// Call invokes an exported function by name.
func (r *Runtime) Call(ctx context.Context, name string, params ...uint64) ([]uint64, error) {
	p := func(i int) int32 { return int32(uint32(params[i])) }

	var fn func(*base.Module) int32
	switch name {
	case "allocate":
		if len(params) != 1 {
			return nil, fmt.Errorf("allocate expects 1 param, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.Allocate(m, p(0)) }
	case "rten_get_error_len":
		if len(params) != 0 {
			return nil, fmt.Errorf("rten_get_error_len expects 0 params, got %d", len(params))
		}
		fn = genwasm.RtenGetErrorLen
	case "rten_get_error":
		if len(params) != 2 {
			return nil, fmt.Errorf("rten_get_error expects 2 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetError(m, p(0), p(1)) }
	case "rten_load_model":
		if len(params) != 2 {
			return nil, fmt.Errorf("rten_load_model expects 2 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenLoadModel(m, p(0), p(1)) }
	case "rten_free_model":
		if len(params) != 1 {
			return nil, fmt.Errorf("rten_free_model expects 1 param, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenFreeModel(m, p(0)) }
	case "rten_get_input_count":
		if len(params) != 1 {
			return nil, fmt.Errorf("rten_get_input_count expects 1 param, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetInputCount(m, p(0)) }
	case "rten_get_output_count":
		if len(params) != 1 {
			return nil, fmt.Errorf("rten_get_output_count expects 1 param, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetOutputCount(m, p(0)) }
	case "rten_get_input_dims":
		if len(params) != 4 {
			return nil, fmt.Errorf("rten_get_input_dims expects 4 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetInputDims(m, p(0), p(1), p(2), p(3)) }
	case "rten_run":
		if len(params) != 5 {
			return nil, fmt.Errorf("rten_run expects 5 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenRun(m, p(0), p(1), p(2), p(3), p(4)) }
	case "rten_create_float_tensor":
		if len(params) != 4 {
			return nil, fmt.Errorf("rten_create_float_tensor expects 4 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenCreateFloatTensor(m, p(0), p(1), p(2), p(3)) }
	case "rten_create_int_tensor":
		if len(params) != 4 {
			return nil, fmt.Errorf("rten_create_int_tensor expects 4 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenCreateIntTensor(m, p(0), p(1), p(2), p(3)) }
	case "rten_get_tensor_ndim":
		if len(params) != 1 {
			return nil, fmt.Errorf("rten_get_tensor_ndim expects 1 param, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetTensorNdim(m, p(0)) }
	case "rten_get_tensor_shape":
		if len(params) != 3 {
			return nil, fmt.Errorf("rten_get_tensor_shape expects 3 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetTensorShape(m, p(0), p(1), p(2)) }
	case "rten_get_tensor_dtype":
		if len(params) != 1 {
			return nil, fmt.Errorf("rten_get_tensor_dtype expects 1 param, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetTensorDtype(m, p(0)) }
	case "rten_get_float_data":
		if len(params) != 3 {
			return nil, fmt.Errorf("rten_get_float_data expects 3 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetFloatData(m, p(0), p(1), p(2)) }
	case "rten_get_int_data":
		if len(params) != 3 {
			return nil, fmt.Errorf("rten_get_int_data expects 3 params, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenGetIntData(m, p(0), p(1), p(2)) }
	case "rten_free_tensor":
		if len(params) != 1 {
			return nil, fmt.Errorf("rten_free_tensor expects 1 param, got %d", len(params))
		}
		fn = func(m *base.Module) int32 { return genwasm.RtenFreeTensor(m, p(0)) }
	default:
		return nil, fmt.Errorf("%s function not found", name)
	}

	result, err := r.call0(fn)
	if err != nil {
		return nil, err
	}
	return []uint64{result}, nil
}

// Allocate allocates memory in the module.
func (r *Runtime) Allocate(ctx context.Context, size uint32) (uint32, error) {
	results, err := r.Call(ctx, "allocate", uint64(size))
	if err != nil {
		return 0, fmt.Errorf("allocate failed: %w", err)
	}

	return uint32(results[0]), nil
}

// Deallocate frees memory in the module.
func (r *Runtime) Deallocate(ctx context.Context, ptr, size uint32) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("wasm trap: %v", rec)
		}
	}()
	genwasm.Deallocate(r.mod, int32(ptr), int32(size))
	return nil
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
