package wasi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"syscall"
)

func makeErrno(err error) Errno {
	if err == nil {
		return ESUCCESS
	}
	if err == syscall.EAGAIN {
		return EAGAIN
	}
	return makeErrnoSlow(err)
}

func makeErrnoSlow(err error) Errno {
	switch {
	case errors.Is(err, context.Canceled):
		return ECANCELED
	case errors.Is(err, context.DeadlineExceeded):
		return ETIMEDOUT
	case errors.Is(err, io.ErrUnexpectedEOF),
		errors.Is(err, fs.ErrClosed),
		errors.Is(err, net.ErrClosed):
		return EIO
	}

	var sysErrno syscall.Errno
	if errors.As(err, &sysErrno) {
		if sysErrno == 0 {
			return ESUCCESS
		}
		return syscallErrnoToWASI(sysErrno)
	}

	var wasiErrno Errno
	if errors.As(err, &wasiErrno) {
		return wasiErrno
	}

	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) {
		if timeout.Timeout() {
			return ETIMEDOUT
		}
	}

	panic(fmt.Errorf("unexpected error: %v", err))
}

func errnoToSyscall(errno Errno) syscall.Errno {
	switch errno {
	case ESUCCESS:
		return 0
	case E2BIG:
		return syscall.E2BIG
	case EACCES:
		return syscall.EACCES
	case EADDRINUSE:
		return syscall.EADDRINUSE
	case EADDRNOTAVAIL:
		return syscall.EADDRNOTAVAIL
	case EAFNOSUPPORT:
		return syscall.EAFNOSUPPORT
	case EAGAIN:
		return syscall.EAGAIN
	case EALREADY:
		return syscall.EALREADY
	case EBADF:
		return syscall.EBADF
	case EBADMSG:
		return syscall.EBADMSG
	case EBUSY:
		return syscall.EBUSY
	case ECANCELED:
		return syscall.ECANCELED
	case ECHILD:
		return syscall.ECHILD
	case ECONNABORTED:
		return syscall.ECONNABORTED
	case ECONNREFUSED:
		return syscall.ECONNREFUSED
	case ECONNRESET:
		return syscall.ECONNRESET
	case EDEADLK:
		return syscall.EDEADLK
	case EDESTADDRREQ:
		return syscall.EDESTADDRREQ
	case EDOM:
		return syscall.EDOM
	case EDQUOT:
		return syscall.EDQUOT
	case EEXIST:
		return syscall.EEXIST
	case EFAULT:
		return syscall.EFAULT
	case EFBIG:
		return syscall.EFBIG
	case EHOSTUNREACH:
		return syscall.EHOSTUNREACH
	case EIDRM:
		return syscall.EIDRM
	case EILSEQ:
		return syscall.EILSEQ
	case EINPROGRESS:
		return syscall.EINPROGRESS
	case EINTR:
		return syscall.EINTR
	case EINVAL:
		return syscall.EINVAL
	case EIO:
		return syscall.EIO
	case EISCONN:
		return syscall.EISCONN
	case EISDIR:
		return syscall.EISDIR
	case ELOOP:
		return syscall.ELOOP
	case EMFILE:
		return syscall.EMFILE
	case EMLINK:
		return syscall.EMLINK
	case EMSGSIZE:
		return syscall.EMSGSIZE
	case EMULTIHOP:
		return syscall.EMULTIHOP
	case ENAMETOOLONG:
		return syscall.ENAMETOOLONG
	case ENETDOWN:
		return syscall.ENETDOWN
	case ENETRESET:
		return syscall.ENETRESET
	case ENETUNREACH:
		return syscall.ENETUNREACH
	case ENFILE:
		return syscall.ENFILE
	case ENOBUFS:
		return syscall.ENOBUFS
	case ENODEV:
		return syscall.ENODEV
	case ENOENT:
		return syscall.ENOENT
	case ENOEXEC:
		return syscall.ENOEXEC
	case ENOLCK:
		return syscall.ENOLCK
	case ENOLINK:
		return syscall.ENOLINK
	case ENOMEM:
		return syscall.ENOMEM
	case ENOMSG:
		return syscall.ENOMSG
	case ENOPROTOOPT:
		return syscall.ENOPROTOOPT
	case ENOSPC:
		return syscall.ENOSPC
	case ENOSYS:
		return syscall.ENOSYS
	case ENOTCONN:
		return syscall.ENOTCONN
	case ENOTDIR:
		return syscall.ENOTDIR
	case ENOTEMPTY:
		return syscall.ENOTEMPTY
	case ENOTRECOVERABLE:
		return syscall.ENOTRECOVERABLE
	case ENOTSOCK:
		return syscall.ENOTSOCK
	case ENOTSUP:
		return syscall.ENOTSUP
	case ENOTTY:
		return syscall.ENOTTY
	case ENXIO:
		return syscall.ENXIO
	case EOVERFLOW:
		return syscall.EOVERFLOW
	case EOWNERDEAD:
		return syscall.EOWNERDEAD
	case EPERM:
		return syscall.EPERM
	case EPIPE:
		return syscall.EPIPE
	case EPROTO:
		return syscall.EPROTO
	case EPROTONOSUPPORT:
		return syscall.EPROTONOSUPPORT
	case EPROTOTYPE:
		return syscall.EPROTOTYPE
	case ERANGE:
		return syscall.ERANGE
	case EROFS:
		return syscall.EROFS
	case ESPIPE:
		return syscall.ESPIPE
	case ESRCH:
		return syscall.ESRCH
	case ESTALE:
		return syscall.ESTALE
	case ETIMEDOUT:
		return syscall.ETIMEDOUT
	case ETXTBSY:
		return syscall.ETXTBSY
	case EXDEV:
		return syscall.EXDEV
	case ENOTCAPABLE:
		return syscall.EPERM
	default:
		panic("unsupport wasi errno: " + errno.Error())
	}
}
