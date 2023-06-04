package wasi

import (
	"context"
	"encoding/json"
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

	// SockConnect connects a socket to an address, returning the local socket
	// address that the connection was made from.
	//
	// The implementation must not retain the socket address.
	//
	// Note: This is similar to connect in POSIX.
	SockConnect(ctx context.Context, fd FD, addr SocketAddress) (SocketAddress, Errno)

	// SockListen allows the socket to accept connections with SockAccept.
	//
	// Note: This is similar to listen in POSIX.
	SockListen(ctx context.Context, fd FD, backlog int) Errno

	// SockSendTo sends a message on a socket.
	//
	// It's similar to SockSend, but accepts an additional SocketAddress.
	//
	// Note: This is similar to sendto in POSIX, though it also supports
	// writing the data from multiple buffers in the manner of writev.
	SockSendTo(ctx context.Context, fd FD, iovecs []IOVec, flags SIFlags, addr SocketAddress) (Size, Errno)

	// SockRecvFrom receives a message from a socket.
	//
	// It's similar to SockRecv, but returns an additional SocketAddress.
	//
	// Note: This is similar to recvfrom in POSIX, though it also supports reading
	// the data into multiple buffers in the manner of readv.
	SockRecvFrom(ctx context.Context, fd FD, iovecs []IOVec, flags RIFlags) (Size, ROFlags, SocketAddress, Errno)

	// SockGetOptInt gets a socket option.
	//
	// Note: This is similar to getsockopt in POSIX.
	SockGetOptInt(ctx context.Context, fd FD, level SocketOptionLevel, option SocketOption) (int, Errno)

	// SockSetOptInt sets a socket option.
	//
	// Note: This is similar to setsockopt in POSIX.
	SockSetOptInt(ctx context.Context, fd FD, level SocketOptionLevel, option SocketOption, value int) Errno

	// SockLocalAddress gets the local address of the socket.
	//
	// The returned address is only valid until the next call on this
	// interface. Assume that any method may invalidate the address.
	//
	// Note: This is similar to getsockname in POSIX.
	SockLocalAddress(ctx context.Context, fd FD) (SocketAddress, Errno)

	// SockRemoteAddress gets the address of the peer when the socket is a
	// connection.
	//
	// The returned address is only valid until the next call on this
	// interface. Assume that any method may invalidate the address.
	//
	// Note: This is similar to getpeername in POSIX.
	SockRemoteAddress(ctx context.Context, fd FD) (SocketAddress, Errno)
}

// Port is a port.
type Port uint32

// SocketAddress is a socket address.
type SocketAddress interface {
	Network() string
	String() string

	sockaddr()
}

// These interfaces are declared in encoding/json and gopkg.in/yaml.v3,
// but we redeclare them here to avoid taking a dependency on those packages.
type jsonMarshaler interface{ MarshalJSON() ([]byte, error) }
type yamlMarshaler interface{ MarshalYAML() (any, error) }

type Inet4Address struct {
	Port int
	Addr [4]byte
}

func (a *Inet4Address) sockaddr() {}

func (a *Inet4Address) Network() string {
	return "ip4"
}

func (a *Inet4Address) String() string {
	return fmt.Sprintf(`%d.%d.%d.%d:%d`, a.Addr[0], a.Addr[1], a.Addr[2], a.Addr[3], a.Port)
}

func (a *Inet4Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%d.%d.%d.%d:%d"`, a.Addr[0], a.Addr[1], a.Addr[2], a.Addr[3], a.Port)), nil
}

func (a *Inet4Address) MarshalYAML() (any, error) {
	return a.String(), nil
}

var (
	_ jsonMarshaler = (*Inet4Address)(nil)
	_ yamlMarshaler = (*Inet4Address)(nil)
)

type Inet6Address struct {
	Port int
	Addr [16]byte
}

func (a *Inet6Address) sockaddr() {}

func (a *Inet6Address) Network() string {
	return "ip6"
}

func (a *Inet6Address) String() string {
	return net.JoinHostPort(net.IP(a.Addr[:]).String(), strconv.Itoa(a.Port))
}

func (a *Inet6Address) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

func (a *Inet6Address) MarshalYAML() (any, error) {
	return a.String(), nil
}

var (
	_ jsonMarshaler = (*Inet6Address)(nil)
	_ yamlMarshaler = (*Inet6Address)(nil)
)

type UnixAddress struct {
	Name string
}

func (a *UnixAddress) sockaddr() {}

func (a *UnixAddress) Network() string {
	return "unix"
}

func (a *UnixAddress) String() string {
	return a.Name
}

func (a *UnixAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Name)
}

func (a *UnixAddress) MarshalYAML() (any, error) {
	return a.Name, nil
}

var (
	_ jsonMarshaler = (*UnixAddress)(nil)
	_ yamlMarshaler = (*UnixAddress)(nil)
)

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
	OOBInline
	Linger
	RecvLowWatermark
	RecvTimeout
	SendTimeout
	QueryAcceptConnections
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
	case OOBInline:
		return "OOBInline"
	case Linger:
		return "Linger"
	case RecvLowWatermark:
		return "RecvLowWatermark"
	case RecvTimeout:
		return "RecvTimeout"
	case SendTimeout:
		return "SendTimeout"
	case QueryAcceptConnections:
		return "QueryAcceptConnections"
	default:
		return fmt.Sprintf("SocketOption(%d)", so)
	}
}
