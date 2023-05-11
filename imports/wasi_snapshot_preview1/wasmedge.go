package wasi_snapshot_preview1

import (
	"context"
	"encoding/binary"
	"io"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wazergo"
	. "github.com/stealthrocket/wazergo/types"
	"github.com/stealthrocket/wazergo/wasm"
	"github.com/tetratelabs/wazero/api"
)

// WasmEdge is the WasmEdge sockets extension to WASI preview 1.
var WasmEdge = Extension{
	"sock_open":         wazergo.F3((*Module).WasmEdgeSockOpen),
	"sock_bind":         wazergo.F3((*Module).WasmEdgeSockBind),
	"sock_connect":      wazergo.F3((*Module).WasmEdgeSockConnect),
	"sock_listen":       wazergo.F2((*Module).WasmEdgeSockListen),
	"sock_getsockopt":   wazergo.F5((*Module).WasmEdgeSockGetOpt),
	"sock_setsockopt":   wazergo.F5((*Module).WasmEdgeSockSetOpt),
	"sock_getlocaladdr": wazergo.F3((*Module).WasmEdgeSockLocalAddr),
	"sock_getpeeraddr":  wazergo.F3((*Module).WasmEdgeSockPeerAddr),
}

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
	socketAddr, ok := m.wasmEdgeGetSocketAddress(addr.Load(), int(port))
	if !ok {
		return Errno(wasi.EINVAL)
	}
	return Errno(s.SockBind(ctx, wasi.FD(fd), socketAddr))
}

func (m *Module) WasmEdgeSockConnect(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Uint32) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	socketAddr, ok := m.wasmEdgeGetSocketAddress(addr.Load(), int(port))
	if !ok {
		return Errno(wasi.EINVAL)
	}
	return Errno(s.SockConnect(ctx, wasi.FD(fd), socketAddr))
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

func (m *Module) WasmEdgeSockLocalAddr(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Pointer[Uint32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	sa, errno := s.SockLocalAddress(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	portint, ok := m.wasmEdgePutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	port.Store(Uint32(portint))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeSockPeerAddr(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Pointer[Uint32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	sa, errno := s.SockPeerAddress(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	portint, ok := m.wasmEdgePutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	port.Store(Uint32(portint))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) wasmEdgeGetSocketAddress(b wasmEdgeAddress, port int) (sa wasi.SocketAddress, ok bool) {
	if len(b) == 128 {
		switch wasi.ProtocolFamily(binary.LittleEndian.Uint16(b)) {
		case wasi.Inet:
			b = b[2:6]
		case wasi.Inet6:
			b = b[2:18]
		default:
			return // not implemented
		}
	}
	switch len(b) {
	case 4:
		m.inet4addr.Port = port
		copy(m.inet4addr.Addr[:], b)
		sa = &m.inet4addr
		ok = true
	case 16:
		m.inet6addr.Port = port
		copy(m.inet6addr.Addr[:], b)
		sa = &m.inet6addr
		ok = true
	}
	return
}

func (m *Module) wasmEdgePutSocketAddress(b wasmEdgeAddress, sa wasi.SocketAddress) (port int, ok bool) {
	// WasmEdge sockets v2 only supports 128 byte buffers.
	if len(b) != 128 {
		return
	}
	switch t := sa.(type) {
	case *wasi.Inet4Address:
		binary.LittleEndian.PutUint16(b, uint16(wasi.Inet))
		copy(b[2:], t.Addr[:])
		port = t.Port
		ok = true
	case *wasi.Inet6Address:
		binary.LittleEndian.PutUint16(b, uint16(wasi.Inet6))
		copy(b[2:], t.Addr[:])
		port = t.Port
		ok = true
	}
	return
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
