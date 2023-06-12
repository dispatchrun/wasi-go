package unix

import (
	"context"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/stealthrocket/wasi-go"
	"golang.org/x/sys/unix"
)

// System is a WASI preview 1 implementation for Unix.
//
// An instance of System is not safe for concurrent use.
type System struct {
	// Args are the environment variables accessible via ArgsGet.
	Args []string

	// Environ is the environment variables accessible via EnvironGet.
	Environ []string

	// Realtime returns the realtime clock value.
	Realtime          func(context.Context) (uint64, error)
	RealtimePrecision time.Duration

	// Monotonic returns the monotonic clock value.
	Monotonic          func(context.Context) (uint64, error)
	MonotonicPrecision time.Duration

	// Yield is called when SchedYield is called. If Yield is nil,
	// SchedYield is a noop.
	Yield func(context.Context) error

	// Exit is called with an exit code when ProcExit is called.
	// If Exit is nil, ProcExit is a noop.
	Exit func(context.Context, int) error

	// Raise is called with a signal when ProcRaise is called.
	// If Raise is nil, ProcRaise is a noop.
	Raise func(context.Context, int) error

	// Rand is the source for RandomGet.
	Rand io.Reader

	wasi.FileTable[FD]

	pollfds   []unix.PollFd
	unixInet4 unix.SockaddrInet4
	unixInet6 unix.SockaddrInet6
	unixUnix  unix.SockaddrUnix
	inet4Addr wasi.Inet4Address
	inet4Peer wasi.Inet4Address
	inet6Addr wasi.Inet6Address
	inet6Peer wasi.Inet6Address
	unixAddr  wasi.UnixAddress
	unixPeer  wasi.UnixAddress

	// shutfds are a pair of file descriptors allocated to the read and write
	// ends of a pipe. They are used to asynchronously interrupting calls to
	// poll(2) by closing the write end of the pipe, causing the read end to
	// become reading for reading and any polling on the fd to return.
	mutex   sync.Mutex
	shutfds [2]int
}

var _ wasi.System = (*System)(nil)

func (s *System) ArgsSizesGet(ctx context.Context) (argCount, stringBytes int, errno wasi.Errno) {
	argCount, stringBytes = wasi.SizesGet(s.Args)
	return
}

func (s *System) ArgsGet(ctx context.Context) ([]string, wasi.Errno) {
	return s.Args, wasi.ESUCCESS
}

func (s *System) EnvironSizesGet(ctx context.Context) (envCount, stringBytes int, errno wasi.Errno) {
	envCount, stringBytes = wasi.SizesGet(s.Environ)
	return
}

func (s *System) EnvironGet(ctx context.Context) ([]string, wasi.Errno) {
	return s.Environ, wasi.ESUCCESS
}

func (s *System) ClockResGet(ctx context.Context, id wasi.ClockID) (wasi.Timestamp, wasi.Errno) {
	switch id {
	case wasi.Realtime:
		return wasi.Timestamp(s.RealtimePrecision), wasi.ESUCCESS
	case wasi.Monotonic:
		return wasi.Timestamp(s.MonotonicPrecision), wasi.ESUCCESS
	case wasi.ProcessCPUTimeID, wasi.ThreadCPUTimeID:
		return 0, wasi.ENOTSUP
	default:
		return 0, wasi.EINVAL
	}
}

func (s *System) ClockTimeGet(ctx context.Context, id wasi.ClockID, precision wasi.Timestamp) (wasi.Timestamp, wasi.Errno) {
	switch id {
	case wasi.Realtime:
		if s.Realtime == nil {
			return 0, wasi.ENOTSUP
		}
		t, err := s.Realtime(ctx)
		return wasi.Timestamp(t), makeErrno(err)
	case wasi.Monotonic:
		if s.Monotonic == nil {
			return 0, wasi.ENOTSUP
		}
		t, err := s.Monotonic(ctx)
		return wasi.Timestamp(t), makeErrno(err)
	case wasi.ProcessCPUTimeID, wasi.ThreadCPUTimeID:
		return 0, wasi.ENOTSUP
	default:
		return 0, wasi.EINVAL
	}
}

