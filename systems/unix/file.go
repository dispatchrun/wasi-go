package unix

import (
	"context"

	"github.com/stealthrocket/wasi-go"
	"golang.org/x/sys/unix"
)

type FD int

func (fd FD) FDAdvise(ctx context.Context, offset, length wasi.FileSize, advice wasi.Advice) wasi.Errno {
	err := ignoreEINTR(func() error { return fdadvise(int(fd), int64(offset), int64(length), advice) })
	return makeErrno(err)
}

func (fd FD) FDAllocate(ctx context.Context, offset, length wasi.FileSize) wasi.Errno {
	err := ignoreEINTR(func() error { return fallocate(int(fd), int64(offset), int64(length)) })
	return makeErrno(err)
}

func (fd FD) FDClose(ctx context.Context) wasi.Errno {
	// It's unclear what to do for EINTR on Linux, so do nothing and assume the
	// file descriptor has been closed.
	//
	// See:
	// - https://man7.org/linux/man-pages/man2/close.2.html
	// - https://lwn.net/Articles/576478/
	err := closeTraceEBADF(int(fd))
	return makeErrno(err)
}

func (fd FD) FDDataSync(ctx context.Context) wasi.Errno {
	err := ignoreEINTR(func() error { return fdatasync(int(fd)) })
	return makeErrno(err)
}

func (fd FD) FDStatSetFlags(ctx context.Context, flags wasi.FDFlags) wasi.Errno {
	fl, err := ignoreEINTR2(func() (int, error) {
		return unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	})
	if err != nil {
		return makeErrno(err)
	}
	if flags.Has(wasi.Append) {
		fl |= unix.O_APPEND
	} else {
		fl &^= unix.O_APPEND
	}
	if flags.Has(wasi.NonBlock) {
		fl |= unix.O_NONBLOCK
	} else {
		fl &^= unix.O_NONBLOCK
	}
	_, err = ignoreEINTR2(func() (int, error) {
		return unix.FcntlInt(uintptr(fd), unix.F_SETFL, fl)
	})
	return makeErrno(err)
}

func (fd FD) FDFileStatGet(ctx context.Context) (wasi.FileStat, wasi.Errno) {
	var sysStat unix.Stat_t
	if err := ignoreEINTR(func() error { return unix.Fstat(int(fd), &sysStat) }); err != nil {
		return wasi.FileStat{}, makeErrno(err)
	}
	stat := makeFileStat(&sysStat)
	return stat, wasi.ESUCCESS
}

func (fd FD) FDFileStatSetSize(ctx context.Context, size wasi.FileSize) wasi.Errno {
	err := ignoreEINTR(func() error { return unix.Ftruncate(int(fd), int64(size)) })
	return makeErrno(err)
}

func (fd FD) FDFileStatSetTimes(ctx context.Context, accessTime, modifyTime wasi.Timestamp, flags wasi.FSTFlags) wasi.Errno {
	ts := [2]unix.Timespec{
		{Nsec: __UTIME_OMIT},
		{Nsec: __UTIME_OMIT},
	}
	if flags.Has(wasi.AccessTime) {
		if flags.Has(wasi.AccessTimeNow) {
			ts[0] = unix.Timespec{Nsec: __UTIME_NOW}
		} else {
			ts[0] = unix.NsecToTimespec(int64(accessTime))
		}
	}
	if flags.Has(wasi.ModifyTime) {
		if flags.Has(wasi.ModifyTimeNow) {
			ts[1] = unix.Timespec{Nsec: __UTIME_NOW}
		} else {
			ts[1] = unix.NsecToTimespec(int64(modifyTime))
		}
	}
	err := ignoreEINTR(func() error { return futimens(int(fd), &ts) })
	return makeErrno(err)
}

func (fd FD) FDPread(ctx context.Context, iovecs []wasi.IOVec, offset wasi.FileSize) (wasi.Size, wasi.Errno) {
	n, err := handleEINTR(func() (int, error) { return preadv(int(fd), makeIOVecs(iovecs), int64(offset)) })
	return wasi.Size(n), makeErrno(err)
}

func (fd FD) FDPwrite(ctx context.Context, iovecs []wasi.IOVec, offset wasi.FileSize) (wasi.Size, wasi.Errno) {
	n, err := handleEINTR(func() (int, error) { return pwritev(int(fd), makeIOVecs(iovecs), int64(offset)) })
	return wasi.Size(n), makeErrno(err)
}

