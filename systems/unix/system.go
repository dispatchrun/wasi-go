package unix

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/internal/descriptor"
	"golang.org/x/sys/unix"
)

// System is a WASI preview 1 implementation for Unix.
//
// It implements the wasi.System and wasi.SocketsExtension interfaces.
//
// An instance of System is not safe for concurrent use.
type System struct {
	// Args are the environment variables accessible via ArgsGet.
	Args []string

	// Environ is the environment variables accessible via EnvironGet.
	Environ []string

	// Realtime returns the realtime clock value.
	Realtime          func(context.Context) (uint64, error)
	RealtimePrecision time.Duration

	// Monotonic returns the monotonic clock value.
	Monotonic          func(context.Context) (uint64, error)
	MonotonicPrecision time.Duration

	// Yield is called when SchedYield is called. If Yield is nil,
	// SchedYield is a noop.
	Yield func(context.Context) error

	// Exit is called with an exit code when ProcExit is called.
	// If Exit is nil, ProcExit is a noop.
	Exit func(context.Context, int) error

	// Raise is called with a signal when ProcRaise is called.
	// If Raise is nil, ProcRaise is a noop.
	Raise func(context.Context, int) error

	// Rand is the source for RandomGet.
	Rand io.Reader

	fds      descriptor.Table[wasi.FD, fdinfo]
	preopens descriptor.Table[wasi.FD, string]

	pollfds   []unix.PollFd
	unixInet4 unix.SockaddrInet4
	unixInet6 unix.SockaddrInet6
	unixUnix  unix.SockaddrUnix
	wasiInet4 wasi.Inet4Address
	wasiInet6 wasi.Inet6Address
	wasiUnix  wasi.UnixAddress

	// shutfds are a pair of file descriptors allocated to the read and write
	// ends of a pipe. They are used to asynchronously interrupting calls to
	// poll(2) by closing the write end of the pipe, causing the read end to
	// become reading for reading and any polling on the fd to return.
	mutex   sync.Mutex
	shutfds [2]int
}

var _ wasi.System = (*System)(nil)
var _ wasi.SocketsExtension = (*System)(nil)

type fdinfo struct {
	// fd is the underlying OS file descriptor.
	fd int

	// stat is cached information about the file descriptor.
	stat wasi.FDStat

	// dir is lazily allocated when FDReadDir is called, it maintains the state
	// of the directory iterator.
	dir *dirbuf
}

// Preopen adds an open file to the list of pre-opens.
func (s *System) Preopen(hostfd int, path string, fdstat wasi.FDStat) wasi.FD {
	fd := s.Register(hostfd, fdstat)
	s.preopens.Assign(fd, path)
	return fd
}

func (s *System) Register(hostfd int, fdstat wasi.FDStat) wasi.FD {
	fdstat.RightsBase &= wasi.AllRights
	fdstat.RightsInheriting &= wasi.AllRights
	return s.fds.Insert(fdinfo{
		fd:   hostfd,
		stat: fdstat,
	})
}

func (s *System) isPreopen(fd wasi.FD) bool {
	return s.preopens.Access(fd) != nil
}

func (s *System) lookupFD(guestfd wasi.FD, rights wasi.Rights) (*fdinfo, wasi.Errno) {
	f := s.fds.Access(guestfd)
	if f == nil {
		return nil, wasi.EBADF
	}
	if !f.stat.RightsBase.Has(rights) {
		return nil, wasi.ENOTCAPABLE
	}
	return f, wasi.ESUCCESS
}

func (s *System) lookupPreopenPath(guestfd wasi.FD) (string, wasi.Errno) {
	path, ok := s.preopens.Lookup(guestfd)
	if !ok {
		return "", wasi.EBADF
	}
	f, errno := s.lookupFD(guestfd, 0)
	if errno != wasi.ESUCCESS {
		return "", errno
	}
	if f.stat.FileType != wasi.DirectoryType {
		return "", wasi.ENOTDIR
	}
	return path, wasi.ESUCCESS
}

func (s *System) lookupSocketFD(guestfd wasi.FD, rights wasi.Rights) (*fdinfo, wasi.Errno) {
	f, errno := s.lookupFD(guestfd, rights)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	switch f.stat.FileType {
	case wasi.SocketStreamType, wasi.SocketDGramType:
		return f, wasi.ESUCCESS
	default:
		return nil, wasi.ENOTSOCK
	}
}

func (s *System) ArgsSizesGet(ctx context.Context) (argCount, stringBytes int, errno wasi.Errno) {
	argCount = len(s.Args)
	for _, arg := range s.Args {
		stringBytes += len(arg) + 1
	}
	return argCount, stringBytes, wasi.ESUCCESS
}

func (s *System) ArgsGet(ctx context.Context) ([]string, wasi.Errno) {
	return s.Args, wasi.ESUCCESS
}

func (s *System) EnvironSizesGet(ctx context.Context) (envCount, stringBytes int, errno wasi.Errno) {
	envCount = len(s.Environ)
	for _, env := range s.Environ {
		stringBytes += len(env) + 1
	}
	return envCount, stringBytes, wasi.ESUCCESS
}

func (s *System) EnvironGet(ctx context.Context) ([]string, wasi.Errno) {
	return s.Environ, wasi.ESUCCESS
}