func (s *System) PollOneOff(ctx context.Context, subscriptions []wasi.Subscription, events []wasi.Event) (int, wasi.Errno) {
	if len(subscriptions) == 0 || len(events) < len(subscriptions) {
		return 0, wasi.EINVAL
	}
	wakefd, err := s.init()
	if err != nil {
		return 0, makeErrno(err)
	}
	epoch := time.Duration(0)
	timeout := time.Duration(-1)
	s.pollfds = s.pollfds[:0]
	s.pollfds = append(s.pollfds, unix.PollFd{
		Fd:     int32(wakefd),
		Events: unix.POLLIN | unix.POLLERR | unix.POLLHUP,
	})

	numEvents := 0
	for i := range events {
		events[i] = wasi.Event{}
	}

	for i := range subscriptions {
		sub := &subscriptions[i]

		switch sub.EventType {
		case wasi.FDReadEvent, wasi.FDWriteEvent:
			fd, _, errno := s.LookupFD(sub.GetFDReadWrite().FD, wasi.PollFDReadWriteRight)
			if errno != wasi.ESUCCESS {
				events[i] = errorEvent(sub, errno)
				numEvents++
				continue
			}
			var pollevent int16 = unix.POLLIN
			if sub.EventType == wasi.FDWriteEvent {
				pollevent = unix.POLLOUT
			}
			s.pollfds = append(s.pollfds, unix.PollFd{
				Fd:     int32(fd),
				Events: pollevent,
			})

		case wasi.ClockEvent:
			c := sub.GetClock()
			if c.ID != wasi.Monotonic || s.Monotonic == nil {
				events[i] = errorEvent(sub, wasi.ENOSYS)
				numEvents++
				continue
			}
			if epoch == 0 {
				// Only capture the current time if the program requested a
				// clock subscription; it allows programs that never ask for
				// a timeout to run with a system which does not have a
				// monotonic clock configured.
				now, err := s.Monotonic(ctx)
				if err != nil {
					return 0, makeErrno(err)
				}
				epoch = time.Duration(now)
			}
			t := c.Timeout.Duration()
			if c.Flags.Has(wasi.Abstime) {
				// If the subscription asks for an absolute monotonic time point
				// we can honnor it by computing its relative delta to the poll
				// epoch.
				t -= epoch
			}
			switch {
			case timeout < 0:
				timeout = t
			case t < timeout:
				timeout = t
			}
		}
	}

	var timeoutMillis int
	// We set the timeout to zero when we already produced events due to
	// invalid subscriptions; this is useful to still make progress on I/O
	// completion.
	if numEvents == 0 {
		if timeout < 0 {
			timeoutMillis = -1
		} else {
			timeoutMillis = int(timeout.Milliseconds())
		}
	}

	n, err := unix.Poll(s.pollfds, timeoutMillis)
	if err != nil {
		return 0, makeErrno(err)
	}

	if n > 0 && s.pollfds[0].Revents != 0 {
		// If the wake fd was notified it means the system was shut down,
		// we report this by cancelling all subscriptions.
		//
		// Technically we might be erasing events that had already gathered
		// errors in the first loop prior to the call to unix.Poll; this is
		// not a concern since at this time the program would likely be
		// terminating and should not be bothered with handling other errors.
		for i := range subscriptions {
			events[i] = wasi.Event{
				UserData:  subscriptions[i].UserData,
				EventType: subscriptions[i].EventType,
				Errno:     wasi.ECANCELED,
			}
		}
		return len(subscriptions), wasi.ESUCCESS
	}

	var now time.Duration
	if timeout >= 0 {
		t, err := s.Monotonic(ctx)
		if err != nil {
			return 0, makeErrno(err)
		}
		now = time.Duration(t)
	}

	j := 1
	for i := range subscriptions {
		sub := &subscriptions[i]
		e := wasi.Event{UserData: sub.UserData, EventType: sub.EventType + 1}

		if events[i].EventType != 0 {
			continue
		}

		switch sub.EventType {
		case wasi.ClockEvent:
			c := sub.GetClock()
			t := c.Timeout.Duration()
			if !c.Flags.Has(wasi.Abstime) {
				t += epoch
			}
			if t >= now {
				events[i] = e
			}

		case wasi.FDReadEvent, wasi.FDWriteEvent:
			pf := &s.pollfds[j]
			j++
			if pf.Revents == 0 {
				continue
			}

			if e.EventType == wasi.FDReadEvent && (pf.Revents&unix.POLLIN) != 0 {
				e.FDReadWrite.NBytes = 1 // we don't know how many, so just say 1
			}
			if e.EventType == wasi.FDWriteEvent && (pf.Revents&unix.POLLOUT) != 0 {
				e.FDReadWrite.NBytes = 1 // we don't know how many, so just say 1
			}
			if (pf.Revents & unix.POLLERR) != 0 {
				e.Errno = wasi.ECANCELED // we don't know what error, just pass something
			}
			if (pf.Revents & unix.POLLHUP) != 0 {
				e.FDReadWrite.Flags |= wasi.Hangup
			}
			events[i] = e
		}
	}

	// A 1:1 correspondance between the subscription and events arrays is used
	// to track the completion of events, including the completion of invalid
	// subscriptions, clock events, and I/O notifications coming from poll(2).
	//
	// We use zero as the marker on events for subscriptions that have not been
	// fulfilled, but because the zero event type is used to represent clock
	// subscriptions, we mark completed events with the event type + 1.
	//
	// The event type is finally restored to its correct value in the loop below
	// when we pack all completed events at the front of the output buffer.
	n = 0

	for _, e := range events[:len(subscriptions)] {
		if e.EventType != 0 {
			e.EventType--
			events[n] = e
			n++
		}
	}

	return n, wasi.ESUCCESS
}

