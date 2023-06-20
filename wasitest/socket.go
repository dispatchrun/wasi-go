package wasitest

import (
	"context"
	"math"
	"testing"

	"github.com/stealthrocket/wasi-go"
)

var (
	localIPv4 = [4]byte{127, 0, 0, 1}
	localIPv6 = [16]byte{15: 1}

	unknownIPv4 = [4]byte{123, 234, 123, 234}
	unknownIPv6 = [16]byte{15: 2}
)

var socket = testSuite{
	"can create a tcp socket for ipv4": testSocketOpenOK(
		wasi.InetFamily, wasi.StreamSocket, wasi.TCPProtocol,
	),

	"can create a tcp socket for ipv6": testSocketOpenOK(
		wasi.Inet6Family, wasi.StreamSocket, wasi.TCPProtocol,
	),

	"can create a udp socket for ipv4": testSocketOpenOK(
		wasi.InetFamily, wasi.DatagramSocket, wasi.UDPProtocol,
	),

	"can create a udp socket for ipv6": testSocketOpenOK(
		wasi.Inet6Family, wasi.DatagramSocket, wasi.UDPProtocol,
	),

	"can create a stream socket for ipv4 with the default protocol": testSocketOpenOK(
		wasi.InetFamily, wasi.StreamSocket, 0,
	),

	"can create a stream socket for ipv6 with the default protocol": testSocketOpenOK(
		wasi.Inet6Family, wasi.StreamSocket, 0,
	),

	"can create a datagram socket for ipv4 with the default protocol": testSocketOpenOK(
		wasi.InetFamily, wasi.DatagramSocket, 0,
	),

	"can create a datagram socket for ipv6 with the default protocol": testSocketOpenOK(
		wasi.Inet6Family, wasi.DatagramSocket, 0,
	),

	"can create a stream socket for unix with the default protocol": testSocketOpenOK(
		wasi.UnixFamily, wasi.StreamSocket, 0,
	),

	"cannot create an ipv4 stream socket with the udp protocol": testSocketOpenError(
		wasi.InetFamily, wasi.StreamSocket, wasi.UDPProtocol, wasi.EPROTONOSUPPORT,
	),

	"cannot create an ipv4 datagram socket with the tcp protocol": testSocketOpenError(
		wasi.InetFamily, wasi.DatagramSocket, wasi.TCPProtocol, wasi.EPROTONOSUPPORT,
	),

	"cannot create an ipv6 stream socket with the udp protocol": testSocketOpenError(
		wasi.Inet6Family, wasi.StreamSocket, wasi.UDPProtocol, wasi.EPROTONOSUPPORT,
	),

	"cannot create an ipv6 datagram socket with the tcp protocol": testSocketOpenError(
		wasi.Inet6Family, wasi.DatagramSocket, wasi.TCPProtocol, wasi.EPROTONOSUPPORT,
	),

	"cannot create a unix stream socket with the tcp protocol": testSocketOpenError(
		wasi.UnixFamily, wasi.StreamSocket, wasi.TCPProtocol, wasi.EPROTONOSUPPORT,
	),

	"cannot create a unix stream socket with the udp protocol": testSocketOpenError(
		wasi.UnixFamily, wasi.StreamSocket, wasi.UDPProtocol, wasi.EPROTONOSUPPORT,
	),

	"cannot create a unix datagram socket with the tcp protocol": testSocketOpenError(
		wasi.UnixFamily, wasi.DatagramSocket, wasi.TCPProtocol, wasi.EPROTONOSUPPORT,
	),

	"cannot create a unix datagram socket with the udp protocol": testSocketOpenError(
		wasi.UnixFamily, wasi.DatagramSocket, wasi.UDPProtocol, wasi.EPROTONOSUPPORT,
	),

	"tcp sockets for ipv4 are of stream type": testSocketType(
		wasi.InetFamily, wasi.StreamSocket, wasi.TCPProtocol,
	),

	"tcp sockets for ipv6 are of stream type": testSocketType(
		wasi.Inet6Family, wasi.StreamSocket, wasi.TCPProtocol,
	),

	"udp sockets for ipv4 are of datagram type": testSocketType(
		wasi.InetFamily, wasi.DatagramSocket, wasi.UDPProtocol,
	),

	"udp sockets for ipv6 are of datagram type": testSocketType(
		wasi.Inet6Family, wasi.DatagramSocket, wasi.UDPProtocol,
	),

	"bind an ipv4 stream socket to a port selects that port": testSocketBindOK(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4, Port: 41200},
	),

	"bind an ipv4 datagram socket to a port selects that port": testSocketBindOK(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{Addr: localIPv4, Port: 41201},
	),

	"bind an ipv6 stream socket to a port selects that port": testSocketBindOK(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6, Port: 41202},
	),

	"bind an ipv6 datagram socket to a port selects that port": testSocketBindOK(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{Addr: localIPv6, Port: 41203},
	),

	"bind an ipv4 stream socket to port zero selects a random port": testSocketBindOK(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"bind an ipv4 datagram socket to port zero selects a random port": testSocketBindOK(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"bind an ipv6 stream socket to port zero selects a random port": testSocketBindOK(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"bind an ipv6 datagram socket to port zero selects a random port": testSocketBindOK(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"bind an ipv4 stream socket to address zero selects any address": testSocketBindOK(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{},
	),

	"bind an ipv4 datagram socket to address zero selects any address": testSocketBindOK(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{},
	),

	"bind an ipv6 stream socket to address zero selects any address": testSocketBindOK(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{},
	),

	"bind an ipv6 datagram socket to address zero selects any address": testSocketBindOK(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{},
	),

	"cannot bind an ipv4 stream socket to an address which does not exist": testSocketBindError(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: unknownIPv4}, wasi.EADDRNOTAVAIL,
	),

	"cannot bind an ipv4 datagram socket to an address which does not exist": testSocketBindError(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{Addr: unknownIPv4}, wasi.EADDRNOTAVAIL,
	),

	"cannot bind an ipv6 stream socket to an address which does not exist": testSocketBindError(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: unknownIPv6}, wasi.EADDRNOTAVAIL,
	),

	"cannot bind an ipv6 datagram socket to an address which does not exist": testSocketBindError(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{Addr: unknownIPv6}, wasi.EADDRNOTAVAIL,
	),

	"cannot bind an ipv4 stream socket to a port which does not exist": testSocketBindError(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4, Port: -1}, wasi.EINVAL,
	),

	"cannot bind an ipv4 datagram socket to a port which does not exist": testSocketBindError(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{Addr: localIPv4, Port: -1}, wasi.EINVAL,
	),

	"cannot bind an ipv6 stream socket to a port which does not exist": testSocketBindError(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6, Port: -1}, wasi.EINVAL,
	),

	"cannot bind an ipv6 datagram socket to a port which does not exist": testSocketBindError(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{Addr: localIPv6, Port: -1}, wasi.EINVAL,
	),

	"cannot bind an ipv4 stream socket that was already bound": testSocketBindAfterBind(
		wasi.InetFamily, wasi.StreamSocket,
		&wasi.Inet4Address{Addr: localIPv4},
		&wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot bind an ipv6 stream socket that was already bound": testSocketBindAfterBind(
		wasi.Inet6Family, wasi.StreamSocket,
		&wasi.Inet6Address{Addr: localIPv6},
		&wasi.Inet6Address{Addr: localIPv6},
	),

	"cannot bind an ipv4 datagram socket that was already bound": testSocketBindAfterBind(
		wasi.InetFamily, wasi.DatagramSocket,
		&wasi.Inet4Address{Addr: localIPv4},
		&wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot bind an ipv6 datagram socket that was already bound": testSocketBindAfterBind(
		wasi.Inet6Family, wasi.DatagramSocket,
		&wasi.Inet6Address{Addr: localIPv6},
		&wasi.Inet6Address{Addr: localIPv6},
	),

	"cannot bind an ipv4 datagram socket that was already connected": testSocketBindAfterConnect(
		wasi.InetFamily, wasi.DatagramSocket,
		&wasi.Inet4Address{Addr: localIPv4, Port: 53},
		&wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot bind an ipv6 datagram socket that was already connected": testSocketBindAfterConnect(
		wasi.Inet6Family, wasi.DatagramSocket,
		&wasi.Inet6Address{Addr: localIPv6, Port: 53},
		&wasi.Inet6Address{Addr: localIPv6},
	),

	"can listen on ipv4 stream sockets": testSocketListenOK(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"can listen on ipv6 stream sockets": testSocketListenOK(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"can connect two ipv4 stream sockets": testSocketConnectAndAccept(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"can connect two ipv6 stream sockets": testSocketConnectAndAccept(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"can connect a ipv4 datagram socket": testSocketConnectOK(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{Addr: localIPv4, Port: 53},
	),

	"can connect a ipv6 datagram socket": testSocketConnectOK(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{Addr: localIPv6, Port: 53},
	),

	"failing to connect sets the socket error and getting the socket error clears it on ipv4 stream sockets": testSocketConnectError(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4, Port: 62431},
	),

	"failing to connect sets the socket error and getting the socket error clears it on ipv6 stream sockets": testSocketConnectError(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6, Port: 62432},
	),

	"cannot connect a listening ipv4 stream socket": testSocketConnectAfterListen(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot connect a listening ipv6 stream socket": testSocketConnectAfterListen(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"cannot connect a connected ipv4 stream socket": testSocketConnectAfterConnect(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot connect a connected ipv6 stream socket": testSocketConnectAfterConnect(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"listen on an unbound ipv4 stream socket automatically binds it": testSocketListenBeforeBind(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"listen on an unbound ipv6 stream socket automatically binds it": testSocketListenBeforeBind(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"listen on a listening ipv4 stream socket is supported": testSocketListenAfterListen(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"listen on a listening ipv6 stream socket is supported": testSocketListenAfterListen(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"listen with a negative backlog on an ipv4 address is invalid": testSocketListenNegativeBacklog(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"listen with a negative backlog on an ipv6 address is invalid": testSocketListenNegativeBacklog(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"listen with a large backlog on an ipv4 address is supported": testSocketListenLargeBacklog(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"listen with a large backlog on an ipv6 address is supported": testSocketListenLargeBacklog(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot shutdown an ipv4 stream socket with an invalid argument": testSocketShutdownInvalidArgument(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"cannot shutdown an ipv6 stream socket with an invalid argument": testSocketShutdownInvalidArgument(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot shutdown an ipv4 datagram socket with an invalid argument": testSocketShutdownInvalidArgument(
		wasi.InetFamily, wasi.DatagramSocket,
	),

	"cannot shutdown an ipv6 datagram socket with an invalid argument": testSocketShutdownInvalidArgument(
		wasi.Inet6Family, wasi.DatagramSocket,
	),

	"cannot shutdown an ipv4 stream socket which is not connected": testSocketShutdownBeforeConnect(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"cannot shutdown an ipv6 stream socket which is not connected": testSocketShutdownBeforeConnect(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot shutdown an ipv4 datagram socket which is not connected": testSocketShutdownBeforeConnect(
		wasi.InetFamily, wasi.DatagramSocket,
	),

	"cannot shutdown an ipv6 datagram socket which is not connected": testSocketShutdownBeforeConnect(
		wasi.Inet6Family, wasi.DatagramSocket,
	),

	"cannot shutdown an ipv4 stream socket which is listening": testSocketShutdownAfterListen(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot shutdown an ipv6 stream socket which is listening": testSocketShutdownAfterListen(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"can shutdown ipv4 stream socket after accepting": testSocketConnectAndShutdown(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"the default buffer sizes are not zero on ipv4 stream sockets": testSocketDefaultBufferSizes(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"the default buffer sizes are not zero on ipv6 stream sockets": testSocketDefaultBufferSizes(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"can set the buffer sizes of ipv4 stream sockets": testSocketSetBufferSizes(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"can set the buffer sizes of ipv6 stream sockets": testSocketSetBufferSizes(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot option of ipv4 stream socket with invalid level": testSocketSetOptionInvalidLevel(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"cannot option of ipv6 stream socket with invalid level": testSocketSetOptionInvalidLevel(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot option of ipv4 stream socket with invalid argument": testSocketSetOptionInvalidArgument(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"cannot option of ipv6 stream socket with invalid argument": testSocketSetOptionInvalidArgument(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot bind a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			_, errno := sys.SockBind(ctx, fd, &wasi.Inet4Address{Addr: localIPv4})
			return errno
		},
	),

	"cannot listen on a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			return sys.SockListen(ctx, fd, 0)
		},
	),

	"cannot receive on a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			_, _, errno := sys.SockRecv(ctx, fd, []wasi.IOVec{nil}, 0)
			return errno
		},
	),

	"cannot send on a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			_, errno := sys.SockSend(ctx, fd, []wasi.IOVec{nil}, 0)
			return errno
		},
	),

	"cannot shutdown a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			return sys.SockShutdown(ctx, fd, wasi.ShutdownRD|wasi.ShutdownWR)
		},
	),

	"cannot accept on a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			_, _, _, errno := sys.SockAccept(ctx, fd, 0)
			return errno
		},
	),

	"cannot get socket options on a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			_, errno := sys.SockGetOpt(ctx, fd, wasi.SocketLevel, wasi.QuerySocketType)
			return errno
		},
	),

	"cannot set socket options on a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			return sys.SockSetOpt(ctx, fd, wasi.SocketLevel, wasi.SendBufferSize, wasi.IntValue(4096))
		},
	),
}

