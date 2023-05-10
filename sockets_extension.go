package wasi

// SocketsExtension is a sockets extension for WASI preview 1.
type SocketsExtension interface {
	SockOpen(family ProtocolFamily, socketType SocketType) (FD, Errno)
	SockBind(fd FD, addr SocketAddress, port Port) Errno
	SockConnect(fd FD, addr SocketAddress, port Port) Errno
	SockListen(fd FD, backlog int) Errno
	SockGetOptInt(fd FD, level SocketOptionLevel, option SocketOption) (int, Errno)
	SockSetOptInt(fd FD, level SocketOptionLevel, option SocketOption, value int) Errno
}

type Port uint32

type SocketAddress []byte

type ProtocolFamily int32

const (
	_ ProtocolFamily = iota
	Inet
	Inet6
)

type Protocol int32

const (
	IPProtocol Protocol = iota
	TCPProtocol
	UDPProtocol
)

type SocketType int32

const (
	_ SocketType = iota
	DatagramSocket
	StreamSocket
)

type SocketOptionLevel int32

const (
	SocketLevel SocketOptionLevel = iota
)

type SocketOption int32

const (
	ReuseAddress SocketOption = iota
	QuerySocketType
	QuerySocketError
)
