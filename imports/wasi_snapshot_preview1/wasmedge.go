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
	rightsBase := wasi.SockListenRights | wasi.SockConnectionRights
	rightsInheriting := wasi.SockConnectionRights
	result, errno := m.WASI.SockOpen(ctx, wasi.ProtocolFamily(family), wasi.SocketType(sockType), wasi.IPProtocol, rightsBase, rightsInheriting)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	openfd.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeSockBind(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Uint32) Errno {
	socketAddr, ok := m.wasmEdgeGetSocketAddress(addr.Load(), int(port))
	if !ok {
		return Errno(wasi.EINVAL)
	}
	_, errno := m.WASI.SockBind(ctx, wasi.FD(fd), socketAddr)
	return Errno(errno)
}

func (m *Module) WasmEdgeSockConnect(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], port Uint32) Errno {
	socketAddr, ok := m.wasmEdgeGetSocketAddress(addr.Load(), int(port))
	if !ok {
		return Errno(wasi.EINVAL)
	}
	_, errno := m.WASI.SockConnect(ctx, wasi.FD(fd), socketAddr)
	return Errno(errno)
}

func (m *Module) WasmEdgeSockListen(ctx context.Context, fd Int32, backlog Int32) Errno {
	return Errno(m.WASI.SockListen(ctx, wasi.FD(fd), int(backlog)))
}

func (m *Module) WasmEdgeSockSendTo(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], addr Pointer[wasmEdgeAddress], port Int32, flags Uint32, nwritten Pointer[Int32]) Errno {
	socketAddr, ok := m.wasmEdgeGetSocketAddress(addr.Load(), int(port))
	if !ok {
		return Errno(wasi.EINVAL)
	}
	m.iovecs = iovecs.Append(m.iovecs[:0])
	size, errno := m.WASI.SockSendTo(ctx, wasi.FD(fd), m.iovecs, wasi.SIFlags(flags), socketAddr)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(size))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV1SockRecvFrom(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], addr Pointer[wasmEdgeAddress], iflags Uint32, nread Pointer[Int32], oflags Pointer[Int32]) Errno {
	m.iovecs = iovecs.Append(m.iovecs[:0])
	size, roflags, sa, errno := m.WASI.SockRecvFrom(ctx, wasi.FD(fd), m.iovecs, wasi.RIFlags(iflags))
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
	m.iovecs = iovecs.Append(m.iovecs[:0])
	size, roflags, sa, errno := m.WASI.SockRecvFrom(ctx, wasi.FD(fd), m.iovecs, wasi.RIFlags(iflags))
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
	// See socket.go
	switch wasi.SocketOptionLevel(level) {
	case wasi.TcpLevel:
		option += 0x1000
	}
	// Only int options are supported for now.
	switch wasi.SocketOption(option) {
	case wasi.Linger, wasi.RecvTimeout, wasi.SendTimeout, wasi.BindToDevice:
		// These accept struct linger / struct timeval / string.
		return Errno(wasi.ENOTSUP)
	}
	if valueLen != 4 {
		return Errno(wasi.EINVAL)
	}
	return Errno(m.WASI.SockSetOpt(ctx, wasi.FD(fd), wasi.SocketOptionLevel(level), wasi.SocketOption(option), wasi.IntValue(value.Load())))
}

