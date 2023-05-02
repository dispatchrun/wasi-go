//go:build darwin

package wasiunix

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/stealthrocket/wasi"
	"golang.org/x/sys/unix"
)

const (
	__UTIME_NOW  = -1
	__UTIME_OMIT = -2
)

func prepareTimesAndAttrs(ts *[2]unix.Timespec) (attrs, size int, times [2]unix.Timespec) {
	const sizeOfTimespec = int(unsafe.Sizeof(times[0]))
	i := 0
	if ts[1].Nsec != __UTIME_OMIT {
		attrs |= unix.ATTR_CMN_MODTIME
		times[i] = ts[1]
		i++
	}
	if ts[0].Nsec != __UTIME_OMIT {
		attrs |= unix.ATTR_CMN_ACCTIME
		times[i] = ts[0]
		i++
	}
	return attrs, i * sizeOfTimespec, times
}

func futimens(fd int, ts *[2]unix.Timespec) error {
	attrs, size, times := prepareTimesAndAttrs(ts)
	attrlist := unix.Attrlist{
		Bitmapcount: unix.ATTR_BIT_MAP_COUNT,
		Commonattr:  uint32(attrs),
	}
	return fsetattrlist(fd, &attrlist, unsafe.Pointer(&times), size, 0)
}

func fsetattrlist(fd int, attrlist *unix.Attrlist, attrbuf unsafe.Pointer, attrbufsize int, options uint32) error {
	_, _, e := unix.Syscall6(
		uintptr(unix.SYS_FSETATTRLIST),
		uintptr(fd),
		uintptr(unsafe.Pointer(attrlist)),
		uintptr(attrbuf),
		uintptr(attrbufsize),
		uintptr(options),
		uintptr(0),
	)
	if e != 0 {
		return e
	}
	return nil
}

const minIovec = 8

func appendBytes(vecs []unix.Iovec, bs [][]byte) []unix.Iovec {
	for _, b := range bs {
		vecs = append(vecs, unix.Iovec{
			Base: unsafe.SliceData(b),
			Len:  uint64(len(b)),
		})
	}
	return vecs
}

func fdadvise(fd int, offset, length int64, advice wasi.Advice) error {
	// Since posix_fadvise is not available, just ignore the hint.
	return nil
}

func fallocate(fd int, offset, length int64) error {
	var sysStat unix.Stat_t
	if err := unix.Fstat(fd, &sysStat); err != nil {
		return err
	}
	if offset != sysStat.Size {
		return wasi.ENOSYS
	}
	err := unix.FcntlFstore(uintptr(fd), unix.F_PREALLOCATE, &unix.Fstore_t{
		Flags:   unix.F_ALLOCATEALL | unix.F_ALLOCATECONTIG,
		Posmode: unix.F_PEOFPOSMODE,
		Offset:  0,
		Length:  length,
	})
	if err != nil {
		return err
	}
	return unix.Ftruncate(fd, sysStat.Size+length)
}

func fdatasync(fd int) error {
	_, _, err := unix.Syscall(unix.SYS_FDATASYNC, uintptr(fd), 0, 0)
	if err != 0 {
		return err
	}
	return nil
}

func fsync(fd int) error {
	// See https://twitter.com/TigerBeetleDB/status/1422854887113732097
	_, err := unix.FcntlInt(uintptr(fd), unix.F_FULLFSYNC, 0)
	return err
}

func lseek(fd int, offset int64, whence int) (int64, error) {
	// TODO: there is an issue with unix.Seek where it returns random error
	// values for delta >= 2^32-1; syscall.Seek does not appear to suffer from
	// this problem, nor does using unix.Syscall directly.
	//
	// The standard syscall package uses a special syscallX function to call
	// lseek, which x/sys/unix does not, here is the reason (copied from
	// src/runtime/sys_darwin.go):
	//
	//  The X versions of syscall expect the libc call to return a 64-bit result.
	//  Otherwise (the non-X version) expects a 32-bit result.
	//  This distinction is required because an error is indicated by returning -1,
	//  and we need to know whether to check 32 or 64 bits of the result.
	//  (Some libc functions that return 32 bits put junk in the upper 32 bits of AX.)
	//
	// return unix.Seek(f.FD, int64(delta), sysWhence)
	return syscall.Seek(fd, offset, whence)
}

func readv(fd int, iovs [][]byte) (int, error) {
	iovecs := make([]unix.Iovec, 0, minIovec)
	iovecs = appendBytes(iovecs, iovs)
	n, _, err := unix.Syscall(
		uintptr(unix.SYS_READV),
		uintptr(fd),
		uintptr(unsafe.Pointer(unsafe.SliceData(iovecs))),
		uintptr(len(iovecs)),
	)
	if err != 0 {
		return int(n), err
	}
	return int(n), nil
}

func writev(fd int, iovs [][]byte) (int, error) {
	iovecs := make([]unix.Iovec, 0, minIovec)
	iovecs = appendBytes(iovecs, iovs)
	n, _, err := unix.Syscall(
		uintptr(unix.SYS_WRITEV),
		uintptr(fd),
		uintptr(unsafe.Pointer(unsafe.SliceData(iovecs))),
		uintptr(len(iovecs)),
	)
	if err != 0 {
		return int(n), err
	}
	return int(n), nil
}

func preadv(fd int, iovs [][]byte, offset int64) (int, error) {
	read := 0
	for _, iov := range iovs {
		n, err := unix.Pread(fd, iov, offset)
		offset += int64(n)
		read += n
		if err != nil {
			return read, err
		}
	}
	return read, nil
}

func pwritev(fd int, iovs [][]byte, offset int64) (int, error) {
	written := 0
	for _, iov := range iovs {
		n, err := unix.Pwrite(fd, iov, offset)
		offset += int64(n)
		written += n
		if err != nil {
			return written, err
		}
	}
	return written, nil
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
	case unix.EOPNOTSUPP:
		// There's no wasi.EOPNOTSUPP, but on Linux ENOTSUP==EOPNOTSUPP.
		return wasi.ENOTSUP
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

	// Omitted because there's no equivalent wasi.Errno:
	// case unix.EAUTH:
	// case unix.EBADARCH:
	// case unix.EBADEXEC:
	// case unix.EBADMACHO:
	// case unix.EBADRPC:
	// case unix.EDEVERR:
	// case unix.EFTYPE:
	// case unix.EHOSTDOWN:
	// case unix.ELAST:
	// case unix.ENEEDAUTH:
	// case unix.ENOATTR:
	// case unix.ENODATA:
	// case unix.ENOPOLICY:
	// case unix.ENOSR:
	// case unix.ENOSTR:
	// case unix.ENOTBLK:
	// case unix.EPFNOSUPPORT:
	// case unix.EPROCLIM:
	// case unix.EPROCUNAVAIL:
	// case unix.EPROGMISMATCH:
	// case unix.EPROGUNAVAIL:
	// case unix.EPWROFF:
	// case unix.EQFULL:
	// case unix.EUSERS:
	// case unix.EREMOTE:
	// case unix.ERPCMISMATCH:
	// case unix.ESHLIBVERS:
	// case unix.ESHUTDOWN:
	// case unix.ESOCKTNOSUPPORT:
	// case unix.ETIME:
	// case unix.ETOOMANYREFS:

	default:
		panic(fmt.Errorf("unexpected unix.Errno(%d): %v", int(err), err))
	}
}
