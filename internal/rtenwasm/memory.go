package rtengowasm

// MemoryRead reads bytes from wasm memory.
func (r *Runtime) MemoryRead(ptr, size uint32) ([]byte, bool) {
	return r.memory.Read(ptr, size)
}

// MemoryWrite writes bytes to wasm memory.
func (r *Runtime) MemoryWrite(ptr uint32, data []byte) bool {
	return r.memory.Write(ptr, data)
}