func testNotSocket(test func(context.Context, wasi.System, wasi.FD) wasi.Errno) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		assertEqual(t, test(ctx, sys, 0), wasi.ENOTSOCK)
		assertEqual(t, test(ctx, sys, 1), wasi.ENOTSOCK)
		assertEqual(t, test(ctx, sys, 2), wasi.ENOTSOCK)
	}
}

func testSocketType(family wasi.ProtocolFamily, typ wasi.SocketType, proto wasi.Protocol) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, wasi.AnySocket, proto)
		assertEqual(t, errno, wasi.ESUCCESS)

		opt, errno := sys.SockGetOpt(ctx, sock, wasi.SocketLevel, wasi.QuerySocketType)
		assertEqual(t, errno, wasi.ESUCCESS)

		val, ok := opt.(wasi.IntValue)
		assertEqual(t, ok, true)
		assertEqual(t, wasi.SocketType(val), typ)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketOpenOK(family wasi.ProtocolFamily, typ wasi.SocketType, proto wasi.Protocol) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, proto)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketOpenError(family wasi.ProtocolFamily, typ wasi.SocketType, proto wasi.Protocol, want wasi.Errno) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, proto)
		assertEqual(t, sock, ^wasi.FD(0))
		assertEqual(t, errno, want)
	}
}