func (s *System) ClockResGet(ctx context.Context, id wasi.ClockID) (wasi.Timestamp, wasi.Errno) {
	switch id {
	case wasi.Realtime:
		return wasi.Timestamp(s.RealtimePrecision), wasi.ESUCCESS
	case wasi.Monotonic:
		return wasi.Timestamp(s.MonotonicPrecision), wasi.ESUCCESS
	case wasi.ProcessCPUTimeID, wasi.ThreadCPUTimeID:
		return 0, wasi.ENOTSUP
	default:
		return 0, wasi.EINVAL
	}
}

func (s *System) ClockTimeGet(ctx context.Context, id wasi.ClockID, precision wasi.Timestamp) (wasi.Timestamp, wasi.Errno) {
	switch id {
	case wasi.Realtime:
		if s.Realtime == nil {
			return 0, wasi.ENOTSUP
		}
		t, err := s.Realtime(ctx)
		return wasi.Timestamp(t), makeErrno(err)
	case wasi.Monotonic:
		if s.Monotonic == nil {
			return 0, wasi.ENOTSUP
		}
		t, err := s.Monotonic(ctx)
		return wasi.Timestamp(t), makeErrno(err)
	case wasi.ProcessCPUTimeID, wasi.ThreadCPUTimeID:
		return 0, wasi.ENOTSUP
	default:
		return 0, wasi.EINVAL
	}
}