func (fd FD) FDRead(ctx context.Context, iovecs []wasi.IOVec) (wasi.Size, wasi.Errno) {
	n, err := handleEINTR(func() (int, error) { return readv(int(fd), makeIOVecs(iovecs)) })
	return wasi.Size(n), makeErrno(err)
}

func (fd FD) FDWrite(ctx context.Context, iovecs []wasi.IOVec) (wasi.Size, wasi.Errno) {
	n, err := handleEINTR(func() (int, error) { return writev(int(fd), makeIOVecs(iovecs)) })
	return wasi.Size(n), makeErrno(err)
}

func (fd FD) FDOpenDir(ctx context.Context) (wasi.Dir, wasi.Errno) {
	if _, err := ignoreEINTR2(func() (int64, error) {
		return lseek(int(fd), 0, 0)
	}); err != nil {
		return nil, makeErrno(err)
	}
	return &dirbuf{fd: int(fd)}, wasi.ESUCCESS
}

func (fd FD) FDSync(ctx context.Context) wasi.Errno {
	err := ignoreEINTR(func() error { return fsync(int(fd)) })
	return makeErrno(err)
}

func (fd FD) FDSeek(ctx context.Context, delta wasi.FileDelta, whence wasi.Whence) (wasi.FileSize, wasi.Errno) {
	var sysWhence int
	switch whence {
	case wasi.SeekStart:
		sysWhence = unix.SEEK_SET
	case wasi.SeekCurrent:
		sysWhence = unix.SEEK_CUR
	case wasi.SeekEnd:
		sysWhence = unix.SEEK_END
	default:
		return 0, wasi.EINVAL
	}
	off, err := ignoreEINTR2(func() (int64, error) { return lseek(int(fd), int64(delta), sysWhence) })
	return wasi.FileSize(off), makeErrno(err)
}

func (fd FD) PathCreateDirectory(ctx context.Context, path string) wasi.Errno {
	err := ignoreEINTR(func() error { return unix.Mkdirat(int(fd), path, 0755) })
	return makeErrno(err)
}

func (fd FD) PathFileStatGet(ctx context.Context, flags wasi.LookupFlags, path string) (wasi.FileStat, wasi.Errno) {
	var sysStat unix.Stat_t
	var sysFlags int
	if !flags.Has(wasi.SymlinkFollow) {
		sysFlags |= unix.AT_SYMLINK_NOFOLLOW
	}
	err := ignoreEINTR(func() error { return unix.Fstatat(int(fd), path, &sysStat, sysFlags) })
	return makeFileStat(&sysStat), makeErrno(err)
}

func (fd FD) PathFileStatSetTimes(ctx context.Context, lookupFlags wasi.LookupFlags, path string, accessTime, modifyTime wasi.Timestamp, fstFlags wasi.FSTFlags) wasi.Errno {
	var sysFlags int
	if !lookupFlags.Has(wasi.SymlinkFollow) {
		sysFlags |= unix.AT_SYMLINK_NOFOLLOW
	}
	ts := [2]unix.Timespec{
		{Nsec: __UTIME_OMIT},
		{Nsec: __UTIME_OMIT},
	}
	if fstFlags.Has(wasi.AccessTime) {
		if fstFlags.Has(wasi.AccessTimeNow) {
			ts[0] = unix.Timespec{Nsec: __UTIME_NOW}
		} else {
			ts[0] = unix.NsecToTimespec(int64(accessTime))
		}
	}
	if fstFlags.Has(wasi.ModifyTime) {
		if fstFlags.Has(wasi.ModifyTimeNow) {
			ts[1] = unix.Timespec{Nsec: __UTIME_NOW}
		} else {
			ts[1] = unix.NsecToTimespec(int64(modifyTime))
		}
	}
	err := ignoreEINTR(func() error { return unix.UtimesNanoAt(int(fd), path, ts[:], sysFlags) })
	return makeErrno(err)
}

func (fd FD) PathLink(ctx context.Context, flags wasi.LookupFlags, oldPath string, newDir FD, newPath string) wasi.Errno {
	var sysFlags int
	if flags.Has(wasi.SymlinkFollow) {
		sysFlags |= unix.AT_SYMLINK_FOLLOW
	}
	err := ignoreEINTR(func() error { return unix.Linkat(int(fd), oldPath, int(newDir), newPath, sysFlags) })
	return makeErrno(err)
}

