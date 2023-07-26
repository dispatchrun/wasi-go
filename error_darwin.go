package wasi

import (
	"syscall"
)

func syscallErrnoToWASI(err syscall.Errno) Errno {
	switch err {
	case syscall.E2BIG:
		return E2BIG
	case syscall.EACCES:
		return EACCES
	case syscall.EADDRINUSE:
		return EADDRINUSE
	case syscall.EADDRNOTAVAIL:
		return EADDRNOTAVAIL
	case syscall.EAFNOSUPPORT:
		return EAFNOSUPPORT
	case syscall.EAGAIN:
		return EAGAIN
	case syscall.EALREADY:
		return EALREADY
	case syscall.EBADF:
		return EBADF
	case syscall.EBADMSG:
		return EBADMSG
	case syscall.EBUSY:
		return EBUSY
	case syscall.ECANCELED:
		return ECANCELED
	case syscall.ECHILD:
		return ECHILD
	case syscall.ECONNABORTED:
		return ECONNABORTED
	case syscall.ECONNREFUSED:
		return ECONNREFUSED
	case syscall.ECONNRESET:
		return ECONNRESET
	case syscall.EDEADLK:
		return EDEADLK
	case syscall.EDESTADDRREQ:
		return EDESTADDRREQ
	case syscall.EDOM:
		return EDOM
	case syscall.EDQUOT:
		return EDQUOT
	case syscall.EEXIST:
		return EEXIST
	case syscall.EFAULT:
		return EFAULT
	case syscall.EFBIG:
		return EFBIG
	case syscall.EHOSTUNREACH:
		return EHOSTUNREACH
	case syscall.EIDRM:
		return EIDRM
	case syscall.EILSEQ:
		return EILSEQ
	case syscall.EINPROGRESS:
		return EINPROGRESS
	case syscall.EINTR:
		return EINTR
	case syscall.EINVAL:
		return EINVAL
	case syscall.EIO:
		return EIO
	case syscall.EISCONN:
		return EISCONN
	case syscall.EISDIR:
		return EISDIR
	case syscall.ELOOP:
		return ELOOP
	case syscall.EMFILE:
		return EMFILE
	case syscall.EMLINK:
		return EMLINK
	case syscall.EMSGSIZE:
		return EMSGSIZE
	case syscall.EMULTIHOP:
		return EMULTIHOP
	case syscall.ENAMETOOLONG:
		return ENAMETOOLONG
	case syscall.ENETDOWN:
		return ENETDOWN
	case syscall.ENETRESET:
		return ENETRESET
	case syscall.ENETUNREACH:
		return ENETUNREACH
	case syscall.ENFILE:
		return ENFILE
	case syscall.ENOBUFS:
		return ENOBUFS
	case syscall.ENODEV:
		return ENODEV
	case syscall.ENOENT:
		return ENOENT
	case syscall.ENOEXEC:
		return ENOEXEC
	case syscall.ENOLCK:
		return ENOLCK
	case syscall.ENOLINK:
		return ENOLINK
	case syscall.ENOMEM:
		return ENOMEM
	case syscall.ENOMSG:
		return ENOMSG
	case syscall.ENOPROTOOPT:
		return ENOPROTOOPT
	case syscall.ENOSPC:
		return ENOSPC
	case syscall.ENOSYS:
		return ENOSYS
	case syscall.ENOTCONN:
		return ENOTCONN
	case syscall.ENOTDIR:
		return ENOTDIR
	case syscall.ENOTEMPTY:
		return ENOTEMPTY
	case syscall.ENOTRECOVERABLE:
		return ENOTRECOVERABLE
	case syscall.ENOTSOCK:
		return ENOTSOCK
	case syscall.ENOTSUP:
		return ENOTSUP
	case syscall.ENOTTY:
		return ENOTTY
	case syscall.ENXIO:
		return ENXIO
	case syscall.EOPNOTSUPP:
		// There's no EOPNOTSUPP, but on Linux ENOTSUP==EOPNOTSUPP.
		return ENOTSUP
	case syscall.EOVERFLOW:
		return EOVERFLOW
	case syscall.EOWNERDEAD:
		return EOWNERDEAD
	case syscall.EPERM:
		return EPERM
	case syscall.EPIPE:
		return EPIPE
	case syscall.EPROTO:
		return EPROTO
	case syscall.EPROTONOSUPPORT:
		return EPROTONOSUPPORT
	case syscall.EPROTOTYPE:
		return EPROTOTYPE
	case syscall.ERANGE:
		return ERANGE
	case syscall.EROFS:
		return EROFS
	case syscall.ESPIPE:
		return ESPIPE
	case syscall.ESRCH:
		return ESRCH
	case syscall.ESTALE:
		return ESTALE
	case syscall.ETIMEDOUT:
		return ETIMEDOUT
	case syscall.ETXTBSY:
		return ETXTBSY
	case syscall.EXDEV:
		return EXDEV

	// Omitted because they're duplicates:
	// case syscall.EWOULDBLOCK: (EAGAIN)

	// Omitted because there's no equivalent Errno:
	// case syscall.EAUTH:
	// case syscall.EBADARCH:
	// case syscall.EBADEXEC:
	// case syscall.EBADMACHO:
	// case syscall.EBADRPC:
	// case syscall.EDEVERR:
	// case syscall.EFTYPE:
	// case syscall.EHOSTDOWN:
	// case syscall.ELAST:
	// case syscall.ENEEDAUTH:
	// case syscall.ENOATTR:
	// case syscall.ENODATA:
	// case syscall.ENOPOLICY:
	// case syscall.ENOSR:
	// case syscall.ENOSTR:
	// case syscall.ENOTBLK:
	// case syscall.EPFNOSUPPORT:
	// case syscall.EPROCLIM:
	// case syscall.EPROCUNAVAIL:
	// case syscall.EPROGMISMATCH:
	// case syscall.EPROGUNAVAIL:
	// case syscall.EPWROFF:
	// case syscall.EQFULL:
	// case syscall.EUSERS:
	// case syscall.EREMOTE:
	// case syscall.ERPCMISMATCH:
	// case syscall.ESHLIBVERS:
	// case syscall.ESHUTDOWN:
	// case syscall.ESOCKTNOSUPPORT:
	// case syscall.ETIME:
	// case syscall.ETOOMANYREFS:

	default:
		panic("unsupported syscall errno: " + err.Error())
	}
}
