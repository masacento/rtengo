use std::alloc::{alloc, dealloc, Layout};
use std::cell::RefCell;
use std::panic::{catch_unwind, AssertUnwindSafe};
use std::slice;

use rten::{DataType, Dimension, Model, NodeId, Value, ValueOrView, ValueType};
use rten_tensor::Layout as TensorLayout;
use rten_tensor::prelude::*;

const DTYPE_FLOAT32: u32 = 1;
const DTYPE_INT32: u32 = 2;
const DTYPE_INT8: u32 = 3;
const DTYPE_UINT8: u32 = 4;

struct State {
    models: Vec<Option<Model>>,
    tensors: Vec<Option<Value>>,
    last_error: String,
}

impl State {
    fn new() -> Self {
        Self {
            models: vec![None],
            tensors: vec![None],
            last_error: String::new(),
        }
    }

    fn clear_error(&mut self) {
        self.last_error.clear();
    }

    fn set_error(&mut self, err: impl ToString) {
        self.last_error = err.to_string();
    }

    fn insert_model(&mut self, model: Model) -> u32 {
        insert_slot(&mut self.models, model)
    }

    fn insert_tensor(&mut self, tensor: Value) -> u32 {
        insert_slot(&mut self.tensors, tensor)
    }
}

thread_local! {
    static STATE: RefCell<State> = RefCell::new(State::new());
}

fn insert_slot<T>(slots: &mut Vec<Option<T>>, value: T) -> u32 {
    for (index, slot) in slots.iter_mut().enumerate().skip(1) {
        if slot.is_none() {
            *slot = Some(value);
            return index as u32;
        }
    }
    slots.push(Some(value));
    (slots.len() - 1) as u32
}

fn shape_from_raw(ptr: *const u32, len: usize) -> Result<Vec<usize>, String> {
    if len > 0 && ptr.is_null() {
        return Err("shape pointer is null".to_string());
    }
    let shape = unsafe { slice::from_raw_parts(ptr, len) };
    Ok(shape.iter().map(|&dim| dim as usize).collect())
}

fn run_with_state<R: Copy>(default: R, f: impl FnOnce(&mut State) -> Result<R, String>) -> R {
    match catch_unwind(AssertUnwindSafe(|| {
        STATE.with(|state| {
            let mut state = state.borrow_mut();
            state.clear_error();
            match f(&mut state) {
                Ok(value) => value,
                Err(err) => {
                    state.set_error(err);
                    default
                }
            }
        })
    })) {
        Ok(value) => value,
        Err(_) => {
            STATE.with(|state| state.borrow_mut().set_error("panic in RTen WASM runtime"));
            default
        }
    }
}

fn tensor_dtype(value: &Value) -> u32 {
    match value.dtype() {
        ValueType::Tensor(DataType::Float) => DTYPE_FLOAT32,
        ValueType::Tensor(DataType::Int32) => DTYPE_INT32,
        ValueType::Tensor(DataType::Int8) => DTYPE_INT8,
        ValueType::Tensor(DataType::UInt8) => DTYPE_UINT8,
        _ => 0,
    }
}

fn copy_shape(shape: &[usize], out_ptr: *mut u32, max_dims: usize) -> Result<u32, String> {
    if shape.len() > max_dims {
        return Err(format!(
            "shape has {} dims but output buffer only has {}",
            shape.len(),
            max_dims
        ));
    }
    if !shape.is_empty() && out_ptr.is_null() {
        return Err("shape output pointer is null".to_string());
    }
    let out = unsafe { slice::from_raw_parts_mut(out_ptr, max_dims) };
    for (dst, &dim) in out.iter_mut().zip(shape) {
        *dst = dim as u32;
    }
    Ok(shape.len() as u32)
}

#[no_mangle]
pub extern "C" fn allocate(size: usize) -> *mut u8 {
    if size == 0 {
        return std::ptr::null_mut();
    }
    let layout = match Layout::array::<u8>(size) {
        Ok(layout) => layout,
        Err(_) => return std::ptr::null_mut(),
    };
    unsafe { alloc(layout) }
}

#[no_mangle]
pub extern "C" fn deallocate(ptr: *mut u8, size: usize) {
    if ptr.is_null() || size == 0 {
        return;
    }
    if let Ok(layout) = Layout::array::<u8>(size) {
        unsafe { dealloc(ptr, layout) };
    }
}

#[no_mangle]
pub extern "C" fn rten_init() -> u32 {
    STATE.with(|state| *state.borrow_mut() = State::new());
    1
}

#[no_mangle]
pub extern "C" fn rten_load_model(model_ptr: *const u8, model_len: usize) -> u32 {
    run_with_state(0, |state| {
        if model_len > 0 && model_ptr.is_null() {
            return Err("model pointer is null".to_string());
        }
        let model_data = unsafe { slice::from_raw_parts(model_ptr, model_len) }.to_vec();
        let model = Model::load(model_data).map_err(|err| err.to_string())?;
        Ok(state.insert_model(model))
    })
}

