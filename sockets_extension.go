package wasi

import (
	"context"
	"fmt"
	"net"
	"strconv"
)

// SocketsExtension is a sockets extension for WASI preview 1.
type SocketsExtension interface {
	// SockOpen opens a socket.
	//
	// Note: This is similar to socket in POSIX.
	SockOpen(ctx context.Context, family ProtocolFamily, socketType SocketType, protocol Protocol, rightsBase, rightsInheriting Rights) (FD, Errno)

	// SockBind binds a socket to an address.
	//
	// The implementation must not retain the socket address.
	//
	// Note: This is similar to bind in POSIX.
	SockBind(ctx context.Context, fd FD, addr SocketAddress) Errno

	// SockConnect connects a socket to an address.
	//
	// The implementation must not retain the socket address.
	//
	// Note: This is similar to connect in POSIX.
	SockConnect(ctx context.Context, fd FD, addr SocketAddress) Errno

	// SockListen allows the socket to accept connections with SockAccept.
	//
	// Note: This is similar to listen in POSIX.
	SockListen(ctx context.Context, fd FD, backlog int) Errno

	// SockGetOptInt gets a socket option.
	//
	// Note: This is similar to getsockopt in POSIX.
	SockGetOptInt(ctx context.Context, fd FD, level SocketOptionLevel, option SocketOption) (int, Errno)

	// SockSetOptInt sets a socket option.
	//
	// Note: This is similar to setsockopt in POSIX.
	SockSetOptInt(ctx context.Context, fd FD, level SocketOptionLevel, option SocketOption, value int) Errno

	// SockPeerName gets the address of the peer when the socket is a
	// connection.
	//
	// The returned address is only valid until the next call on this
	// interface. Assume that any method may invalidate the address.
	//
	// Note: This is similar to getpeername in POSIX.
	SockPeerName(ctx context.Context, fd FD) (SocketAddress, Errno)
}

// Port is a port.
type Port uint32

// SocketAddress is a socket address.
type SocketAddress interface {
	Network() string
	String() string

	sockaddr()
}

type Inet4Address struct {
	Port int
	Addr [4]byte
}

func (a *Inet4Address) sockaddr() {}

func (a *Inet4Address) Network() string { return "ip4" }

func (a *Inet4Address) String() string {
	return fmt.Sprintf("%s:%d", net.IP(a.Addr[:]), a.Port)
}

type Inet6Address struct {
	Port int
	Addr [16]byte
}

func (a *Inet6Address) sockaddr() {}

func (a *Inet6Address) Network() string { return "ip6" }

func (a *Inet6Address) String() string {
	return net.JoinHostPort(net.IP(a.Addr[:]).String(), strconv.Itoa(a.Port))
}

// ProtocolFamily is a socket protocol family.
type ProtocolFamily int32

const (
	_ ProtocolFamily = iota
	Inet
	Inet6
)

func (pf ProtocolFamily) String() string {
	switch pf {
	case Inet:
		return "Inet"
	case Inet6:
		return "Inet6"
	default:
		return fmt.Sprintf("ProtocolFamily(%d)", pf)
	}
}

// Protocol is a socket protocol.
type Protocol int32

const (
	IPProtocol Protocol = iota
	TCPProtocol
	UDPProtocol
)

func (p Protocol) String() string {
	switch p {
	case IPProtocol:
		return "IPProtocol"
	case TCPProtocol:
		return "TCPProtocol"
	case UDPProtocol:
		return "UDPProtocol"
	default:
		return fmt.Sprintf("Protocol(%d)", p)
	}
}

// SocketType is a type of socket.
type SocketType int32

const (
	_ SocketType = iota
	DatagramSocket
	StreamSocket
)

func (st SocketType) String() string {
	switch st {
	case DatagramSocket:
		return "DatagramSocket"
	case StreamSocket:
		return "StreamSocket"
	default:
		return fmt.Sprintf("SocketType(%d)", st)
	}
}

// SocketOptionLevel controls the level that a socket option is applied
// at or queried from.
type SocketOptionLevel int32

const (
	SocketLevel SocketOptionLevel = iota
)

func (sl SocketOptionLevel) String() string {
	switch sl {
	case SocketLevel:
		return "SocketLevel"
	default:
		return fmt.Sprintf("SocketOptionLevel(%d)", sl)
	}
}

// SocketOption is a socket option that can be queried or set.
type SocketOption int32

const (
	ReuseAddress SocketOption = iota
	QuerySocketType
	QuerySocketError
	DontRoute
	Broadcast
	SendBufferSize
	RecvBufferSize
	KeepAlive
)

func (so SocketOption) String() string {
	switch so {
	case ReuseAddress:
		return "ReuseAddress"
	case QuerySocketType:
		return "QuerySocketType"
	case QuerySocketError:
		return "QuerySocketError"
	case DontRoute:
		return "DontRoute"
	case Broadcast:
		return "Broadcast"
	case SendBufferSize:
		return "SendBufferSize"
	case RecvBufferSize:
		return "RecvBufferSize"
	case KeepAlive:
		return "KeepAlive"
	default:
		return fmt.Sprintf("SocketOption(%d)", so)
	}
}
