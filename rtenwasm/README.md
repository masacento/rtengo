# rtenwasm

RTen WebAssembly wrapper used by the Go runtime in this repository.

This crate exports the C-style ABI expected by `internal/rtenwasm`, including
`rten_load_model`, `rten_run`, tensor constructors and tensor data accessors.

The WASM module is not executed directly. It is AOT-compiled to Go code with
[wasm2go](https://github.com/goccy/wasm2go) and the generated code is committed
under `internal/genwasm`.

Regenerate the Go module:

```sh
rustup target add wasm32-wasip1
go install github.com/goccy/wasm2go/cmd/wasm2go@latest
make -C rtenwasm gen
```

Note: SIMD (`simd128`) must stay disabled because wasm2go does not support
v128 instructions. Do not add `RUSTFLAGS="-C target-feature=+simd128"` to the
build.

This wrapper only provides the ABI used by the Go runtime. It does not add
support for AVX2-specific RTen quantized models in WebAssembly.