func errorEvent(s *wasi.Subscription, err wasi.Errno) wasi.Event {
	return wasi.Event{
		UserData:  s.UserData,
		EventType: s.EventType + 1,
		Errno:     err,
	}
}

func (s *System) ProcExit(ctx context.Context, code wasi.ExitCode) wasi.Errno {
	if s.Exit != nil {
		return makeErrno(s.Exit(ctx, int(code)))
	}
	return wasi.ENOSYS
}

func (s *System) ProcRaise(ctx context.Context, signal wasi.Signal) wasi.Errno {
	if s.Raise != nil {
		return makeErrno(s.Raise(ctx, int(signal)))
	}
	return wasi.ENOSYS
}

func (s *System) SchedYield(ctx context.Context) wasi.Errno {
	if s.Yield != nil {
		return makeErrno(s.Yield(ctx))
	}
	return wasi.ENOSYS
}

func (s *System) RandomGet(ctx context.Context, b []byte) wasi.Errno {
	if _, err := io.ReadFull(s.Rand, b); err != nil {
		return wasi.EIO
	}
	return wasi.ESUCCESS
}

func (s *System) SockAccept(ctx context.Context, fd wasi.FD, flags wasi.FDFlags) (wasi.FD, wasi.SocketAddress, wasi.SocketAddress, wasi.Errno) {
	socket, stat, errno := s.LookupSocketFD(fd, wasi.SockAcceptRight)
	if errno != wasi.ESUCCESS {
		return -1, nil, nil, errno
	}
	if (flags & ^wasi.NonBlock) != 0 {
		return -1, nil, nil, wasi.EINVAL
	}
	addr, errno := s.SockLocalAddress(ctx, fd)
	if errno != wasi.ESUCCESS {
		return -1, nil, nil, errno
	}
	connflags := 0
	if (flags & wasi.NonBlock) != 0 {
		connflags |= unix.O_NONBLOCK
	}
	connfd, sa, err := accept(int(socket), connflags)
	if err != nil {
		return -1, nil, nil, makeErrno(err)
	}
	peer, ok := s.makeRemoteSocketAddress(sa)
	if !ok {
		unix.Close(connfd)
		return -1, nil, nil, wasi.ENOTSUP
	}
	guestfd := s.Register(FD(connfd), wasi.FDStat{
		FileType:         wasi.SocketStreamType,
		Flags:            flags,
		RightsBase:       stat.RightsInheriting,
		RightsInheriting: stat.RightsInheriting,
	})
	return guestfd, peer, addr, wasi.ESUCCESS
}

