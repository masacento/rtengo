package rtengo

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
)

// Model represents a loaded machine learning model.
type Model struct {
	runtime *Runtime
	id      uint32
}

// LoadModel loads a model from binary data.
func (r *Runtime) LoadModel(ctx context.Context, model []byte) (*Model, error) {
	if len(model) == 0 {
		return nil, errors.New("model data is empty")
	}

	modelPtr, err := r.rten.Allocate(ctx, uint32(len(model)))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for model: %w", err)
	}

	if !r.rten.MemoryWrite(modelPtr, model) {
		_ = r.rten.Deallocate(ctx, modelPtr, uint32(len(model)))
		return nil, errors.New("failed to write model data to WASM memory")
	}

	results, err := r.rten.Call(ctx, "rten_load_model", uint64(modelPtr), uint64(len(model)))
	if err != nil {
		_ = r.rten.Deallocate(ctx, modelPtr, uint32(len(model)))
		return nil, fmt.Errorf("rten_load_model failed: %w", err)
	}

	modelID := uint32(results[0])
	if modelID == 0 {
		errMsg := r.rten.GetError(ctx)
		if errMsg != "" {
			return nil, fmt.Errorf("failed to load model: %s", errMsg)
		}
		return nil, errors.New("failed to load model")
	}

	return &Model{
		runtime: r,
		id:      modelID,
	}, nil
}

// GetInputCount returns the number of model inputs.
func (m *Model) GetInputCount(ctx context.Context) (uint32, error) {
	results, err := m.runtime.rten.Call(ctx, "rten_get_input_count", uint64(m.id))
	if err != nil {
		return 0, fmt.Errorf("rten_get_input_count failed: %w", err)
	}

	return uint32(results[0]), nil
}

// GetOutputCount returns the number of model outputs.
func (m *Model) GetOutputCount(ctx context.Context) (uint32, error) {
	results, err := m.runtime.rten.Call(ctx, "rten_get_output_count", uint64(m.id))
	if err != nil {
		return 0, fmt.Errorf("rten_get_output_count failed: %w", err)
	}

	return uint32(results[0]), nil
}

// GetInputDims returns the dimensions of the specified input.
func (m *Model) GetInputDims(ctx context.Context, inputIndex uint32) ([]int32, error) {
	const maxDims = 8

	dimsPtr, err := m.runtime.rten.Allocate(ctx, uint32(maxDims*4))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for dimensions: %w", err)
	}
	defer m.runtime.rten.Deallocate(ctx, dimsPtr, uint32(maxDims*4))

	results, err := m.runtime.rten.Call(ctx, "rten_get_input_dims", uint64(m.id), uint64(inputIndex), uint64(dimsPtr), uint64(maxDims))
	if err != nil {
		return nil, fmt.Errorf("rten_get_input_dims failed: %w", err)
	}

	dimCount := uint32(results[0])
	if dimCount == 0 {
		return nil, nil
	}

	dimsBytes, ok := m.runtime.rten.MemoryRead(dimsPtr, dimCount*4)
	if !ok {
		return nil, errors.New("failed to read dimensions from WASM memory")
	}

	dims := make([]int32, dimCount)
	for i := uint32(0); i < dimCount; i++ {
		dims[i] = int32(binary.LittleEndian.Uint32(dimsBytes[i*4:]))
	}

	return dims, nil
}

// Run executes the model with the given input tensors.
func (m *Model) Run(ctx context.Context, inputs ...*Tensor) ([]*Tensor, error) {
	if len(inputs) == 0 {
		return nil, errors.New("at least one input tensor required")
	}

	inputIdsPtr, err := m.runtime.rten.Allocate(ctx, uint32(len(inputs)*4))
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for input IDs: %w", err)
	}
	defer m.runtime.rten.Deallocate(ctx, inputIdsPtr, uint32(len(inputs)*4))

	inputIdsBytes := make([]byte, len(inputs)*4)
	for i, t := range inputs {
		binary.LittleEndian.PutUint32(inputIdsBytes[i*4:], t.id)
	}
	if !m.runtime.rten.MemoryWrite(inputIdsPtr, inputIdsBytes) {
		return nil, errors.New("failed to write input IDs to WASM memory")
	}

	outputCount, err := m.GetOutputCount(ctx)
	if err != nil {
		return nil, err
	}
	if outputCount == 0 {
		outputCount = 16
	}

	outputIdsPtr, err := m.runtime.rten.Allocate(ctx, outputCount*4)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory for output IDs: %w", err)
	}
	defer m.runtime.rten.Deallocate(ctx, outputIdsPtr, outputCount*4)

	results, err := m.runtime.rten.Call(ctx,
		"rten_run",
		uint64(m.id),
		uint64(inputIdsPtr), uint64(len(inputs)),
		uint64(outputIdsPtr), uint64(outputCount),
	)
	if err != nil {
		return nil, fmt.Errorf("rten_run failed: %w", err)
	}

	numOutputs := uint32(results[0])
	if numOutputs == 0 {
		errMsg := m.runtime.rten.GetError(ctx)
		if errMsg != "" {
			return nil, fmt.Errorf("inference failed: %s", errMsg)
		}
		return nil, errors.New("inference failed")
	}

	outputIdsBytes, ok := m.runtime.rten.MemoryRead(outputIdsPtr, numOutputs*4)
	if !ok {
		return nil, errors.New("failed to read output IDs from WASM memory")
	}

	outputs := make([]*Tensor, numOutputs)
	for i := uint32(0); i < numOutputs; i++ {
		tensorID := binary.LittleEndian.Uint32(outputIdsBytes[i*4:])

		shape, err := m.runtime.rten.GetTensorShape(ctx, tensorID)
		if err != nil {
			return nil, fmt.Errorf("failed to get output tensor shape: %w", err)
		}

		dtype := DataType(m.runtime.rten.GetTensorDType(ctx, tensorID))

		outputs[i] = &Tensor{
			runtime: m.runtime,
			id:      tensorID,
			shape:   shape,
			dtype:   dtype,
		}
	}

	return outputs, nil
}

// Close releases resources associated with the model.
func (m *Model) Close(ctx context.Context) error {
	_, err := m.runtime.rten.Call(ctx, "rten_free_model", uint64(m.id))
	return err
}
