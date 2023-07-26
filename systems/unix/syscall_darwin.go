package unix

import (
	"syscall"
	"unsafe"

	"github.com/stealthrocket/wasi-go"
	"golang.org/x/sys/unix"
)

func accept(socket, flags int) (int, unix.Sockaddr, error) {
	conn, addr, err := acceptCloseOnExec(socket)
	if err != nil {
		return -1, addr, err
	}
	if (flags & unix.O_NONBLOCK) != 0 {
		if err := unix.SetNonblock(conn, true); err != nil {
			closeTraceEBADF(conn)
			return -1, addr, err
		}
	}
	return conn, addr, nil
}

func acceptCloseOnExec(socket int) (int, unix.Sockaddr, error) {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()
	// This must only be called on non-blocking sockets or we may prevent
	// other goroutines from spawning processes.
	conn, addr, err := unix.Accept(socket)
	if err != nil {
		return -1, addr, err
	}
	unix.CloseOnExec(conn)
	return conn, addr, nil
}

func pipe(fds []int, flags int) error {
	if err := pipeCloseOnExec(fds); err != nil {
		return err
	}
	if (flags & unix.O_NONBLOCK) != 0 {
		if err := unix.SetNonblock(fds[1], true); err != nil {
			closePipe(fds)
			return err
		}
		if err := unix.SetNonblock(fds[0], true); err != nil {
			closePipe(fds)
			return err
		}
	}
	return nil
}

func pipeCloseOnExec(fds []int) error {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()

	if err := unix.Pipe(fds); err != nil {
		return err
	}
	unix.CloseOnExec(fds[0])
	unix.CloseOnExec(fds[1])
	return nil
}

func closePipe(fds []int) {
	closeTraceEBADF(fds[1])
	closeTraceEBADF(fds[0])
	fds[0] = -1
	fds[1] = -1
}

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
	// Note: there is an issue with unix.Seek where it returns random error
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

func getsocketdomain(fd int) (int, error) {
	return 0, unix.ENOSYS
}