func (s *System) SockRecv(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.RIFlags) (wasi.Size, wasi.ROFlags, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, wasi.FDReadRight)
	if errno != wasi.ESUCCESS {
		return 0, 0, errno
	}
	var sysIFlags int
	if flags.Has(wasi.RecvPeek) {
		sysIFlags |= unix.MSG_PEEK
	}
	if flags.Has(wasi.RecvWaitAll) {
		sysIFlags |= unix.MSG_WAITALL
	}
	n, _, sysOFlags, _, err := unix.RecvmsgBuffers(int(socket), makeIOVecs(iovecs), nil, sysIFlags)
	var roflags wasi.ROFlags
	if (sysOFlags & unix.MSG_TRUNC) != 0 {
		roflags |= wasi.RecvDataTruncated
	}
	return wasi.Size(n), roflags, makeErrno(err)
}

func (s *System) SockSend(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.SIFlags) (wasi.Size, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := unix.SendmsgBuffers(int(socket), makeIOVecs(iovecs), nil, nil, 0)
	return wasi.Size(n), makeErrno(err)
}

func (s *System) SockShutdown(ctx context.Context, fd wasi.FD, flags wasi.SDFlags) wasi.Errno {
	socket, _, errno := s.LookupSocketFD(fd, wasi.SockShutdownRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	var sysHow int
	switch {
	case flags.Has(wasi.ShutdownRD | wasi.ShutdownWR):
		sysHow = unix.SHUT_RDWR
	case flags.Has(wasi.ShutdownRD):
		sysHow = unix.SHUT_RD
	case flags.Has(wasi.ShutdownWR):
		sysHow = unix.SHUT_WR
	default:
		return wasi.EINVAL
	}
	err := unix.Shutdown(int(socket), sysHow)
	return makeErrno(err)
}

func (s *System) SockOpen(ctx context.Context, pf wasi.ProtocolFamily, socketType wasi.SocketType, protocol wasi.Protocol, rightsBase, rightsInheriting wasi.Rights) (wasi.FD, wasi.Errno) {
	var sysDomain int
	switch pf {
	case wasi.InetFamily:
		sysDomain = unix.AF_INET
	case wasi.Inet6Family:
		sysDomain = unix.AF_INET6
	default:
		return -1, wasi.EINVAL
	}
	var fdType wasi.FileType
	var sysType int
	switch socketType {
	case wasi.DatagramSocket:
		sysType = unix.SOCK_DGRAM
		fdType = wasi.SocketDGramType
	case wasi.StreamSocket:
		sysType = unix.SOCK_STREAM
		fdType = wasi.SocketStreamType
	default:
		return -1, wasi.EINVAL
	}
	var sysProtocol int
	switch protocol {
	case wasi.IPProtocol:
		sysProtocol = unix.IPPROTO_IP
	case wasi.TCPProtocol:
		sysProtocol = unix.IPPROTO_TCP
	case wasi.UDPProtocol:
		sysProtocol = unix.IPPROTO_UDP
	default:
		return -1, wasi.EINVAL
	}
	fd, err := unix.Socket(sysDomain, sysType, sysProtocol)
	if err != nil {
		return -1, makeErrno(err)
	}
	guestfd := s.Register(FD(fd), wasi.FDStat{
		FileType:         fdType,
		RightsBase:       rightsBase,
		RightsInheriting: rightsInheriting,
	})
	return guestfd, wasi.ESUCCESS
}

func (s *System) SockBind(ctx context.Context, fd wasi.FD, addr wasi.SocketAddress) (wasi.SocketAddress, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, wasi.SockAcceptRight)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	sa, ok := s.toUnixSockAddress(addr)
	if !ok {
		return nil, wasi.EINVAL
	}
	if err := unix.Bind(int(socket), sa); err != nil {
		return nil, makeErrno(err)
	}
	return s.SockLocalAddress(ctx, fd)
}