func testSocketBindOK(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, addr.Family(), bind.Family())

		switch a := addr.(type) {
		case *wasi.Inet4Address:
			b := bind.(*wasi.Inet4Address)
			assertEqual(t, a.Addr, b.Addr)
			if b.Port == 0 {
				assertNotEqual(t, a.Port, 0)
			} else {
				assertEqual(t, a.Port, b.Port)
			}
		case *wasi.Inet6Address:
			b := bind.(*wasi.Inet6Address)
			assertEqual(t, a.Addr, b.Addr)
			if b.Port == 0 {
				assertNotEqual(t, a.Port, 0)
			} else {
				assertEqual(t, a.Port, b.Port)
			}
		default:
			t.Errorf("socket bound to address of unknown type %T", a)
		}

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketBindError(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress, want wasi.Errno) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, addr, nil)
		assertEqual(t, errno, want)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketListenOK(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 10), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketConnectOK(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockConnect(ctx, sock, bind)
		assertNotEqual(t, addr, nil)
		assertEqual(t, addr.Family(), bind.Family())
		if errno != wasi.ESUCCESS {
			assertEqual(t, errno, wasi.EINPROGRESS)
		}

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(42, wasi.FDWriteEvent, wasi.SubscriptionFDReadWrite{
				FD: sock,
			}),
		}
		evs := make([]wasi.Event, len(subs))

		numEvents, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)

		assertEqual(t, evs[0], wasi.Event{
			UserData:  42,
			EventType: wasi.FDWriteEvent,
		})
	}
}

