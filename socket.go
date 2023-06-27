package wasi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"
)

// RIFlags are flags provided to SockRecv.
type RIFlags uint16

const (
	// RecvPeek indicates that SockRecv should return the message without
	// removing it from the socket's receive queue.
	RecvPeek RIFlags = 1 << iota

	// RecvWaitAll indicates that on byte-stream sockets, SockRecv should block
	// until the full amount of data can be returned.
	RecvWaitAll
)

// Has is true if the flag is set.
func (flags RIFlags) Has(f RIFlags) bool {
	return (flags & f) == f
}

var riflagsStrings = [...]string{
	"RecvPeek",
	"RecvWaitAll",
}

func (flags RIFlags) String() (s string) {
	if flags == 0 {
		return "RIFlags(0)"
	}
	for i, name := range riflagsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("RIFlags(%d)", flags)
	}
	return
}

// ROFlags are flags returned by SockRecv.
type ROFlags uint16

const (
	// RecvDataTruncated indicates that message data has been truncated.
	RecvDataTruncated ROFlags = 1 << iota
)

// Has is true if the flag is set.
func (flags ROFlags) Has(f ROFlags) bool {
	return (flags & f) == f
}

func (flags ROFlags) String() string {
	switch flags {
	case RecvDataTruncated:
		return "RecvDataTruncated"
	default:
		return fmt.Sprintf("ROFlags(%d)", flags)
	}
}

// SIFlags are flags provided to SockSend.
//
// As there are currently no flags defined, it must be set to zero.
type SIFlags uint16

// Has is true if the flag is set.
func (flags SIFlags) Has(f SIFlags) bool {
	return (flags & f) == f
}

func (flags SIFlags) String() string {
	return fmt.Sprintf("SIFlags(%d)", flags)
}

// SDFlags are flags provided to SockShutdown which indicate which channels
// on a socket to shut down.
type SDFlags uint16

const (
	// ShutdownRD disables further receive operations.
	ShutdownRD SDFlags = 1 << iota

	// ShutdownWR disables further send operations.
	ShutdownWR
)

// Has is true if the flag is set.
func (flags SDFlags) Has(f SDFlags) bool {
	return (flags & f) == f
}

var sdflagsStrings = [...]string{
	"ShutdownRD",
	"ShutdownWR",
}

func (flags SDFlags) String() (s string) {
	if flags == 0 {
		return "SDFlags(0)"
	}
	for i, name := range sdflagsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("SDFlags(%d)", flags)
	}
	return
}

// =============================================================================
// The types and functions below are used in socket extensions to the initial
// WASI preview 1 specification.
// =============================================================================

// Port is a port.
type Port uint32

// SocketAddress is a socket address.
type SocketAddress interface {
	Network() string
	String() string
	Family() ProtocolFamily
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

func (a *Inet4Address) Family() ProtocolFamily {
	return InetFamily
}

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

func (a *Inet6Address) Family() ProtocolFamily {
	return Inet6Family
}

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

func (a *UnixAddress) Family() ProtocolFamily {
	return UnixFamily
}

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
	UnspecifiedFamily ProtocolFamily = iota
	InetFamily
	Inet6Family
	UnixFamily
)

func (pf ProtocolFamily) String() string {
	switch pf {
	case UnspecifiedFamily:
		return "UnspecifiedFamily"
	case InetFamily:
		return "InetFamily"
	case Inet6Family:
		return "Inet6Family"
	case UnixFamily:
		return "UnixFamily"
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
	AnySocket SocketType = iota
	DatagramSocket
	StreamSocket
)

func (st SocketType) String() string {
	switch st {
	case AnySocket:
		return "AnySocket"
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
	SocketLevel   SocketOptionLevel = 0 // SOL_SOCKET
	TcpLevel      SocketOptionLevel = 6 // IPPROTO_TCP
	ReservedLevel SocketOptionLevel = 0x74696d65
)

func (sl SocketOptionLevel) String() string {
	switch sl {
	case SocketLevel:
		return "SocketLevel"
	case TcpLevel:
		return "TcpLevel"
	case ReservedLevel:
		return "ReservedLevel"
	default:
		return fmt.Sprintf("SocketOptionLevel(%d)", sl)
	}
}

// SocketOption is a socket option that can be queried or set.
type SocketOption int32

const (
	// SOL_SOCKET level options.
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
	BindToDevice

	// 0x1000 + iota are IPPROTO_TCP level options.
	TcpNoDelay SocketOption = 0x1000 + iota

	// >= 0x9000 are reserved.
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
	case BindToDevice:
		return "BindToDevice"
	case TcpNoDelay:
		return "TcpNoDelay"
	default:
		return fmt.Sprintf("SocketOption(%d)", so)
	}
}

// AddressInfo is information about an address.
type AddressInfo struct {
	Flags         AddressInfoFlags
	Family        ProtocolFamily
	SocketType    SocketType
	Protocol      Protocol
	Address       SocketAddress
	CanonicalName string
}

// AddressInfoFlags are AddressInfo flags.
type AddressInfoFlags uint16

const (
	Passive AddressInfoFlags = 1 << iota
	CanonicalName
	NumericHost
	NumericService
	V4Mapped
	QueryAll
	AddressConfigured
)

// Has is true if the flag is set. If multiple flags are specified, Has returns
// true if all flags are set.
func (flags AddressInfoFlags) Has(f AddressInfoFlags) bool {
	return (flags & f) == f
}

// HasAny is true if any of the specified flags are set.
func (flags AddressInfoFlags) HasAny(f AddressInfoFlags) bool {
	return (flags & f) != 0
}

var addressInfoFlagsStrings = [...]string{
	"Passive",
	"CanonicalName",
	"NumericHost",
	"NumericService",
	"V4Mapped",
	"QueryAll",
	"AddressConfigured",
}

func (flags AddressInfoFlags) String() (s string) {
	for i, name := range addressInfoFlagsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("AddressInfoFlags(%d)", flags)
	}
	return
}