func (s *System) SockConnect(ctx context.Context, fd wasi.FD, peer wasi.SocketAddress) (wasi.SocketAddress, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	sa, ok := s.toUnixSockAddress(peer)
	if !ok {
		return nil, wasi.EINVAL
	}
	err := unix.Connect(int(socket), sa)
	if err != nil && err != unix.EINPROGRESS {
		return nil, makeErrno(err)
	}
	addr, errno := s.SockLocalAddress(ctx, fd)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	return addr, makeErrno(err)
}

func (s *System) SockListen(ctx context.Context, fd wasi.FD, backlog int) wasi.Errno {
	socket, _, errno := s.LookupSocketFD(fd, wasi.SockAcceptRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Listen(int(socket), backlog)
	return makeErrno(err)
}

func (s *System) SockSendTo(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.SIFlags, addr wasi.SocketAddress) (wasi.Size, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	sa, ok := s.toUnixSockAddress(addr)
	if !ok {
		return 0, wasi.EINVAL
	}
	n, err := unix.SendmsgBuffers(int(socket), makeIOVecs(iovecs), nil, sa, 0)
	return wasi.Size(n), makeErrno(err)
}

func (s *System) SockRecvFrom(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.RIFlags) (wasi.Size, wasi.ROFlags, wasi.SocketAddress, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, wasi.FDReadRight)
	if errno != wasi.ESUCCESS {
		return 0, 0, nil, errno
	}
	var sysIFlags int
	if flags.Has(wasi.RecvPeek) {
		sysIFlags |= unix.MSG_PEEK
	}
	if flags.Has(wasi.RecvWaitAll) {
		sysIFlags |= unix.MSG_WAITALL
	}
	n, _, sysOFlags, sa, err := unix.RecvmsgBuffers(int(socket), makeIOVecs(iovecs), nil, sysIFlags)
	var addr wasi.SocketAddress
	if sa != nil {
		var ok bool
		addr, ok = s.makeRemoteSocketAddress(sa)
		if !ok {
			errno = wasi.ENOTSUP
		}
	}
	var roflags wasi.ROFlags
	if (sysOFlags & unix.MSG_TRUNC) != 0 {
		roflags |= wasi.RecvDataTruncated
	}
	return wasi.Size(n), roflags, addr, makeErrno(err)
}