func (fd FD) PathOpen(ctx context.Context, lookupFlags wasi.LookupFlags, path string, openFlags wasi.OpenFlags, rightsBase, rightsInheriting wasi.Rights, fdFlags wasi.FDFlags) (FD, wasi.Errno) {
	oflags := unix.O_CLOEXEC
	if openFlags.Has(wasi.OpenDirectory) {
		oflags |= unix.O_DIRECTORY
		rightsBase &= wasi.DirectoryRights
	}
	if openFlags.Has(wasi.OpenCreate) {
		oflags |= unix.O_CREAT
	}
	if openFlags.Has(wasi.OpenExclusive) {
		oflags |= unix.O_EXCL
	}
	if openFlags.Has(wasi.OpenTruncate) {
		oflags |= unix.O_TRUNC
	}
	if fdFlags.Has(wasi.Append) {
		oflags |= unix.O_APPEND
	}
	if fdFlags.Has(wasi.DSync) {
		oflags |= unix.O_DSYNC
	}
	if fdFlags.Has(wasi.Sync) {
		oflags |= unix.O_SYNC
	}
	if fdFlags.Has(wasi.RSync) {
		// O_RSYNC is not widely supported, and in many cases is an
		// alias for O_SYNC.
		oflags |= unix.O_SYNC
	}
	if fdFlags.Has(wasi.NonBlock) {
		oflags |= unix.O_NONBLOCK
	}
	if !lookupFlags.Has(wasi.SymlinkFollow) {
		oflags |= unix.O_NOFOLLOW
	}
	switch {
	case openFlags.Has(wasi.OpenDirectory):
		oflags |= unix.O_RDONLY
	case rightsBase.Has(wasi.FDReadRight) && rightsBase.Has(wasi.FDWriteRight):
		oflags |= unix.O_RDWR
	case rightsBase.Has(wasi.FDReadRight):
		oflags |= unix.O_RDONLY
	case rightsBase.Has(wasi.FDWriteRight):
		oflags |= unix.O_WRONLY
	default:
		oflags |= unix.O_RDONLY
	}

	mode := uint32(0644)
	if (oflags & unix.O_DIRECTORY) != 0 {
		mode = 0
	}
	hostfd, err := ignoreEINTR2(func() (int, error) {
		return unix.Openat(int(fd), path, oflags, mode)
	})
	return FD(hostfd), makeErrno(err)
}

func (fd FD) PathReadLink(ctx context.Context, path string, buffer []byte) (int, wasi.Errno) {
	n, err := ignoreEINTR2(func() (int, error) {
		return unix.Readlinkat(int(fd), path, buffer)
	})
	if err != nil {
		return n, makeErrno(err)
	} else if n == len(buffer) {
		return n, wasi.ERANGE
	} else {
		return n, wasi.ESUCCESS
	}
}

func (fd FD) PathRemoveDirectory(ctx context.Context, path string) wasi.Errno {
	err := ignoreEINTR(func() error { return unix.Unlinkat(int(fd), path, unix.AT_REMOVEDIR) })
	return makeErrno(err)
}

func (fd FD) PathRename(ctx context.Context, oldPath string, newDir FD, newPath string) wasi.Errno {
	err := ignoreEINTR(func() error { return unix.Renameat(int(fd), oldPath, int(newDir), newPath) })
	return makeErrno(err)
}

func (fd FD) PathSymlink(ctx context.Context, oldPath string, newPath string) wasi.Errno {
	err := ignoreEINTR(func() error { return unix.Symlinkat(oldPath, int(fd), newPath) })
	return makeErrno(err)
}

func (fd FD) PathUnlinkFile(ctx context.Context, path string) wasi.Errno {
	err := ignoreEINTR(func() error { return unix.Unlinkat(int(fd), path, 0) })
	return makeErrno(err)
}

func (d *dirbuf) FDReadDir(ctx context.Context, entries []wasi.DirEntry, cookie wasi.DirCookie, bufferSizeBytes int) (int, wasi.Errno) {
	n, err := d.readDirEntries(entries, cookie, bufferSizeBytes)
	return n, makeErrno(err)
}

func (d *dirbuf) FDCloseDir(ctx context.Context) wasi.Errno {
	return wasi.ESUCCESS
}
