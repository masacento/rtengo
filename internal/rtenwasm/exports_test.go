package rtengowasm

import "testing"

func TestEmbeddedWASMNonEmpty(t *testing.T) {
	wasm1 := EmbeddedWASM()
	if len(wasm1) == 0 {
		t.Fatal("expected embedded WASM to be non-empty")
	}

	wasm2 := EmbeddedWASM()
	if len(wasm2) != len(wasm1) {
		t.Fatalf("expected embedded WASM size to be stable: %d vs %d", len(wasm1), len(wasm2))
	}

	for i := 0; i < len(wasm1); i++ {
		if wasm1[i] != wasm2[i] {
			t.Fatalf("embedded WASM differs at byte %d", i)
		}
	}
}