func testSocketConnectError(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockConnect(ctx, sock, bind)
		assertNotEqual(t, addr, nil)
		assertEqual(t, addr.Family(), bind.Family())
		assertEqual(t, errno, wasi.EINPROGRESS)

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(42, wasi.FDWriteEvent, wasi.SubscriptionFDReadWrite{
				FD: sock,
			}),
		}
		evs := make([]wasi.Event, len(subs))

		numEvents, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)

		assertEqual(t, evs[0], wasi.Event{
			UserData:  42,
			EventType: wasi.FDWriteEvent,
		})

		t.Run("the error is reported after polling", func(t *testing.T) {
			errno := sockErrno(t, ctx, sys, sock)
			assertEqual(t, errno, wasi.ECONNREFUSED)
		})

		t.Run("the error is cleared on the second read", func(t *testing.T) {
			errno := sockErrno(t, ctx, sys, sock)
			assertEqual(t, errno, wasi.ESUCCESS)
		})
	}
}

func testSocketConnectAndAccept(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		server, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		serverAddr, errno := sys.SockBind(ctx, server, bind)
		assertNotEqual(t, serverAddr, nil)
		assertEqual(t, serverAddr.Family(), bind.Family())
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, server, 10), wasi.ESUCCESS)

		client, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		clientAddr, errno := sys.SockConnect(ctx, client, serverAddr)
		if errno != wasi.ESUCCESS {
			assertEqual(t, errno, wasi.EINPROGRESS)
		}
		assertNotEqual(t, clientAddr, nil)
		assertEqual(t, clientAddr.Family(), bind.Family())

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(2, wasi.FDWriteEvent, wasi.SubscriptionFDReadWrite{
				FD: client,
			}),
		}
		evs := make([]wasi.Event, len(subs))

		numEvents, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  2,
			EventType: wasi.FDWriteEvent,
		})

		subs = []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(1, wasi.FDReadEvent, wasi.SubscriptionFDReadWrite{
				FD: server,
			}),
		}
		numEvents, errno = sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  1,
			EventType: wasi.FDReadEvent,
		})

		accept, remoteAddr, localAddr, errno := sys.SockAccept(ctx, server, wasi.NonBlock)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertNotEqual(t, accept, ^wasi.FD(0))
		assertDeepEqual(t, localAddr, serverAddr)
		assertDeepEqual(t, remoteAddr, clientAddr)
		assertEqual(t, sockIsNonBlocking(t, ctx, sys, accept), true)

		localAddr, errno = sys.SockLocalAddress(ctx, accept)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertDeepEqual(t, localAddr, serverAddr)

		remoteAddr, errno = sys.SockRemoteAddress(ctx, accept)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertDeepEqual(t, remoteAddr, clientAddr)

		assertEqual(t, sys.FDClose(ctx, accept), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, client), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, server), wasi.ESUCCESS)
	}
}

