package wasitest

import (
	"bytes"
	"context"
	"math"
	"testing"
	"time"

	"github.com/stealthrocket/wasi-go"
)

var (
	localIPv4 = [4]byte{127, 0, 0, 1}
	localIPv6 = [16]byte{15: 1}

	unknownIPv4 = [4]byte{123, 234, 123, 234}
	unknownIPv6 = [16]byte{15: 2}

	randomPort = 20000
)

func nextPort() int {
	randomPort++
	return randomPort
}

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

	"unconnected ipv4 stream sockets are not ready for reading or writing": testSocketPollBeforeConnectStream(wasi.InetFamily),

	"unconnected ipv6 stream sockets are not ready for reading or writing": testSocketPollBeforeConnectStream(wasi.Inet6Family),

	"unconnected ipv4 datagram sockets are ready for writing but not for reading": testSocketPollBeforeConnectDatagram(wasi.InetFamily),

	"unconnected ipv6 datagram sockets are ready for writing but not for reading": testSocketPollBeforeConnectDatagram(wasi.Inet6Family),

	"bind an ipv4 stream socket to a port selects that port": testSocketBindOK(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"bind an ipv4 datagram socket to a port selects that port": testSocketBindOK(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"bind an ipv6 stream socket to a port selects that port": testSocketBindOK(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"bind an ipv6 datagram socket to a port selects that port": testSocketBindOK(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
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
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
		&wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot bind an ipv6 datagram socket that was already connected": testSocketBindAfterConnect(
		wasi.Inet6Family, wasi.DatagramSocket,
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
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

	"can accpet on ipv4 stream sockets in blocking mode": testSocketConnectAndAcceptBlocking(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"can accept on ipv6 stream sockets in blocking mode": testSocketConnectAndAcceptBlocking(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"can connect a ipv4 datagram socket": testSocketConnectOK(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"can connect a ipv6 datagram socket": testSocketConnectOK(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"failing to connect sets the socket error and getting the socket error clears it on ipv4 stream sockets": testSocketConnectError(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"failing to connect sets the socket error and getting the socket error clears it on ipv6 stream sockets": testSocketConnectError(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"cannot connect a listening ipv4 stream socket": testSocketConnectAfterListen(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot connect a listening ipv6 stream socket": testSocketConnectAfterListen(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"cannot connect a connected ipv4 stream socket": testSocketReconnectStream(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot connect a connected ipv6 stream socket": testSocketReconnectStream(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"can reconnect an ipv4 datagram socket": testSocketReconnectDatagram(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"can reconnect an ipv6 datagram socket": testSocketReconnectDatagram(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"cannot connect to a connected ipv4 stream socket": testSocketConnectToConnected(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot connect to a connected ipv6 stream socket": testSocketConnectToConnected(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"cannot connect an ipv4 stream socket to an address of the wrong family": testSocketConnectWrongFamily(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"cannot connect an ipv6 stream socket to an address of the wrong family": testSocketConnectWrongFamily(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot connect an ipv4 datagram socket to an address of the wrong family": testSocketConnectWrongFamily(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"cannot connect an ipv6 datagram socket to an address of the wrong family": testSocketConnectWrongFamily(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot listen on a connected ipv4 stream socket": testSocketListenAfterConnect(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot listen on a connected ipv6 stream socket": testSocketListenAfterConnect(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
	),

	"cannot listen on ipv4 datagram sockets": testSocketListenDatagram(
		wasi.InetFamily, wasi.DatagramSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"can listen on ipv6 datagram sockets": testSocketListenDatagram(
		wasi.Inet6Family, wasi.DatagramSocket, &wasi.Inet6Address{Addr: localIPv6},
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

	"cannot accept on an ipv4 stream socket which is not listening": testSocketAcceptBeforeListen(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"cannot accept on an ipv6 stream socket which is not listening": testSocketAcceptBeforeListen(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot accept on an ipv4 datagram socket": testSocketAcceptDatagram(
		wasi.InetFamily,
	),

	"cannot accept on an ipv6 datagram socket": testSocketAcceptDatagram(
		wasi.Inet6Family,
	),

	"cannot accept on a connected ipv4 stream socket": testSocketAcceptAfterConnect(
		wasi.InetFamily, wasi.StreamSocket, &wasi.Inet4Address{Addr: localIPv4},
	),

	"cannot accept on a connected ipv6 stream socket": testSocketAcceptAfterConnect(
		wasi.Inet6Family, wasi.StreamSocket, &wasi.Inet6Address{Addr: localIPv6},
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

	"the default buffer sizes are not zero on ipv4 datagram sockets": testSocketDefaultBufferSizes(
		wasi.InetFamily, wasi.DatagramSocket,
	),

	"the default buffer sizes are not zero on ipv6 datagram sockets": testSocketDefaultBufferSizes(
		wasi.Inet6Family, wasi.DatagramSocket,
	),

	"cannot set option of ipv4 stream socket with invalid level": testSocketSetOptionInvalidLevel(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"cannot set option of ipv6 stream socket with invalid level": testSocketSetOptionInvalidLevel(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot set option of ipv4 datagram socket with invalid level": testSocketSetOptionInvalidLevel(
		wasi.InetFamily, wasi.DatagramSocket,
	),

	"cannot set option of ipv6 datagram socket with invalid level": testSocketSetOptionInvalidLevel(
		wasi.Inet6Family, wasi.DatagramSocket,
	),

	"cannot set option of ipv4 stream socket with invalid argument": testSocketSetOptionInvalidArgument(
		wasi.InetFamily, wasi.StreamSocket,
	),

	"cannot set option of ipv6 stream socket with invalid argument": testSocketSetOptionInvalidArgument(
		wasi.Inet6Family, wasi.StreamSocket,
	),

	"cannot set option of ipv4 datagram socket with invalid argument": testSocketSetOptionInvalidArgument(
		wasi.InetFamily, wasi.DatagramSocket,
	),

	"cannot set option of ipv6 datagram socket with invalid argument": testSocketSetOptionInvalidArgument(
		wasi.Inet6Family, wasi.DatagramSocket,
	),

	"connected ipv4 stream sockets can send and receive data": testSocketSendAndReceiveStream(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"connected ipv6 stream sockets can send and receive data": testSocketSendAndReceiveStream(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"connected ipv4 stream sockets can send and peek data": testSocketSendAndPeekStream(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"connected ipv6 stream sockets can send and peek data": testSocketSendAndPeekStream(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"connected ipv4 stream sockets can send and receive data in blocking mode": testSocketSendAndReceiveStreamBlocking(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"connected ipv6 stream sockets can send and receive data in blocking mode": testSocketSendAndReceiveStreamBlocking(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"timeout unblocks ipv4 stream sockets waiting for data in blocking mode": testSocketTimeoutStreamBlocking(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"timeout unblocks ipv6 stream sockets waiting for data in blocking mode": testSocketTimeoutStreamBlocking(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"timeout unblocks ipv4 datagram sockets waiting for data in blocking mode": testSocketTimeoutDatagramBlocking(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"timeout unblocks ipv6 datagram sockets waiting for data in blocking mode": testSocketTimeoutDatagramBlocking(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"connected ipv4 datagram sockets can send and receive data": testSocketSendAndReceiveConnectedDatagram(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"connected ipv6 datagram sockets can send and receive data": testSocketSendAndReceiveConnectedDatagram(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"connected ipv4 datagram sockets cannot send data to a specific address": testSocketSendToConnectedDatagram(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"connected ipv6 datagram sockets cannot send data to a specific address": testSocketSendToConnectedDatagram(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"connected ipv4 datagram sockets can send and peek data": testSocketSendAndPeekConnectedDatagram(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"connected ipv6 datagram sockets can send and peek data": testSocketSendAndPeekConnectedDatagram(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"connected ipv4 datagram sockets can send and receive data in blocking mode": testSocketSendAndReceiveConnectedDatagramBlocking(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4},
	),

	"connected ipv6 datagram sockets can send and receive data in blocking mode": testSocketSendAndReceiveConnectedDatagramBlocking(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6},
	),

	"unconnected ipv4 datagram sockets can send and receive data": testSocketSendAndReceiveNotConnectedDatagram(
		wasi.InetFamily,
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"unconnected ipv6 datagram sockets can send and receive data": testSocketSendAndReceiveNotConnectedDatagram(
		wasi.Inet6Family,
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"unconnected ipv4 datagram sockets can send and receive data in blocking mode": testSocketSendAndReceiveNotConnectedDatagramBlocking(
		wasi.InetFamily,
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"unconnected ipv6 datagram sockets can send and receive data in blocking mode": testSocketSendAndReceiveNotConnectedDatagramBlocking(
		wasi.Inet6Family,
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"large messages are truncated when sent on ipv4 datagram sockets": testSocketSendAndReceiveTruncatedDatagram(
		wasi.InetFamily,
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"large messages are truncated when sent on ipv6 datagram sockets": testSocketSendAndReceiveTruncatedDatagram(
		wasi.Inet6Family,
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"can send messages to unbound addresses on a ipv4 datagram socket": testSocketSendDatagramToNowhere(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"can send messages to unbound addresses on a ipv6 datagram socket": testSocketSendDatagramToNowhere(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"cannot bind an ipv4 datagram socket after sending a message": testSocketBindAfterSendDatagram(
		wasi.InetFamily, &wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"cannot bind an ipv6 datagram socket after sending a message": testSocketBindAfterSendDatagram(
		wasi.Inet6Family, &wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"cannot receive on an unbound ipv4 datagram socket": testSocketReceiveBeforeBindDatagram(
		wasi.InetFamily,
	),

	"cannot receive on an unbound ipv6 datagram socket": testSocketReceiveBeforeBindDatagram(
		wasi.Inet6Family,
	),

	"drop messages larger than the ipv4 datagram socket receive buffer size": testSocketSendAndReceiveLargerThanRecvBufferSize(
		wasi.InetFamily,
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"drop messages larger than the ipv6 datagram socket receive buffer size": testSocketSendAndReceiveLargerThanRecvBufferSize(
		wasi.Inet6Family,
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
	),

	"cannot send a message larger than the ipv4 datagram socket send buffer size": testSocketSendAndReceiveLargerThanSendBufferSize(
		wasi.InetFamily,
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
		&wasi.Inet4Address{Addr: localIPv4, Port: nextPort()},
	),

	"cannot send a message larger than the ipv6 datagram socket send buffer size": testSocketSendAndReceiveLargerThanSendBufferSize(
		wasi.Inet6Family,
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
		&wasi.Inet6Address{Addr: localIPv6, Port: nextPort()},
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
			_, errno := sys.SockGetOpt(ctx, fd, wasi.QuerySocketType)
			return errno
		},
	),

	"cannot set socket options on a file descriptor which is not a socket": testNotSocket(
		func(ctx context.Context, sys wasi.System, fd wasi.FD) wasi.Errno {
			return sys.SockSetOpt(ctx, fd, wasi.SendBufferSize, wasi.IntValue(4096))
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

		opt, errno := sys.SockGetOpt(ctx, sock, wasi.QuerySocketType)
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

func testSocketPollBeforeConnectStream(family wasi.ProtocolFamily) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{
			Now: time.Now,
		})

		sock, errno := sockOpen(t, ctx, sys, family, wasi.StreamSocket, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionClock(
				wasi.UserData(1),
				wasi.SubscriptionClock{ID: wasi.Monotonic, Timeout: 0, Precision: 1},
			),
			wasi.MakeSubscriptionFDReadWrite(
				wasi.UserData(2),
				wasi.FDReadEvent,
				wasi.SubscriptionFDReadWrite{FD: sock},
			),
			wasi.MakeSubscriptionFDReadWrite(
				wasi.UserData(3),
				wasi.FDWriteEvent,
				wasi.SubscriptionFDReadWrite{FD: sock},
			),
		}

		evs := make([]wasi.Event, len(subs))

		n, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, errno, wasi.ESUCCESS)
		switch n {
		default:
			t.Fatalf("wrong number of events: want 1 or 3 but got %d", n)
		case 1:
			// Darwin reports that sockets are not ready for read/write before
			// being connected.
			assertEqual(t, evs[0], wasi.Event{
				UserData:  1,
				EventType: wasi.ClockEvent,
			})
		case 3:
			// Linux reports that sockets are ready for read/write before being
			// connected.
			assertEqual(t, evs[0], wasi.Event{
				UserData:  1,
				EventType: wasi.ClockEvent,
			})
			assertEqual(t, evs[1], wasi.Event{
				UserData:  2,
				EventType: wasi.FDReadEvent,
			})
			assertEqual(t, evs[2], wasi.Event{
				UserData:  3,
				EventType: wasi.FDWriteEvent,
			})
		}
	}
}

func testSocketPollBeforeConnectDatagram(family wasi.ProtocolFamily) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{
			Now: time.Now,
		})

		sock, errno := sockOpen(t, ctx, sys, family, wasi.DatagramSocket, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionClock(
				wasi.UserData(1),
				wasi.SubscriptionClock{ID: wasi.Monotonic, Timeout: 0, Precision: 1},
			),
			wasi.MakeSubscriptionFDReadWrite(
				wasi.UserData(2),
				wasi.FDReadEvent,
				wasi.SubscriptionFDReadWrite{FD: sock},
			),
			wasi.MakeSubscriptionFDReadWrite(
				wasi.UserData(3),
				wasi.FDWriteEvent,
				wasi.SubscriptionFDReadWrite{FD: sock},
			),
		}

		evs := make([]wasi.Event, len(subs))

		n, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, n, 2)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  1,
			EventType: wasi.ClockEvent,
		})
		assertEqual(t, evs[1], wasi.Event{
			UserData:  3,
			EventType: wasi.FDWriteEvent,
		})
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

func testSocketListenDatagram(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 10), wasi.ENOTSUP)
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
		if typ == wasi.StreamSocket {
			assertEqual(t, errno, wasi.EINPROGRESS)
		} else {
			assertEqual(t, errno, wasi.ESUCCESS)
		}

		sockPoll(t, ctx, sys, sock, wasi.FDWriteEvent)
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

		sockPoll(t, ctx, sys, sock, wasi.FDWriteEvent)

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
		if typ == wasi.StreamSocket {
			assertEqual(t, errno, wasi.EINPROGRESS)
		} else {
			assertEqual(t, errno, wasi.ESUCCESS)
		}
		assertNotEqual(t, clientAddr, nil)
		assertEqual(t, clientAddr.Family(), bind.Family())

		sockPoll(t, ctx, sys, client, wasi.FDWriteEvent)
		sockPoll(t, ctx, sys, server, wasi.FDReadEvent)

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

func testSocketConnectAndAcceptBlocking(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		server, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, server, false)

		serverAddr, errno := sys.SockBind(ctx, server, bind)
		assertNotEqual(t, serverAddr, nil)
		assertEqual(t, serverAddr.Family(), bind.Family())
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, server, 10), wasi.ESUCCESS)

		client, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, client, false)

		clientAddr, errno := sys.SockConnect(ctx, client, serverAddr)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertNotEqual(t, clientAddr, nil)
		assertEqual(t, clientAddr.Family(), bind.Family())

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
		if typ == wasi.StreamSocket {
			assertEqual(t, errno, wasi.EINPROGRESS)
		} else {
			assertEqual(t, errno, wasi.ESUCCESS)
		}

		sockPoll(t, ctx, sys, client, wasi.FDWriteEvent)
		sockPoll(t, ctx, sys, server, wasi.FDReadEvent)

		accept, _, _, errno := sys.SockAccept(ctx, server, wasi.NonBlock)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sockIsNonBlocking(t, ctx, sys, accept), true)
		assertEqual(t, sys.SockShutdown(ctx, accept, wasi.ShutdownWR), wasi.ESUCCESS)

		sockPoll(t, ctx, sys, client, wasi.FDReadEvent)

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

		sockPoll(t, ctx, sys, accept, wasi.FDReadEvent)

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
		if typ == wasi.StreamSocket {
			assertEqual(t, errno, wasi.EINPROGRESS)
		} else {
			assertEqual(t, errno, wasi.ESUCCESS)
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

func testSocketReconnectStream(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.StreamSocket

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
		switch errno {
		case wasi.EALREADY:
		case wasi.EISCONN:
		case wasi.ESUCCESS:
		default:
			t.Errorf("invalid error code returned on second call to connect: %s", errno)
		}

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
	}
}

func testSocketReconnectDatagram(family wasi.ProtocolFamily, addr wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, sock, addr)
		assertEqual(t, errno, wasi.ESUCCESS)

		switch a := addr.(type) {
		case *wasi.Inet4Address:
			a.Port = nextPort()
		case *wasi.Inet6Address:
			a.Port = nextPort()
		}

		_, errno = sys.SockConnect(ctx, sock, addr)
		assertEqual(t, errno, wasi.ESUCCESS)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketConnectToConnected(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 0), wasi.ESUCCESS)

		conn1, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr1, errno := sys.SockConnect(ctx, conn1, addr)
		assertEqual(t, errno, wasi.EINPROGRESS)
		assertNotEqual(t, addr1, nil)

		conn2, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr2, errno := sys.SockConnect(ctx, conn2, addr1)
		assertEqual(t, errno, wasi.EINPROGRESS)
		assertNotEqual(t, addr2, nil)

		sockPoll(t, ctx, sys, conn2, wasi.FDWriteEvent)

		assertEqual(t, sockErrno(t, ctx, sys, conn2), wasi.ECONNREFUSED)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, conn1), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, conn2), wasi.ESUCCESS)
	}
}

func testSocketConnectWrongFamily(family wasi.ProtocolFamily, typ wasi.SocketType, addr wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, sock, addr)
		assertEqual(t, errno, wasi.EAFNOSUPPORT)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
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

func testSocketListenAfterConnect(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
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
		assertEqual(t, sys.SockListen(ctx, conn, 0), wasi.EINVAL)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
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

func testSocketAcceptBeforeListen(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, peer, addr, errno := sys.SockAccept(ctx, sock, wasi.NonBlock)
		assertEqual(t, conn, ^wasi.FD(0))
		assertEqual(t, peer, nil)
		assertEqual(t, addr, nil)
		assertEqual(t, errno, wasi.EINVAL)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketAcceptDatagram(family wasi.ProtocolFamily) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, peer, addr, errno := sys.SockAccept(ctx, sock, wasi.NonBlock)
		assertEqual(t, conn, ^wasi.FD(0))
		assertEqual(t, peer, nil)
		assertEqual(t, addr, nil)
		assertEqual(t, errno, wasi.ENOTSUP)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketAcceptAfterConnect(family wasi.ProtocolFamily, typ wasi.SocketType, bind wasi.SocketAddress) testFunc {
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

		fd, peer, addr, errno := sys.SockAccept(ctx, conn, wasi.NonBlock)
		assertEqual(t, fd, ^wasi.FD(0))
		assertEqual(t, peer, nil)
		assertEqual(t, addr, nil)
		assertEqual(t, errno, wasi.EINVAL)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
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

func testSocketSendAndReceiveStream(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.StreamSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 10), wasi.ESUCCESS)

		conn1, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, conn1, addr)
		assertEqual(t, errno, wasi.EINPROGRESS)

		sockPoll(t, ctx, sys, conn1, wasi.FDWriteEvent)
		sockPoll(t, ctx, sys, sock, wasi.FDReadEvent)

		conn2, _, _, errno := sys.SockAccept(ctx, sock, wasi.NonBlock)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 32)

		size1, errno := sys.FDWrite(ctx, conn1, []wasi.IOVec{buffer1})
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn2, wasi.FDReadEvent)
		size2, errno := sys.FDRead(ctx, conn2, []wasi.IOVec{buffer2})
		assertEqual(t, size2, size1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))

		buffer1 = []byte("How are you?")
		size3, errno := sys.FDWrite(ctx, conn2, []wasi.IOVec{buffer1})
		assertEqual(t, size3, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn1, wasi.FDReadEvent)
		size4, errno := sys.FDRead(ctx, conn1, []wasi.IOVec{buffer2})
		assertEqual(t, size4, size3)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))

		assertEqual(t, sys.FDClose(ctx, conn2), wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn1, wasi.FDReadEvent)
		size5, errno := sys.FDRead(ctx, conn1, []wasi.IOVec{buffer2})
		assertEqual(t, size5, 0) // EOF
		assertEqual(t, errno, wasi.ESUCCESS)

		assertEqual(t, sys.FDClose(ctx, conn1), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndReceiveStreamBlocking(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.StreamSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, sock, false)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 10), wasi.ESUCCESS)

		conn1, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, conn1, false)

		_, errno = sys.SockConnect(ctx, conn1, addr)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn2, _, _, errno := sys.SockAccept(ctx, sock, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 32)

		size1, errno := sys.FDWrite(ctx, conn1, []wasi.IOVec{buffer1})
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn2, wasi.FDReadEvent)
		size2, errno := sys.FDRead(ctx, conn2, []wasi.IOVec{buffer2})
		assertEqual(t, size2, size1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))

		buffer1 = []byte("How are you?")
		size3, errno := sys.FDWrite(ctx, conn2, []wasi.IOVec{buffer1})
		assertEqual(t, size3, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn1, wasi.FDReadEvent)
		size4, errno := sys.FDRead(ctx, conn1, []wasi.IOVec{buffer2})
		assertEqual(t, size4, size3)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))

		assertEqual(t, sys.FDClose(ctx, conn2), wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn1, wasi.FDReadEvent)
		size5, errno := sys.FDRead(ctx, conn1, []wasi.IOVec{buffer2})
		assertEqual(t, size5, 0) // EOF
		assertEqual(t, errno, wasi.ESUCCESS)

		assertEqual(t, sys.FDClose(ctx, conn1), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndPeekStream(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.StreamSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 10), wasi.ESUCCESS)

		conn1, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, conn1, addr)
		assertEqual(t, errno, wasi.EINPROGRESS)

		sockPoll(t, ctx, sys, conn1, wasi.FDWriteEvent)
		sockPoll(t, ctx, sys, sock, wasi.FDReadEvent)

		conn2, _, _, errno := sys.SockAccept(ctx, sock, wasi.NonBlock)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 32)

		size1, errno := sys.FDWrite(ctx, conn1, []wasi.IOVec{buffer1})
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn2, wasi.FDReadEvent)

		for i := range buffer1 {
			size2, flags, errno := sys.SockRecv(ctx, conn2, []wasi.IOVec{buffer2[:i+1]}, wasi.RecvPeek)
			assertEqual(t, size2, wasi.Size(i+1))
			assertEqual(t, flags, wasi.ROFlags(0))
			assertEqual(t, errno, wasi.ESUCCESS)
			assertEqual(t, string(buffer1[:size2]), string(buffer2[:size2]))
		}

		size3, flags, errno := sys.SockRecv(ctx, conn2, []wasi.IOVec{buffer2[:7]}, 0)
		assertEqual(t, size3, wasi.Size(7))
		assertEqual(t, flags, wasi.ROFlags(0))
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer1[:size3]), string(buffer2[:size3]))

		for i := range buffer1[size3:] {
			size4, flags, errno := sys.SockRecv(ctx, conn2, []wasi.IOVec{buffer2[:i+1]}, wasi.RecvPeek)
			assertEqual(t, size4, wasi.Size(i+1))
			assertEqual(t, flags, wasi.ROFlags(0))
			assertEqual(t, errno, wasi.ESUCCESS)
			assertEqual(t, string(buffer1[size3:size3+size4]), string(buffer2[:size4]))
		}

		assertEqual(t, sys.FDClose(ctx, conn1), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketTimeoutStreamBlocking(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.StreamSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, sock, false)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, sys.SockListen(ctx, sock, 10), wasi.ESUCCESS)

		conn1, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, conn1, false)

		_, errno = sys.SockConnect(ctx, conn1, addr)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn2, _, _, errno := sys.SockAccept(ctx, sock, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		const recvTimeout = 20 * time.Millisecond
		const sendTimeout = 40 * time.Millisecond

		errno = sys.SockSetOpt(ctx, conn1,
			wasi.RecvTimeout,
			wasi.TimeValue(recvTimeout))
		assertEqual(t, errno, wasi.ESUCCESS)
		errno = sys.SockSetOpt(ctx, conn1,
			wasi.SendTimeout,
			wasi.TimeValue(sendTimeout),
		)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockRecvTimeout := sockOption[wasi.TimeValue](t, ctx, sys, conn1, wasi.RecvTimeout)
		assertEqual(t, sockRecvTimeout, wasi.TimeValue(recvTimeout))
		sockSendTimeout := sockOption[wasi.TimeValue](t, ctx, sys, conn1, wasi.SendTimeout)
		assertEqual(t, sockSendTimeout, wasi.TimeValue(sendTimeout))

		buffer := make([]byte, 10)
		start := time.Now()

		n, _, errno := sys.SockRecv(ctx, conn1, []wasi.IOVec{buffer}, 0)
		assertEqual(t, n, ^wasi.Size(0))
		assertEqual(t, errno, wasi.EAGAIN)

		delay := time.Since(start)
		assertEqual(t, delay >= recvTimeout && delay < sendTimeout, true)

		assertEqual(t, sys.FDClose(ctx, conn2), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, conn1), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketTimeoutDatagramBlocking(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, sock, false)

		const recvTimeout = 20 * time.Millisecond
		const sendTimeout = 40 * time.Millisecond

		errno = sys.SockSetOpt(ctx, sock,
			wasi.RecvTimeout,
			wasi.TimeValue(recvTimeout))
		assertEqual(t, errno, wasi.ESUCCESS)
		errno = sys.SockSetOpt(ctx, sock,
			wasi.SendTimeout,
			wasi.TimeValue(sendTimeout),
		)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockRecvTimeout := sockOption[wasi.TimeValue](t, ctx, sys, sock, wasi.RecvTimeout)
		assertEqual(t, sockRecvTimeout, wasi.TimeValue(recvTimeout))
		sockSendTimeout := sockOption[wasi.TimeValue](t, ctx, sys, sock, wasi.SendTimeout)
		assertEqual(t, sockSendTimeout, wasi.TimeValue(sendTimeout))

		buffer := make([]byte, 10)
		start := time.Now()

		n, _, errno := sys.SockRecv(ctx, sock, []wasi.IOVec{buffer}, 0)
		assertEqual(t, n, ^wasi.Size(0))
		assertEqual(t, errno, wasi.EAGAIN)

		delay := time.Since(start)
		assertEqual(t, delay >= recvTimeout && delay < sendTimeout, true)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndReceiveConnectedDatagram(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockAddr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		connAddr, errno := sys.SockConnect(ctx, conn, sockAddr)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 32)

		size1, errno := sys.SockSend(ctx, conn, []wasi.IOVec{buffer1}, 0)
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, sock, wasi.FDReadEvent)
		size2, roflags, raddr, errno := sys.SockRecvFrom(ctx, sock, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size2, size1)
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))
		assertDeepEqual(t, raddr, connAddr)

		buffer1 = []byte("How are you?")
		sockPoll(t, ctx, sys, sock, wasi.FDWriteEvent)
		size3, errno := sys.SockSendTo(ctx, sock, []wasi.IOVec{buffer1}, 0, connAddr)
		assertEqual(t, size3, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn, wasi.FDReadEvent)
		size4, roflags, raddr, errno := sys.SockRecvFrom(ctx, conn, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size4, size3)
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))
		assertDeepEqual(t, raddr, sockAddr)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendToConnectedDatagram(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockAddr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, conn, sockAddr)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer := []byte("Hello, World!")

		size1, errno := sys.SockSendTo(ctx, conn, []wasi.IOVec{buffer}, 0, sockAddr)
		assertEqual(t, size1, wasi.Size(0))
		assertEqual(t, errno, wasi.EISCONN)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndReceiveConnectedDatagramBlocking(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, sock, false)

		sockAddr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, conn, false)

		connAddr, errno := sys.SockConnect(ctx, conn, sockAddr)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 32)

		size1, errno := sys.SockSend(ctx, conn, []wasi.IOVec{buffer1}, 0)
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		size2, roflags, raddr, errno := sys.SockRecvFrom(ctx, sock, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size2, size1)
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))
		assertDeepEqual(t, raddr, connAddr)

		buffer1 = []byte("How are you?")
		size3, errno := sys.SockSendTo(ctx, sock, []wasi.IOVec{buffer1}, 0, connAddr)
		assertEqual(t, size3, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		size4, roflags, raddr, errno := sys.SockRecvFrom(ctx, conn, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size4, size3)
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))
		assertDeepEqual(t, raddr, sockAddr)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndReceiveNotConnectedDatagram(family wasi.ProtocolFamily, addr1, addr2 wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockAddr, errno := sys.SockBind(ctx, sock, addr1)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		connAddr, errno := sys.SockBind(ctx, conn, addr2)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 32)

		size1, errno := sys.SockSendTo(ctx, conn, []wasi.IOVec{buffer1}, 0, sockAddr)
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, sock, wasi.FDReadEvent)
		size2, roflags, raddr, errno := sys.SockRecvFrom(ctx, sock, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size2, size1)
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))
		assertDeepEqual(t, raddr, connAddr)

		buffer1 = []byte("How are you?")
		sockPoll(t, ctx, sys, sock, wasi.FDWriteEvent)
		size3, errno := sys.SockSendTo(ctx, sock, []wasi.IOVec{buffer1}, 0, connAddr)
		assertEqual(t, size3, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, conn, wasi.FDReadEvent)
		size4, roflags, raddr, errno := sys.SockRecvFrom(ctx, conn, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size4, size3)
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))
		assertDeepEqual(t, raddr, sockAddr)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndReceiveNotConnectedDatagramBlocking(family wasi.ProtocolFamily, addr1, addr2 wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, sock, false)

		sockAddr, errno := sys.SockBind(ctx, sock, addr1)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		setNonBlock(t, ctx, sys, conn, false)

		connAddr, errno := sys.SockBind(ctx, conn, addr2)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 32)

		size1, errno := sys.SockSendTo(ctx, conn, []wasi.IOVec{buffer1}, 0, sockAddr)
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		size2, roflags, raddr, errno := sys.SockRecvFrom(ctx, sock, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size2, size1)
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))
		assertDeepEqual(t, raddr, connAddr)

		buffer1 = []byte("How are you?")
		size3, errno := sys.SockSendTo(ctx, sock, []wasi.IOVec{buffer1}, 0, connAddr)
		assertEqual(t, size3, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		size4, roflags, raddr, errno := sys.SockRecvFrom(ctx, conn, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size4, size3)
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2[:len(buffer1)]), string(buffer1))
		assertDeepEqual(t, raddr, sockAddr)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndReceiveTruncatedDatagram(family wasi.ProtocolFamily, addr1, addr2 wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockAddr, errno := sys.SockBind(ctx, sock, addr1)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		connAddr, errno := sys.SockBind(ctx, conn, addr2)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 10)

		size1, errno := sys.SockSendTo(ctx, conn, []wasi.IOVec{buffer1}, 0, sockAddr)
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, sock, wasi.FDReadEvent)
		size2, roflags, raddr, errno := sys.SockRecvFrom(ctx, sock, []wasi.IOVec{buffer2}, 0)
		assertEqual(t, size2, wasi.Size(len(buffer2)))
		assertEqual(t, roflags, wasi.RecvDataTruncated)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer2), string(buffer1[:len(buffer2)]))
		assertDeepEqual(t, raddr, connAddr)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndPeekConnectedDatagram(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		_, errno = sys.SockConnect(ctx, conn, addr)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer1 := []byte("Hello, World!")
		buffer2 := make([]byte, 32)

		size1, errno := sys.FDWrite(ctx, conn, []wasi.IOVec{buffer1})
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, sock, wasi.FDReadEvent)

		for i := range buffer1 {
			n := i + 1
			wantFlags := wasi.ROFlags(0)
			if n < len(buffer1) {
				wantFlags = wasi.RecvDataTruncated
			}
			size2, flags, errno := sys.SockRecv(ctx, sock, []wasi.IOVec{buffer2[:n]}, wasi.RecvPeek)
			assertEqual(t, size2, wasi.Size(n))
			assertEqual(t, flags, wantFlags)
			assertEqual(t, errno, wasi.ESUCCESS)
			assertEqual(t, string(buffer1[:n]), string(buffer2[:n]))
		}

		size3, flags, errno := sys.SockRecv(ctx, sock, []wasi.IOVec{buffer2[:7]}, 0)
		assertEqual(t, size3, wasi.Size(7))
		assertEqual(t, flags, wasi.RecvDataTruncated)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, string(buffer1[:size3]), string(buffer2[:size3]))

		size4, flags, errno := sys.SockRecv(ctx, sock, []wasi.IOVec{buffer2}, wasi.RecvPeek)
		assertEqual(t, size4, ^wasi.Size(0))
		assertEqual(t, flags, wasi.ROFlags(0))
		assertEqual(t, errno, wasi.EAGAIN)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendDatagramToNowhere(family wasi.ProtocolFamily, addr wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket
		msg := []byte("Hello, World!")

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		size, errno := sys.SockSendTo(ctx, sock, []wasi.IOVec{msg}, 0, addr)
		assertEqual(t, size, wasi.Size(len(msg)))
		assertEqual(t, errno, wasi.ESUCCESS)

		laddr, errno := sys.SockLocalAddress(ctx, sock)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, laddr.Family(), addr.Family())

		switch a := laddr.(type) {
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

func testSocketBindAfterSendDatagram(family wasi.ProtocolFamily, bind wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket
		msg := []byte("Hello, World!")

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		size, errno := sys.SockSendTo(ctx, sock, []wasi.IOVec{msg}, 0, bind)
		assertEqual(t, size, wasi.Size(len(msg)))
		assertEqual(t, errno, wasi.ESUCCESS)

		addr, errno := sys.SockBind(ctx, sock, bind)
		assertEqual(t, errno, wasi.EINVAL)
		assertDeepEqual(t, addr, nil)

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketReceiveBeforeBindDatagram(family wasi.ProtocolFamily) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket
		buf := make([]byte, 32)

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		size, roflags, raddr, errno := sys.SockRecvFrom(ctx, sock, []wasi.IOVec{buf}, 0)
		assertEqual(t, size, ^wasi.Size(0))
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.EAGAIN)
		assertDeepEqual(t, raddr, nil)

		laddr, errno := sys.SockLocalAddress(ctx, sock)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, laddr.Family(), family)

		switch a := laddr.(type) {
		case *wasi.Inet4Address:
			var zero [4]byte
			assertEqual(t, a.Addr, zero)
			assertEqual(t, a.Port, 0)
		case *wasi.Inet6Address:
			var zero [16]byte
			assertEqual(t, a.Addr, zero)
			assertEqual(t, a.Port, 0)
		default:
			t.Errorf("invalid socket address type: %T", a)
		}

		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndReceiveLargerThanRecvBufferSize(family wasi.ProtocolFamily, addr1, addr2 wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockAddr, errno := sys.SockBind(ctx, sock, addr1)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		connAddr, errno := sys.SockBind(ctx, conn, addr2)
		assertEqual(t, errno, wasi.ESUCCESS)

		errno = sys.SockSetOpt(ctx, sock, wasi.RecvBufferSize, wasi.IntValue(4096))
		assertEqual(t, errno, wasi.ESUCCESS)

		recvBufferSize := sockOption[wasi.IntValue](t, ctx, sys, sock, wasi.RecvBufferSize)
		buffer1 := bytes.Repeat([]byte{'@'}, int(recvBufferSize+1))
		buffer2 := make([]byte, len(buffer1))

		size1, errno := sys.SockSendTo(ctx, conn, []wasi.IOVec{buffer1}, 0, sockAddr)
		assertEqual(t, size1, wasi.Size(len(buffer1)))
		assertEqual(t, errno, wasi.ESUCCESS)

		size2, errno := sys.SockSendTo(ctx, conn, []wasi.IOVec{[]byte("42")}, 0, sockAddr)
		assertEqual(t, size2, 2)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockPoll(t, ctx, sys, sock, wasi.FDReadEvent)

		size3, roflags, raddr, errno := sys.SockRecvFrom(ctx, sock, []wasi.IOVec{buffer2}, 0)
		// The actual behavior here is not portable, the message may or may not
		// be dropped depending on the implementation. However, there are only
		// two valid behaviors, either the message is received or it's dropped,
		// so we accept either.
		if int(size3) == len(buffer1) {
			assertEqual(t, string(buffer2[:size3]), string(buffer1[:size3]))
		} else {
			assertEqual(t, size3, 2)
			assertEqual(t, string(buffer2[:2]), "42")
		}
		assertEqual(t, roflags, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertDeepEqual(t, raddr, connAddr)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSendAndReceiveLargerThanSendBufferSize(family wasi.ProtocolFamily, addr1, addr2 wasi.SocketAddress) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		typ := wasi.DatagramSocket

		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		sockAddr, errno := sys.SockBind(ctx, sock, addr1)
		assertEqual(t, errno, wasi.ESUCCESS)

		conn, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)

		sendBufferSize := sockOption[wasi.IntValue](t, ctx, sys, conn, wasi.RecvBufferSize)
		buffer1 := bytes.Repeat([]byte{'@'}, int(sendBufferSize+1))

		size1, errno := sys.SockSendTo(ctx, conn, []wasi.IOVec{buffer1}, 0, sockAddr)
		assertEqual(t, size1, 0)
		assertEqual(t, errno, wasi.EMSGSIZE)

		assertEqual(t, sys.FDClose(ctx, conn), wasi.ESUCCESS)
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
				bufferSize := sockOption[wasi.IntValue](t, ctx, sys, sock, test.option)
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

				defaultBufferSize := sockOption[wasi.IntValue](t, ctx, sys, sock, test.option)
				assertNotEqual(t, defaultBufferSize, 0)

				setBufferSize := func(size wasi.IntValue) {
					t.Helper()
					assertEqual(t, sys.SockSetOpt(ctx, sock, test.option, size), wasi.ESUCCESS)
				}

				getBufferSize := func() wasi.IntValue {
					t.Helper()
					return sockOption[wasi.IntValue](t, ctx, sys, sock, test.option)
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
					assertEqual(t, sys.SockSetOpt(ctx, sock, test.option, wasi.IntValue(-1)), wasi.EINVAL)
					size := getBufferSize()
					assertEqual(t, size, want)
				})

				t.Run("small socket buffer sizes are capped to a minimum value", func(t *testing.T) {
					assertEqual(t, sys.SockSetOpt(ctx, sock, test.option, wasi.IntValue(0)), wasi.ESUCCESS)
					size := getBufferSize()
					assertNotEqual(t, size, 0)
				})

				t.Run("large socket buffer sizes are capped to a maximum value", func(t *testing.T) {
					assertEqual(t, sys.SockSetOpt(ctx, sock, test.option, wasi.IntValue(math.MaxInt32)), wasi.ESUCCESS)
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
		const option = (-1) << 32
		assertEqual(t, sys.SockSetOpt(ctx, sock, option, wasi.IntValue(0)), wasi.EINVAL)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func testSocketSetOptionInvalidArgument(family wasi.ProtocolFamily, typ wasi.SocketType) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		sock, errno := sockOpen(t, ctx, sys, family, typ, 0)
		assertEqual(t, errno, wasi.ESUCCESS)
		const option = -1
		assertEqual(t, sys.SockSetOpt(ctx, sock, option, wasi.IntValue(0)), wasi.EINVAL)
		assertEqual(t, sys.FDClose(ctx, sock), wasi.ESUCCESS)
	}
}

