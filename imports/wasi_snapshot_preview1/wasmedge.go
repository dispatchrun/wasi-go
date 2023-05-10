package wasi_snapshot_preview1

import (
	"context"
	"encoding/binary"
	"io"

	"github.com/stealthrocket/wasi-go"
	. "github.com/stealthrocket/wazergo/types"
	"github.com/stealthrocket/wazergo/wasm"
	"github.com/tetratelabs/wazero/api"
)

func (m *Module) WasmEdgeSockOpen(ctx context.Context, family Int32, sockType Int32, openfd Pointer[Int32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	rightsBase := wasi.SockListenRights | wasi.SockConnectionRights
	rightsInheriting := wasi.SockConnectionRights
	result, errno := s.SockOpen(ctx, wasi.ProtocolFamily(family), wasi.SocketType(sockType), wasi.IPProtocol, rightsBase, rightsInheriting)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	openfd.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeSockBind(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Uint32) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	b := addr.Load()
	var sa wasi.SocketAddress
	switch len(b) {
	case 4:
		sa = &wasi.Inet4Address{Port: int(port), Addr: [4]byte(b)}
	case 16:
		sa = &wasi.Inet6Address{Port: int(port), Addr: [16]byte(b)}
	default:
		return Errno(wasi.EINVAL)
	}
	return Errno(s.SockBind(ctx, wasi.FD(fd), sa))
}

func (m *Module) WasmEdgeSockConnect(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Uint32) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	b := addr.Load()
	var sa wasi.SocketAddress
	switch len(b) {
	case 4:
		sa = &wasi.Inet4Address{Port: int(port), Addr: [4]byte(b)}
	case 16:
		sa = &wasi.Inet6Address{Port: int(port), Addr: [16]byte(b)}
	default:
		return Errno(wasi.EINVAL)
	}
	return Errno(s.SockConnect(ctx, wasi.FD(fd), sa))
}

func (m *Module) WasmEdgeSockListen(ctx context.Context, fd Int32, backlog Int32) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	return Errno(s.SockListen(ctx, wasi.FD(fd), int(backlog)))
}

func (m *Module) WasmEdgeSockSetOpt(ctx context.Context, fd Int32, level Int32, option Int32, value Pointer[Int32], valueLen Int32) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	if valueLen != 4 {
		// Only int options are supported for now.
		return Errno(wasi.EINVAL)
	}
	return Errno(s.SockSetOptInt(ctx, wasi.FD(fd), wasi.SocketOptionLevel(level), wasi.SocketOption(option), int(value.Load())))
}

func (m *Module) WasmEdgeSockGetOpt(ctx context.Context, fd Int32, level Int32, option Int32, value Pointer[Int32], valueLen Int32) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	if valueLen != 4 {
		// Only int options are supported for now.
		return Errno(wasi.EINVAL)
	}
	result, errno := s.SockGetOptInt(ctx, wasi.FD(fd), wasi.SocketOptionLevel(level), wasi.SocketOption(option))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	value.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

type wasmEdgeAddress []byte

func (arg wasmEdgeAddress) ObjectSize() int {
	return 8
}

func (arg wasmEdgeAddress) LoadObject(memory api.Memory, object []byte) wasmEdgeAddress {
	offset := binary.LittleEndian.Uint32(object[:4])
	length := binary.LittleEndian.Uint32(object[4:])
	return wasm.Read(memory, offset, length)
}

func (arg wasmEdgeAddress) StoreObject(memory api.Memory, object []byte) {
	panic("BUG: socket addresses cannot be stored back to wasm memory")
}

func (arg wasmEdgeAddress) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	Bytes(arg.LoadObject(memory, object)).Format(w)
}