func testSocketConnectAndShutdown(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		server, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, server, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, server, 10), wasi.ESUCCESS)

		client, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, client, addr)
		if errno != wasi.ESUCCESS {
			assertEqual(t, errno, wasi.EINPROGRESS)
		}

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(1, wasi.FDWriteEvent, wasi.SubscriptionFDReadWrite{
				FD: client,
			}),
		}
		evs := make([]wasi.Event, len(subs))

		numEvents, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  1,
			EventType: wasi.FDWriteEvent,
		})

		subs = []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(1, wasi.FDReadEvent, wasi.SubscriptionFDReadWrite{
				FD: server,
			}),
		}
		numEvents, errno = sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  1,
			EventType: wasi.FDReadEvent,
		})

		accept, _, _, errno := sys.SockAccept(ctx, server, wasi.NonBlock)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sockIsNonBlocking(t, ctx, sys, accept), true)
		assertEqual(t, sys.SockShutdown(ctx, accept, wasi.ShutdownWR), wasi.ESUCCESS)

		subs = []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(1, wasi.FDReadEvent, wasi.SubscriptionFDReadWrite{
				FD: client,
			}),
		}
		numEvents, errno = sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  1,
			EventType: wasi.FDReadEvent,
		})
		assertEqual(t, sys.SockShutdown(ctx, client, wasi.ShutdownWR), wasi.ESUCCESS)

		// Darwin and Linux disagree on when to return ENOTCONN on shutdown(2);
		// on Darwin, the error is returned for read and write directions
		// independently, while on Linux, the error is only returned after
		// shutting down both read and write directions. We have not way of
		// managing this so we only test the Linux behavior which is less strict
		// than Darwin, and expect ENOTCONN only after both the read and write
		// ends of the socket have been shut down.
		assertEqual(t, sys.SockShutdown(ctx, client, wasi.ShutdownRD), wasi.ENOTCONN)
		assertEqual(t, sys.SockShutdown(ctx, client, wasi.ShutdownWR), wasi.ENOTCONN)

		subs = []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(1, wasi.FDReadEvent, wasi.SubscriptionFDReadWrite{
				FD: accept,
			}),
		}
		numEvents, errno = sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  1,
			EventType: wasi.FDReadEvent,
		})

		assertEqual(t, sockErrno(t, ctx, sys, client), wasi.ESUCCESS)
		assertEqual(t, sockErrno(t, ctx, sys, accept), wasi.ESUCCESS)

		assertEqual(t, sys.FDClose(ctx, accept), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, client), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, server), wasi.ESUCCESS)
	}
}

