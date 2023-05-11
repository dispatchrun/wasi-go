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

// WasmEdgeV1 is the original WasmEdge sockets extension to WASI preview 1.
var WasmEdgeV1 = Extension{
	"sock_accept":       wazergo.F2((*Module).WasmEdgeV1SockAccept),
	"sock_open":         wazergo.F3((*Module).WasmEdgeSockOpen),
	"sock_bind":         wazergo.F3((*Module).WasmEdgeSockBind),
	"sock_connect":      wazergo.F3((*Module).WasmEdgeSockConnect),
	"sock_listen":       wazergo.F2((*Module).WasmEdgeSockListen),
	"sock_send_to":      wazergo.F6((*Module).WasmEdgeSockSendTo),
	"sock_recv_from":    wazergo.F6((*Module).WasmEdgeV1SockRecvFrom),
	"sock_getsockopt":   wazergo.F5((*Module).WasmEdgeSockGetOpt),
	"sock_setsockopt":   wazergo.F5((*Module).WasmEdgeSockSetOpt),
	"sock_getlocaladdr": wazergo.F4((*Module).WasmEdgeV1SockLocalAddr),
	"sock_getpeeraddr":  wazergo.F4((*Module).WasmEdgeV1SockPeerAddr),
	"sock_getaddrinfo":  wazergo.F8((*Module).WasmEdgeSockAddrInfo),
}

// WasmEdgeV2 is V2 of the WasmEdge sockets extension to WASI preview 1.
//
// Version 2 has a sock_accept function that's compatible with the WASI
// preview 1 specification, and adds support for AF_UNIX addresses.
//
// TODO: support AF_UNIX addresses
// TODO: support SO_LINGER, SO_RCVTIMEO, SO_SNDTIMEO, SO_BINDTODEVICE socket options
var WasmEdgeV2 = Extension{
	"sock_open":         wazergo.F3((*Module).WasmEdgeSockOpen),
	"sock_bind":         wazergo.F3((*Module).WasmEdgeSockBind),
	"sock_connect":      wazergo.F3((*Module).WasmEdgeSockConnect),
	"sock_listen":       wazergo.F2((*Module).WasmEdgeSockListen),
	"sock_send_to":      wazergo.F6((*Module).WasmEdgeSockSendTo),
	"sock_recv_from":    wazergo.F7((*Module).WasmEdgeV2SockRecvFrom),
	"sock_getsockopt":   wazergo.F5((*Module).WasmEdgeSockGetOpt),
	"sock_setsockopt":   wazergo.F5((*Module).WasmEdgeSockSetOpt),
	"sock_getlocaladdr": wazergo.F3((*Module).WasmEdgeV2SockLocalAddr),
	"sock_getpeeraddr":  wazergo.F3((*Module).WasmEdgeV2SockPeerAddr),
	"sock_getaddrinfo":  wazergo.F8((*Module).WasmEdgeSockAddrInfo),
}

func (m *Module) WasmEdgeV1SockAccept(ctx context.Context, fd Int32, connfd Pointer[Int32]) Errno {
	// V1 sock_accept was not compatible with WASI preview 1, as the
	// fdflags param was missing. This was corrected in V2.
	return m.SockAccept(ctx, fd, 0, connfd)
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

func (m *Module) WasmEdgeSockSendTo(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], addr Pointer[wasmEdgeAddress], port Int32, flags Uint32, nwritten Pointer[Int32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	socketAddr, ok := m.wasmEdgeGetSocketAddress(addr.Load(), int(port))
	if !ok {
		return Errno(wasi.EINVAL)
	}
	m.iovecs = iovecs.Append(m.iovecs[:0])
	size, errno := s.SockSendTo(ctx, wasi.FD(fd), m.iovecs, socketAddr, wasi.SIFlags(flags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(size))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV1SockRecvFrom(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], addr Pointer[Uint8], iflags Uint32, nread Pointer[Uint32], oflags Pointer[Uint32]) Errno {
	// TODO: implement sock_recv_from (v1)
	return Errno(wasi.ENOSYS)
}

func (m *Module) WasmEdgeV2SockRecvFrom(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], addr Pointer[Uint8], iflags Uint32, port Pointer[Uint32], nread Pointer[Uint32], oflags Pointer[Uint32]) Errno {
	// TODO: implement sock_recv_from (v2)
	return Errno(wasi.ENOSYS)
}

func (m *Module) WasmEdgeSockSetOpt(ctx context.Context, fd Int32, level Int32, option Int32, value Pointer[Int32], valueLen Int32) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	// Only int options are supported for now.
	switch wasi.SocketOption(option) {
	case wasi.Linger, wasi.RecvTimeout, wasi.SendTimeout:
		// These accept struct linger / struct timeval.
		return Errno(wasi.ENOTSUP)
	}
	if valueLen != 4 {
		return Errno(wasi.EINVAL)
	}
	return Errno(s.SockSetOptInt(ctx, wasi.FD(fd), wasi.SocketOptionLevel(level), wasi.SocketOption(option), int(value.Load())))
}

