package unix

import (
	"context"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
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

	pollfds []unix.PollFd
	inet4   unix.SockaddrInet4
	inet6   unix.SockaddrInet6
	unix    unix.SockaddrUnix

	mutex sync.Mutex
	wake  [2]*os.File
	shut  atomic.Bool
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
	r, _, err := s.init()
	if err != nil {
		return 0, makeErrno(err)
	}
	s.pollfds = append(s.pollfds[:0], unix.PollFd{
		Fd:     int32(r.Fd()),
		Events: unix.POLLIN | unix.POLLHUP,
	})

	realtimeEpoch := time.Duration(0)
	monotonicEpoch := time.Duration(0)

	timeout := time.Duration(-1)
	timeoutEventIndex := -1

	numEvents := 0
	for i := range events {
		events[i] = wasi.Event{}
	}

	for i := range subscriptions {
		sub := &subscriptions[i]

		var pollEvent int16 = unix.POLLIN | unix.POLLHUP
		switch sub.EventType {
		case wasi.FDWriteEvent:
			pollEvent = unix.POLLOUT
			fallthrough
		case wasi.FDReadEvent:
			fd, _, errno := s.LookupFD(sub.GetFDReadWrite().FD, wasi.PollFDReadWriteRight)
			if errno != wasi.ESUCCESS {
				events[i] = errorEvent(sub, errno)
				numEvents++
				continue
			}
			s.pollfds = append(s.pollfds, unix.PollFd{
				Fd:     int32(fd),
				Events: pollEvent,
			})

		case wasi.ClockEvent:
			c := sub.GetClock()

			var epoch *time.Duration
			var gettime func(context.Context) (uint64, error)
			switch c.ID {
			case wasi.Realtime:
				epoch, gettime = &realtimeEpoch, s.Realtime
			case wasi.Monotonic:
				epoch, gettime = &monotonicEpoch, s.Monotonic
			}
			if gettime == nil {
				events[i] = errorEvent(sub, wasi.ENOTSUP)
				numEvents++
				continue
			}

			t := c.Timeout.Duration() + c.Precision.Duration()
			if c.Flags.Has(wasi.Abstime) {
				// Only capture the current time if the program requested a
				// clock subscription; it allows programs that never ask for
				// a timeout to run with a system which does not have a
				// monotonic clock configured.
				if *epoch == 0 {
					t, err := gettime(ctx)
					if err != nil {
						events[i] = errorEvent(sub, wasi.MakeErrno(err))
						numEvents++
						continue
					}
					*epoch = time.Duration(t)
				}
				// If the subscription asks for an absolute monotonic time point
				// we can honnor it by computing its relative delta to the poll
				// epoch.
				t -= *epoch
			}

			if t < 0 {
				t = 0
			}
			switch {
			case timeout < 0:
				timeout = t
				timeoutEventIndex = i
			case t < timeout:
				timeout = t
				timeoutEventIndex = i
			}
		}
	}

	// We set the timeout to zero when we already produced events due to
	// invalid subscriptions; this is useful to still make progress on I/O
	// completion.
	var deadline time.Time
	if numEvents > 0 {
		timeout = 0
	}
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	// This loops until either the deadline is reached or at least one event is
	// reported.
	for {
		timeoutMillis := 0
		switch {
		case timeout < 0:
			timeoutMillis = -1
		case !deadline.IsZero():
			timeoutMillis = int(time.Until(deadline).Milliseconds())
		}

		n, err := unix.Poll(s.pollfds, timeoutMillis)
		if err != nil && err != unix.EINTR {
			return 0, makeErrno(err)
		}

		// poll(2) may cause spurious wake up, so we verify that the system
		// has indeed been shutdown instead of relying on reading the events
		// reported on the first pollfd.
		if s.shut.Load() {
			// If the wake fd was notified it means the system was shut down,
			// we report this by cancelling all subscriptions.
			//
			// Technically we might be erasing events that had already gathered
			// errors in the first loop prior to the call to unix.Poll; this is
			// not a concern since at this time the program would likely be
			// terminating and should not be bothered with handling other
			// errors.
			for i := range subscriptions {
				events[i] = wasi.Event{
					UserData:  subscriptions[i].UserData,
					EventType: subscriptions[i].EventType,
					Errno:     wasi.ECANCELED,
				}
			}
			return len(subscriptions), wasi.ESUCCESS
		}

		if timeoutEventIndex >= 0 && deadline.Before(time.Now().Add(time.Millisecond)) {
			events[timeoutEventIndex] = wasi.Event{
				UserData:  subscriptions[timeoutEventIndex].UserData,
				EventType: subscriptions[timeoutEventIndex].EventType + 1,
			}
		}

		j := 1
		for i := range subscriptions {
			if events[i].EventType != 0 {
				continue
			}
			switch sub := &subscriptions[i]; sub.EventType {
			case wasi.FDReadEvent, wasi.FDWriteEvent:
				pf := &s.pollfds[j]
				j++
				if pf.Revents == 0 {
					continue
				}
				// Linux never reports POLLHUP for disconnected sockets,
				// so there is no reliable mechanism to set wasi.Hanghup.
				// We optimize for portability here and just report that
				// the file descriptor is ready for reading or writing,
				// and let the application deal with the conditions it
				// sees from the following calles to read/write/etc...
				events[i] = wasi.Event{
					UserData:  sub.UserData,
					EventType: sub.EventType + 1,
				}
			}
		}

		// A 1:1 correspondance between the subscription and events arrays is
		// used to track the completion of events, including the completion of
		// invalid subscriptions, clock events, and I/O notifications coming
		// from poll(2).
		//
		// We use zero as the marker on events for subscriptions that have not
		// been fulfilled, but because the zero event type is used to represent
		// clock subscriptions, we mark completed events with the event type+1.
		//
		// The event type is finally restored to its correct value in the loop
		// below when we pack all completed events at the front of the output
		// buffer.
		n = 0

		for _, e := range events[:len(subscriptions)] {
			if e.EventType != 0 {
				e.EventType--
				events[n] = e
				n++
			}
		}

		if n > 0 {
			return n, wasi.ESUCCESS
		}
	}
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
	peer := makeSocketAddress(sa)
	if peer == nil {
		_ = closeTraceEBADF(connfd)
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
	for {
		n, _, sysOFlags, _, err := unix.RecvmsgBuffers(int(socket), makeIOVecs(iovecs), nil, sysIFlags)
		if err == unix.EINTR {
			continue
		}
		var roflags wasi.ROFlags
		if (sysOFlags & unix.MSG_TRUNC) != 0 {
			roflags |= wasi.RecvDataTruncated
		}
		return wasi.Size(n), roflags, makeErrno(err)
	}
}

func (s *System) SockSend(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.SIFlags) (wasi.Size, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := handleEINTR(func() (int, error) {
		return unix.SendmsgBuffers(int(socket), makeIOVecs(iovecs), nil, nil, 0)
	})
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
	// Linux allows calling shutdown(2) on listening sockets, but not Darwin.
	// To provide a portable behavior we align on the POSIX behavior which says
	// that shutting down non-connected sockets must return ENOTCONN.
	//
	// Note that this may cause issues in the future if applications need a way
	// to break out of a blocking accept(2) call. We could relax this limitation
	// down the line, tho keep in mind that applications may be better served by
	// not relying on system-specific behaviors and should use synchronization
	// mechanisms is user-space to maximize portability.
	//
	// For more context see: https://bugzilla.kernel.org/show_bug.cgi?id=106241
	if runtime.GOOS == "linux" {
		v, err := ignoreEINTR2(func() (int, error) {
			return unix.GetsockoptInt(int(socket), unix.SOL_SOCKET, unix.SO_ACCEPTCONN)
		})
		if err != nil {
			return makeErrno(err)
		}
		if v != 0 {
			return wasi.ENOTCONN
		}
	}
	err := ignoreEINTR(func() error { return unix.Shutdown(int(socket), sysHow) })
	return makeErrno(err)
}

func (s *System) SockOpen(ctx context.Context, pf wasi.ProtocolFamily, socketType wasi.SocketType, protocol wasi.Protocol, rightsBase, rightsInheriting wasi.Rights) (wasi.FD, wasi.Errno) {
	var sysDomain int
	switch pf {
	case wasi.InetFamily:
		sysDomain = unix.AF_INET
	case wasi.Inet6Family:
		sysDomain = unix.AF_INET6
	case wasi.UnixFamily:
		sysDomain = unix.AF_UNIX
	default:
		return -1, wasi.EINVAL
	}

	if socketType == wasi.AnySocket {
		switch protocol {
		case wasi.TCPProtocol:
			socketType = wasi.StreamSocket
		case wasi.UDPProtocol:
			socketType = wasi.DatagramSocket
		}
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

	fd, err := ignoreEINTR2(func() (int, error) {
		return unix.Socket(sysDomain, sysType, sysProtocol)
	})
	if err != nil {
		// Darwin gives EPROTOTYPE when the socket type and protocol do
		// not match, which differs from the Linux behavior which returns
		// EPROTONOSUPPORT. Since there is no real use case for dealing
		// with the error differently, and valid applications will not
		// invoke SockOpen with invalid parameters, we align on the Linux
		// behavior for simplicity.
		if err == unix.EPROTOTYPE {
			err = unix.EPROTONOSUPPORT
		}
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
	err := ignoreEINTR(func() error { return unix.Bind(int(socket), sa) })
	if err != nil {
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
	err := ignoreEINTR(func() error { return unix.Connect(int(socket), sa) })
	if err != nil && err != unix.EINPROGRESS {
		// Darwin gives EOPNOTSUPP when trying to connect a socket that is
		// already connected or already listening. Align on the Linux behavior
		// here and convert the error to EISCONN.
		if err == unix.EOPNOTSUPP {
			err = unix.EISCONN
		}
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
	err := ignoreEINTR(func() error { return unix.Listen(int(socket), backlog) })
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
	n, err := handleEINTR(func() (int, error) {
		return unix.SendmsgBuffers(int(socket), makeIOVecs(iovecs), nil, sa, 0)
	})
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
	for {
		n, _, sysOFlags, sa, err := unix.RecvmsgBuffers(int(socket), makeIOVecs(iovecs), nil, sysIFlags)
		if err == unix.EINTR {
			continue
		}
		var addr wasi.SocketAddress
		if sa != nil {
			addr = makeSocketAddress(sa)
			if addr == nil {
				errno = wasi.ENOTSUP
			}
		}
		var roflags wasi.ROFlags
		if (sysOFlags & unix.MSG_TRUNC) != 0 {
			roflags |= wasi.RecvDataTruncated
		}
		return wasi.Size(n), roflags, addr, makeErrno(err)
	}
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
	case wasi.TcpLevel:
		sysLevel = unix.IPPROTO_TCP
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
	case wasi.TcpNoDelay:
		sysOption = unix.TCP_NODELAY
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

	value, err := ignoreEINTR2(func() (int, error) {
		return unix.GetsockoptInt(int(socket), sysLevel, sysOption)
	})
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
	case wasi.RecvBufferSize, wasi.SendBufferSize:
		// Linux doubles the socket buffer sizes, so we adjust the value here
		// to ensure the behavior is portable across operating systems.
		if runtime.GOOS == "linux" {
			value /= 2
		}
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
	case wasi.TcpLevel:
		sysLevel = unix.IPPROTO_TCP
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
	case wasi.TcpNoDelay:
		sysOption = unix.TCP_NODELAY
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

	// Linux allows setting the socket buffer size to zero, but this is not
	// portable so we instead refuse to support this option here.
	if runtime.GOOS == "linux" {
		switch option {
		case wasi.RecvBufferSize, wasi.SendBufferSize:
			if intval <= 0 {
				return wasi.EINVAL
			}
		}
	}

	err := ignoreEINTR(func() error {
		return unix.SetsockoptInt(int(socket), sysLevel, sysOption, int(intval))
	})
	return makeErrno(err)
}

func (s *System) SockLocalAddress(ctx context.Context, fd wasi.FD) (wasi.SocketAddress, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	sa, err := ignoreEINTR2(func() (unix.Sockaddr, error) {
		return unix.Getsockname(int(socket))
	})
	if err != nil {
		return nil, makeErrno(err)
	}
	addr := makeSocketAddress(sa)
	if addr == nil {
		return nil, wasi.ENOTSUP
	}
	return addr, wasi.ESUCCESS
}

func (s *System) SockRemoteAddress(ctx context.Context, fd wasi.FD) (wasi.SocketAddress, wasi.Errno) {
	socket, _, errno := s.LookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	sa, err := ignoreEINTR2(func() (unix.Sockaddr, error) {
		return unix.Getpeername(int(socket))
	})
	if err != nil {
		return nil, makeErrno(err)
	}
	addr := makeSocketAddress(sa)
	if addr == nil {
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
			inet4Addr := &wasi.Inet4Address{Port: port}
			copy(inet4Addr.Addr[:], ipv4)
			results[0].Address = inet4Addr
		} else {
			inet6Addr := &wasi.Inet6Address{Port: port}
			copy(inet6Addr.Addr[:], ip)
			results[0].Address = inet6Addr
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
	s.shut.Store(true)
	s.mutex.Lock()
	r := s.wake[0]
	w := s.wake[1]
	s.wake[0] = nil
	s.wake[1] = nil
	s.mutex.Unlock()

	if r != nil {
		r.Close()
	}
	if w != nil {
		w.Close()
	}
	return s.FileTable.Close(ctx)
}

// Shutdown may be called asynchronously to cancel all blocking operations on
// the system, causing calls such as PollOneOff to unblock and return an
// error indicating that the system is shutting down.
func (s *System) Shutdown(ctx context.Context) error {
	_, w, err := s.init()
	if err != nil {
		if err == context.Canceled {
			err = nil // already shutdown
		}
		return err
	}
	s.shut.Store(true)
	return w.Close()
}

func (s *System) init() (*os.File, *os.File, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.wake[0] == nil {
		if s.shut.Load() {
			return nil, nil, context.Canceled
		}
		r, w, err := os.Pipe()
		if err != nil {
			return nil, nil, err
		}
		s.wake[0] = r
		s.wake[1] = w
	}

	return s.wake[0], s.wake[1], nil
}

func (s *System) toUnixSockAddress(addr wasi.SocketAddress) (sa unix.Sockaddr, ok bool) {
	switch t := addr.(type) {
	case *wasi.Inet4Address:
		s.inet4.Port = t.Port
		s.inet4.Addr = t.Addr
		sa = &s.inet4
	case *wasi.Inet6Address:
		s.inet6.Port = t.Port
		s.inet6.Addr = t.Addr
		sa = &s.inet6
	case *wasi.UnixAddress:
		s.unix.Name = t.Name
		sa = &s.unix
	default:
		return nil, false
	}
	return sa, true
}

func makeSocketAddress(sa unix.Sockaddr) wasi.SocketAddress {
	switch t := sa.(type) {
	case *unix.SockaddrInet4:
		return &wasi.Inet4Address{
			Addr: t.Addr,
			Port: t.Port,
		}
	case *unix.SockaddrInet6:
		return &wasi.Inet6Address{
			Addr: t.Addr,
			Port: t.Port,
		}
	case *unix.SockaddrUnix:
		name := t.Name
		if len(name) == 0 {
			// For consistency across platforms, replace empty unix socket
			// addresses with @. On Linux, addresses where the first byte is
			// a null byte are considered abstract unix sockets, and the first
			// byte is replaced with @.
			name = "@"
		}
		return &wasi.UnixAddress{
			Name: name,
		}
	default:
		return nil
	}
}