func testSocketBindAfterBind(family wasi.ProtocolFamily, typ wasi.SocketType, bind1, bind2 wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockBind(ctx, sock, bind1)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockBind(ctx, sock, bind2)
		assertEqual(t, errno, wasi.EINVAL)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketBindAfterConnect(family wasi.ProtocolFamily, typ wasi.SocketType, bind1, bind2 wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, sock, bind1)
		if errno != wasi.ESUCCESS {
			assertEqual(t, errno, wasi.EINPROGRESS)
		}

		_, errno = sys.SockBind(ctx, sock, bind2)
		assertEqual(t, errno, wasi.EINVAL)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketConnectAfterListen(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 10), wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, sock, bind)
		assertEqual(t, errno, wasi.EISCONN)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketConnectAfterConnect(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 0), wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, conn, addr)
		assertEqual(t, errno, wasi.EINPROGRESS)

		// The second call to connect(2) may race since the connection is done
		// asynchronously, so we have to tolerate ESUCCESS but also want to make
		// sure that the only other possible error is EALREADY.
		_, errno = sys.SockConnect(ctx, conn, addr)
		if errno != wasi.ESUCCESS {
			assertEqual(t, errno, wasi.EALREADY)
		}

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
	}
}

func testSocketListenBeforeBind(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 10), wasi.ESUCCESS)

		addr, errno := sys.SockLocalAddress(ctx, sock)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertNotEqual(t, addr, nil)

		switch a := addr.(type) {
		case *wasi.Inet4Address:
			var zero [4]byte
			assertEqual(t, a.Addr, zero)
			assertNotEqual(t, a.Port, 0)
		case *wasi.Inet6Address:
			var zero [16]byte
			assertEqual(t, a.Addr, zero)
			assertNotEqual(t, a.Port, 0)
		default:
			t.Errorf("invalid socket address type: %T", a)
		}

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketListenAfterListen(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 0), wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 1), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketListenNegativeBacklog(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, -1), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketListenLargeBacklog(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, math.MaxInt32), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketShutdownInvalidArgument(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockShutdown(ctx, sock, ^(wasi.ShutdownRD|wasi.ShutdownWR)), wasi.EINVAL)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketShutdownBeforeConnect(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockShutdown(ctx, sock, wasi.ShutdownRD|wasi.ShutdownWR), wasi.ENOTCONN)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketShutdownAfterListen(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 0), wasi.ESUCCESS)

		assertEqual(t, sys.SockShutdown(ctx, sock, wasi.ShutdownRD|wasi.ShutdownWR), wasi.ENOTCONN)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketDefaultBufferSizes(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		tests := []struct {
			scenario string
			option   wasi.SocketOption
		}{
			{scenario: "recv buffer size", option: wasi.RecvBufferSize},
			{scenario: "send buffer size", option: wasi.SendBufferSize},
		}

		for _, test := range tests {
			t.Run(test.scenario, func(t *testing.T) {
				bufferSize := sockOption[wasi.IntValue](t, ctx, sys, sock, wasi.SocketLevel, test.option)
				assertNotEqual(t, bufferSize, 0)
			})
		}

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSetBufferSizes(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		tests := []struct {
			scenario string
			option   wasi.SocketOption
		}{
			{scenario: "recv buffer size", option: wasi.RecvBufferSize},
			{scenario: "send buffer size", option: wasi.SendBufferSize},
		}

		for _, test := range tests {
			t.Run(test.scenario, func(t *testing.T) {
				sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
				assertEqual(t, errno, wasi.ESUCCESS)

				defaultBufferSize := sockOption[wasi.IntValue](t, ctx, sys, sock, wasi.SocketLevel, test.option)
				assertNotEqual(t, defaultBufferSize, 0)

				setBufferSize := func(size wasi.IntValue) {
					t.Helper()
					assertEqual(t, sys.SockSetOpt(ctx, sock, wasi.SocketLevel, test.option, size), wasi.ESUCCESS)
				}

				getBufferSize := func() wasi.IntValue {
					t.Helper()
					return sockOption[wasi.IntValue](t, ctx, sys, sock, wasi.SocketLevel, test.option)
				}

				t.Run("grow the socket buffer size", func(t *testing.T) {
					want := 2 * defaultBufferSize
					setBufferSize(want)
					size := getBufferSize()
					assertEqual(t, size, want)
				})

				t.Run("shrink the socket buffer size", func(t *testing.T) {
					want := defaultBufferSize / 2
					setBufferSize(want)
					size := getBufferSize()
					assertEqual(t, size, want)
				})

				t.Run("negative socket buffer size are fobidden", func(t *testing.T) {
					want := getBufferSize()
					assertEqual(t, sys.SockSetOpt(ctx, sock, wasi.SocketLevel, test.option, wasi.IntValue(-1)), wasi.EINVAL)
					size := getBufferSize()
					assertEqual(t, size, want)
				})

				t.Run("small socket buffer sizes are capped to a minimum value", func(t *testing.T) {
					assertEqual(t, sys.SockSetOpt(ctx, sock, wasi.SocketLevel, test.option, wasi.IntValue(0)), wasi.ESUCCESS)
					size := getBufferSize()
					assertNotEqual(t, size, 0)
				})

				t.Run("large socket buffer sizes are capped to a maximum value", func(t *testing.T) {
					assertEqual(t, sys.SockSetOpt(ctx, sock, wasi.SocketLevel, test.option, wasi.IntValue(math.MaxInt32)), wasi.ESUCCESS)
					size := getBufferSize()
					assertNotEqual(t, size, math.MaxInt32)
				})

				assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
			})
		}
	}
}

