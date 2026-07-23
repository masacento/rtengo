WASM_TARGET ?= wasm32-wasip1
GEN_OUT ?= ../internal/genwasm
GEN_IMPORT ?= github.com/masacento/rtengo/internal/genwasm

.PHONY: check build gen clean

check:
	cargo check

# Build the WASM module.
# Note: SIMD (simd128) must stay disabled because wasm2go does not support
# v128 instructions, so no RUSTFLAGS=+simd128 here.
build:
	cargo build --release --target $(WASM_TARGET)

# Regenerate the Go module under internal/genwasm from the built WASM.
gen: build
	wasm2go -i target/$(WASM_TARGET)/release/rtenwasm.wasm -out-dir $(GEN_OUT) -pkg genwasm -import $(GEN_IMPORT)

clean:
	cargo clean