func (s *System) SockGetOpt(ctx context.Context, fd wasi.FD, level wasi.SocketOptionLevel, option wasi.SocketOption) (wasi.SocketOptionValue, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	var sysLevel int
	switch level {
	case wasi.SocketLevel:
		sysLevel = unix.SOL_SOCKET
	default:
		return nil, wasi.EINVAL
	}
	var sysOption int
	switch option {
	case wasi.ReuseAddress:
		sysOption = unix.SO_REUSEADDR
	case wasi.QuerySocketType:
		sysOption = unix.SO_TYPE
	case wasi.QuerySocketError:
		sysOption = unix.SO_ERROR
	case wasi.DontRoute:
		sysOption = unix.SO_DONTROUTE
	case wasi.Broadcast:
		sysOption = unix.SO_BROADCAST
	case wasi.SendBufferSize:
		sysOption = unix.SO_SNDBUF
	case wasi.RecvBufferSize:
		sysOption = unix.SO_RCVBUF
	case wasi.KeepAlive:
		sysOption = unix.SO_KEEPALIVE
	case wasi.OOBInline:
		sysOption = unix.SO_OOBINLINE
	case wasi.RecvLowWatermark:
		sysOption = unix.SO_RCVLOWAT
	case wasi.QueryAcceptConnections:
		sysOption = unix.SO_ACCEPTCONN
	case wasi.Linger:
		// This returns a struct linger value.
		return nil, wasi.ENOTSUP // TODO: implement SO_LINGER
	case wasi.RecvTimeout, wasi.SendTimeout:
		// These return a struct timeval value.
		return nil, wasi.ENOTSUP // TODO: implement SO_RCVTIMEO, SO_SNDTIMEO
	case wasi.BindToDevice:
		// This returns a string value.
		return nil, wasi.ENOTSUP // TODO: implement SO_BINDTODEVICE
	default:
		return nil, wasi.EINVAL
	}

	value, err := unix.GetsockoptInt(int(socket), sysLevel, sysOption)
	if err != nil {
		return nil, makeErrno(err)
	}

	errno = wasi.ESUCCESS
	switch option {
	case wasi.QuerySocketType:
		switch value {
		case unix.SOCK_DGRAM:
			value = int(wasi.DatagramSocket)
		case unix.SOCK_STREAM:
			value = int(wasi.StreamSocket)
		default:
			value = -1
			errno = wasi.ENOTSUP
		}
	case wasi.QuerySocketError:
		value = int(makeErrno(unix.Errno(value)))
	}

	return wasi.IntValue(value), errno
}

func (s *System) SockSetOpt(ctx context.Context, fd wasi.FD, level wasi.SocketOptionLevel, option wasi.SocketOption, value wasi.SocketOptionValue) wasi.Errno {
	socket, _, errno := s.LookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	var sysLevel int
	switch level {
	case wasi.SocketLevel:
		sysLevel = unix.SOL_SOCKET
	default:
		return wasi.EINVAL
	}
	var sysOption int
	switch option {
	case wasi.ReuseAddress:
		sysOption = unix.SO_REUSEADDR
	case wasi.QuerySocketType:
		sysOption = unix.SO_TYPE
	case wasi.QuerySocketError:
		sysOption = unix.SO_ERROR
	case wasi.DontRoute:
		sysOption = unix.SO_DONTROUTE
	case wasi.Broadcast:
		sysOption = unix.SO_BROADCAST
	case wasi.SendBufferSize:
		sysOption = unix.SO_SNDBUF
	case wasi.RecvBufferSize:
		sysOption = unix.SO_RCVBUF
	case wasi.KeepAlive:
		sysOption = unix.SO_KEEPALIVE
	case wasi.OOBInline:
		sysOption = unix.SO_OOBINLINE
	case wasi.RecvLowWatermark:
		sysOption = unix.SO_RCVLOWAT
	case wasi.QueryAcceptConnections:
		sysOption = unix.SO_ACCEPTCONN
	case wasi.Linger:
		// This accepts a struct linger value.
		return wasi.ENOTSUP // TODO: implement SO_LINGER
	case wasi.RecvTimeout, wasi.SendTimeout:
		// These accept a struct timeval value.
		return wasi.ENOTSUP // TODO: implement SO_RCVTIMEO, SO_SNDTIMEO
	case wasi.BindToDevice:
		// This accepts a string value.
		return wasi.ENOTSUP // TODO: implement SO_BINDTODEVICE
	default:
		return wasi.EINVAL
	}
	intval, ok := value.(wasi.IntValue)
	if !ok {
		return wasi.EINVAL
	}
	err := unix.SetsockoptInt(int(socket), sysLevel, sysOption, int(intval))
	return makeErrno(err)
}