#[no_mangle]
pub extern "C" fn rten_free_model(model_id: u32) -> u32 {
    run_with_state(0, |state| {
        if let Some(slot) = state.models.get_mut(model_id as usize) {
            *slot = None;
            Ok(1)
        } else {
            Err("invalid model id".to_string())
        }
    })
}

#[no_mangle]
pub extern "C" fn rten_get_input_count(model_id: u32) -> u32 {
    run_with_state(0, |state| {
        let model = state
            .models
            .get(model_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid model id")?;
        Ok(model.input_ids().len() as u32)
    })
}

#[no_mangle]
pub extern "C" fn rten_get_output_count(model_id: u32) -> u32 {
    run_with_state(0, |state| {
        let model = state
            .models
            .get(model_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid model id")?;
        Ok(model.output_ids().len() as u32)
    })
}

#[no_mangle]
pub extern "C" fn rten_get_input_id(model_id: u32, input_index: u32) -> u32 {
    run_with_state(0, |state| {
        let model = state
            .models
            .get(model_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid model id")?;
        model
            .input_ids()
            .get(input_index as usize)
            .map(|id| id.as_u32())
            .ok_or_else(|| "invalid input index".to_string())
    })
}

#[no_mangle]
pub extern "C" fn rten_get_output_id(model_id: u32, output_index: u32) -> u32 {
    run_with_state(0, |state| {
        let model = state
            .models
            .get(model_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid model id")?;
        model
            .output_ids()
            .get(output_index as usize)
            .map(|id| id.as_u32())
            .ok_or_else(|| "invalid output index".to_string())
    })
}

#[no_mangle]
pub extern "C" fn rten_get_input_dims(
    model_id: u32,
    input_index: u32,
    dims_ptr: *mut u32,
    max_dims: usize,
) -> u32 {
    run_with_state(0, |state| {
        let model = state
            .models
            .get(model_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid model id")?;
        let dims = model
            .input_shape(input_index as usize)
            .ok_or("input shape is unavailable")?;
        let fixed_dims: Vec<usize> = dims
            .into_iter()
            .map(|dim| match dim {
                Dimension::Fixed(size) => size,
                Dimension::Symbolic(_) => usize::MAX,
            })
            .collect();
        copy_shape(&fixed_dims, dims_ptr, max_dims)
    })
}

#[no_mangle]
pub extern "C" fn rten_create_float_tensor(
    shape_ptr: *const u32,
    shape_len: usize,
    data_ptr: *const f32,
    data_len: usize,
) -> u32 {
    run_with_state(0, |state| {
        if data_len > 0 && data_ptr.is_null() {
            return Err("float data pointer is null".to_string());
        }
        let shape = shape_from_raw(shape_ptr, shape_len)?;
        let data = unsafe { slice::from_raw_parts(data_ptr, data_len) }.to_vec();
        let tensor = Value::from_shape(shape, data).map_err(|err| err.to_string())?;
        Ok(state.insert_tensor(tensor))
    })
}

#[no_mangle]
pub extern "C" fn rten_create_int_tensor(
    shape_ptr: *const u32,
    shape_len: usize,
    data_ptr: *const i32,
    data_len: usize,
) -> u32 {
    run_with_state(0, |state| {
        if data_len > 0 && data_ptr.is_null() {
            return Err("int data pointer is null".to_string());
        }
        let shape = shape_from_raw(shape_ptr, shape_len)?;
        let data = unsafe { slice::from_raw_parts(data_ptr, data_len) }.to_vec();
        let tensor = Value::from_shape(shape, data).map_err(|err| err.to_string())?;
        Ok(state.insert_tensor(tensor))
    })
}

#[no_mangle]
pub extern "C" fn rten_free_tensor(tensor_id: u32) -> u32 {
    run_with_state(0, |state| {
        if let Some(slot) = state.tensors.get_mut(tensor_id as usize) {
            *slot = None;
            Ok(1)
        } else {
            Err("invalid tensor id".to_string())
        }
    })
}

#[no_mangle]
pub extern "C" fn rten_run(
    model_id: u32,
    input_ids_ptr: *const u32,
    input_count: usize,
    output_ids_ptr: *mut u32,
    max_outputs: usize,
) -> u32 {
    run_with_state(0, |state| {
        if input_count > 0 && input_ids_ptr.is_null() {
            return Err("input IDs pointer is null".to_string());
        }
        if max_outputs > 0 && output_ids_ptr.is_null() {
            return Err("output IDs pointer is null".to_string());
        }

        let model = state
            .models
            .get(model_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid model id")?;
        let input_tensor_ids = unsafe { slice::from_raw_parts(input_ids_ptr, input_count) };
        if input_tensor_ids.len() != model.input_ids().len() {
            return Err(format!(
                "expected {} inputs, got {}",
                model.input_ids().len(),
                input_tensor_ids.len()
            ));
        }

        let mut inputs = Vec::with_capacity(input_tensor_ids.len());
        for (&node_id, &tensor_id) in model.input_ids().iter().zip(input_tensor_ids) {
            let tensor = state
                .tensors
                .get(tensor_id as usize)
                .and_then(Option::as_ref)
                .ok_or("invalid input tensor id")?;
            inputs.push((node_id, ValueOrView::View(tensor.as_view())));
        }

        let output_node_ids: Vec<NodeId> = model.output_ids().to_vec();
        if output_node_ids.len() > max_outputs {
            return Err(format!(
                "model has {} outputs but output buffer only has {} slots",
                output_node_ids.len(),
                max_outputs
            ));
        }

        let outputs = model
            .run(inputs, &output_node_ids, None)
            .map_err(|err| format!("{err:?}"))?;

        let output_ids = unsafe { slice::from_raw_parts_mut(output_ids_ptr, max_outputs) };
        for (index, output) in outputs.into_iter().enumerate() {
            output_ids[index] = state.insert_tensor(output);
        }
        Ok(output_node_ids.len() as u32)
    })
}

#[no_mangle]
pub extern "C" fn rten_get_tensor_ndim(tensor_id: u32) -> u32 {
    run_with_state(0, |state| {
        let tensor = state
            .tensors
            .get(tensor_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid tensor id")?;
        Ok(tensor.ndim() as u32)
    })
}

#[no_mangle]
pub extern "C" fn rten_get_tensor_shape(
    tensor_id: u32,
    shape_ptr: *mut u32,
    max_dims: usize,
) -> u32 {
    run_with_state(0, |state| {
        let tensor = state
            .tensors
            .get(tensor_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid tensor id")?;
        copy_shape(&tensor.shape(), shape_ptr, max_dims)
    })
}

#[no_mangle]
pub extern "C" fn rten_get_tensor_dtype(tensor_id: u32) -> u32 {
    run_with_state(0, |state| {
        let tensor = state
            .tensors
            .get(tensor_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid tensor id")?;
        Ok(tensor_dtype(tensor))
    })
}

#[no_mangle]
pub extern "C" fn rten_get_tensor_len(tensor_id: u32) -> u32 {
    run_with_state(0, |state| {
        let tensor = state
            .tensors
            .get(tensor_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid tensor id")?;
        Ok(tensor.len() as u32)
    })
}

#[no_mangle]
pub extern "C" fn rten_get_float_data(tensor_id: u32, data_ptr: *mut f32, data_len: usize) -> u32 {
    run_with_state(0, |state| {
        if data_len > 0 && data_ptr.is_null() {
            return Err("float output pointer is null".to_string());
        }
        let tensor = state
            .tensors
            .get(tensor_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid tensor id")?;
        let view = tensor
            .as_tensor_view::<f32>()
            .ok_or("tensor is not float32")?;
        let data = view.to_vec();
        if data.len() > data_len {
            return Err(format!(
                "output buffer has room for {} floats but tensor has {}",
                data_len,
                data.len()
            ));
        }
        let out = unsafe { slice::from_raw_parts_mut(data_ptr, data_len) };
        out[..data.len()].copy_from_slice(&data);
        Ok(data.len() as u32)
    })
}

#[no_mangle]
pub extern "C" fn rten_get_int_data(tensor_id: u32, data_ptr: *mut i32, data_len: usize) -> u32 {
    run_with_state(0, |state| {
        if data_len > 0 && data_ptr.is_null() {
            return Err("int output pointer is null".to_string());
        }
        let tensor = state
            .tensors
            .get(tensor_id as usize)
            .and_then(Option::as_ref)
            .ok_or("invalid tensor id")?;
        let view = tensor
            .as_tensor_view::<i32>()
            .ok_or("tensor is not int32")?;
        let data = view.to_vec();
        if data.len() > data_len {
            return Err(format!(
                "output buffer has room for {} ints but tensor has {}",
                data_len,
                data.len()
            ));
        }
        let out = unsafe { slice::from_raw_parts_mut(data_ptr, data_len) };
        out[..data.len()].copy_from_slice(&data);
        Ok(data.len() as u32)
    })
}

#[no_mangle]
pub extern "C" fn rten_get_error_len() -> u32 {
    STATE.with(|state| state.borrow().last_error.len() as u32)
}

#[no_mangle]
pub extern "C" fn rten_get_error(out_ptr: *mut u8, out_len: usize) -> u32 {
    STATE.with(|state| {
        let state = state.borrow();
        let err = state.last_error.as_bytes();
        if err.is_empty() {
            return 0;
        }
        if out_ptr.is_null() || out_len < err.len() {
            return 0;
        }
        let out = unsafe { slice::from_raw_parts_mut(out_ptr, out_len) };
        out[..err.len()].copy_from_slice(err);
        1
    })
}
