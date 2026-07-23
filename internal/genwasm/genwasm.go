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
	if __memcap < 1245184 {
		__memcap = 1245184
	}
	m.Memory = make([]byte, 1245184, __memcap)
	m.MemMu = &sync.Mutex{}
	m.MemSize = &atomic.Uint64{}
	m.Threads = &base.ThreadPool{}
	m.MemSize.Store(1245184)
	m.M = unsafe.Pointer(unsafe.SliceData(m.Memory))
	m.MaxMem = 4294967296
	m.T0 = make([]any, 3027)
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
	InitElemSeg_2_10(m)
	m.DataEnd = 1184188
	initData_0(m)
	return m
}
// NewWithWASI constructs a *Module with a custom
// wasi_snapshot_preview1 implementation and a default initial
// linear-memory reservation. Use NewWithWASIReserve to pre-size
// the reservation (e.g. to cover an interpreter's whole boot and
// avoid reallocating/copying linear memory on the first grow).
func NewWithWASI(wasi_snapshot_preview1 base.Wasi_snapshot_preview1Imports) *base.Module {
	return NewWithWASIReserve(wasi_snapshot_preview1, 1556480)
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
	m.T0 = make([]any, 3027)
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
	InitElemSeg_2_10(m)
	m.DataEnd = 1184188
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
	m.T0 = make([]any, 3027)
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
	InitElemSeg_2_10(m)
	m.DataEnd = 1184188
	base.RestoreGlobals(m, globals)
	return m
}
func initData_0(m *base.Module) {
	copy(m.Memory[1048576:], wasm2goData_data_bin[0:135612])
}
func Allocate(m *base.Module, l0 int32) int32 {
	return Fn5837(m, l0)
}
func Deallocate(m *base.Module, l0 int32, l1 int32) {
	Fn5838(m, l0, l1)
}
func RtenCreateFloatTensor(m *base.Module, l0 int32, l1 int32, l2 int32, l3 int32) int32 {
	return Fn5839(m, l0, l1, l2, l3)
}
func RtenCreateIntTensor(m *base.Module, l0 int32, l1 int32, l2 int32, l3 int32) int32 {
	return Fn5840(m, l0, l1, l2, l3)
}
func RtenFreeModel(m *base.Module, l0 int32) int32 {
	return Fn5841(m, l0)
}
func RtenFreeTensor(m *base.Module, l0 int32) int32 {
	return Fn5842(m, l0)
}
func RtenGetError(m *base.Module, l0 int32, l1 int32) int32 {
	return Fn5843(m, l0, l1)
}
func RtenGetErrorLen(m *base.Module) int32 {
	return Fn5844(m)
}
func RtenGetFloatData(m *base.Module, l0 int32, l1 int32, l2 int32) int32 {
	return Fn5845(m, l0, l1, l2)
}
func RtenGetInputCount(m *base.Module, l0 int32) int32 {
	return Fn5846(m, l0)
}
func RtenGetInputDims(m *base.Module, l0 int32, l1 int32, l2 int32, l3 int32) int32 {
	return Fn5847(m, l0, l1, l2, l3)
}
func RtenGetInputId(m *base.Module, l0 int32, l1 int32) int32 {
	return Fn5848(m, l0, l1)
}
func RtenGetIntData(m *base.Module, l0 int32, l1 int32, l2 int32) int32 {
	return Fn5849(m, l0, l1, l2)
}
func RtenGetOutputCount(m *base.Module, l0 int32) int32 {
	return Fn5850(m, l0)
}
func RtenGetOutputId(m *base.Module, l0 int32, l1 int32) int32 {
	return Fn5851(m, l0, l1)
}
func RtenGetTensorDtype(m *base.Module, l0 int32) int32 {
	return Fn5852(m, l0)
}
func RtenGetTensorLen(m *base.Module, l0 int32) int32 {
	return Fn5853(m, l0)
}
func RtenGetTensorNdim(m *base.Module, l0 int32) int32 {
	return Fn5854(m, l0)
}
func RtenGetTensorShape(m *base.Module, l0 int32, l1 int32, l2 int32) int32 {
	return Fn5855(m, l0, l1, l2)
}
func RtenInit(m *base.Module) int32 {
	return Fn5856(m)
}
func RtenLoadModel(m *base.Module, l0 int32, l1 int32) int32 {
	return Fn5857(m, l0, l1)
}
func RtenRun(m *base.Module, l0 int32, l1 int32, l2 int32, l3 int32, l4 int32) int32 {
	return Fn5858(m, l0, l1, l2, l3, l4)
}
func Memory(m *base.Module) []byte {
	return m.Memory
}
//go:embed data.bin
var wasm2goData_data_bin []byte
