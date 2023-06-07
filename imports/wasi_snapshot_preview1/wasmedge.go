package wasi_snapshot_preview1

import (
	"context"
	"encoding/binary"
	"io"
	"unsafe"

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
	"sock_getaddrinfo":  wazergo.F6((*Module).WasmEdgeSockAddrInfo),
}

// WasmEdgeV2 is V2 of the WasmEdge sockets extension to WASI preview 1.
//
// Version 2 has a sock_accept function that's compatible with the WASI
// preview 1 specification. It widens addresses so that additional
// address families could be supported in future (e.g. AF_UNIX).
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
	"sock_getaddrinfo":  wazergo.F6((*Module).WasmEdgeSockAddrInfo),
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
	_, errno := s.SockBind(ctx, wasi.FD(fd), socketAddr)
	return Errno(errno)
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
	_, errno := s.SockConnect(ctx, wasi.FD(fd), socketAddr)
	return Errno(errno)
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
	size, errno := s.SockSendTo(ctx, wasi.FD(fd), m.iovecs, wasi.SIFlags(flags), socketAddr)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(size))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV1SockRecvFrom(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], addr Pointer[wasmEdgeAddress], iflags Uint32, nread Pointer[Int32], oflags Pointer[Int32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	m.iovecs = iovecs.Append(m.iovecs[:0])
	size, roflags, sa, errno := s.SockRecvFrom(ctx, wasi.FD(fd), m.iovecs, wasi.RIFlags(iflags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	if _, _, ok := m.wasmEdgeV1PutSocketAddress(addr.Load(), sa); !ok {
		return Errno(wasi.EINVAL)
	}
	nread.Store(Int32(size))
	oflags.Store(Int32(roflags))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV2SockRecvFrom(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], addr Pointer[wasmEdgeAddress], iflags Uint32, port Pointer[Uint32], nread Pointer[Int32], oflags Pointer[Int32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	m.iovecs = iovecs.Append(m.iovecs[:0])
	size, roflags, sa, errno := s.SockRecvFrom(ctx, wasi.FD(fd), m.iovecs, wasi.RIFlags(iflags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	portint, ok := m.wasmEdgeV2PutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	port.Store(Uint32(portint))
	nread.Store(Int32(size))
	oflags.Store(Int32(roflags))
	return Errno(wasi.ESUCCESS)
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
	portint, at, ok := m.wasmEdgeV1PutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	addrType.Store(Uint32(at))
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
	sa, errno := s.SockRemoteAddress(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	portint, at, ok := m.wasmEdgeV1PutSocketAddress(addr.Load(), sa)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	addrType.Store(Uint32(at))
	port.Store(Uint32(portint))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV2SockPeerAddr(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Pointer[Uint32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	sa, errno := s.SockRemoteAddress(ctx, wasi.FD(fd))
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

func (m *Module) WasmEdgeSockAddrInfo(ctx context.Context, node String, service String, hintsPtr Pointer[wasmEdgeAddressInfo], resPtr Pointer[wasmEdgeAddressInfo], maxResLength Uint32, resLengthPtr Pointer[Uint32]) Errno {
	s, ok := m.WASI.(wasi.SocketsExtension)
	if !ok {
		return Errno(wasi.ENOSYS)
	}
	// WasmEdge appends null bytes. Remove them.
	if len(node) > 0 && node[len(node)-1] == 0 {
		node = node[:len(node)-1]
	}
	if len(service) > 0 && service[len(service)-1] == 0 {
		service = service[:len(service)-1]
	}
	var hints *wasi.AddressInfo
	if hintsPtr.Offset() != 0 {
		rawhints := hintsPtr.Load()
		hints = &m.addrhint
		hints.Flags = wasi.AddressInfoFlags(rawhints.Flags)
		hints.Family = wasi.ProtocolFamily(rawhints.Family)
		hints.SocketType = wasi.SocketType(rawhints.SocketType)
		hints.Protocol = wasi.Protocol(rawhints.Protocol)
	}
	if int(maxResLength) > cap(m.addrinfo) {
		m.addrinfo = make([]wasi.AddressInfo, int(maxResLength))
	}
	n, errno := s.SockAddressInfo(ctx, string(node), string(service), hints, m.addrinfo[:maxResLength])
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	// TODO: write sock_getaddrinfo results
	results := m.addrinfo[:n]
	_ = results
	panic("not implemented")
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

func (m *Module) wasmEdgeV1PutSocketAddress(b wasmEdgeAddress, sa wasi.SocketAddress) (port, addressType int, ok bool) {
	if len(b) != 16 {
		return
	}
	switch t := sa.(type) {
	case *wasi.Inet4Address:
		binary.LittleEndian.PutUint16(b, uint16(wasi.Inet))
		copy(b, t.Addr[:])
		addressType = 4
		port = t.Port
		ok = true
	case *wasi.Inet6Address:
		binary.LittleEndian.PutUint16(b, uint16(wasi.Inet6))
		copy(b, t.Addr[:])
		addressType = 6
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

type wasmEdgeAddressInfo struct {
	Flags               uint16
	Family              uint8
	SocketType          uint8
	Protocol            uint32
	AddressLength       uint32
	Address             uint32
	CanonicalName       uint32
	CanonicalNameLength uint32
	Next                uint32
}

func (a wasmEdgeAddressInfo) ObjectSize() int {
	return int(unsafe.Sizeof(wasmEdgeAddressInfo{}))
}

func (a wasmEdgeAddressInfo) LoadObject(_ api.Memory, b []byte) wasmEdgeAddressInfo {
	return UnsafeLoadObject[wasmEdgeAddressInfo](b)
}

func (a wasmEdgeAddressInfo) StoreObject(_ api.Memory, b []byte) {
	UnsafeStoreObject(b, a)
}

func (a wasmEdgeAddressInfo) FormatObject(w io.Writer, _ api.Memory, b []byte) {
	Format(w, a.LoadObject(nil, b))
}
