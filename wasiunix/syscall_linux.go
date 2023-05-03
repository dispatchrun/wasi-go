//go:build linux

package wasiunix

import (
	"fmt"
	"unsafe"

	"github.com/stealthrocket/wasi"
	"golang.org/x/sys/unix"
)

func accept(socket, flags int) (int, unix.Sockaddr, error) {
	return unix.Accep4(socket, flags|unix.O_CLOEXEC)
}

func pipe(fds []int, flags int) error {
	return unix.Pipe2(fds, flags|unix.O_CLOEXEC)
}

func futimens(fd int, ts *[2]unix.Timespec) error {
	// https://github.com/bminor/glibc/blob/master/sysdeps/unix/sysv/linux/futimens.c
	_, _, err := unix.Syscall6(
		uintptr(unix.SYS_UTIMENSAT),
		uintptr(fd),
		uintptr(0), // path=NULL
		uintptr(unsafe.Pointer(ts)),
		uintptr(0),
		uintptr(0),
		uintptr(0),
	)
	if err != 0 {
		return err
	}
	return nil
}

func fdadvise(fd int, offset, length int64, advice wasi.Advice) error {
	var sysAdvice int
	switch advice {
	case wasi.Normal:
		sysAdvice = unix.FADV_NORMAL
	case wasi.Sequential:
		sysAdvice = unix.FADV_SEQUENTIAL
	case wasi.Random:
		sysAdvice = unix.FADV_RANDOM
	case wasi.WillNeed:
		sysAdvice = unix.FADV_WILLNEED
	case wasi.DontNeed:
		sysAdvice = unix.FADV_DONTNEED
	case wasi.NoReuse:
		sysAdvice = unix.FADV_NOREUSE
	default:
		return wasi.EINVAL
	}
	return unix.Fadvise(fd, offset, length, sysAdvice)
}

func fallocate(fd int, offset, length int64) error {
	return unix.Fallocate(fd, 0, offset, length)
}

func fdatasync(fd int) error {
	return unix.Fdatasync(fd)
}

func fsync(fd int) error {
	return unix.Fsync(fd)
}

func lseek(fd int, offset int64, whence int) (int64, error) {
	return unix.Seek(fd, offset, whence)
}

func readv(fd int, iovs [][]byte) (int, error) {
	return unix.Readv(fd, iovs)
}

func writev(fd int, iovs [][]byte) (int, error) {
	return unix.Writev(fd, iovs)
}

func preadv(fd int, iovs [][]byte, offset int64) (int, error) {
	return unix.Preadv(fd, iovs, offset)
}

func pwritev(fd int, iovs [][]byte, offset int64) (int, error) {
	return unix.Pwritev(fd, iovs, offset)
}