func (s *System) SockLocalAddress(ctx context.Context, fd wasi.FD) (wasi.SocketAddress, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	sa, err := unix.Getsockname(int(socket))
	if err != nil {
		return nil, makeErrno(err)
	}
	addr, ok := s.makeLocalSocketAddress(sa)
	if !ok {
		return nil, wasi.ENOTSUP
	}
	return addr, wasi.ESUCCESS
}

func (s *System) SockRemoteAddress(ctx context.Context, fd wasi.FD) (wasi.SocketAddress, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	sa, err := unix.Getpeername(int(socket))
	if err != nil {
		return nil, makeErrno(err)
	}
	addr, ok := s.makeRemoteSocketAddress(sa)
	if !ok {
		return nil, wasi.ENOTSUP
	}
	return addr, wasi.ESUCCESS
}

func (s *System) SockAddressInfo(ctx context.Context, name, service string, hints wasi.AddressInfo, results []wasi.AddressInfo) (int, wasi.Errno) {
	if cap(results) == 0 {
		return 0, wasi.EINVAL
	}
	// TODO: support AI_ADDRCONFIG, AI_CANONNAME, AI_V4MAPPED, AI_V4MAPPED_CFG, AI_ALL

	var network string
	f, p, t := hints.Family, hints.Protocol, hints.SocketType
	switch {
	case t == wasi.StreamSocket && p != wasi.UDPProtocol:
		switch f {
		case wasi.UnspecifiedFamily:
			network = "tcp"
		case wasi.InetFamily:
			network = "tcp4"
		case wasi.Inet6Family:
			network = "tcp6"
		default:
			return 0, wasi.ENOTSUP // EAI_FAMILY
		}
	case t == wasi.DatagramSocket && p != wasi.TCPProtocol:
		switch f {
		case wasi.UnspecifiedFamily:
			network = "udp"
		case wasi.InetFamily:
			network = "udp4"
		case wasi.Inet6Family:
			network = "udp6"
		default:
			return 0, wasi.ENOTSUP // EAI_FAMILY
		}
	case t == wasi.AnySocket:
		switch f {
		case wasi.UnspecifiedFamily:
			network = "ip"
		case wasi.InetFamily:
			network = "ip4"
		case wasi.Inet6Family:
			network = "ip6"
		default:
			return 0, wasi.ENOTSUP // EAI_FAMILY
		}
	default:
		return 0, wasi.ENOTSUP // EAI_SOCKTYPE / EAI_PROTOCOL
	}

	var port int
	var err error
	if hints.Flags.Has(wasi.NumericService) {
		port, err = strconv.Atoi(service)
	} else {
		port, err = net.LookupPort(network, service)
	}
	if err != nil || port < 0 || port > 65535 {
		return 0, wasi.EINVAL // EAI_NONAME / EAI_SERVICE
	}

	var ip net.IP
	if hints.Flags.Has(wasi.NumericHost) {
		ip = net.ParseIP(name)
		if ip == nil {
			return 0, wasi.EINVAL
		}
	} else if name == "" {
		if !hints.Flags.Has(wasi.Passive) {
			return 0, wasi.EINVAL
		}
		if hints.Family == wasi.Inet6Family {
			ip = net.IPv6zero
		} else {
			ip = net.IPv4zero
		}
	}
	if ip != nil {
		results = results[:1]
		results[0] = wasi.AddressInfo{}
		if ipv4 := ip.To4(); ipv4 != nil {
			s.inet4Addr.Port = port
			copy(s.inet4Addr.Addr[:], ipv4)
			results[0].Address = &s.inet4Addr
		} else {
			s.inet6Addr.Port = port
			copy(s.inet6Addr.Addr[:], ip)
			results[0].Address = &s.inet6Addr
		}
		return 1, wasi.ESUCCESS
	}

	ips, err := net.LookupIP(name)
	if err != nil {
		return 0, wasi.ECANCELED // TODO: better errors on name resolution failure
	}

	if len(ips) > cap(results) {
		ips = ips[:cap(results)]
	}
	results = results[:0]
	for _, ip := range ips {
		var addr wasi.AddressInfo
		if ipv4 := ip.To4(); ipv4 != nil {
			if hints.Family == wasi.Inet6Family {
				continue
			}
			inet4Addr := wasi.Inet4Address{Port: port}
			copy(inet4Addr.Addr[:], ip)
			addr.Address = &inet4Addr
		} else {
			if hints.Family == wasi.InetFamily {
				continue
			}
			inet6Addr := wasi.Inet6Address{Port: port}
			copy(inet6Addr.Addr[:], ip)
			addr.Address = &inet6Addr
		}
		results = append(results, addr)
	}
	return len(results), wasi.ESUCCESS
}