func (m *Module) WasmEdgeSockGetOpt(ctx context.Context, fd Int32, level Int32, option Int32, value Pointer[Int32], valueLen Int32) Errno {
	// See socket.go
	switch wasi.SocketOptionLevel(level) {
	case wasi.TcpLevel:
		option += 0x1000
	}
	// Only int options are supported for now.
	switch wasi.SocketOption(option) {
	case wasi.Linger, wasi.RecvTimeout, wasi.SendTimeout, wasi.BindToDevice:
		// These accept struct linger / struct timeval / string.
		return Errno(wasi.ENOTSUP)
	}
	if valueLen != 4 {
		return Errno(wasi.EINVAL)
	}
	result, errno := m.WASI.SockGetOpt(ctx, wasi.FD(fd), wasi.SocketOptionLevel(level), wasi.SocketOption(option))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	intval, ok := result.(wasi.IntValue)
	if !ok {
		return Errno(wasi.EINVAL)
	}
	value.Store(Int32(intval))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) WasmEdgeV1SockLocalAddr(ctx context.Context, fd Int32, addr Pointer[wasmEdgeAddress], addrType Pointer[Uint32], port Pointer[Uint32]) Errno {
	sa, errno := m.WASI.SockLocalAddress(ctx, wasi.FD(fd))
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
	sa, errno := m.WASI.SockLocalAddress(ctx, wasi.FD(fd))
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
	sa, errno := m.WASI.SockRemoteAddress(ctx, wasi.FD(fd))
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
	sa, errno := m.WASI.SockRemoteAddress(ctx, wasi.FD(fd))
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

func (m *Module) WasmEdgeSockAddrInfo(ctx context.Context, name String, service String, hintsPtr Pointer[wasmEdgeAddressInfo], resPtrPtr Pointer[Pointer[wasmEdgeAddressInfo]], maxResLength Uint32, resLengthPtr Pointer[Uint32]) Errno {
	if len(name) == 0 && len(service) == 0 || maxResLength == 1 {
		return Errno(wasi.EINVAL)
	}
	// WasmEdge appends null bytes. Remove them.
	if len(name) > 0 && name[len(name)-1] == 0 {
		name = name[:len(name)-1]
	}
	if len(service) > 0 && service[len(service)-1] == 0 {
		service = service[:len(service)-1]
	}
	var hints wasi.AddressInfo
	rawhints := hintsPtr.Load()
	hints.Flags = wasi.AddressInfoFlags(rawhints.Flags)
	hints.Family = wasi.ProtocolFamily(rawhints.Family)
	hints.SocketType = wasi.SocketType(rawhints.SocketType)
	hints.Protocol = wasi.Protocol(rawhints.Protocol)

	if int(maxResLength) > cap(m.addrinfo) {
		m.addrinfo = make([]wasi.AddressInfo, int(maxResLength))
	}
	n, errno := m.WASI.SockAddressInfo(ctx, string(name), string(service), hints, m.addrinfo[:maxResLength])
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	} else if n == 0 {
		resLengthPtr.Store(0)
		return Errno(wasi.ESUCCESS)
	}

	// This is a very poorly designed interface. The results pointer points to
	// an addrinfo style struct, and you're supposed to treat it as a linked
	// list? The socket address isn't a sockaddr style struct, but rather a
	// pointer to some other struct which has a length and some more indirection
	// (github.com/second-state/wasmedge_wasi_socket/blob/7e49c11/src/socket.rs#L78).
	// We have no idea here how much space the guest has allocated for socket
	// addresses and canonical names.
	// There's an addrlen (github.com/second-state/wasmedge_wasi_socket/blob/7e49c11/src/socket.rs#L112)
	// field, but it isn't set by the WasmEdge sockets lib. It's not clear
	// whether that's the length of the object that addr points to, or whether
	// object always points to a WasiSockaddr. If it's the latter, WasiSockaddr
	// has its own sa_data_len field? Why is sa_data_len=14 but the sockets lib
	// allocates 26 bytes of space (github.com/second-state/wasmedge_wasi_socket/blob/7e49c11/src/socket.rs#L172)?
	// Same thing with the canonical name. The sockets lib allocates 30 bytes of space,
	// but then doesn't set ai_canonnamelen... Argh.
	mem := resPtrPtr.Memory()
	resPtr := resPtrPtr.Load()
	results := m.addrinfo[:n]
	count := 0
	for {
		res := resPtr.Load()
		if res.Address == 0 {
			return Errno(wasi.EFAULT)
		}
		res.AddressLength = 16 // sizeof(WasiSockaddr)
		addrDataFamily, ok := mem.Read(res.Address, 1)
		if !ok {
			return Errno(wasi.EFAULT)
		}
		addrDataLen, ok := mem.ReadUint32Le(res.Address + 4)
		if !ok {
			return Errno(wasi.EFAULT)
		}
		// WasmEdge lies.
		if addrDataLen == 14 {
			addrDataLen = 26
		}
		addrDataPtr, ok := mem.ReadUint32Le(res.Address + 8)
		if !ok {
			return Errno(wasi.EFAULT)
		}
		addrData, ok := mem.Read(addrDataPtr, addrDataLen)
		if !ok {
			return Errno(wasi.EFAULT)
		}
		switch addr := results[0].Address.(type) {
		case *wasi.Inet4Address:
			if len(addrData) < 6 {
				return Errno(wasi.EFAULT)
			}
			binary.BigEndian.PutUint16(addrData, uint16(addr.Port))
			copy(addrData[2:], addr.Addr[:])
			addrDataFamily[0] = uint8(wasi.InetFamily)
			// mem.WriteUint32Le(res.Address+4, 6) // WasmEdge writes 16?
		case *wasi.Inet6Address:
			if len(addrData) < 18 {
				return Errno(wasi.EFAULT)
			}
			binary.BigEndian.PutUint16(addrData, uint16(addr.Port))
			copy(addrData[2:], addr.Addr[:])
			addrDataFamily[0] = uint8(wasi.Inet6Family)
			// mem.WriteUint32Le(res.Address+4, 18) // WasmEdge writes 26?
		}
		res.CanonicalNameLength = 0 // Not yet supported
		resPtr.Store(res)
		count++
		results = results[1:]
		if res.Next == 0 || len(results) == 0 {
			break
		}
		resPtr = Ptr[wasmEdgeAddressInfo](resPtr.Memory(), res.Next)
	}
	resLengthPtr.Store(Uint32(count))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) wasmEdgeGetSocketAddress(b wasmEdgeAddress, port int) (sa wasi.SocketAddress, ok bool) {
	// V2 addresses.
	if len(b) == 128 {
		switch wasi.ProtocolFamily(binary.LittleEndian.Uint16(b)) {
		case wasi.InetFamily:
			b = b[2:6] // fallthrough to v1 parser below
		case wasi.Inet6Family:
			b = b[2:18] // fallthrough to v1 parser below
		case wasi.UnixFamily:
			b = b[2:]
			n := 0
			for n < len(b) && b[n] != 0 {
				n++
			}
			if n == len(b) || b[n] != 0 {
				return
			}
			m.unixaddr.Name = string(b[:n])
			return &m.unixaddr, true
		default:
			return
		}
	}

	// V1 addresses.
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
		binary.LittleEndian.PutUint16(b, uint16(wasi.InetFamily))
		copy(b, t.Addr[:])
		addressType = 4
		port = t.Port
		ok = true
	case *wasi.Inet6Address:
		binary.LittleEndian.PutUint16(b, uint16(wasi.Inet6Family))
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
		binary.LittleEndian.PutUint16(b, uint16(wasi.InetFamily))
		copy(b[2:], t.Addr[:])
		port = t.Port
		ok = true
	case *wasi.Inet6Address:
		binary.LittleEndian.PutUint16(b, uint16(wasi.Inet6Family))
		copy(b[2:], t.Addr[:])
		port = t.Port
		ok = true
	case *wasi.UnixAddress:
		binary.LittleEndian.PutUint16(b, uint16(wasi.UnixFamily))
		if len(t.Name) > 125 {
			return
		}
		copy(b[2:], t.Name[:])
		b[2+len(t.Name)] = 0
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
