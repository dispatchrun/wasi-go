package wasiunix

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/stealthrocket/wasi"
	"github.com/stealthrocket/wasi/wasiunix/internal/descriptor"
	"golang.org/x/sys/unix"
)

// Provider is a WASI preview 1 implementation for Unix systems.
//
// It implements the wasi.Provider interface.
//
// The provider is not safe for concurrent use.
type Provider struct {
	// Args are the environment variables accessible via ArgsGet.
	Args []string

	// Environ is the environment variables accessible via EnvironGet.
	Environ []string

	// Realtime returns the realtime clock value.
	Realtime          func() uint64
	RealtimePrecision time.Duration

	// Monotonic returns the monotonic clock value.
	Monotonic          func() uint64
	MonotonicPrecision time.Duration

	// Yield is called when SchedYield is called. If Yield is nil,
	// SchedYield is a noop.
	Yield func()

	// Exit is called with an exit code when ProcExit is called.
	// If Exit is nil, ProcExit is a noop.
	Exit func(int)

	// Raise is called with a signal when ProcRaise is called.
	// If Raise is nil, ProcRaise is a noop.
	Raise func(int)

	// Rand is the source for RandomGet.
	Rand io.Reader

	fds descriptor.Table[wasi.FD, *fdinfo]

	pollfds []unix.PollFd
}

type fdinfo struct {
	// Path is the path of the file.
	Path string

	// FD is the underlying OS file descriptor.
	FD int

	// FDStat is cached information about the file descriptor.
	FDStat wasi.FDStat

	// DirEntries are cached directory entries.
	DirEntries []os.DirEntry
}

func (p *Provider) lookupFD(guestfd wasi.FD, rights wasi.Rights) (*fdinfo, wasi.Errno) {
	f, ok := p.fds.Lookup(guestfd)
	if !ok {
		return nil, wasi.EBADF
	}
	if !f.FDStat.RightsBase.Has(rights) {
		return nil, wasi.ENOTCAPABLE
	}
	return f, wasi.ESUCCESS
}

func (p *Provider) lookupPreopenFD(guestfd wasi.FD, rights wasi.Rights) (*fdinfo, wasi.Errno) {
	f, errno := p.lookupFD(guestfd, rights)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	// TODO: check that it's a preopen
	if f.FDStat.FileType != wasi.DirectoryType {
		return nil, wasi.ENOTDIR
	}
	return f, wasi.ESUCCESS
}

func (p *Provider) lookupSocketFD(guestfd wasi.FD, rights wasi.Rights) (*fdinfo, wasi.Errno) {
	f, errno := p.lookupFD(guestfd, rights)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	switch f.FDStat.FileType {
	case wasi.SocketStreamType, wasi.SocketDGramType:
		return f, wasi.ESUCCESS
	default:
		return nil, wasi.ENOTSOCK
	}
}

// RegisterFD registers an open file.
func (p *Provider) RegisterFD(hostfd int, path string, fdstat wasi.FDStat) wasi.FD {
	fdstat.RightsBase &= wasi.AllRights
	fdstat.RightsInheriting &= wasi.AllRights
	return p.fds.Insert(&fdinfo{
		Path:   path,
		FD:     hostfd,
		FDStat: fdstat,
	})
}

func (p *Provider) ArgsGet() ([]string, wasi.Errno) {
	return p.Args, wasi.ESUCCESS
}

func (p *Provider) EnvironGet() ([]string, wasi.Errno) {
	return p.Environ, wasi.ESUCCESS
}

func (p *Provider) ClockResGet(id wasi.ClockID) (wasi.Timestamp, wasi.Errno) {
	switch id {
	case wasi.Realtime:
		return wasi.Timestamp(p.RealtimePrecision), wasi.ESUCCESS
	case wasi.Monotonic:
		return wasi.Timestamp(p.MonotonicPrecision), wasi.ESUCCESS
	case wasi.ProcessCPUTimeID, wasi.ThreadCPUTimeID:
		return 0, wasi.ENOTSUP
	default:
		return 0, wasi.EINVAL
	}
}