// SocketOptionValue is a socket option value.
type SocketOptionValue interface {
	String() string

	sockopt()
}

// IntValue is an integer value.
type IntValue int

func (IntValue) sockopt() {}

func (i IntValue) String() string {
	return strconv.Itoa(int(i))
}

// TimeValue is used to represent socket options with a duration value.
type TimeValue Timestamp

func (TimeValue) sockopt() {}

func (tv TimeValue) String() string {
	return time.Duration(tv).String()
}

// StringValue is used to represent an arbitrary socket option value.
type StringValue string

func (StringValue) sockopt() {}

func (s StringValue) String() string {
	return string(s)
}

// SocketsNotSupported is a helper type intended to be embeded in
// implementations of the Sytem interface that do not support sockets.
//
// The type defines all socket-related methods to return ENOSYS, allowing
// the type to implement the interface but indicating to callers that the
// functionality is not supported.
type SocketsNotSupported struct{}

func (SocketsNotSupported) SockOpen(ctx context.Context, family ProtocolFamily, socketType SocketType, protocol Protocol, rightsBase, rightsInheriting Rights) (FD, Errno) {
	return -1, ENOSYS
}

func (SocketsNotSupported) SockBind(ctx context.Context, fd FD, addr SocketAddress) (SocketAddress, Errno) {
	return nil, ENOSYS
}

func (SocketsNotSupported) SockConnect(ctx context.Context, fd FD, addr SocketAddress) (SocketAddress, Errno) {
	return nil, ENOSYS
}

func (SocketsNotSupported) SockListen(ctx context.Context, fd FD, backlog int) Errno {
	return ENOSYS
}

func (SocketsNotSupported) SockAccept(ctx context.Context, fd FD, flags FDFlags) (FD, SocketAddress, SocketAddress, Errno) {
	return -1, nil, nil, ENOSYS
}

func (SocketsNotSupported) SockRecv(ctx context.Context, fd FD, iovecs []IOVec, flags RIFlags) (Size, ROFlags, Errno) {
	return 0, 0, ENOSYS
}

func (SocketsNotSupported) SockSend(ctx context.Context, fd FD, iovecs []IOVec, flags SIFlags) (Size, Errno) {
	return 0, ENOSYS
}

func (SocketsNotSupported) SockSendTo(ctx context.Context, fd FD, iovecs []IOVec, flags SIFlags, addr SocketAddress) (Size, Errno) {
	return 0, ENOSYS
}

func (SocketsNotSupported) SockRecvFrom(ctx context.Context, fd FD, iovecs []IOVec, flags RIFlags) (Size, ROFlags, SocketAddress, Errno) {
	return 0, 0, nil, ENOSYS
}

func (SocketsNotSupported) SockGetOpt(ctx context.Context, fd FD, level SocketOptionLevel, option SocketOption) (SocketOptionValue, Errno) {
	return nil, ENOSYS
}

func (SocketsNotSupported) SockSetOpt(ctx context.Context, fd FD, level SocketOptionLevel, option SocketOption, value SocketOptionValue) Errno {
	return ENOSYS
}

func (SocketsNotSupported) SockLocalAddress(ctx context.Context, fd FD) (SocketAddress, Errno) {
	return nil, ENOSYS
}

func (SocketsNotSupported) SockRemoteAddress(ctx context.Context, fd FD) (SocketAddress, Errno) {
	return nil, ENOSYS
}

func (SocketsNotSupported) SockAddressInfo(ctx context.Context, name, service string, hints AddressInfo, results []AddressInfo) (int, Errno) {
	return 0, ENOSYS
}

func (SocketsNotSupported) SockShutdown(ctx context.Context, fd FD, flags SDFlags) Errno {
	return ENOSYS
}