func sockOpen(t *testing.T, ctx context.Context, sys wasi.System, family wasi.ProtocolFamily, typ wasi.SocketType, proto wasi.Protocol) (wasi.FD, wasi.Errno) {
	t.Helper()
	sock, errno := sys.SockOpen(ctx, family, typ, proto, wasi.AllRights, wasi.AllRights)
	skipIfNotImplemented(t, errno)
	if errno == wasi.ESUCCESS {
		setNonBlock(t, ctx, sys, sock, true)
	}
	return sock, errno
}

func setNonBlock(t *testing.T, ctx context.Context, sys wasi.System, sock wasi.FD, nonBlock bool) {
	flags := wasi.FDFlags(0)
	if nonBlock {
		flags |= wasi.NonBlock
	}
	assertEqual(t, sys.FDStatSetFlags(ctx, sock, flags), wasi.ESUCCESS)
	assertEqual(t, sockIsNonBlocking(t, ctx, sys, sock), nonBlock)
	assertEqual(t, sockErrno(t, ctx, sys, sock), wasi.ESUCCESS)
}

func sockOption[T wasi.SocketOptionValue](t *testing.T, ctx context.Context, sys wasi.System, sock wasi.FD, option wasi.SocketOption) T {
	t.Helper()
	opt, errno := sys.SockGetOpt(ctx, sock, option)
	assertEqual(t, errno, wasi.ESUCCESS)
	val, ok := opt.(T)
	assertEqual(t, ok, true)
	return val
}