func testSocketSetOptionInvalidLevel(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		const level = -1
		const option = 0
		assertEqual(t, sys.SockSetOpt(ctx, sock, level, option, wasi.IntValue(0)), wasi.EINVAL)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSetOptionInvalidArgument(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		const option = -1
		assertEqual(t, sys.SockSetOpt(ctx, sock, wasi.SocketLevel, option, wasi.IntValue(0)), wasi.EINVAL)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func sockOpen(t *testing.T, ctx context.Context, sys wasi.System, family wasi.ProtocolFamily, typ wasi.SocketType, proto wasi.Protocol) (wasi.FD, wasi.Errno) {
	t.Helper()
	sock, errno := sys.SockOpen(ctx, family, typ, proto, wasi.AllRights, wasi.AllRights)
	skipIfNotImplemented(t, errno)
	if errno == wasi.ESUCCESS {
		assertEqual(t, sys.FDStatSetFlags(ctx, sock, wasi.NonBlock), wasi.ESUCCESS)
		assertEqual(t, sockIsNonBlocking(t, ctx, sys, sock), true)
		assertEqual(t, sockErrno(t, ctx, sys, sock), wasi.ESUCCESS)
	}
	return sock, errno
}

func sockOption[T wasi.SocketOptionValue](t *testing.T, ctx context.Context, sys wasi.System, sock wasi.FD, level wasi.SocketOptionLevel, option wasi.SocketOption) T {
	t.Helper()
	opt, errno := sys.SockGetOpt(ctx, sock, level, option)
	assertEqual(t, errno, wasi.ESUCCESS)
	val, ok := opt.(T)
	assertEqual(t, ok, true)
	return val
}

func sockErrno(t *testing.T, ctx context.Context, sys wasi.System, sock wasi.FD) wasi.Errno {
	t.Helper()
	return wasi.Errno(sockOption[wasi.IntValue](t, ctx, sys, sock, wasi.SocketLevel, wasi.QuerySocketError))
}

func sockIsNonBlocking(t *testing.T, ctx context.Context, sys wasi.System, sock wasi.FD) bool {
	t.Helper()
	stat, errno := sys.FDStatGet(ctx, sock)
	assertEqual(t, errno, wasi.ESUCCESS)
	return stat.Flags.Has(wasi.NonBlock)
}
