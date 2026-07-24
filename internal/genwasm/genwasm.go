package genwasm

import (
	base "github.com/masacento/rtengo/internal/genwasm/base"
	"sync"
	"sync/atomic"
	"unsafe"
	_ "github.com/masacento/rtengo/internal/genwasm/p2"
	_ "embed"
)

func NewWithWASIReserve(wasi_snapshot_preview1 base.Wasi_snapshot_preview1Imports, reserveBytes int) *base.Module {
	m := &base.Module{Wasi_snapshot_preview1: wasi_snapshot_preview1}
	__memcap := reserveBytes
	if __memcap < 1179648 {
		__memcap = 1179648
	}
	m.Memory = make([]byte, 1179648, __memcap)
	m.MemMu = &sync.Mutex{}
	m.MemSize = &atomic.Uint64{}
	m.Threads = &base.ThreadPool{}
	m.MemSize.Store(1179648)
	m.M = unsafe.Pointer(unsafe.SliceData(m.Memory))
	m.MaxMem = 4294967296
	m.T0 = make([]any, 2598)
	m.G0 = int32(1048576)
	InitElemSeg_0_0(m)
	InitElemSeg_1_0(m)
	InitElemSeg_2_0(m)
	InitElemSeg_2_1(m)
	InitElemSeg_2_2(m)
	InitElemSeg_2_3(m)
	InitElemSeg_2_4(m)
	InitElemSeg_2_5(m)
	InitElemSeg_2_6(m)
	InitElemSeg_2_7(m)
	InitElemSeg_2_8(m)
	InitElemSeg_2_9(m)
	m.DataEnd = 1169736
	initData_0(m)
	return m
}
// NewWithWASI constructs a *Module with a custom
// wasi_snapshot_preview1 implementation and a default initial
// linear-memory reservation. Use NewWithWASIReserve to pre-size
// the reservation (e.g. to cover an interpreter's whole boot and
// avoid reallocating/copying linear memory on the first grow).
func NewWithWASI(wasi_snapshot_preview1 base.Wasi_snapshot_preview1Imports) *base.Module {
	return NewWithWASIReserve(wasi_snapshot_preview1, 1474560)
}
// New constructs a *Module using DefaultWASI() for the
// wasi_snapshot_preview1 import. Use NewWithWASI to plug in a
// custom implementation (sandboxed FS, captured stdout, ...).
func New() *base.Module {
	return NewWithWASI(base.DefaultWASI())
}
func NewWithMemory(wasi_snapshot_preview1 base.Wasi_snapshot_preview1Imports, memory []byte, memSize uint64) *base.Module {
	m := &base.Module{Wasi_snapshot_preview1: wasi_snapshot_preview1}
	m.Memory = memory
	m.MemMu = &sync.Mutex{}
	m.MemSize = &atomic.Uint64{}
	m.Threads = &base.ThreadPool{}
	m.MemSize.Store(memSize)
	m.M = unsafe.Pointer(unsafe.SliceData(m.Memory))
	m.MaxMem = uint64(len(memory))
	m.T0 = make([]any, 2598)
	m.G0 = int32(1048576)
	InitElemSeg_0_0(m)
	InitElemSeg_1_0(m)
	InitElemSeg_2_0(m)
	InitElemSeg_2_1(m)
	InitElemSeg_2_2(m)
	InitElemSeg_2_3(m)
	InitElemSeg_2_4(m)
	InitElemSeg_2_5(m)
	InitElemSeg_2_6(m)
	InitElemSeg_2_7(m)
	InitElemSeg_2_8(m)
	InitElemSeg_2_9(m)
	m.DataEnd = 1169736
	return m
}
func NewFromSnapshot(wasi_snapshot_preview1 base.Wasi_snapshot_preview1Imports, memory []byte, memSize uint64, globals []uint64) *base.Module {
	m := &base.Module{Wasi_snapshot_preview1: wasi_snapshot_preview1}
	m.Memory = memory
	m.MemMu = &sync.Mutex{}
	m.MemSize = &atomic.Uint64{}
	m.Threads = &base.ThreadPool{}
	m.MemSize.Store(memSize)
	m.M = unsafe.Pointer(unsafe.SliceData(m.Memory))
	m.MaxMem = uint64(len(memory))
	m.T0 = make([]any, 2598)
	m.G0 = int32(1048576)
	InitElemSeg_0_0(m)
	InitElemSeg_1_0(m)
	InitElemSeg_2_0(m)
	InitElemSeg_2_1(m)
	InitElemSeg_2_2(m)
	InitElemSeg_2_3(m)
	InitElemSeg_2_4(m)
	InitElemSeg_2_5(m)
	InitElemSeg_2_6(m)
	InitElemSeg_2_7(m)
	InitElemSeg_2_8(m)
	InitElemSeg_2_9(m)
	m.DataEnd = 1169736
	base.RestoreGlobals(m, globals)
	return m
}
func initData_0(m *base.Module) {
	copy(m.Memory[1048576:], wasm2goData_data_bin[0:121160])
}
func Allocate(m *base.Module, l0 int32) int32 {
	return Fn5406(m, l0)
}
func Deallocate(m *base.Module, l0 int32, l1 int32) {
	Fn5407(m, l0, l1)
}
func RtenCreateFloatTensor(m *base.Module, l0 int32, l1 int32, l2 int32, l3 int32) int32 {
	return Fn5408(m, l0, l1, l2, l3)
}
func RtenCreateIntTensor(m *base.Module, l0 int32, l1 int32, l2 int32, l3 int32) int32 {
	return Fn5409(m, l0, l1, l2, l3)
}
func RtenFreeModel(m *base.Module, l0 int32) int32 {
	return Fn5410(m, l0)
}
func RtenFreeTensor(m *base.Module, l0 int32) int32 {
	return Fn5411(m, l0)
}
func RtenGetError(m *base.Module, l0 int32, l1 int32) int32 {
	return Fn5412(m, l0, l1)
}
func RtenGetErrorLen(m *base.Module) int32 {
	return Fn5413(m)
}
func RtenGetFloatData(m *base.Module, l0 int32, l1 int32, l2 int32) int32 {
	return Fn5414(m, l0, l1, l2)
}
func RtenGetInputCount(m *base.Module, l0 int32) int32 {
	return Fn5415(m, l0)
}
func RtenGetInputDims(m *base.Module, l0 int32, l1 int32, l2 int32, l3 int32) int32 {
	return Fn5416(m, l0, l1, l2, l3)
}
func RtenGetInputId(m *base.Module, l0 int32, l1 int32) int32 {
	return Fn5417(m, l0, l1)
}
func RtenGetIntData(m *base.Module, l0 int32, l1 int32, l2 int32) int32 {
	return Fn5418(m, l0, l1, l2)
}
func RtenGetOutputCount(m *base.Module, l0 int32) int32 {
	return Fn5419(m, l0)
}
func RtenGetOutputId(m *base.Module, l0 int32, l1 int32) int32 {
	return Fn5420(m, l0, l1)
}
func RtenGetTensorDtype(m *base.Module, l0 int32) int32 {
	return Fn5421(m, l0)
}
func RtenGetTensorLen(m *base.Module, l0 int32) int32 {
	return Fn5422(m, l0)
}
func RtenGetTensorNdim(m *base.Module, l0 int32) int32 {
	return Fn5423(m, l0)
}
func RtenGetTensorShape(m *base.Module, l0 int32, l1 int32, l2 int32) int32 {
	return Fn5424(m, l0, l1, l2)
}
func RtenInit(m *base.Module) int32 {
	return Fn5425(m)
}
func RtenLoadModel(m *base.Module, l0 int32, l1 int32) int32 {
	return Fn5426(m, l0, l1)
}
func RtenRun(m *base.Module, l0 int32, l1 int32, l2 int32, l3 int32, l4 int32) int32 {
	return Fn5427(m, l0, l1, l2, l3, l4)
}
func Memory(m *base.Module) []byte {
	return m.Memory
}
//go:embed data.bin
var wasm2goData_data_bin []byte
