package rtengowasm

import (
	"github.com/masacento/rtengo/internal/genwasm"
)

// MemoryRead reads bytes from module memory.
// The returned slice is a copy; the module's memory may be relocated by a
// later memory.grow.
func (r *Runtime) MemoryRead(ptr, size uint32) ([]byte, bool) {
	mem := genwasm.Memory(r.mod)
	end := uint64(ptr) + uint64(size)
	if end > uint64(len(mem)) {
		return nil, false
	}

	out := make([]byte, size)
	copy(out, mem[ptr:end])
	return out, true
}

// MemoryWrite writes bytes to module memory.
func (r *Runtime) MemoryWrite(ptr uint32, data []byte) bool {
	mem := genwasm.Memory(r.mod)
	end := uint64(ptr) + uint64(len(data))
	if end > uint64(len(mem)) {
		return false
	}

	copy(mem[ptr:end], data)
	return true
}