func (s *System) Close(ctx context.Context) error {
	err := s.FileTable.Close(ctx)

	s.mutex.Lock()
	fd0 := s.shutfds[0]
	fd1 := s.shutfds[1]
	s.shutfds[0] = -1
	s.shutfds[1] = -1
	s.mutex.Unlock()

	if fd0 != 0 || fd1 != 0 { // true if the system was initialized
		unix.Close(fd0)
		unix.Close(fd1)
	}
	return err
}

// Shutdown may be called to asynchronously cancel all blocking operations on
// the system, causing calls such as PollOneOff to unblock and return an
// error indicating that the system is shutting down.
func (s *System) Shutdown(ctx context.Context) error {
	_, err := s.init()
	if err != nil {
		return err
	}
	s.shutdown()
	return nil
}

func (s *System) init() (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.shutfds[0] == 0 && s.shutfds[1] == 0 {
		if err := pipe(s.shutfds[:], unix.O_NONBLOCK); err != nil {
			return -1, err
		}
	}
	return s.shutfds[0], nil
}

func (s *System) shutdown() {
	s.mutex.Lock()
	fd := s.shutfds[1]
	s.shutfds[1] = -1
	s.mutex.Unlock()
	unix.Close(fd)
}

func (s *System) toUnixSockAddress(addr wasi.SocketAddress) (sa unix.Sockaddr, ok bool) {
	switch t := addr.(type) {
	case *wasi.Inet4Address:
		s.unixInet4.Port = t.Port
		s.unixInet4.Addr = t.Addr
		sa = &s.unixInet4
	case *wasi.Inet6Address:
		s.unixInet6.Port = t.Port
		s.unixInet6.Addr = t.Addr
		sa = &s.unixInet6
	case *wasi.UnixAddress:
		s.unixUnix.Name = t.Name
		sa = &s.unixUnix
	default:
		return nil, false
	}
	return sa, true
}

func (s *System) makeLocalSocketAddress(sa unix.Sockaddr) (wasi.SocketAddress, bool) {
	return s.makeSocketAddress(sa, &s.inet4Addr, &s.inet6Addr, &s.unixAddr)
}

func (s *System) makeRemoteSocketAddress(sa unix.Sockaddr) (wasi.SocketAddress, bool) {
	return s.makeSocketAddress(sa, &s.inet4Peer, &s.inet6Peer, &s.unixPeer)
}

func (s *System) makeSocketAddress(sa unix.Sockaddr, in4 *wasi.Inet4Address, in6 *wasi.Inet6Address, un *wasi.UnixAddress) (wasi.SocketAddress, bool) {
	switch t := sa.(type) {
	case *unix.SockaddrInet4:
		in4.Addr = t.Addr
		in4.Port = t.Port
		return in4, true
	case *unix.SockaddrInet6:
		in6.Addr = t.Addr
		in6.Port = t.Port
		return in6, true
	case *unix.SockaddrUnix:
		un.Name = t.Name
		return un, true
	default:
		return nil, false
	}
}
