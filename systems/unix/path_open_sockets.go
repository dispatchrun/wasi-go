package unix

import (
	"context"
	"net/url"
	"strings"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/internal/sockets"
)

// PathOpenSockets is an extension to WASI preview 1 that adds the ability to
// create TCP sockets. It works by proxying calls to path_open. If fd<0
// and the path is of the form:
//
//	<network>:<operation>://<host>:<port>[?options=value[&option=value]*
//
// where network is one of "tcp", "tcp4" or "tcp6", and operation is either
// "listen" or "dial", the extension will open a socket that either listens
// on, or connects to, the specified host:port address. Otherwise, the
// extension passes the arguments to the underlying WASI implementation to open
// a file or directory as normal.
//
// The following options are available
// - nonblock=<0|1>:  Open the socket in non-blocking mode. Default is 1.
// - nodelay=<0|1>:   Set TCP_NODELAY. Default is 1.
// - reuseaddr=<0|1>: Set SO_REUSEADDR. Default is 1.
// - backlog=<N>:     Set the listen(2) backlog. Default is 128.
type PathOpenSockets struct{ wasi.System }

func (p *PathOpenSockets) PathOpen(ctx context.Context, fd wasi.FD, lookupFlags wasi.LookupFlags, path string, openFlags wasi.OpenFlags, rightsBase, rightsInheriting wasi.Rights, fdFlags wasi.FDFlags) (wasi.FD, wasi.Errno) {
	addr, op, ok := parseURI(path)
	if !ok || fd >= 0 {
		return p.System.PathOpen(ctx, fd, lookupFlags, path, openFlags, rightsBase, rightsInheriting, fdFlags)
	}
	var sockfd int
	var err error
	switch op {
	case "listen":
		sockfd, err = sockets.Listen(addr)
	case "dial":
		sockfd, err = sockets.Dial(addr)
	}
	errno := wasi.ESUCCESS
	if err != nil {
		errno = makeErrno(err)
		if errno != wasi.EINPROGRESS {
			return -1, errno
		}
	}
	return p.Register(sockfd, wasi.FDStat{
		FileType:         wasi.SocketStreamType,
		Flags:            fdFlags,
		RightsBase:       rightsBase,
		RightsInheriting: rightsInheriting,
	}), errno
}

func parseURI(path string) (network string, op string, ok bool) {
	u, err := url.Parse(path)
	if err != nil {
		return
	}
	network, op, ok = strings.Cut(u.Scheme, "+")
	if !ok || (op != "listen" && op != "dial") {
		return
	}
	u.Scheme = network
	return u.String(), op, true
}
