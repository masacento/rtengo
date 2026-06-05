# rtenwasm

RTen WebAssembly wrapper used by the Go runtime in this repository.

This crate exports the C-style ABI expected by `internal/rtenwasm`, including
`rten_load_model`, `rten_run`, tensor constructors and tensor data accessors.

Build and install the embedded WASM:

```sh
rustup target add wasm32-wasip1
make -C rtenwasm install
```

This wrapper only provides the ABI used by the Go runtime. It does not add
support for AVX2-specific RTen quantized models in WebAssembly.
