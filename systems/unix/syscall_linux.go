package unix

import (
	"unsafe"

	"github.com/stealthrocket/wasi-go"
	"golang.org/x/sys/unix"
)

const (
	__UTIME_NOW  = unix.UTIME_NOW
	__UTIME_OMIT = unix.UTIME_OMIT
)

func accept(socket, flags int) (int, unix.Sockaddr, error) {
	return unix.Accept4(socket, flags|unix.O_CLOEXEC)
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
