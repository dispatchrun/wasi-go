package unix

import (
	"context"
	"errors"
	"fmt"
	"unsafe"

	"github.com/stealthrocket/wasi-go"
	"golang.org/x/sys/unix"
)

func makeErrno(err error) wasi.Errno {
	if err == nil {
		return wasi.ESUCCESS
	}
	if err == context.Canceled {
		return wasi.ECANCELED
	}
	var sysErrno unix.Errno
	if errors.As(err, &sysErrno) {
		if sysErrno == 0 {
			return wasi.ESUCCESS
		}
		return syscallErrnoToWASI(sysErrno)
	}
	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) {
		if timeout.Timeout() {
			return wasi.ETIMEDOUT
		}
	}
	panic(fmt.Errorf("unexpected error: %v", err))
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
