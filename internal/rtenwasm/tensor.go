package rtengowasm

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

// CreateFloatTensor creates a float32 tensor and returns its ID.
func (r *Runtime) CreateFloatTensor(ctx context.Context, shape []uint64, data []float32) (uint32, error) {
	expectedLen := uint64(1)
	for _, dim := range shape {
		expectedLen *= dim
	}
	if uint64(len(data)) != expectedLen {
		return 0, fmt.Errorf("data length %d does not match shape (expected %d)", len(data), expectedLen)
	}

	shapePtr, err := r.Allocate(ctx, uint32(len(shape)*4))
	if err != nil {
		return 0, fmt.Errorf("failed to allocate memory for shape: %w", err)
	}
	defer r.Deallocate(ctx, shapePtr, uint32(len(shape)*4))

	shapeBytes := make([]byte, len(shape)*4)
	for i, dim := range shape {
		binary.LittleEndian.PutUint32(shapeBytes[i*4:], uint32(dim))
	}
	if !r.MemoryWrite(shapePtr, shapeBytes) {
		return 0, errors.New("failed to write shape to WASM memory")
	}

	dataPtr, err := r.Allocate(ctx, uint32(len(data)*4))
	if err != nil {
		return 0, fmt.Errorf("failed to allocate memory for data: %w", err)
	}
	defer r.Deallocate(ctx, dataPtr, uint32(len(data)*4))

	dataBytes := make([]byte, len(data)*4)
	for i, val := range data {
		binary.LittleEndian.PutUint32(dataBytes[i*4:], math.Float32bits(val))
	}
	if !r.MemoryWrite(dataPtr, dataBytes) {
		return 0, errors.New("failed to write data to WASM memory")
	}

	results, err := r.Call(ctx,
		"rten_create_float_tensor",
		uint64(shapePtr), uint64(len(shape)),
		uint64(dataPtr), uint64(len(data)),
	)
	if err != nil {
		return 0, fmt.Errorf("rten_create_float_tensor failed: %w", err)
	}

	tensorID := uint32(results[0])
	if tensorID == 0 {
		errMsg := r.GetError(ctx)
		if errMsg != "" {
			return 0, fmt.Errorf("failed to create tensor: %s", errMsg)
		}
		return 0, errors.New("failed to create tensor")
	}

	return tensorID, nil
}

// CreateIntTensor creates an int32 tensor and returns its ID.
func (r *Runtime) CreateIntTensor(ctx context.Context, shape []uint64, data []int32) (uint32, error) {
	expectedLen := uint64(1)
	for _, dim := range shape {
		expectedLen *= dim
	}
	if uint64(len(data)) != expectedLen {
		return 0, fmt.Errorf("data length %d does not match shape (expected %d)", len(data), expectedLen)
	}

	shapePtr, err := r.Allocate(ctx, uint32(len(shape)*4))
	if err != nil {
		return 0, fmt.Errorf("failed to allocate memory for shape: %w", err)
	}
	defer r.Deallocate(ctx, shapePtr, uint32(len(shape)*4))

	shapeBytes := make([]byte, len(shape)*4)
	for i, dim := range shape {
		binary.LittleEndian.PutUint32(shapeBytes[i*4:], uint32(dim))
	}
	if !r.MemoryWrite(shapePtr, shapeBytes) {
		return 0, errors.New("failed to write shape to WASM memory")
	}

	dataPtr, err := r.Allocate(ctx, uint32(len(data)*4))
	if err != nil {
		return 0, fmt.Errorf("failed to allocate memory for data: %w", err)
	}
	defer r.Deallocate(ctx, dataPtr, uint32(len(data)*4))

	dataBytes := make([]byte, len(data)*4)
	for i, val := range data {
		binary.LittleEndian.PutUint32(dataBytes[i*4:], uint32(val))
	}
	if !r.MemoryWrite(dataPtr, dataBytes) {
		return 0, errors.New("failed to write data to WASM memory")
	}

	results, err := r.Call(ctx,
		"rten_create_int_tensor",
		uint64(shapePtr), uint64(len(shape)),
		uint64(dataPtr), uint64(len(data)),
	)
	if err != nil {
		return 0, fmt.Errorf("rten_create_int_tensor failed: %w", err)
	}

	tensorID := uint32(results[0])
	if tensorID == 0 {
		errMsg := r.GetError(ctx)
		if errMsg != "" {
			return 0, fmt.Errorf("failed to create tensor: %s", errMsg)
		}
		return 0, errors.New("failed to create tensor")
	}

	return tensorID, nil
}

