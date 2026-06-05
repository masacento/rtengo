WASM_TARGET ?= wasm32-wasip1
OUT ?= ../internal/rtenwasm/rten.wasm.gz

.PHONY: check build install clean

check:
	cargo check

build:
	RUSTFLAGS="-C target-feature=+simd128" cargo build --release --target $(WASM_TARGET)

install: build
	gzip -9 -c target/$(WASM_TARGET)/release/rtenwasm.wasm > $(OUT)

clean:
	cargo clean
