package wasi

import (
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/stealthrocket/wazergo/types"
	"github.com/stealthrocket/wazergo/wasm"
	"github.com/tetratelabs/wazero/api"
)

func (FDStat) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (FDStat) LoadObject(memory api.Memory, object []byte) (f FDStat) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&f)), f.ObjectSize()), object)
	return
}

func (f FDStat) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&f)), f.ObjectSize()))
}

func (FDStat) ObjectSize() int {
	return int(unsafe.Sizeof(FDStat{}))
}

func (FileStat) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (FileStat) LoadObject(memory api.Memory, object []byte) (f FileStat) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&f)), f.ObjectSize()), object)
	return
}

func (f FileStat) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&f)), f.ObjectSize()))
}

func (FileStat) ObjectSize() int {
	return int(unsafe.Sizeof(FileStat{}))
}

func (PreStat) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (PreStat) LoadObject(memory api.Memory, object []byte) (p PreStat) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p)), p.ObjectSize()), object)
	return
}

func (p PreStat) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&p)), p.ObjectSize()))
}

func (PreStat) ObjectSize() int {
	return int(unsafe.Sizeof(PreStat{}))
}

func (arg IOVec) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	types.Bytes(arg.LoadObject(memory, object)).Format(w)
}

func (arg IOVec) LoadObject(memory api.Memory, object []byte) IOVec {
	offset := binary.LittleEndian.Uint32(object[:4])
	length := binary.LittleEndian.Uint32(object[4:])
	return wasm.Read(memory, offset, length)
}

func (arg IOVec) StoreObject(memory api.Memory, object []byte) {
	panic("NOT IMPLEMENTED")
}

func (arg IOVec) ObjectSize() int {
	return 8
}

func (Subscription) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (Subscription) LoadObject(memory api.Memory, object []byte) (s Subscription) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&s)), s.ObjectSize()), object)
	return
}

func (s Subscription) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&s)), s.ObjectSize()))
}

func (Subscription) ObjectSize() int {
	return int(unsafe.Sizeof(Subscription{}))
}

func (Event) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (Event) LoadObject(memory api.Memory, object []byte) (e Event) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&e)), e.ObjectSize()), object)
	return
}

func (e Event) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&e)), e.ObjectSize()))
}

func (Event) ObjectSize() int {
	return int(unsafe.Sizeof(Event{}))
}