// GetTensorShape returns the tensor shape.
func (r *Runtime) GetTensorShape(ctx context.Context, tensorID uint32) ([]uint64, error) {
	results, err := r.Call(ctx, "rten_get_tensor_ndim", uint64(tensorID))
	if err != nil {
		return nil, err
	}

	ndim := uint32(results[0])
	if ndim == 0 {
		return []uint64{}, nil
	}

	shapePtr, err := r.Allocate(ctx, ndim*4)
	if err != nil {
		return nil, err
	}
	defer r.Deallocate(ctx, shapePtr, ndim*4)

	_, err = r.Call(ctx, "rten_get_tensor_shape", uint64(tensorID), uint64(shapePtr), uint64(ndim))
	if err != nil {
		return nil, err
	}

	shapeBytes, ok := r.MemoryRead(shapePtr, ndim*4)
	if !ok {
		return nil, errors.New("failed to read shape from WASM memory")
	}

	shape := make([]uint64, ndim)
	for i := uint32(0); i < ndim; i++ {
		shape[i] = uint64(binary.LittleEndian.Uint32(shapeBytes[i*4:]))
	}

	return shape, nil
}

// GetTensorDType returns the tensor data type.
func (r *Runtime) GetTensorDType(ctx context.Context, tensorID uint32) int {
	results, err := r.Call(ctx, "rten_get_tensor_dtype", uint64(tensorID))
	if err != nil {
		return 0
	}

	return int(results[0])
}

// GetFloatData returns float32 data from a tensor.
func (r *Runtime) GetFloatData(ctx context.Context, tensorID uint32, dataLen uint64) ([]float32, error) {
	dataPtr, err := r.Allocate(ctx, uint32(dataLen*4))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for data: %w", err)
	}
	defer r.Deallocate(ctx, dataPtr, uint32(dataLen*4))

	results, err := r.Call(ctx, "rten_get_float_data", uint64(tensorID), uint64(dataPtr), uint64(dataLen))
	if err != nil {
		return nil, fmt.Errorf("rten_get_float_data failed: %w", err)
	}

	copiedLen := uint32(results[0])
	if copiedLen == 0 {
		errMsg := r.GetError(ctx)
		if errMsg != "" {
			return nil, fmt.Errorf("failed to get tensor data: %s", errMsg)
		}
		return nil, errors.New("failed to get tensor data")
	}

	dataBytes, ok := r.MemoryRead(dataPtr, copiedLen*4)
	if !ok {
		return nil, errors.New("failed to read data from WASM memory")
	}

	data := make([]float32, copiedLen)
	for i := uint32(0); i < copiedLen; i++ {
		bits := binary.LittleEndian.Uint32(dataBytes[i*4:])
		data[i] = math.Float32frombits(bits)
	}

	return data, nil
}

// GetIntData returns int32 data from a tensor.
func (r *Runtime) GetIntData(ctx context.Context, tensorID uint32, dataLen uint64) ([]int32, error) {
	dataPtr, err := r.Allocate(ctx, uint32(dataLen*4))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for data: %w", err)
	}
	defer r.Deallocate(ctx, dataPtr, uint32(dataLen*4))

	results, err := r.Call(ctx, "rten_get_int_data", uint64(tensorID), uint64(dataPtr), uint64(dataLen))
	if err != nil {
		return nil, fmt.Errorf("rten_get_int_data failed: %w", err)
	}

	copiedLen := uint32(results[0])
	if copiedLen == 0 {
		errMsg := r.GetError(ctx)
		if errMsg != "" {
			return nil, fmt.Errorf("failed to get tensor data: %s", errMsg)
		}
		return nil, errors.New("failed to get tensor data")
	}

	dataBytes, ok := r.MemoryRead(dataPtr, copiedLen*4)
	if !ok {
		return nil, errors.New("failed to read data from WASM memory")
	}

	data := make([]int32, copiedLen)
	for i := uint32(0); i < copiedLen; i++ {
		data[i] = int32(binary.LittleEndian.Uint32(dataBytes[i*4:]))
	}

	return data, nil
}

// FreeTensor releases tensor resources.
func (r *Runtime) FreeTensor(ctx context.Context, tensorID uint32) error {
	_, err := r.Call(ctx, "rten_free_tensor", uint64(tensorID))
	return err
}