func (p *Provider) ClockTimeGet(id wasi.ClockID, precision wasi.Timestamp) (wasi.Timestamp, wasi.Errno) {
	switch id {
	case wasi.Realtime:
		if p.Realtime == nil {
			return 0, wasi.ENOTSUP
		}
		return wasi.Timestamp(p.Realtime()), wasi.ESUCCESS
	case wasi.Monotonic:
		if p.Monotonic == nil {
			return 0, wasi.ENOTSUP
		}
		return wasi.Timestamp(p.Monotonic()), wasi.ESUCCESS
	case wasi.ProcessCPUTimeID, wasi.ThreadCPUTimeID:
		return 0, wasi.ENOTSUP
	default:
		return 0, wasi.EINVAL
	}
}

func (p *Provider) FDAdvise(fd wasi.FD, offset wasi.FileSize, length wasi.FileSize, advice wasi.Advice) wasi.Errno {
	f, errno := p.lookupFD(fd, wasi.FDAdviseRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := fdadvise(f.FD, int64(offset), int64(length), advice)
	return makeErrno(err)
}

func (p *Provider) FDAllocate(fd wasi.FD, offset wasi.FileSize, length wasi.FileSize) wasi.Errno {
	f, errno := p.lookupFD(fd, wasi.FDAllocateRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := fallocate(f.FD, int64(offset), int64(length))
	return makeErrno(err)
}

func (p *Provider) FDClose(fd wasi.FD) wasi.Errno {
	// TODO: don't allow closing of preopens
	f, errno := p.lookupFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	p.fds.Delete(fd)
	err := unix.Close(f.FD)
	return makeErrno(err)
}

func (p *Provider) FDDataSync(fd wasi.FD) wasi.Errno {
	f, errno := p.lookupFD(fd, wasi.FDDataSyncRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := fdatasync(f.FD)
	return makeErrno(err)
}

func (p *Provider) FDStatGet(fd wasi.FD) (wasi.FDStat, wasi.Errno) {
	f, errno := p.lookupFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return wasi.FDStat{}, errno
	}
	return f.FDStat, wasi.ESUCCESS
}

func (p *Provider) FDStatSetFlags(fd wasi.FD, flags wasi.FDFlags) wasi.Errno {
	f, errno := p.lookupFD(fd, wasi.FDStatSetFlagsRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	changes := flags ^ f.FDStat.Flags
	if changes == 0 {
		return wasi.ESUCCESS
	}
	if changes.Has(wasi.Sync | wasi.DSync | wasi.RSync) {
		return wasi.ENOSYS // TODO: support changing {Sync,DSync,Rsync}
	}
	fl, err := unix.FcntlInt(uintptr(f.FD), unix.F_GETFL, 0)
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
	if _, err := unix.FcntlInt(uintptr(f.FD), unix.F_SETFL, fl); err != nil {
		return makeErrno(err)
	}
	f.FDStat.Flags ^= changes
	return wasi.ESUCCESS
}

func (p *Provider) FDStatSetRights(fd wasi.FD, rightsBase, rightsInheriting wasi.Rights) wasi.Errno {
	f, errno := p.lookupFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	// Rights can only be preserved or removed, not added.
	rightsBase &= wasi.AllRights
	rightsInheriting &= wasi.AllRights
	if (rightsBase &^ f.FDStat.RightsBase) != 0 {
		return wasi.ENOTCAPABLE
	}
	if (rightsInheriting &^ f.FDStat.RightsInheriting) != 0 {
		return wasi.ENOTCAPABLE
	}
	f.FDStat.RightsBase &= rightsBase
	f.FDStat.RightsInheriting &= rightsInheriting
	return wasi.ESUCCESS
}

func (p *Provider) FDFileStatGet(fd wasi.FD) (wasi.FileStat, wasi.Errno) {
	f, errno := p.lookupFD(fd, wasi.FDFileStatGetRight)
	if errno != wasi.ESUCCESS {
		return wasi.FileStat{}, errno
	}
	var sysStat unix.Stat_t
	if err := unix.Fstat(f.FD, &sysStat); err != nil {
		return wasi.FileStat{}, makeErrno(err)
	}
	stat := makeFileStat(&sysStat)
	switch f.FD {
	case syscall.Stdin, syscall.Stdout, syscall.Stderr:
		// Override stdio size/times.
		stat.Size = 0
		stat.AccessTime = 0
		stat.ModifyTime = 0
		stat.ChangeTime = 0
	}
	return stat, wasi.ESUCCESS
}

func (p *Provider) FDFileStatSetSize(fd wasi.FD, size wasi.FileSize) wasi.Errno {
	f, errno := p.lookupFD(fd, wasi.FDFileStatSetSizeRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Ftruncate(f.FD, int64(size))
	return makeErrno(err)
}

func (p *Provider) FDFileStatSetTimes(fd wasi.FD, accessTime, modifyTime wasi.Timestamp, flags wasi.FSTFlags) wasi.Errno {
	f, errno := p.lookupFD(fd, wasi.FDFileStatSetTimesRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	var sysStat unix.Stat_t
	if err := unix.Fstat(f.FD, &sysStat); err != nil {
		return makeErrno(err)
	}
	ts := [2]unix.Timespec{sysStat.Atim, sysStat.Mtim}
	if flags.Has(wasi.AccessTimeNow) || flags.Has(wasi.ModifyTimeNow) {
		if p.Monotonic == nil {
			return wasi.ENOSYS
		}
		now := p.Monotonic()
		if flags.Has(wasi.AccessTimeNow) {
			accessTime = wasi.Timestamp(now)
		}
		if flags.Has(wasi.ModifyTimeNow) {
			modifyTime = wasi.Timestamp(now)
		}
	}
	if flags.Has(wasi.AccessTime) || flags.Has(wasi.AccessTimeNow) {
		ts[0] = unix.NsecToTimespec(int64(accessTime))
	}
	if flags.Has(wasi.ModifyTime) || flags.Has(wasi.ModifyTimeNow) {
		ts[1] = unix.NsecToTimespec(int64(modifyTime))
	}
	err := futimens(f.FD, &ts)
	return makeErrno(err)
}

func (p *Provider) FDPread(fd wasi.FD, iovecs []wasi.IOVec, offset wasi.FileSize) (wasi.Size, wasi.Errno) {
	f, errno := p.lookupFD(fd, wasi.FDReadRight|wasi.FDSeekRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := preadv(f.FD, makeIOVecs(iovecs), int64(offset))
	return wasi.Size(n), makeErrno(err)
}

func (p *Provider) FDPreStatGet(fd wasi.FD) (wasi.PreStat, wasi.Errno) {
	f, errno := p.lookupPreopenFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return wasi.PreStat{}, errno
	}
	// TODO: error if the file is not a preopen
	stat := wasi.PreStat{
		Type: wasi.PreOpenDir,
		PreStatDir: wasi.PreStatDir{
			NameLength: wasi.Size(len(f.Path)),
		},
	}
	return stat, wasi.ESUCCESS
}

func (p *Provider) FDPreStatDirName(fd wasi.FD) (string, wasi.Errno) {
	f, errno := p.lookupPreopenFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return "", errno
	}
	return f.Path, wasi.ESUCCESS
}

func (p *Provider) FDPwrite(fd wasi.FD, iovecs []wasi.IOVec, offset wasi.FileSize) (wasi.Size, wasi.Errno) {
	f, errno := p.lookupFD(fd, wasi.FDWriteRight|wasi.FDSeekRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := pwritev(f.FD, makeIOVecs(iovecs), int64(offset))
	return wasi.Size(n), makeErrno(err)
}

func (p *Provider) FDRead(fd wasi.FD, iovecs []wasi.IOVec) (wasi.Size, wasi.Errno) {
	f, errno := p.lookupFD(fd, wasi.FDReadRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := readv(f.FD, makeIOVecs(iovecs))
	return wasi.Size(n), makeErrno(err)
}

func (p *Provider) FDReadDir(fd wasi.FD, buffer []wasi.DirEntryName, bufferSizeBytes int, cookie wasi.DirCookie) ([]wasi.DirEntryName, wasi.Errno) {
	f, errno := p.lookupFD(fd, wasi.FDReadDirRight)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}

	// TODO: use a readdir iterator
	// This is all very tricky to get right, so let's cheat for now
	// and use os.ReadDir.
	if cookie == 0 {
		entries, err := os.ReadDir(f.Path)
		if err != nil {
			return buffer, makeErrno(err)
		}
		f.DirEntries = entries
		// Add . and .. entries, since they're stripped by os.ReadDir
		if info, err := os.Stat(f.Path); err == nil {
			f.DirEntries = append(f.DirEntries, &statDirEntry{".", info})
		}
		if info, err := os.Stat(filepath.Join(f.Path, "..")); err == nil {
			f.DirEntries = append(f.DirEntries, &statDirEntry{"..", info})
		}
	}
	if cookie > math.MaxInt {
		return buffer, wasi.EINVAL
	}
	var n int
	pos := int(cookie)
	for ; pos < len(f.DirEntries) && n < bufferSizeBytes; pos++ {
		e := f.DirEntries[pos]
		name := e.Name()
		info, err := e.Info()
		if err != nil {
			return buffer, makeErrno(err)
		}
		s := info.Sys().(*syscall.Stat_t)
		buffer = append(buffer, wasi.DirEntryName{
			Entry: wasi.DirEntry{
				Type:       makeFileType(uint32(s.Mode)),
				INode:      wasi.INode(s.Ino),
				NameLength: wasi.DirNameLength(len(name)),
				Next:       wasi.DirCookie(pos + 1),
			},
			Name: name,
		})
		n += int(unsafe.Sizeof(wasi.DirEntry{})) + len(name)
	}
	return buffer, wasi.ESUCCESS
}

func (p *Provider) FDRenumber(from, to wasi.FD) wasi.Errno {
	f, errno := p.lookupFD(from, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	// TODO: limit max file descriptor number
	f, replaced := p.fds.Assign(to, f)
	if replaced {
		unix.Close(f.FD)
	}
	return wasi.ENOSYS
}

func (p *Provider) FDSync(fd wasi.FD) wasi.Errno {
	f, errno := p.lookupFD(fd, wasi.FDSyncRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := fsync(f.FD)
	return makeErrno(err)
}

func (p *Provider) FDSeek(fd wasi.FD, delta wasi.FileDelta, whence wasi.Whence) (wasi.FileSize, wasi.Errno) {
	return p.fdseek(fd, wasi.FDSeekRight, delta, whence)
}

func (p *Provider) FDTell(fd wasi.FD) (wasi.FileSize, wasi.Errno) {
	return p.fdseek(fd, wasi.FDTellRight, 0, wasi.SeekCurrent)
}

func (p *Provider) fdseek(fd wasi.FD, rights wasi.Rights, delta wasi.FileDelta, whence wasi.Whence) (wasi.FileSize, wasi.Errno) {
	f, errno := p.lookupFD(fd, rights)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
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
	off, err := lseek(f.FD, int64(delta), sysWhence)
	return wasi.FileSize(off), makeErrno(err)
}

func (p *Provider) FDWrite(fd wasi.FD, iovecs []wasi.IOVec) (wasi.Size, wasi.Errno) {
	f, errno := p.lookupFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := writev(f.FD, makeIOVecs(iovecs))
	return wasi.Size(n), makeErrno(err)
}

func (p *Provider) PathCreateDirectory(fd wasi.FD, path string) wasi.Errno {
	d, errno := p.lookupFD(fd, wasi.PathCreateDirectoryRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Mkdirat(d.FD, path, 0755)
	return makeErrno(err)
}

func (p *Provider) PathFileStatGet(fd wasi.FD, flags wasi.LookupFlags, path string) (wasi.FileStat, wasi.Errno) {
	d, errno := p.lookupFD(fd, wasi.PathFileStatGetRight)
	if errno != wasi.ESUCCESS {
		return wasi.FileStat{}, errno
	}
	var sysStat unix.Stat_t
	var sysFlags int
	if !flags.Has(wasi.SymlinkFollow) {
		sysFlags |= unix.AT_SYMLINK_NOFOLLOW
	}
	err := unix.Fstatat(d.FD, path, &sysStat, sysFlags)
	return makeFileStat(&sysStat), makeErrno(err)
}

func (p *Provider) PathFileStatSetTimes(fd wasi.FD, lookupFlags wasi.LookupFlags, path string, accessTime, modifyTime wasi.Timestamp, fstFlags wasi.FSTFlags) wasi.Errno {
	d, errno := p.lookupFD(fd, wasi.PathFileStatSetTimesRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	if fstFlags.Has(wasi.AccessTimeNow) || fstFlags.Has(wasi.ModifyTimeNow) {
		now := wasi.Timestamp(time.Now().UnixNano())
		if fstFlags.Has(wasi.AccessTimeNow) {
			accessTime = now
		}
		if fstFlags.Has(wasi.ModifyTimeNow) {
			modifyTime = now
		}
	}
	var sysFlags int
	if !lookupFlags.Has(wasi.SymlinkFollow) {
		sysFlags |= unix.AT_SYMLINK_NOFOLLOW
	}
	var ts [2]unix.Timespec
	changeAccessTime := fstFlags.Has(wasi.AccessTime) || fstFlags.Has(wasi.AccessTimeNow)
	changeModifyTime := fstFlags.Has(wasi.ModifyTime) || fstFlags.Has(wasi.ModifyTimeNow)
	if !changeAccessTime || !changeModifyTime {
		var stat unix.Stat_t
		err := unix.Fstatat(d.FD, path, &stat, sysFlags)
		if err != nil {
			return makeErrno(err)
		}
		ts[0] = stat.Atim
		ts[1] = stat.Mtim
	}
	if changeAccessTime {
		ts[0] = unix.NsecToTimespec(int64(accessTime))
	}
	if changeModifyTime {
		ts[1] = unix.NsecToTimespec(int64(modifyTime))
	}
	err := unix.UtimesNanoAt(d.FD, path, ts[:], sysFlags)
	return makeErrno(err)
}

func (p *Provider) PathLink(fd wasi.FD, flags wasi.LookupFlags, oldPath string, newFD wasi.FD, newPath string) wasi.Errno {
	oldDir, errno := p.lookupFD(fd, wasi.PathLinkSourceRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	newDir, errno := p.lookupFD(newFD, wasi.PathLinkTargetRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	sysFlags := 0
	if flags.Has(wasi.SymlinkFollow) {
		sysFlags |= unix.AT_SYMLINK_FOLLOW
	}
	err := unix.Linkat(oldDir.FD, oldPath, newDir.FD, newPath, sysFlags)
	return makeErrno(err)
}

func (p *Provider) PathOpen(fd wasi.FD, lookupFlags wasi.LookupFlags, path string, openFlags wasi.OpenFlags, rightsBase, rightsInheriting wasi.Rights, fdFlags wasi.FDFlags) (wasi.FD, wasi.Errno) {
	d, errno := p.lookupFD(fd, wasi.PathOpenRight)
	if errno != wasi.ESUCCESS {
		return -1, errno
	}
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "../") {
		return -1, wasi.EPERM
	}

	// Rights can only be preserved or removed, not added.
	rightsBase &= wasi.AllRights
	rightsInheriting &= wasi.AllRights
	if (rightsBase &^ d.FDStat.RightsInheriting) != 0 {
		return -1, wasi.ENOTCAPABLE
	} else if (rightsInheriting &^ d.FDStat.RightsInheriting) != 0 {
		return -1, wasi.ENOTCAPABLE
	}
	rightsBase &= d.FDStat.RightsInheriting
	rightsInheriting &= d.FDStat.RightsInheriting

	oflags := unix.O_CLOEXEC
	if openFlags.Has(wasi.OpenDirectory) {
		oflags |= unix.O_DIRECTORY
		rightsBase &^= wasi.FDSeekRight
	}
	if openFlags.Has(wasi.OpenCreate) {
		if !d.FDStat.RightsBase.Has(wasi.PathCreateFileRight) {
			return -1, wasi.ENOTCAPABLE
		}
		oflags |= unix.O_CREAT
	}
	if openFlags.Has(wasi.OpenExclusive) {
		oflags |= unix.O_EXCL
	}
	if openFlags.Has(wasi.OpenTruncate) {
		if !d.FDStat.RightsBase.Has(wasi.PathFileStatSetSizeRight) {
			return -1, wasi.ENOTCAPABLE
		}
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
	// TODO: handle O_RSYNC
	if fdFlags.Has(wasi.NonBlock) {
		oflags |= unix.O_NONBLOCK
	}
	if !lookupFlags.Has(wasi.SymlinkFollow) {
		oflags |= unix.O_NOFOLLOW
	}
	switch {
	case openFlags.Has(wasi.OpenDirectory):
		oflags |= unix.O_RDONLY
	case rightsBase.HasAny(wasi.ReadRights) && rightsBase.HasAny(wasi.WriteRights):
		oflags |= unix.O_RDWR
	case rightsBase.HasAny(wasi.ReadRights):
		oflags |= unix.O_RDONLY
	case rightsBase.HasAny(wasi.WriteRights):
		oflags |= unix.O_WRONLY
	default:
		oflags |= unix.O_RDONLY
	}

	mode := uint32(0644)
	fileType := wasi.RegularFileType
	if (oflags & unix.O_DIRECTORY) != 0 {
		fileType = wasi.DirectoryType
		mode = 0
	}
	hostfd, err := unix.Openat(d.FD, path, oflags, mode)
	if err != nil {
		return -1, makeErrno(err)
	}

	guestfd := p.RegisterFD(hostfd, filepath.Join(d.Path, path), wasi.FDStat{
		FileType:         fileType,
		Flags:            fdFlags,
		RightsBase:       rightsBase,
		RightsInheriting: rightsInheriting,
	})
	return guestfd, wasi.ESUCCESS
}

func (p *Provider) PathReadLink(fd wasi.FD, path string) (string, wasi.Errno) {
	d, errno := p.lookupFD(fd, wasi.PathReadLinkRight)
	if errno != wasi.ESUCCESS {
		return "", errno
	}
	var buf [1024]byte // TODO: receive buffer as argument, return length
	n, err := unix.Readlinkat(d.FD, path, buf[:])
	if err != nil {
		return "", makeErrno(err)
	}
	return string(buf[:n]), wasi.ESUCCESS
}

func (p *Provider) PathRemoveDirectory(fd wasi.FD, path string) wasi.Errno {
	d, errno := p.lookupFD(fd, wasi.PathRemoveDirectoryRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Unlinkat(d.FD, path, unix.AT_REMOVEDIR)
	return makeErrno(err)
}

func (p *Provider) PathRename(fd wasi.FD, oldPath string, newFD wasi.FD, newPath string) wasi.Errno {
	oldDir, errno := p.lookupFD(fd, wasi.PathRenameSourceRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	newDir, errno := p.lookupFD(newFD, wasi.PathRenameTargetRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Renameat(oldDir.FD, oldPath, newDir.FD, newPath)
	return makeErrno(err)
}

func (p *Provider) PathSymlink(oldPath string, fd wasi.FD, newPath string) wasi.Errno {
	d, errno := p.lookupFD(fd, wasi.PathSymlinkRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Symlinkat(oldPath, d.FD, newPath)
	return makeErrno(err)
}

func (p *Provider) PathUnlinkFile(fd wasi.FD, path string) wasi.Errno {
	d, errno := p.lookupFD(fd, wasi.PathUnlinkFileRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Unlinkat(d.FD, path, 0)
	return makeErrno(err)
}

func (p *Provider) PollOneOff(subscriptions []wasi.Subscription, events []wasi.Event) ([]wasi.Event, wasi.Errno) {
	if len(subscriptions) == 0 {
		return events, wasi.EINVAL
	}
	timeout := time.Duration(-1)
	p.pollfds = p.pollfds[:0]
	for i := range subscriptions {
		s := &subscriptions[i]

		switch s.EventType {
		case wasi.FDReadEvent, wasi.FDWriteEvent:
			f, errno := p.lookupFD(s.GetFDReadWrite().FD, wasi.PollFDReadWriteRight)
			if errno != wasi.ESUCCESS {
				// TODO: set the error on the event instead of aborting the call
				return events, errno
			}
			var pollevent int16 = unix.POLLIN
			if s.EventType == wasi.FDWriteEvent {
				pollevent = unix.POLLOUT
			}
			p.pollfds = append(p.pollfds, unix.PollFd{
				Fd:     int32(f.FD),
				Events: pollevent,
			})
		case wasi.ClockEvent:
			c := s.GetClock()
			switch {
			case c.ID != wasi.Monotonic || c.Flags.Has(wasi.Abstime):
				return events, wasi.ENOSYS // not implemented
			case timeout < 0:
				timeout = time.Duration(c.Timeout)
			case timeout >= 0 && time.Duration(c.Timeout) < timeout:
				timeout = time.Duration(c.Timeout)
			}
		}
	}

	if len(p.pollfds) == 0 {
		// Just sleep if there's no FD events to poll.
		if timeout >= 0 {
			time.Sleep(timeout)
		}
		return events, wasi.ESUCCESS
	}

	var timeoutMillis int
	if timeout < 0 {
		timeoutMillis = -1
	} else {
		timeoutMillis = int(timeout.Milliseconds())
	}
	n, err := unix.Poll(p.pollfds, timeoutMillis)
	if err != nil {
		return events, makeErrno(err)
	}

	j := 0
	for i := range subscriptions {
		s := &subscriptions[i]
		if s.EventType == wasi.ClockEvent {
			continue
		}
		pf := &p.pollfds[j]
		j++
		if pf.Revents == 0 {
			continue
		}
		e := wasi.Event{UserData: s.UserData, EventType: s.EventType}

		// TODO: review cases where Revents contains many flags
		if s.EventType == wasi.FDReadEvent && (pf.Revents&unix.POLLIN) != 0 {
			e.FDReadWrite.NBytes = 1 // we don't know how many, so just say 1
		}
		if s.EventType == wasi.FDWriteEvent && (pf.Revents&unix.POLLOUT) != 0 {
			e.FDReadWrite.NBytes = 1 // we don't know how many, so just say 1
		}
		if (pf.Revents & unix.POLLERR) != 0 {
			e.Errno = wasi.ECANCELED // we don't know what error, just pass something
		}
		if (pf.Revents & unix.POLLHUP) != 0 {
			e.FDReadWrite.Flags |= wasi.Hangup
		}
		events = append(events, e)
	}
	if n != len(events) {
		panic("unexpected unix.Poll result")
	}
	return events, wasi.ESUCCESS
}

func (p *Provider) ProcExit(code wasi.ExitCode) wasi.Errno {
	if p.Exit != nil {
		p.Exit(int(code))
	}
	return wasi.ESUCCESS
}

func (p *Provider) ProcRaise(signal wasi.Signal) wasi.Errno {
	if p.Raise != nil {
		p.Raise(int(signal))
	}
	return wasi.ESUCCESS
}

func (p *Provider) SchedYield() wasi.Errno {
	if p.Yield != nil {
		p.Yield()
	}
	return wasi.ESUCCESS
}

func (p *Provider) RandomGet(b []byte) wasi.Errno {
	if _, err := io.ReadFull(p.Rand, b); err != nil {
		return wasi.EIO
	}
	return wasi.ESUCCESS
}

func (p *Provider) SockAccept(fd wasi.FD, flags wasi.FDFlags) (wasi.FD, wasi.Errno) {
	socket, errno := p.lookupSocketFD(fd, wasi.SockAcceptRight)
	if errno != wasi.ESUCCESS {
		return -1, errno
	}
	if (flags & ^wasi.NonBlock) != 0 {
		return -1, wasi.EINVAL
	}
	// TODO: use accept4 on linux to set O_CLOEXEC and O_NONBLOCK
	connfd, _, err := unix.Accept(socket.FD)
	if err != nil {
		return -1, makeErrno(err)
	}
	if err := unix.SetNonblock(connfd, flags.Has(wasi.NonBlock)); err != nil {
		unix.Close(connfd)
		return -1, makeErrno(err)
	}
	guestfd := p.RegisterFD(connfd, "", wasi.FDStat{
		FileType:         wasi.SocketStreamType,
		Flags:            flags,
		RightsBase:       socket.FDStat.RightsInheriting,
		RightsInheriting: socket.FDStat.RightsInheriting,
	})
	return guestfd, wasi.ESUCCESS
}

func (p *Provider) SockRecv(fd wasi.FD, iovecs []wasi.IOVec, flags wasi.RIFlags) (wasi.Size, wasi.ROFlags, wasi.Errno) {
	socket, errno := p.lookupSocketFD(fd, wasi.FDReadRight)
	if errno != wasi.ESUCCESS {
		return 0, 0, errno
	}
	_ = socket
	return 0, 0, wasi.ENOSYS // TODO: implement SockRecv
}

func (p *Provider) SockSend(fd wasi.FD, iovecs []wasi.IOVec, flags wasi.SIFlags) (wasi.Size, wasi.Errno) {
	socket, errno := p.lookupSocketFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	_ = socket
	return 0, wasi.ENOSYS // TODO: implement SockSend
}

func (p *Provider) SockShutdown(fd wasi.FD, flags wasi.SDFlags) wasi.Errno {
	socket, errno := p.lookupSocketFD(fd, wasi.SockShutdownRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	var sysHow int
	switch {
	case flags.Has(wasi.ShutdownRD | wasi.ShutdownWR):
		sysHow = unix.SHUT_RDWR
	case flags.Has(wasi.ShutdownRD):
		sysHow = unix.SHUT_RD
	case flags.Has(wasi.ShutdownWR):
		sysHow = unix.SHUT_WR
	default:
		return wasi.EINVAL
	}
	err := unix.Shutdown(socket.FD, sysHow)
	return makeErrno(err)
}

func (p *Provider) Close() error {
	p.fds.Range(func(_ wasi.FD, f *fdinfo) bool {
		unix.Close(f.FD)
		return true
	})
	p.fds.Reset()
	return nil
}

type statDirEntry struct {
	name string
	info os.FileInfo
}

func (d *statDirEntry) Name() string               { return d.name }
func (d *statDirEntry) IsDir() bool                { return d.info.IsDir() }
func (d *statDirEntry) Type() os.FileMode          { return d.info.Mode().Type() }
func (d *statDirEntry) Info() (os.FileInfo, error) { return d.info, nil }