func sockErrno(t *testing.T, ctx context.Context, sys wasi.System, sock wasi.FD) wasi.Errno {
	t.Helper()
	return wasi.Errno(sockOption[wasi.IntValue](t, ctx, sys, sock, wasi.QuerySocketError))
}

func sockIsNonBlocking(t *testing.T, ctx context.Context, sys wasi.System, sock wasi.FD) bool {
	t.Helper()
	stat, errno := sys.FDStatGet(ctx, sock)
	assertEqual(t, errno, wasi.ESUCCESS)
	return stat.Flags.Has(wasi.NonBlock)
}

func sockPoll(t *testing.T, ctx context.Context, sys wasi.System, sock wasi.FD, eventType wasi.EventType) {
	subs := []wasi.Subscription{
		wasi.MakeSubscriptionFDReadWrite(
			wasi.UserData(sock+1),
			eventType,
			wasi.SubscriptionFDReadWrite{FD: sock},
		),
	}
	evs := make([]wasi.Event, len(subs))
	numEvents, errno := sys.PollOneOff(ctx, subs, evs)
	assertEqual(t, numEvents, 1)
	assertEqual(t, errno, wasi.ESUCCESS)
	assertEqual(t, evs[0], wasi.Event{
		UserData:  wasi.UserData(sock + 1),
		EventType: eventType,
	})
}
