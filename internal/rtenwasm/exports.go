package rtengowasm

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"sync"
)

//go:embed rten.wasm.gz
var embeddedWASM []byte

var (
	embeddedWASMOnce         sync.Once
	embeddedWASMDecompressed []byte
	embeddedWASMErr          error
)

// EmbeddedWASM returns the embedded RTen WASM module bytes.
func EmbeddedWASM() []byte {
	embeddedWASMOnce.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(embeddedWASM))
		if err != nil {
			embeddedWASMErr = err
			return
		}
		defer reader.Close()

		embeddedWASMDecompressed, embeddedWASMErr = io.ReadAll(reader)
	})

	return embeddedWASMDecompressed
}
