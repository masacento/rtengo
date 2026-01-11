package rtengo

import (
	"context"
	"errors"
)

// Tensor represents a tensor in the WASM memory.
type Tensor struct {
	runtime *Runtime
	id      uint32
	shape   []uint64
	dtype   DataType
}

// DataType represents the data type of a tensor.
type DataType int

const (
	DTypeUnknown DataType = 0
	DTypeFloat32 DataType = 1
	DTypeInt32   DataType = 2
	DTypeInt8    DataType = 3
	DTypeUint8   DataType = 4
)

// NewTensorFloat32 creates a float32 tensor.
func (r *Runtime) NewTensorFloat32(shape []uint64, data []float32) (*Tensor, error) {
	id, err := r.rten.CreateFloatTensor(context.Background(), shape, data)
	if err != nil {
		return nil, err
	}

	return &Tensor{
		runtime: r,
		id:      id,
		shape:   shape,
		dtype:   DTypeFloat32,
	}, nil
}

// NewTensorInt32 creates an int32 tensor.
func (r *Runtime) NewTensorInt32(shape []uint64, data []int32) (*Tensor, error) {
	id, err := r.rten.CreateIntTensor(context.Background(), shape, data)
	if err != nil {
		return nil, err
	}

	return &Tensor{
		runtime: r,
		id:      id,
		shape:   shape,
		dtype:   DTypeInt32,
	}, nil
}

func (r *Runtime) createFloatTensor(ctx context.Context, shape []uint64, data []float32) (*Tensor, error) {
	id, err := r.rten.CreateFloatTensor(ctx, shape, data)
	if err != nil {
		return nil, err
	}

	return &Tensor{
		runtime: r,
		id:      id,
		shape:   shape,
		dtype:   DTypeFloat32,
	}, nil
}

func (r *Runtime) createIntTensor(ctx context.Context, shape []uint64, data []int32) (*Tensor, error) {
	id, err := r.rten.CreateIntTensor(ctx, shape, data)
	if err != nil {
		return nil, err
	}

	return &Tensor{
		runtime: r,
		id:      id,
		shape:   shape,
		dtype:   DTypeInt32,
	}, nil
}

// Shape returns the shape of the tensor.
func (t *Tensor) Shape() []uint64 {
	return t.shape
}

// DataType returns the data type of the tensor.
func (t *Tensor) DataType() DataType {
	return t.dtype
}

// Float32Data returns the float32 data of the tensor.
func (t *Tensor) Float32Data(ctx context.Context) ([]float32, error) {
	if t.dtype != DTypeFloat32 {
		return nil, errors.New("tensor is not float32 type")
	}

	dataLen := uint64(1)
	for _, dim := range t.shape {
		dataLen *= dim
	}
	return t.runtime.rten.GetFloatData(ctx, t.id, dataLen)
}

// Int32Data returns the int32 data of the tensor.
func (t *Tensor) Int32Data(ctx context.Context) ([]int32, error) {
	if t.dtype != DTypeInt32 {
		return nil, errors.New("tensor is not int32 type")
	}

	dataLen := uint64(1)
	for _, dim := range t.shape {
		dataLen *= dim
	}
	return t.runtime.rten.GetIntData(ctx, t.id, dataLen)
}

// Close releases resources associated with the tensor.
func (t *Tensor) Close(ctx context.Context) error {
	return t.runtime.rten.FreeTensor(ctx, t.id)
}