func (s *System) FDAdvise(ctx context.Context, fd wasi.FD, offset wasi.FileSize, length wasi.FileSize, advice wasi.Advice) wasi.Errno {
	f, errno := s.lookupFD(fd, wasi.FDAdviseRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := fdadvise(f.fd, int64(offset), int64(length), advice)
	return makeErrno(err)
}

func (s *System) FDAllocate(ctx context.Context, fd wasi.FD, offset wasi.FileSize, length wasi.FileSize) wasi.Errno {
	f, errno := s.lookupFD(fd, wasi.FDAllocateRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := fallocate(f.fd, int64(offset), int64(length))
	return makeErrno(err)
}

func (s *System) FDClose(ctx context.Context, fd wasi.FD) wasi.Errno {
	f, errno := s.lookupFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Close(f.fd)
	s.fds.Delete(fd)
	// Note: closing pre-opens is allowed.
	// See github.com/WebAssembly/wasi-testsuite/blob/1b1d4a5/tests/rust/src/bin/close_preopen.rs
	s.preopens.Delete(fd)
	return makeErrno(err)
}

func (s *System) FDDataSync(ctx context.Context, fd wasi.FD) wasi.Errno {
	f, errno := s.lookupFD(fd, wasi.FDDataSyncRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := fdatasync(f.fd)
	return makeErrno(err)
}

func (s *System) FDStatGet(ctx context.Context, fd wasi.FD) (wasi.FDStat, wasi.Errno) {
	f, errno := s.lookupFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return wasi.FDStat{}, errno
	}
	return f.stat, wasi.ESUCCESS
}

func (s *System) FDStatSetFlags(ctx context.Context, fd wasi.FD, flags wasi.FDFlags) wasi.Errno {
	f, errno := s.lookupFD(fd, wasi.FDStatSetFlagsRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	changes := flags ^ f.stat.Flags
	if changes == 0 {
		return wasi.ESUCCESS
	}
	if changes.Has(wasi.Sync | wasi.DSync | wasi.RSync) {
		return wasi.ENOSYS // TODO: support changing {Sync,DSync,Rsync}
	}
	fl, err := unix.FcntlInt(uintptr(f.fd), unix.F_GETFL, 0)
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
	if _, err := unix.FcntlInt(uintptr(f.fd), unix.F_SETFL, fl); err != nil {
		return makeErrno(err)
	}
	f.stat.Flags ^= changes
	return wasi.ESUCCESS
}

func (s *System) FDStatSetRights(ctx context.Context, fd wasi.FD, rightsBase, rightsInheriting wasi.Rights) wasi.Errno {
	f, errno := s.lookupFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	// Rights can only be preserved or removed, not added.
	rightsBase &= wasi.AllRights
	rightsInheriting &= wasi.AllRights
	if (rightsBase &^ f.stat.RightsBase) != 0 {
		return wasi.ENOTCAPABLE
	}
	if (rightsInheriting &^ f.stat.RightsInheriting) != 0 {
		return wasi.ENOTCAPABLE
	}
	f.stat.RightsBase &= rightsBase
	f.stat.RightsInheriting &= rightsInheriting
	return wasi.ESUCCESS
}

func (s *System) FDFileStatGet(ctx context.Context, fd wasi.FD) (wasi.FileStat, wasi.Errno) {
	f, errno := s.lookupFD(fd, wasi.FDFileStatGetRight)
	if errno != wasi.ESUCCESS {
		return wasi.FileStat{}, errno
	}
	var sysStat unix.Stat_t
	if err := unix.Fstat(f.fd, &sysStat); err != nil {
		return wasi.FileStat{}, makeErrno(err)
	}
	stat := makeFileStat(&sysStat)
	if fd <= 2 {
		// Override stdio size/times.
		// See github.com/WebAssembly/wasi-testsuite/blob/1b1d4a5/tests/rust/src/bin/fd_filestat_get.rs
		stat.Size = 0
		stat.AccessTime = 0
		stat.ModifyTime = 0
		stat.ChangeTime = 0
	}
	return stat, wasi.ESUCCESS
}

func (s *System) FDFileStatSetSize(ctx context.Context, fd wasi.FD, size wasi.FileSize) wasi.Errno {
	f, errno := s.lookupFD(fd, wasi.FDFileStatSetSizeRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Ftruncate(f.fd, int64(size))
	return makeErrno(err)
}

func (s *System) FDFileStatSetTimes(ctx context.Context, fd wasi.FD, accessTime, modifyTime wasi.Timestamp, flags wasi.FSTFlags) wasi.Errno {
	f, errno := s.lookupFD(fd, wasi.FDFileStatSetTimesRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	var sysStat unix.Stat_t
	if err := unix.Fstat(f.fd, &sysStat); err != nil {
		return makeErrno(err)
	}
	ts := [2]unix.Timespec{sysStat.Atim, sysStat.Mtim}
	if flags.Has(wasi.AccessTimeNow) || flags.Has(wasi.ModifyTimeNow) {
		if s.Monotonic == nil {
			return wasi.ENOSYS
		}
		now, err := s.Monotonic(ctx)
		if err != nil {
			return makeErrno(err)
		}
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
	err := futimens(f.fd, &ts)
	return makeErrno(err)
}

func (s *System) FDPread(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, offset wasi.FileSize) (wasi.Size, wasi.Errno) {
	f, errno := s.lookupFD(fd, wasi.FDReadRight|wasi.FDSeekRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := preadv(f.fd, makeIOVecs(iovecs), int64(offset))
	return wasi.Size(n), makeErrno(err)
}

func (s *System) FDPreStatGet(ctx context.Context, fd wasi.FD) (wasi.PreStat, wasi.Errno) {
	path, errno := s.lookupPreopenPath(fd)
	if errno != wasi.ESUCCESS {
		return wasi.PreStat{}, errno
	}
	stat := wasi.PreStat{
		Type: wasi.PreOpenDir,
		PreStatDir: wasi.PreStatDir{
			NameLength: wasi.Size(len(path)),
		},
	}
	return stat, wasi.ESUCCESS
}

func (s *System) FDPreStatDirName(ctx context.Context, fd wasi.FD) (string, wasi.Errno) {
	return s.lookupPreopenPath(fd)
}

func (s *System) FDPwrite(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, offset wasi.FileSize) (wasi.Size, wasi.Errno) {
	f, errno := s.lookupFD(fd, wasi.FDWriteRight|wasi.FDSeekRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := pwritev(f.fd, makeIOVecs(iovecs), int64(offset))
	return wasi.Size(n), makeErrno(err)
}

func (s *System) FDRead(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec) (wasi.Size, wasi.Errno) {
	f, errno := s.lookupFD(fd, wasi.FDReadRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := readv(f.fd, makeIOVecs(iovecs))
	return wasi.Size(n), makeErrno(err)
}

func (s *System) FDReadDir(ctx context.Context, fd wasi.FD, entries []wasi.DirEntry, cookie wasi.DirCookie, bufferSizeBytes int) (int, wasi.Errno) {
	f, errno := s.lookupFD(fd, wasi.FDReadDirRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	if len(entries) == 0 {
		return 0, wasi.EINVAL
	}
	if f.dir == nil {
		f.dir = new(dirbuf)
	}
	n, err := f.dir.readDirEntries(f.fd, entries, cookie, bufferSizeBytes)
	return n, makeErrno(err)
}

func (s *System) FDRenumber(ctx context.Context, from, to wasi.FD) wasi.Errno {
	if s.isPreopen(from) || s.isPreopen(to) {
		return wasi.ENOTSUP
	}
	f, errno := s.lookupFD(from, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	// TODO: limit max file descriptor number
	g, replaced := s.fds.Assign(to, *f)
	if replaced {
		unix.Close(g.fd)
	}
	s.fds.Delete(from)
	return wasi.ESUCCESS
}

func (s *System) FDSync(ctx context.Context, fd wasi.FD) wasi.Errno {
	f, errno := s.lookupFD(fd, wasi.FDSyncRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := fsync(f.fd)
	return makeErrno(err)
}

func (s *System) FDSeek(ctx context.Context, fd wasi.FD, delta wasi.FileDelta, whence wasi.Whence) (wasi.FileSize, wasi.Errno) {
	return s.fdseek(fd, wasi.FDSeekRight, delta, whence)
}

func (s *System) FDTell(ctx context.Context, fd wasi.FD) (wasi.FileSize, wasi.Errno) {
	return s.fdseek(fd, wasi.FDTellRight, 0, wasi.SeekCurrent)
}

func (s *System) fdseek(fd wasi.FD, rights wasi.Rights, delta wasi.FileDelta, whence wasi.Whence) (wasi.FileSize, wasi.Errno) {
	// Note: FDSeekRight implies FDTellRight. FDTellRight also includes the
	// right to invoke FDSeek in such a way that the file offset remains
	// unaltered.
	f, errno := s.lookupFD(fd, rights)
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
	off, err := lseek(f.fd, int64(delta), sysWhence)
	return wasi.FileSize(off), makeErrno(err)
}

func (s *System) FDWrite(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec) (wasi.Size, wasi.Errno) {
	f, errno := s.lookupFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := writev(f.fd, makeIOVecs(iovecs))
	return wasi.Size(n), makeErrno(err)
}

func (s *System) PathCreateDirectory(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	d, errno := s.lookupFD(fd, wasi.PathCreateDirectoryRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Mkdirat(d.fd, path, 0755)
	return makeErrno(err)
}

func (s *System) PathFileStatGet(ctx context.Context, fd wasi.FD, flags wasi.LookupFlags, path string) (wasi.FileStat, wasi.Errno) {
	d, errno := s.lookupFD(fd, wasi.PathFileStatGetRight)
	if errno != wasi.ESUCCESS {
		return wasi.FileStat{}, errno
	}
	var sysStat unix.Stat_t
	var sysFlags int
	if !flags.Has(wasi.SymlinkFollow) {
		sysFlags |= unix.AT_SYMLINK_NOFOLLOW
	}
	err := unix.Fstatat(d.fd, path, &sysStat, sysFlags)
	return makeFileStat(&sysStat), makeErrno(err)
}

func (s *System) PathFileStatSetTimes(ctx context.Context, fd wasi.FD, lookupFlags wasi.LookupFlags, path string, accessTime, modifyTime wasi.Timestamp, fstFlags wasi.FSTFlags) wasi.Errno {
	d, errno := s.lookupFD(fd, wasi.PathFileStatSetTimesRight)
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
		err := unix.Fstatat(d.fd, path, &stat, sysFlags)
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
	err := unix.UtimesNanoAt(d.fd, path, ts[:], sysFlags)
	return makeErrno(err)
}

func (s *System) PathLink(ctx context.Context, fd wasi.FD, flags wasi.LookupFlags, oldPath string, newFD wasi.FD, newPath string) wasi.Errno {
	oldDir, errno := s.lookupFD(fd, wasi.PathLinkSourceRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	newDir, errno := s.lookupFD(newFD, wasi.PathLinkTargetRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	sysFlags := 0
	if flags.Has(wasi.SymlinkFollow) {
		sysFlags |= unix.AT_SYMLINK_FOLLOW
	}
	err := unix.Linkat(oldDir.fd, oldPath, newDir.fd, newPath, sysFlags)
	return makeErrno(err)
}

func (s *System) PathOpen(ctx context.Context, fd wasi.FD, lookupFlags wasi.LookupFlags, path string, openFlags wasi.OpenFlags, rightsBase, rightsInheriting wasi.Rights, fdFlags wasi.FDFlags) (wasi.FD, wasi.Errno) {
	d, errno := s.lookupFD(fd, wasi.PathOpenRight)
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
	if (rightsBase &^ d.stat.RightsInheriting) != 0 {
		return -1, wasi.ENOTCAPABLE
	} else if (rightsInheriting &^ d.stat.RightsInheriting) != 0 {
		return -1, wasi.ENOTCAPABLE
	}
	rightsBase &= d.stat.RightsInheriting
	rightsInheriting &= d.stat.RightsInheriting

	oflags := unix.O_CLOEXEC
	if openFlags.Has(wasi.OpenDirectory) {
		oflags |= unix.O_DIRECTORY
		rightsBase &= wasi.DirectoryRights
	}
	if openFlags.Has(wasi.OpenCreate) {
		if !d.stat.RightsBase.Has(wasi.PathCreateFileRight) {
			return -1, wasi.ENOTCAPABLE
		}
		oflags |= unix.O_CREAT
	}
	if openFlags.Has(wasi.OpenExclusive) {
		oflags |= unix.O_EXCL
	}
	if openFlags.Has(wasi.OpenTruncate) {
		if !d.stat.RightsBase.Has(wasi.PathFileStatSetSizeRight) {
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
	hostfd, err := unix.Openat(d.fd, path, oflags, mode)
	if err != nil {
		return -1, makeErrno(err)
	}

	guestfd := s.fds.Insert(fdinfo{
		fd: hostfd,
		stat: wasi.FDStat{
			FileType:         fileType,
			Flags:            fdFlags,
			RightsBase:       rightsBase,
			RightsInheriting: rightsInheriting,
		},
	})
	return guestfd, wasi.ESUCCESS
}

func (s *System) PathReadLink(ctx context.Context, fd wasi.FD, path string, buffer []byte) ([]byte, wasi.Errno) {
	d, errno := s.lookupFD(fd, wasi.PathReadLinkRight)
	if errno != wasi.ESUCCESS {
		return buffer, errno
	}
	n, err := unix.Readlinkat(d.fd, path, buffer)
	if err != nil {
		return buffer, makeErrno(err)
	} else if n == len(buffer) {
		return buffer, wasi.ERANGE
	}
	return buffer[:n], wasi.ESUCCESS
}

func (s *System) PathRemoveDirectory(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	d, errno := s.lookupFD(fd, wasi.PathRemoveDirectoryRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Unlinkat(d.fd, path, unix.AT_REMOVEDIR)
	return makeErrno(err)
}

func (s *System) PathRename(ctx context.Context, fd wasi.FD, oldPath string, newFD wasi.FD, newPath string) wasi.Errno {
	oldDir, errno := s.lookupFD(fd, wasi.PathRenameSourceRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	newDir, errno := s.lookupFD(newFD, wasi.PathRenameTargetRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Renameat(oldDir.fd, oldPath, newDir.fd, newPath)
	return makeErrno(err)
}

func (s *System) PathSymlink(ctx context.Context, oldPath string, fd wasi.FD, newPath string) wasi.Errno {
	d, errno := s.lookupFD(fd, wasi.PathSymlinkRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Symlinkat(oldPath, d.fd, newPath)
	return makeErrno(err)
}

func (s *System) PathUnlinkFile(ctx context.Context, fd wasi.FD, path string) wasi.Errno {
	d, errno := s.lookupFD(fd, wasi.PathUnlinkFileRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Unlinkat(d.fd, path, 0)
	return makeErrno(err)
}

func (s *System) PollOneOff(ctx context.Context, subscriptions []wasi.Subscription, events []wasi.Event) (int, wasi.Errno) {
	if len(subscriptions) == 0 || len(events) < len(subscriptions) {
		return 0, wasi.EINVAL
	}
	wakefd, err := s.init()
	if err != nil {
		return 0, makeErrno(err)
	}
	epoch := time.Duration(0)
	timeout := time.Duration(-1)
	s.pollfds = s.pollfds[:0]
	s.pollfds = append(s.pollfds, unix.PollFd{
		Fd:     int32(wakefd),
		Events: unix.POLLIN | unix.POLLERR | unix.POLLHUP,
	})

	numEvents := 0
	for i := range events {
		events[i] = wasi.Event{}
	}

	for i := range subscriptions {
		sub := &subscriptions[i]

		switch sub.EventType {
		case wasi.FDReadEvent, wasi.FDWriteEvent:
			f, errno := s.lookupFD(sub.GetFDReadWrite().FD, wasi.PollFDReadWriteRight)
			if errno != wasi.ESUCCESS {
				events[i] = errorEvent(sub, errno)
				numEvents++
				continue
			}
			var pollevent int16 = unix.POLLIN
			if sub.EventType == wasi.FDWriteEvent {
				pollevent = unix.POLLOUT
			}
			s.pollfds = append(s.pollfds, unix.PollFd{
				Fd:     int32(f.fd),
				Events: pollevent,
			})

		case wasi.ClockEvent:
			c := sub.GetClock()
			if c.ID != wasi.Monotonic || s.Monotonic == nil {
				events[i] = errorEvent(sub, wasi.ENOSYS)
				numEvents++
				continue
			}
			if epoch == 0 {
				// Only capture the current time if the program requested a
				// clock subscription; it allows programs that never ask for
				// a timeout to run with a system which does not have a
				// monotonic clock configured.
				now, err := s.Monotonic(ctx)
				if err != nil {
					return 0, makeErrno(err)
				}
				epoch = time.Duration(now)
			}
			t := c.Timeout.Duration()
			if c.Flags.Has(wasi.Abstime) {
				// If the subscription asks for an absolute monotonic time point
				// we can honnor it by computing its relative delta to the poll
				// epoch.
				t -= epoch
			}
			switch {
			case timeout < 0:
				timeout = t
			case t < timeout:
				timeout = t
			}
		}
	}

	var timeoutMillis int
	// We set the timeout to zero when we already produced events due to
	// invalid subscriptions; this is useful to still make progress on I/O
	// completion.
	if numEvents == 0 {
		if timeout < 0 {
			timeoutMillis = -1
		} else {
			timeoutMillis = int(timeout.Milliseconds())
		}
	}

	n, err := unix.Poll(s.pollfds, timeoutMillis)
	if err != nil {
		return 0, makeErrno(err)
	}

	if n > 0 && s.pollfds[0].Revents != 0 {
		// If the wake fd was notified it means the system was shut down,
		// we report this by cancelling all subscriptions.
		//
		// Technically we might be erasing events that had already gathered
		// errors in the first loop prior to the call to unix.Poll; this is
		// not a concern since at this time the program would likely be
		// terminating and should not be bothered with handling other errors.
		for i := range subscriptions {
			events[i] = wasi.Event{
				UserData:  subscriptions[i].UserData,
				EventType: subscriptions[i].EventType,
				Errno:     wasi.ECANCELED,
			}
		}
		return len(subscriptions), wasi.ESUCCESS
	}

	var now time.Duration
	if timeout >= 0 {
		t, err := s.Monotonic(ctx)
		if err != nil {
			return 0, makeErrno(err)
		}
		now = time.Duration(t)
	}

	j := 1
	for i := range subscriptions {
		sub := &subscriptions[i]
		e := wasi.Event{UserData: sub.UserData, EventType: sub.EventType + 1}

		if events[i].EventType != 0 {
			continue
		}

		switch sub.EventType {
		case wasi.ClockEvent:
			c := sub.GetClock()
			t := c.Timeout.Duration()
			if !c.Flags.Has(wasi.Abstime) {
				t += epoch
			}
			if t >= now {
				events[i] = e
			}

		case wasi.FDReadEvent, wasi.FDWriteEvent:
			pf := &s.pollfds[j]
			j++
			if pf.Revents == 0 {
				continue
			}

			if e.EventType == wasi.FDReadEvent && (pf.Revents&unix.POLLIN) != 0 {
				e.FDReadWrite.NBytes = 1 // we don't know how many, so just say 1
			}
			if e.EventType == wasi.FDWriteEvent && (pf.Revents&unix.POLLOUT) != 0 {
				e.FDReadWrite.NBytes = 1 // we don't know how many, so just say 1
			}
			if (pf.Revents & unix.POLLERR) != 0 {
				e.Errno = wasi.ECANCELED // we don't know what error, just pass something
			}
			if (pf.Revents & unix.POLLHUP) != 0 {
				e.FDReadWrite.Flags |= wasi.Hangup
			}
			events[i] = e
		}
	}

	// A 1:1 correspondance between the subscription and events arrays is used
	// to track the completion of events, including the completion of invalid
	// subscriptions, clock events, and I/O notifications coming from poll(2).
	//
	// We use zero as the marker on events for subscriptions that have not been
	// fulfilled, but because the zero event type is used to represent clock
	// subscriptions, we mark completed events with the event type + 1.
	//
	// The event type is finally restored to its correct value in the loop below
	// when we pack all completed events at the front of the output buffer.
	n = 0

	for _, e := range events[:len(subscriptions)] {
		if e.EventType != 0 {
			e.EventType--
			events[n] = e
			n++
		}
	}

	return n, wasi.ESUCCESS
}

func errorEvent(s *wasi.Subscription, err wasi.Errno) wasi.Event {
	return wasi.Event{
		UserData:  s.UserData,
		EventType: s.EventType + 1,
		Errno:     err,
	}
}

func (s *System) ProcExit(ctx context.Context, code wasi.ExitCode) wasi.Errno {
	if s.Exit != nil {
		return makeErrno(s.Exit(ctx, int(code)))
	}
	return wasi.ENOSYS
}

func (s *System) ProcRaise(ctx context.Context, signal wasi.Signal) wasi.Errno {
	if s.Raise != nil {
		return makeErrno(s.Raise(ctx, int(signal)))
	}
	return wasi.ENOSYS
}

func (s *System) SchedYield(ctx context.Context) wasi.Errno {
	if s.Yield != nil {
		return makeErrno(s.Yield(ctx))
	}
	return wasi.ENOSYS
}

func (s *System) RandomGet(ctx context.Context, b []byte) wasi.Errno {
	if _, err := io.ReadFull(s.Rand, b); err != nil {
		return wasi.EIO
	}
	return wasi.ESUCCESS
}

func (s *System) SockAccept(ctx context.Context, fd wasi.FD, flags wasi.FDFlags) (wasi.FD, wasi.SocketAddress, wasi.Errno) {
	socket, errno := s.lookupSocketFD(fd, wasi.SockAcceptRight)
	if errno != wasi.ESUCCESS {
		return -1, nil, errno
	}
	if (flags & ^wasi.NonBlock) != 0 {
		return -1, nil, wasi.EINVAL
	}
	connflags := 0
	if (flags & wasi.NonBlock) != 0 {
		connflags |= unix.O_NONBLOCK
	}
	connfd, sa, err := accept(socket.fd, connflags)
	if err != nil {
		return -1, nil, makeErrno(err)
	}
	addr, ok := s.fromUnixSockAddress(sa)
	if !ok {
		return -1, nil, wasi.ENOTSUP
	}
	guestfd := s.fds.Insert(fdinfo{
		fd: connfd,
		stat: wasi.FDStat{
			FileType:         wasi.SocketStreamType,
			Flags:            flags,
			RightsBase:       socket.stat.RightsInheriting,
			RightsInheriting: socket.stat.RightsInheriting,
		},
	})
	return guestfd, addr, wasi.ESUCCESS
}

func (s *System) SockRecv(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.RIFlags) (wasi.Size, wasi.ROFlags, wasi.Errno) {
	socket, errno := s.lookupSocketFD(fd, wasi.FDReadRight)
	if errno != wasi.ESUCCESS {
		return 0, 0, errno
	}
	var sysIFlags int
	if flags.Has(wasi.RecvPeek) {
		sysIFlags |= unix.MSG_PEEK
	}
	if flags.Has(wasi.RecvWaitAll) {
		sysIFlags |= unix.MSG_WAITALL
	}
	n, _, sysOFlags, _, err := unix.RecvmsgBuffers(socket.fd, makeIOVecs(iovecs), nil, sysIFlags)
	if err != nil {
		return 0, 0, makeErrno(err)
	}
	var roflags wasi.ROFlags
	if (sysOFlags & unix.MSG_TRUNC) != 0 {
		roflags |= wasi.RecvDataTruncated
	}
	return wasi.Size(n), roflags, wasi.ESUCCESS
}

func (s *System) SockSend(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.SIFlags) (wasi.Size, wasi.Errno) {
	socket, errno := s.lookupSocketFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	n, err := unix.SendmsgBuffers(socket.fd, makeIOVecs(iovecs), nil, nil, 0)
	return wasi.Size(n), makeErrno(err)
}

func (s *System) SockShutdown(ctx context.Context, fd wasi.FD, flags wasi.SDFlags) wasi.Errno {
	socket, errno := s.lookupSocketFD(fd, wasi.SockShutdownRight)
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
	err := unix.Shutdown(socket.fd, sysHow)
	return makeErrno(err)
}

func (s *System) SockOpen(ctx context.Context, pf wasi.ProtocolFamily, socketType wasi.SocketType, protocol wasi.Protocol, rightsBase, rightsInheriting wasi.Rights) (wasi.FD, wasi.Errno) {
	var sysDomain int
	switch pf {
	case wasi.Inet:
		sysDomain = unix.AF_INET
	case wasi.Inet6:
		sysDomain = unix.AF_INET6
	default:
		return -1, wasi.EINVAL
	}
	var fdType wasi.FileType
	var sysType int
	switch socketType {
	case wasi.DatagramSocket:
		sysType = unix.SOCK_DGRAM
		fdType = wasi.SocketDGramType
	case wasi.StreamSocket:
		sysType = unix.SOCK_STREAM
		fdType = wasi.SocketStreamType
	default:
		return -1, wasi.EINVAL
	}
	var sysProtocol int
	switch protocol {
	case wasi.IPProtocol:
		sysProtocol = unix.IPPROTO_IP
	case wasi.TCPProtocol:
		sysProtocol = unix.IPPROTO_TCP
	case wasi.UDPProtocol:
		sysProtocol = unix.IPPROTO_UDP
	default:
		return -1, wasi.EINVAL
	}
	fd, err := unix.Socket(sysDomain, sysType, sysProtocol)
	if err != nil {
		return -1, makeErrno(err)
	}
	guestfd := s.fds.Insert(fdinfo{
		fd: fd,
		stat: wasi.FDStat{
			FileType:         fdType,
			RightsBase:       rightsBase,
			RightsInheriting: rightsInheriting,
		},
	})
	return guestfd, wasi.ESUCCESS
}

func (s *System) SockBind(ctx context.Context, fd wasi.FD, addr wasi.SocketAddress) wasi.Errno {
	socket, errno := s.lookupSocketFD(fd, wasi.SockAcceptRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	sa, ok := s.toUnixSockAddress(addr)
	if !ok {
		return wasi.EINVAL
	}
	err := unix.Bind(socket.fd, sa)
	return makeErrno(err)
}

func (s *System) SockConnect(ctx context.Context, fd wasi.FD, addr wasi.SocketAddress) wasi.Errno {
	socket, errno := s.lookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	sa, ok := s.toUnixSockAddress(addr)
	if !ok {
		return wasi.EINVAL
	}
	err := unix.Connect(socket.fd, sa)
	return makeErrno(err)
}

func (s *System) SockListen(ctx context.Context, fd wasi.FD, backlog int) wasi.Errno {
	socket, errno := s.lookupSocketFD(fd, wasi.SockAcceptRight)
	if errno != wasi.ESUCCESS {
		return errno
	}
	err := unix.Listen(socket.fd, backlog)
	return makeErrno(err)
}

func (s *System) SockSendTo(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.SIFlags, addr wasi.SocketAddress) (wasi.Size, wasi.Errno) {
	socket, errno := s.lookupSocketFD(fd, wasi.FDWriteRight)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	sa, ok := s.toUnixSockAddress(addr)
	if !ok {
		return 0, wasi.EINVAL
	}
	n, err := unix.SendmsgBuffers(socket.fd, makeIOVecs(iovecs), nil, sa, 0)
	return wasi.Size(n), makeErrno(err)
}

func (s *System) SockRecvFrom(ctx context.Context, fd wasi.FD, iovecs []wasi.IOVec, flags wasi.RIFlags) (wasi.Size, wasi.ROFlags, wasi.SocketAddress, wasi.Errno) {
	socket, errno := s.lookupSocketFD(fd, wasi.FDReadRight)
	if errno != wasi.ESUCCESS {
		return 0, 0, nil, errno
	}
	var sysIFlags int
	if flags.Has(wasi.RecvPeek) {
		sysIFlags |= unix.MSG_PEEK
	}
	if flags.Has(wasi.RecvWaitAll) {
		sysIFlags |= unix.MSG_WAITALL
	}
	n, _, sysOFlags, sa, err := unix.RecvmsgBuffers(socket.fd, makeIOVecs(iovecs), nil, sysIFlags)
	if err != nil {
		return 0, 0, nil, makeErrno(err)
	}
	addr, ok := s.fromUnixSockAddress(sa)
	if !ok {
		return 0, 0, nil, wasi.ENOTSUP
	}
	var roflags wasi.ROFlags
	if (sysOFlags & unix.MSG_TRUNC) != 0 {
		roflags |= wasi.RecvDataTruncated
	}
	return wasi.Size(n), roflags, addr, wasi.ESUCCESS
}

func (s *System) SockGetOptInt(ctx context.Context, fd wasi.FD, level wasi.SocketOptionLevel, option wasi.SocketOption) (int, wasi.Errno) {
	socket, errno := s.lookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return 0, errno
	}
	var sysLevel int
	switch level {
	case wasi.SocketLevel:
		sysLevel = unix.SOL_SOCKET
	default:
		return 0, wasi.EINVAL
	}
	var sysOption int
	switch option {
	case wasi.ReuseAddress:
		sysOption = unix.SO_REUSEADDR
	case wasi.QuerySocketType:
		sysOption = unix.SO_TYPE
	case wasi.QuerySocketError:
		sysOption = unix.SO_ERROR
	case wasi.DontRoute:
		sysOption = unix.SO_DONTROUTE
	case wasi.Broadcast:
		sysOption = unix.SO_BROADCAST
	case wasi.SendBufferSize:
		sysOption = unix.SO_SNDBUF
	case wasi.RecvBufferSize:
		sysOption = unix.SO_RCVBUF
	case wasi.KeepAlive:
		sysOption = unix.SO_KEEPALIVE
	case wasi.OOBInline:
		sysOption = unix.SO_OOBINLINE
	case wasi.RecvLowWatermark:
		sysOption = unix.SO_RCVLOWAT
	case wasi.QueryAcceptConnections:
		sysOption = unix.SO_ACCEPTCONN
	case wasi.Linger, wasi.RecvTimeout, wasi.SendTimeout:
		// These accept struct linger / struct timeval.
		return 0, wasi.EINVAL
	default:
		return 0, wasi.EINVAL
	}

	value, err := unix.GetsockoptInt(socket.fd, sysLevel, sysOption)
	if err != nil {
		return 0, makeErrno(err)
	}

	errno = wasi.ESUCCESS
	switch option {
	case wasi.QuerySocketType:
		switch value {
		case unix.SOCK_DGRAM:
			value = int(wasi.DatagramSocket)
		case unix.SOCK_STREAM:
			value = int(wasi.StreamSocket)
		default:
			value = -1
			errno = wasi.ENOTSUP
		}
	case wasi.QuerySocketError:
		value = int(makeErrno(unix.Errno(value)))
	}

	return value, errno
}

func (s *System) SockSetOptInt(ctx context.Context, fd wasi.FD, level wasi.SocketOptionLevel, option wasi.SocketOption, value int) wasi.Errno {
	socket, errno := s.lookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return errno
	}
	var sysLevel int
	switch level {
	case wasi.SocketLevel:
		sysLevel = unix.SOL_SOCKET
	default:
		return wasi.EINVAL
	}
	var sysOption int
	switch option {
	case wasi.ReuseAddress:
		sysOption = unix.SO_REUSEADDR
	case wasi.QuerySocketType:
		sysOption = unix.SO_TYPE
	case wasi.QuerySocketError:
		sysOption = unix.SO_ERROR
	case wasi.DontRoute:
		sysOption = unix.SO_DONTROUTE
	case wasi.Broadcast:
		sysOption = unix.SO_BROADCAST
	case wasi.SendBufferSize:
		sysOption = unix.SO_SNDBUF
	case wasi.RecvBufferSize:
		sysOption = unix.SO_RCVBUF
	case wasi.KeepAlive:
		sysOption = unix.SO_KEEPALIVE
	case wasi.OOBInline:
		sysOption = unix.SO_OOBINLINE
	case wasi.RecvLowWatermark:
		sysOption = unix.SO_RCVLOWAT
	case wasi.QueryAcceptConnections:
		sysOption = unix.SO_ACCEPTCONN
	case wasi.Linger, wasi.RecvTimeout, wasi.SendTimeout:
		// These accept struct linger / struct timeval.
		return wasi.EINVAL
	default:
		return wasi.EINVAL
	}
	err := unix.SetsockoptInt(socket.fd, sysLevel, sysOption, value)
	return makeErrno(err)
}

func (s *System) SockLocalAddress(ctx context.Context, fd wasi.FD) (wasi.SocketAddress, wasi.Errno) {
	socket, errno := s.lookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	sa, err := unix.Getsockname(socket.fd)
	if err != nil {
		return nil, makeErrno(err)
	}
	addr, ok := s.fromUnixSockAddress(sa)
	if !ok {
		return nil, wasi.ENOTSUP
	}
	return addr, wasi.ESUCCESS
}

func (s *System) SockPeerAddress(ctx context.Context, fd wasi.FD) (wasi.SocketAddress, wasi.Errno) {
	socket, errno := s.lookupSocketFD(fd, 0)
	if errno != wasi.ESUCCESS {
		return nil, errno
	}
	sa, err := unix.Getpeername(socket.fd)
	if err != nil {
		return nil, makeErrno(err)
	}
	addr, ok := s.fromUnixSockAddress(sa)
	if !ok {
		return nil, wasi.ENOTSUP
	}
	return addr, wasi.ESUCCESS
}

func (s *System) Close(ctx context.Context) error {
	s.fds.Range(func(fd wasi.FD, f fdinfo) bool {
		unix.Close(f.fd)
		return true
	})
	s.fds.Reset()
	s.preopens.Reset()

	s.mutex.Lock()
	fd0 := s.shutfds[0]
	fd1 := s.shutfds[1]
	s.shutfds[0] = -1
	s.shutfds[1] = -1
	s.mutex.Unlock()

	if fd0 != 0 || fd1 != 0 { // true if the system was initialized
		unix.Close(fd0)
		unix.Close(fd1)
	}
	return nil
}

// Shutdown may be called to asynchronously cancel all blocking operations on
// the system, causing calls such as PollOneOff to unblock and return an
// error indicating that the system is shutting down.
func (s *System) Shutdown(ctx context.Context) error {
	_, err := s.init()
	if err != nil {
		return err
	}
	s.shutdown()
	return nil
}

func (s *System) init() (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.shutfds[0] == 0 && s.shutfds[1] == 0 {
		if err := pipe(s.shutfds[:], unix.O_NONBLOCK); err != nil {
			return -1, err
		}
	}
	return s.shutfds[0], nil
}

func (s *System) shutdown() {
	s.mutex.Lock()
	fd := s.shutfds[1]
	s.shutfds[1] = -1
	s.mutex.Unlock()
	unix.Close(fd)
}

func (s *System) toUnixSockAddress(addr wasi.SocketAddress) (sa unix.Sockaddr, ok bool) {
	switch t := addr.(type) {
	case *wasi.Inet4Address:
		s.unixInet4.Port = t.Port
		s.unixInet4.Addr = t.Addr
		sa = &s.unixInet4
	case *wasi.Inet6Address:
		s.unixInet6.Port = t.Port
		s.unixInet6.Addr = t.Addr
		sa = &s.unixInet6
	case *wasi.UnixAddress:
		s.unixUnix.Name = t.Name
		sa = &s.unixUnix
	default:
		return nil, false
	}
	return sa, true
}

func (s *System) fromUnixSockAddress(sa unix.Sockaddr) (addr wasi.SocketAddress, ok bool) {
	switch t := sa.(type) {
	case *unix.SockaddrInet4:
		s.wasiInet4.Addr = t.Addr
		s.wasiInet4.Port = t.Port
		addr = &s.wasiInet4
	case *unix.SockaddrInet6:
		s.wasiInet6.Addr = t.Addr
		s.wasiInet6.Port = t.Port
		addr = &s.wasiInet6
	case *unix.SockaddrUnix:
		s.wasiUnix.Name = t.Name
		addr = &s.wasiUnix
	default:
		return nil, false
	}
	return addr, true
}
