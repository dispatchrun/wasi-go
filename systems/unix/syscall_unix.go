package unix

import (
	"runtime/debug"
	"unsafe"

	"github.com/stealthrocket/wasi-go"
	"golang.org/x/sys/unix"
)

// This function is used to automtically retry syscalls when they return EINTR
// due to having handled a signal instead of executing. Despite defininig a
// EINTR constant and having proc_raise to trigger signals from the guest, WASI
// does not provide any mechanism for handling signals so masking those errors
// seems like a safer approach to ensure that guest applications will work the
// same regardless of the compiler being used.
func ignoreEINTR(f func() error) error {
	for {
		if err := f(); err != unix.EINTR {
			return err
		}
	}
}

func ignoreEINTR2[F func() (R, error), R any](f F) (R, error) {
	for {
		v, err := f()
		if err != unix.EINTR {
			return v, err
		}
	}
}

// This function is used to handle EINTR returned by syscalls like read(2)
// or write(2). Those syscalls are allowed to transfer partial payloads,
// in which case we silence EINTR and let the caller handle the continuation.
// The syscall is only retried if no data has been transfered at all.
func handleEINTR(f func() (int, error)) (int, error) {
	for {
		n, err := f()
		if err != unix.EINTR {
			return n, err
		}
		if n > 0 {
			return n, nil
		}
	}
}

func closeTraceEBADF(fd int) error {
	if fd < 0 {
		return unix.EBADF
	}
	err := unix.Close(fd)
	if err != nil {
		if err == unix.EBADF {
			println("DEBUG: close", fd, "=> EBADF")
			debug.PrintStack()
		}
	}
	return err
}

func makeErrno(err error) wasi.Errno {
	return wasi.MakeErrno(err)
}

func makeFileStat(s *unix.Stat_t) wasi.FileStat {
	return wasi.FileStat{
		FileType:   makeFileType(uint32(s.Mode)),
		Device:     wasi.Device(s.Dev),
		INode:      wasi.INode(s.Ino),
		NLink:      wasi.LinkCount(s.Nlink),
		Size:       wasi.FileSize(s.Size),
		AccessTime: wasi.Timestamp(s.Atim.Nano()),
		ModifyTime: wasi.Timestamp(s.Mtim.Nano()),
		ChangeTime: wasi.Timestamp(s.Ctim.Nano()),
	}
}

func makeFileType(mode uint32) wasi.FileType {
	switch mode & unix.S_IFMT { // see stat(2)
	case unix.S_IFCHR: // character special
		return wasi.CharacterDeviceType
	case unix.S_IFDIR: // directory
		return wasi.DirectoryType
	case unix.S_IFBLK: // block special
		return wasi.BlockDeviceType
	case unix.S_IFREG: // regular
		return wasi.RegularFileType
	case unix.S_IFLNK: // symbolic link
		return wasi.SymbolicLinkType
	case unix.S_IFSOCK: // socket
		return wasi.SocketStreamType // or wasi.SocketDGramType?
	default:
		// e.g. S_IFIFO, S_IFWHT
		return wasi.UnknownType
	}
}

var _ []byte = (wasi.IOVec)(nil)

func makeIOVecs(iovecs []wasi.IOVec) [][]byte {
	return *(*[][]byte)(unsafe.Pointer(&iovecs))
}
