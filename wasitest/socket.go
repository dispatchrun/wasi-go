package wasitest

import (
	"context"
	"testing"

	"github.com/stealthrocket/wasi-go"
)

var socket = testSuite{
	"can create a tcp socket for ipv4": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.InetFamily, wasi.StreamSocket, wasi.TCPProtocol)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"can create a udp socket for ipv4": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.InetFamily, wasi.DatagramSocket, wasi.UDPProtocol)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"can create a tcp socket for ipv6": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.Inet6Family, wasi.StreamSocket, wasi.TCPProtocol)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"can create a udp socket for ipv6": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.Inet6Family, wasi.DatagramSocket, wasi.UDPProtocol)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"can create a stream socket for ipv4 with the default protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.InetFamily, wasi.StreamSocket, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"can create a datagram socket for ipv6 with the default protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.InetFamily, wasi.DatagramSocket, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"can create a stream socket for unix with the default protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.UnixFamily, wasi.StreamSocket, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"can create a datagram socket for unix with the default protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.UnixFamily, wasi.DatagramSocket, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"cannot create an ipv4 stream socket with the udp protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.InetFamily, wasi.StreamSocket, wasi.UDPProtocol)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, wasi.EPROTOTYPE)
	},

	"cannot create an ipv4 datagram socket with the tcp protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.InetFamily, wasi.DatagramSocket, wasi.TCPProtocol)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, wasi.EPROTOTYPE)
	},

	"cannot create an ipv6 stream socket with the udp protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.Inet6Family, wasi.StreamSocket, wasi.UDPProtocol)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, wasi.EPROTOTYPE)
	},

	"cannot create an ipv6 datagram socket with the tcp protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.Inet6Family, wasi.DatagramSocket, wasi.TCPProtocol)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, wasi.EPROTOTYPE)
	},

	"cannot create a unix stream socket with the udp protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.UnixFamily, wasi.StreamSocket, wasi.UDPProtocol)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, wasi.EPROTONOSUPPORT)
	},

	"cannot create a unix stream socket with the tcp protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.UnixFamily, wasi.StreamSocket, wasi.TCPProtocol)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, wasi.EPROTONOSUPPORT)
	},

	"cannot create a unix datagram socket with the udp protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.UnixFamily, wasi.DatagramSocket, wasi.UDPProtocol)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, wasi.EPROTONOSUPPORT)
	},

	"cannot create a unix datagram socket with the tcp protocol": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.UnixFamily, wasi.DatagramSocket, wasi.TCPProtocol)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, wasi.EPROTONOSUPPORT)
	},

	"tcp sockets for ipv4 are of stream type": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.InetFamily, wasi.AnySocket, wasi.TCPProtocol)
		assertEqual(t, errno, wasi.ESUCCESS)

		opt, errno := sys.SockGetOpt(ctx, sock, wasi.SocketLevel, wasi.QuerySocketType)
		assertEqual(t, errno, wasi.ESUCCESS)

		val, ok := opt.(wasi.IntValue)
		assertEqual(t, ok, true)
		assertEqual(t, wasi.SocketType(val), wasi.StreamSocket)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"tcp sockets for ipv6 are of stream type": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.Inet6Family, wasi.AnySocket, wasi.TCPProtocol)
		assertEqual(t, errno, wasi.ESUCCESS)

		opt, errno := sys.SockGetOpt(ctx, sock, wasi.SocketLevel, wasi.QuerySocketType)
		assertEqual(t, errno, wasi.ESUCCESS)

		val, ok := opt.(wasi.IntValue)
		assertEqual(t, ok, true)
		assertEqual(t, wasi.SocketType(val), wasi.StreamSocket)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"udp sockets for ipv4 are of datagram type": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.InetFamily, wasi.AnySocket, wasi.UDPProtocol)
		assertEqual(t, errno, wasi.ESUCCESS)

		opt, errno := sys.SockGetOpt(ctx, sock, wasi.SocketLevel, wasi.QuerySocketType)
		assertEqual(t, errno, wasi.ESUCCESS)

		val, ok := opt.(wasi.IntValue)
		assertEqual(t, ok, true)
		assertEqual(t, wasi.SocketType(val), wasi.DatagramSocket)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},

	"udp sockets for ipv6 are of datagram type": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, wasi.Inet6Family, wasi.AnySocket, wasi.UDPProtocol)
		assertEqual(t, errno, wasi.ESUCCESS)

		opt, errno := sys.SockGetOpt(ctx, sock, wasi.SocketLevel, wasi.QuerySocketType)
		assertEqual(t, errno, wasi.ESUCCESS)

		val, ok := opt.(wasi.IntValue)
		assertEqual(t, ok, true)
		assertEqual(t, wasi.SocketType(val), wasi.DatagramSocket)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	},
}

func sockOpen(t *testing.T, ctx context.Context, sys wasi.System, family wasi.ProtocolFamily, typ wasi.SocketType, proto wasi.Protocol) (wasi.FD, wasi.Errno) {
	t.Helper()
	sock, errno := sys.SockOpen(ctx, family, typ, proto, wasi.AllRights, wasi.AllRights)
	skipIfNotImplemented(t, errno)
	return sock, errno
}