func (m *Module) WasmEdgeSockGetOpt(ctx context.Context, fd Int32, level Int32, option Int32, value Pointer[Int32], valueLen Int32) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	// Only int options are supported for now.
	switch wasi.SocketOption(option) {
	case wasi.Linger, wasi.RecvTimeout, wasi.SendTimeout:
		// These accept struct linger / struct timeval.
		return Errno(wasi.ENOTSUP)
	}
	if valueLen != 4 {
		return Errno(wasi.EINVAL)
	}
	result, errno := s.SockGetOptInt(ctx, wasi.FD(fd), wasi.SocketOptionLevel(level), wasi.SocketOption(option))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	value.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV1SockLocalAddr(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], addrType Pointer[Uint32], port Pointer[Uint32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	sa, errno := s.SockLocalAddress(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	portint, pf, ok := m.wasmEdgeV1PutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	addrType.Store(Uint32(pf))
	port.Store(Uint32(portint))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV2SockLocalAddr(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Pointer[Uint32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	sa, errno := s.SockLocalAddress(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	portint, ok := m.wasmEdgeV2PutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	port.Store(Uint32(portint))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV1SockPeerAddr(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], addrType Pointer[Uint32], port Pointer[Uint32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	sa, errno := s.SockPeerAddress(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	portint, pf, ok := m.wasmEdgeV1PutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	addrType.Store(Uint32(pf))
	port.Store(Uint32(portint))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV2SockPeerAddr(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Pointer[Uint32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	sa, errno := s.SockPeerAddress(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	portint, ok := m.wasmEdgeV2PutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	port.Store(Uint32(portint))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeSockAddrInfo(ctx context.Context, nodePtr Pointer[Uint8], nodeLen Uint32, servicePtr Pointer[Uint8], serviceLen Uint32, hintsPtr Pointer[Uint8], resPtr Pointer[Uint8], maxResLength Uint32, resLengthPtr Pointer[Uint8]) Errno {
	// TODO: implement sock_getaddrinfo
	return Errno(wasi.ENOSYS)
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

func (m *Module) wasmEdgeV1PutSocketAddress(b wasmEdgeAddress, sa wasi.SocketAddress) (port int, pf wasi.ProtocolFamily, ok bool) {
	if len(b) != 16 {
		return
	}
	switch t := sa.(type) {
	case *wasi.Inet4Address:
		binary.LittleEndian.PutUint16(b, uint16(wasi.Inet))
		copy(b, t.Addr[:])
		pf = wasi.Inet
		port = t.Port
		ok = true
	case *wasi.Inet6Address:
		binary.LittleEndian.PutUint16(b, uint16(wasi.Inet6))
		copy(b, t.Addr[:])
		pf = wasi.Inet6
		port = t.Port
		ok = true
	}
	return
}

func (m *Module) wasmEdgeV2PutSocketAddress(b wasmEdgeAddress, sa wasi.SocketAddress) (port int, ok bool) {
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
