package wasi

// SocketsExtension is a sockets extension for WASI preview 1.
type SocketsExtension interface {
	// SockOpen opens a socket.
	//
	// Note: This is similar to socket in POSIX.
	SockOpen(family ProtocolFamily, socketType SocketType) (FD, Errno)

	// SockBind binds a socket to an address.
	//
	// Note: This is similar to bind in POSIX.
	SockBind(fd FD, addr SocketAddress, port Port) Errno

	// SockConnect connects a socket to an address.
	//
	// Note: This is similar to connect in POSIX.
	SockConnect(fd FD, addr SocketAddress, port Port) Errno

	// SockListen allows the socket to accept connections with SockAccept.
	//
	// Note: This is similar to listen in POSIX.
	SockListen(fd FD, backlog int) Errno

	// SockGetOptInt gets a socket option.
	//
	// Note: This is similar to getsockopt in POSIX.
	SockGetOptInt(fd FD, level SocketOptionLevel, option SocketOption) (int, Errno)

	// SockSetOptInt sets a socket option.
	//
	// Note: This is similar to setsockopt in POSIX.
	SockSetOptInt(fd FD, level SocketOptionLevel, option SocketOption, value int) Errno
}

// Port is a port.
type Port uint32

// SocketAddress is a 4 byte IPv4 address or 16 byte IPv6 address.
type SocketAddress []byte

// ProtocolFamily is a socket protocol family.
type ProtocolFamily int32

const (
	_ ProtocolFamily = iota
	Inet
	Inet6
)

// Protocol is a socket protocol.
type Protocol int32

const (
	IPProtocol Protocol = iota
	TCPProtocol
	UDPProtocol
)

// SocketType is a type of socket.
type SocketType int32

const (
	_ SocketType = iota
	DatagramSocket
	StreamSocket
)

// SocketOptionLevel controls the level that a socket option is applied
// at or queried from.
type SocketOptionLevel int32

const (
	SocketLevel SocketOptionLevel = iota
)

// SocketOption is a socket option that can be queried or set.
type SocketOption int32

const (
	ReuseAddress SocketOption = iota
	QuerySocketType
	QuerySocketError
)