func syscallErrnoToWASI(err unix.Errno) wasi.Errno {
	switch err {
	case unix.E2BIG:
		return wasi.E2BIG
	case unix.EACCES:
		return wasi.EACCES
	case unix.EADDRINUSE:
		return wasi.EADDRINUSE
	case unix.EADDRNOTAVAIL:
		return wasi.EADDRNOTAVAIL
	case unix.EAFNOSUPPORT:
		return wasi.EAFNOSUPPORT
	case unix.EAGAIN:
		return wasi.EAGAIN
	case unix.EALREADY:
		return wasi.EALREADY
	case unix.EBADF:
		return wasi.EBADF
	case unix.EBADMSG:
		return wasi.EBADMSG
	case unix.EBUSY:
		return wasi.EBUSY
	case unix.ECANCELED:
		return wasi.ECANCELED
	case unix.ECHILD:
		return wasi.ECHILD
	case unix.ECONNABORTED:
		return wasi.ECONNABORTED
	case unix.ECONNREFUSED:
		return wasi.ECONNREFUSED
	case unix.ECONNRESET:
		return wasi.ECONNRESET
	case unix.EDEADLK:
		return wasi.EDEADLK
	case unix.EDESTADDRREQ:
		return wasi.EDESTADDRREQ
	case unix.EDOM:
		return wasi.EDOM
	case unix.EDQUOT:
		return wasi.EDQUOT
	case unix.EEXIST:
		return wasi.EEXIST
	case unix.EFAULT:
		return wasi.EFAULT
	case unix.EFBIG:
		return wasi.EFBIG
	case unix.EHOSTUNREACH:
		return wasi.EHOSTUNREACH
	case unix.EIDRM:
		return wasi.EIDRM
	case unix.EILSEQ:
		return wasi.EILSEQ
	case unix.EINPROGRESS:
		return wasi.EINPROGRESS
	case unix.EINTR:
		return wasi.EINTR
	case unix.EINVAL:
		return wasi.EINVAL
	case unix.EIO:
		return wasi.EIO
	case unix.EISCONN:
		return wasi.EISCONN
	case unix.EISDIR:
		return wasi.EISDIR
	case unix.ELOOP:
		return wasi.ELOOP
	case unix.EMFILE:
		return wasi.EMFILE
	case unix.EMLINK:
		return wasi.EMLINK
	case unix.EMSGSIZE:
		return wasi.EMSGSIZE
	case unix.EMULTIHOP:
		return wasi.EMULTIHOP
	case unix.ENAMETOOLONG:
		return wasi.ENAMETOOLONG
	case unix.ENETDOWN:
		return wasi.ENETDOWN
	case unix.ENETRESET:
		return wasi.ENETRESET
	case unix.ENETUNREACH:
		return wasi.ENETUNREACH
	case unix.ENFILE:
		return wasi.ENFILE
	case unix.ENOBUFS:
		return wasi.ENOBUFS
	case unix.ENODEV:
		return wasi.ENODEV
	case unix.ENOENT:
		return wasi.ENOENT
	case unix.ENOEXEC:
		return wasi.ENOEXEC
	case unix.ENOLCK:
		return wasi.ENOLCK
	case unix.ENOLINK:
		return wasi.ENOLINK
	case unix.ENOMEM:
		return wasi.ENOMEM
	case unix.ENOMSG:
		return wasi.ENOMSG
	case unix.ENOPROTOOPT:
		return wasi.ENOPROTOOPT
	case unix.ENOSPC:
		return wasi.ENOSPC
	case unix.ENOSYS:
		return wasi.ENOSYS
	case unix.ENOTCONN:
		return wasi.ENOTCONN
	case unix.ENOTDIR:
		return wasi.ENOTDIR
	case unix.ENOTEMPTY:
		return wasi.ENOTEMPTY
	case unix.ENOTRECOVERABLE:
		return wasi.ENOTRECOVERABLE
	case unix.ENOTSOCK:
		return wasi.ENOTSOCK
	case unix.ENOTSUP:
		return wasi.ENOTSUP
	case unix.ENOTTY:
		return wasi.ENOTTY
	case unix.ENXIO:
		return wasi.ENXIO
	case unix.EOVERFLOW:
		return wasi.EOVERFLOW
	case unix.EOWNERDEAD:
		return wasi.EOWNERDEAD
	case unix.EPERM:
		return wasi.EPERM
	case unix.EPIPE:
		return wasi.EPIPE
	case unix.EPROTO:
		return wasi.EPROTO
	case unix.EPROTONOSUPPORT:
		return wasi.EPROTONOSUPPORT
	case unix.EPROTOTYPE:
		return wasi.EPROTOTYPE
	case unix.ERANGE:
		return wasi.ERANGE
	case unix.EROFS:
		return wasi.EROFS
	case unix.ESPIPE:
		return wasi.ESPIPE
	case unix.ESRCH:
		return wasi.ESRCH
	case unix.ESTALE:
		return wasi.ESTALE
	case unix.ETIMEDOUT:
		return wasi.ETIMEDOUT
	case unix.ETXTBSY:
		return wasi.ETXTBSY
	case unix.EXDEV:
		return wasi.EXDEV

	// Omitted because they're duplicates:
	// case unix.EWOULDBLOCK: (EAGAIN)
	// case unix.EDEADLOCK: (EDEADLK)
	// case unix.EOPNOTSUPP: (ENOTSUP)

	// Omitted because there's no equivalent wasi.Errno:
	// case unix.EADV:
	// case unix.EBADE:
	// case unix.EBADFD:
	// case unix.EBADR:
	// case unix.EBADRQC:
	// case unix.EBADSLT:
	// case unix.EBFONT:
	// case unix.ECHRNG:
	// case unix.ECOMM:
	// case unix.EDOTDOT:
	// case unix.EHOSTDOWN:
	// case unix.EHWPOISON:
	// case unix.EISNAM:
	// case unix.EKEYEXPIRED:
	// case unix.EKEYREJECTED:
	// case unix.EKEYREVOKED:
	// case unix.EL2HLT:
	// case unix.EL2NSYNC:
	// case unix.EL3HLT:
	// case unix.EL3RST:
	// case unix.ELIBACC:
	// case unix.ELIBBAD:
	// case unix.ELIBEXEC:
	// case unix.ELIBMAX:
	// case unix.ELIBSCN:
	// case unix.ELNRNG:
	// case unix.EMEDIUMTYPE:
	// case unix.ENAVAIL:
	// case unix.ENOANO:
	// case unix.ENOCSI:
	// case unix.ENODATA:
	// case unix.ENOKEY:
	// case unix.ENOMEDIUM:
	// case unix.ENONET:
	// case unix.ENOPKG:
	// case unix.ENOSR:
	// case unix.ENOSTR:
	// case unix.ENOTBLK:
	// case unix.ENOTNAM:
	// case unix.ENOTUNIQ:
	// case unix.EPFNOSUPPORT:
	// case unix.EREMCHG:
	// case unix.EREMOTE:
	// case unix.EREMOTEIO:
	// case unix.ERESTART:
	// case unix.ERFKILL:
	// case unix.ESHUTDOWN:
	// case unix.ESOCKTNOSUPPORT:
	// case unix.ESRMNT:
	// case unix.ESTRPIPE:
	// case unix.ETIME:
	// case unix.ETOOMANYREFS:
	// case unix.EUCLEAN:
	// case unix.EUNATCH:
	// case unix.EUSERS:
	// case unix.EXFULL:

	default:
		panic(fmt.Errorf("unexpected unix.Errno(%d): %v", int(err), err))
	}
}
