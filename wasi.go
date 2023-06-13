package wasi

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/stealthrocket/wasi-go/internal/descriptor"
)

// File is an interface used as constraint in the FileType generic type
// parameter.
//
// File implement the WASI functions which operate on a file descriptor number.
type File[T any] interface {
	FDAdvise(ctx context.Context, offset, length FileSize, advice Advice) Errno

	FDAllocate(ctx context.Context, offset, length FileSize) Errno

	FDClose(ctx context.Context) Errno

	FDDataSync(ctx context.Context) Errno

	FDStatSetFlags(ctx context.Context, flags FDFlags) Errno

	FDFileStatGet(ctx context.Context) (FileStat, Errno)

	FDFileStatSetSize(ctx context.Context, size FileSize) Errno

	FDFileStatSetTimes(ctx context.Context, accessTime, modifyTime Timestamp, flags FSTFlags) Errno

	FDPread(ctx context.Context, iovecs []IOVec, offset FileSize) (Size, Errno)

	FDPwrite(ctx context.Context, iovecs []IOVec, offset FileSize) (Size, Errno)

	FDRead(ctx context.Context, iovecs []IOVec) (Size, Errno)

	FDWrite(ctx context.Context, iovecs []IOVec) (Size, Errno)

	FDSync(ctx context.Context) Errno

	FDSeek(ctx context.Context, delta FileDelta, whence Whence) (FileSize, Errno)

	FDOpenDir(ctx context.Context) (Dir, Errno)

	PathCreateDirectory(ctx context.Context, path string) Errno

	PathFileStatGet(ctx context.Context, flags LookupFlags, path string) (FileStat, Errno)

	PathFileStatSetTimes(ctx context.Context, lookupFlags LookupFlags, path string, accessTime, modifyTime Timestamp, flags FSTFlags) Errno

	PathLink(ctx context.Context, flags LookupFlags, oldPath string, newFile T, newPath string) Errno

	PathOpen(ctx context.Context, lookupFlags LookupFlags, path string, openFlags OpenFlags, rightsBase, rightsInheriting Rights, fdFlags FDFlags) (T, Errno)

	PathReadLink(ctx context.Context, path string, buffer []byte) (int, Errno)

	PathRemoveDirectory(ctx context.Context, path string) Errno

	PathRename(ctx context.Context, oldPath string, newFile T, newPath string) Errno

	PathSymlink(ctx context.Context, oldPath string, newPath string) Errno

	PathUnlinkFile(ctx context.Context, path string) Errno
}

// Dir instances are returned by File.FDOpenDir and used to iterate over
type Dir interface {
	FDReadDir(ctx context.Context, entries []DirEntry, cookie DirCookie, bufferSizeBytes int) (int, Errno)

	FDCloseDir(ctx context.Context) Errno
}

// FileTable is a building block used to construct implementations of the System
// interface.
//
// The file table maintains the set of open files and associates them with file
// descriptor numbers.
//
// The type paritally implements the System interface, it is common to embed
// a FileTable field in a struct in order to inherit its methods. The generic
// type allows for specialization of the behavior of files, for example:
//
//	// System embeds a wasi.FileTable to implements most of the wasi.System
//	// interface methods.
//	type System struct {
//		wasi.FileTable[File]
//		...
//	}
//
//	// File implements the wasi.File interface to specialize the behavior of
//	// WASI functions.
//	type File struct {
//		...
//	}
type FileTable[T File[T]] struct {
	files    descriptor.Table[FD, fileEntry[T]]
	preopens descriptor.Table[FD, string]
	dirs     map[FD]Dir
}

type fileEntry[T File[T]] struct {
	file T
	stat FDStat
}

func (t *FileTable[T]) Close(ctx context.Context) error {
	t.files.Range(func(fd FD, f fileEntry[T]) bool {
		f.file.FDClose(ctx)
		return true
	})
	t.files.Reset()
	t.preopens.Reset()
	for _, dir := range t.dirs {
		dir.FDCloseDir(ctx)
	}
	for fd := range t.dirs {
		delete(t.dirs, fd)
	}
	return nil
}

func (t *FileTable[T]) Preopen(file T, path string, stat FDStat) FD {
	fd := t.Register(file, stat)
	t.preopens.Assign(fd, path)
	return fd
}

func (t *FileTable[T]) Register(file T, stat FDStat) FD {
	stat.RightsBase &= AllRights
	stat.RightsInheriting &= AllRights
	return t.files.Insert(fileEntry[T]{file: file, stat: stat})
}

func (t *FileTable[T]) LookupFD(fd FD, rights Rights) (file T, stat FDStat, errno Errno) {
	f, errno := t.lookupFD(fd, rights)
	if f != nil {
		file = f.file
		stat = f.stat
	}
	return file, stat, errno
}

func (t *FileTable[T]) LookupSocketFD(fd FD, rights Rights) (file T, stat FDStat, errno Errno) {
	f, errno := t.lookupSocketFD(fd, rights)
	if f != nil {
		file = f.file
		stat = f.stat
	}
	return file, stat, errno
}

func (t *FileTable[T]) isPreopen(fd FD) bool {
	return t.preopens.Access(fd) != nil
}

func (t *FileTable[T]) lookupFD(fd FD, rights Rights) (*fileEntry[T], Errno) {
	f := t.files.Access(fd)
	if f == nil {
		return nil, EBADF
	}
	if !f.stat.RightsBase.Has(rights) {
		return nil, ENOTCAPABLE
	}
	return f, ESUCCESS
}

func (t *FileTable[T]) lookupPreopenPath(fd FD) (string, Errno) {
	path, ok := t.preopens.Lookup(fd)
	if !ok {
		return "", EBADF
	}
	f := t.files.Access(fd)
	if f == nil {
		return "", EBADF
	}
	if f.stat.FileType != DirectoryType {
		return "", ENOTDIR
	}
	return path, ESUCCESS
}

func (t *FileTable[T]) lookupSocketFD(fd FD, rights Rights) (*fileEntry[T], Errno) {
	f := t.files.Access(fd)
	if f == nil {
		return nil, EBADF
	}
	switch f.stat.FileType {
	case SocketStreamType:
	case SocketDGramType:
	default:
		return nil, ENOTSOCK
	}
	if !f.stat.RightsBase.Has(rights) {
		return nil, ENOTCAPABLE
	}
	return f, ESUCCESS
}

func (t *FileTable[T]) FDAdvise(ctx context.Context, fd FD, offset FileSize, length FileSize, advice Advice) Errno {
	f, errno := t.lookupFD(fd, FDAdviseRight)
	if errno != ESUCCESS {
		return errno
	}
	return f.file.FDAdvise(ctx, offset, length, advice)
}

func (t *FileTable[T]) FDAllocate(ctx context.Context, fd FD, offset FileSize, length FileSize) Errno {
	f, errno := t.lookupFD(fd, FDAllocateRight)
	if errno != ESUCCESS {
		return errno
	}
	return f.file.FDAllocate(ctx, offset, length)
}

func (t *FileTable[T]) FDClose(ctx context.Context, fd FD) Errno {
	f, errno := t.lookupFD(fd, 0)
	if errno != ESUCCESS {
		return errno
	}
	// We capture the file before removing the table entry because f is a
	// pointer into the table and gets erased when the descriptor is deleted.
	file := f.file
	t.files.Delete(fd)
	// Note: closing pre-opens is allowed.
	// See github.com/WebAssembly/wasi-testsuite/blob/1b1d4a5/tests/rust/src/bin/close_preopen.rs
	t.preopens.Delete(fd)
	if dir := t.dirs[fd]; dir != nil {
		delete(t.dirs, fd)
		dir.FDCloseDir(ctx)
	}
	return file.FDClose(ctx)
}

func (t *FileTable[T]) FDDataSync(ctx context.Context, fd FD) Errno {
	f, errno := t.lookupFD(fd, FDDataSyncRight)
	if errno != ESUCCESS {
		return errno
	}
	return f.file.FDDataSync(ctx)
}

func (t *FileTable[T]) FDStatGet(ctx context.Context, fd FD) (FDStat, Errno) {
	f, errno := t.lookupFD(fd, 0)
	if errno != ESUCCESS {
		return FDStat{}, errno
	}
	return f.stat, ESUCCESS
}

func (t *FileTable[T]) FDStatSetFlags(ctx context.Context, fd FD, flags FDFlags) Errno {
	f, errno := t.lookupFD(fd, FDStatSetFlagsRight)
	if errno != ESUCCESS {
		return errno
	}
	changes := flags ^ f.stat.Flags
	if changes == 0 {
		return ESUCCESS
	}
	if changes.Has(Sync | DSync | RSync) {
		return ENOSYS // TODO: support changing {Sync,DSync,Rsync}
	}
	if errno := f.file.FDStatSetFlags(ctx, flags); errno != ESUCCESS {
		return errno
	}
	f.stat.Flags ^= changes
	return ESUCCESS
}

func (t *FileTable[T]) FDStatSetRights(ctx context.Context, fd FD, rightsBase, rightsInheriting Rights) Errno {
	f, errno := t.lookupFD(fd, 0)
	if errno != ESUCCESS {
		return errno
	}
	// Rights can only be preserved or removed, not added.
	rightsBase &= AllRights
	rightsInheriting &= AllRights
	if (rightsBase &^ f.stat.RightsBase) != 0 {
		return ENOTCAPABLE
	}
	if (rightsInheriting &^ f.stat.RightsInheriting) != 0 {
		return ENOTCAPABLE
	}
	f.stat.RightsBase &= rightsBase
	f.stat.RightsInheriting &= rightsInheriting
	return ESUCCESS
}

func (t *FileTable[T]) FDFileStatSetSize(ctx context.Context, fd FD, size FileSize) Errno {
	f, errno := t.lookupFD(fd, FDFileStatSetSizeRight)
	if errno != ESUCCESS {
		return errno
	}
	return f.file.FDFileStatSetSize(ctx, size)
}

func (t *FileTable[T]) FDFileStatGet(ctx context.Context, fd FD) (FileStat, Errno) {
	f, errno := t.lookupFD(fd, FDFileStatGetRight)
	if errno != ESUCCESS {
		return FileStat{}, errno
	}
	s, errno := f.file.FDFileStatGet(ctx)
	if errno != ESUCCESS {
		return FileStat{}, errno
	}
	if fd <= 2 {
		// Override stdio size/times.
		// See github.com/WebAssembly/wasi-testsuite/blob/1b1d4a5/tests/rust/src/bin/fd_filestat_get.rs
		s.Size = 0
		s.AccessTime = 0
		s.ModifyTime = 0
		s.ChangeTime = 0
	}
	return s, ESUCCESS
}

func (t *FileTable[T]) FDFileStatSetTimes(ctx context.Context, fd FD, accessTime, modifyTime Timestamp, flags FSTFlags) Errno {
	f, errno := t.lookupFD(fd, FDFileStatSetTimesRight)
	if errno != ESUCCESS {
		return errno
	}
	return f.file.FDFileStatSetTimes(ctx, accessTime, modifyTime, flags)
}

func (t *FileTable[T]) FDPreStatGet(ctx context.Context, fd FD) (PreStat, Errno) {
	path, errno := t.lookupPreopenPath(fd)
	if errno != ESUCCESS {
		return PreStat{}, errno
	}
	stat := PreStat{
		Type: PreOpenDir,
		PreStatDir: PreStatDir{
			NameLength: Size(len(path)),
		},
	}
	return stat, ESUCCESS
}

func (t *FileTable[T]) FDPreStatDirName(ctx context.Context, fd FD) (string, Errno) {
	return t.lookupPreopenPath(fd)
}

func (t *FileTable[T]) FDPread(ctx context.Context, fd FD, iovecs []IOVec, offset FileSize) (Size, Errno) {
	f, errno := t.lookupFD(fd, FDReadRight|FDSeekRight)
	if errno != ESUCCESS {
		return 0, errno
	}
	return f.file.FDPread(ctx, iovecs, offset)
}

func (t *FileTable[T]) FDPwrite(ctx context.Context, fd FD, iovecs []IOVec, offset FileSize) (Size, Errno) {
	f, errno := t.lookupFD(fd, FDWriteRight|FDSeekRight)
	if errno != ESUCCESS {
		return 0, errno
	}
	return f.file.FDPwrite(ctx, iovecs, offset)
}

func (t *FileTable[T]) FDRead(ctx context.Context, fd FD, iovecs []IOVec) (Size, Errno) {
	f, errno := t.lookupFD(fd, FDReadRight)
	if errno != ESUCCESS {
		return 0, errno
	}
	return f.file.FDRead(ctx, iovecs)
}

func (t *FileTable[T]) FDWrite(ctx context.Context, fd FD, iovecs []IOVec) (Size, Errno) {
	f, errno := t.lookupFD(fd, FDWriteRight)
	if errno != ESUCCESS {
		return 0, errno
	}
	return f.file.FDWrite(ctx, iovecs)
}

func (t *FileTable[T]) FDReadDir(ctx context.Context, fd FD, entries []DirEntry, cookie DirCookie, bufferSizeBytes int) (int, Errno) {
	f, errno := t.lookupFD(fd, FDReadDirRight)
	if errno != ESUCCESS {
		return 0, errno
	}
	if len(entries) == 0 {
		return 0, EINVAL
	}
	d := t.dirs[fd]
	if d == nil {
		d, errno = f.file.FDOpenDir(ctx)
		if errno != ESUCCESS {
			return 0, errno
		}
		if t.dirs == nil {
			t.dirs = make(map[FD]Dir)
		}
		t.dirs[fd] = d
	}
	return d.FDReadDir(ctx, entries, cookie, bufferSizeBytes)
}

func (t *FileTable[T]) FDRenumber(ctx context.Context, from, to FD) Errno {
	if t.isPreopen(from) || t.isPreopen(to) {
		return ENOTSUP
	}
	f, errno := t.lookupFD(from, 0)
	if errno != ESUCCESS {
		return errno
	}
	d := t.dirs[from]
	// TODO: limit max file descriptor number
	g, replaced := t.files.Assign(to, *f)
	if replaced {
		g.file.FDClose(ctx)
		if dir := t.dirs[to]; dir != nil {
			dir.FDCloseDir(ctx)
		}
	}
	t.files.Delete(from)
	if d != nil {
		delete(t.dirs, from)
		t.dirs[to] = d
	}
	return ESUCCESS
}

func (t *FileTable[T]) FDSync(ctx context.Context, fd FD) Errno {
	f, errno := t.lookupFD(fd, FDSyncRight)
	if errno != ESUCCESS {
		return errno
	}
	return f.file.FDSync(ctx)
}

func (t *FileTable[T]) FDSeek(ctx context.Context, fd FD, delta FileDelta, whence Whence) (FileSize, Errno) {
	// Note: FDSeekRight implies FDTellRight. FDTellRight also includes the
	// right to invoke FDSeek in such a way that the file offset remains
	// unaltered.
	f, errno := t.lookupFD(fd, FDSeekRight)
	if errno != ESUCCESS {
		if errno != ENOTCAPABLE || (delta != 0 && whence != SeekCurrent) {
			return 0, errno
		}
		f, errno = t.lookupFD(fd, FDTellRight)
		if errno != ESUCCESS {
			return 0, errno
		}
	}
	return f.file.FDSeek(ctx, delta, whence)
}

func (t *FileTable[T]) FDTell(ctx context.Context, fd FD) (FileSize, Errno) {
	return t.FDSeek(ctx, fd, 0, SeekCurrent)
}

func (t *FileTable[T]) PathCreateDirectory(ctx context.Context, fd FD, path string) Errno {
	d, errno := t.lookupFD(fd, PathCreateDirectoryRight)
	if errno != ESUCCESS {
		return errno
	}
	return d.file.PathCreateDirectory(ctx, path)
}

func (t *FileTable[T]) PathFileStatGet(ctx context.Context, fd FD, lookupFlags LookupFlags, path string) (FileStat, Errno) {
	d, errno := t.lookupFD(fd, PathFileStatGetRight)
	if errno != ESUCCESS {
		return FileStat{}, errno
	}
	return d.file.PathFileStatGet(ctx, lookupFlags, path)
}

func (t *FileTable[T]) PathFileStatSetTimes(ctx context.Context, fd FD, lookupFlags LookupFlags, path string, accessTime, modifyTime Timestamp, fstFlags FSTFlags) Errno {
	d, errno := t.lookupFD(fd, PathFileStatSetTimesRight)
	if errno != ESUCCESS {
		return errno
	}
	return d.file.PathFileStatSetTimes(ctx, lookupFlags, path, accessTime, modifyTime, fstFlags)
}

func (t *FileTable[T]) PathLink(ctx context.Context, fd FD, flags LookupFlags, oldPath string, newFD FD, newPath string) Errno {
	oldDir, errno := t.lookupFD(fd, PathLinkSourceRight)
	if errno != ESUCCESS {
		return errno
	}
	newDir, errno := t.lookupFD(newFD, PathLinkTargetRight)
	if errno != ESUCCESS {
		return errno
	}
	return oldDir.file.PathLink(ctx, flags, oldPath, newDir.file, newPath)
}

func (t *FileTable[T]) PathOpen(ctx context.Context, fd FD, lookupFlags LookupFlags, path string, openFlags OpenFlags, rightsBase, rightsInheriting Rights, fdFlags FDFlags) (FD, Errno) {
	d, errno := t.lookupFD(fd, PathOpenRight)
	if errno != ESUCCESS {
		return -1, errno
	}
	clean := filepath.Clean(path)
	if strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "../") {
		return -1, EPERM
	}

	// Rights can only be preserved or removed, not added.
	rightsBase &= AllRights
	rightsInheriting &= AllRights
	if (rightsBase &^ d.stat.RightsInheriting) != 0 {
		return -1, ENOTCAPABLE
	} else if (rightsInheriting &^ d.stat.RightsInheriting) != 0 {
		return -1, ENOTCAPABLE
	}
	rightsBase &= d.stat.RightsInheriting
	rightsInheriting &= d.stat.RightsInheriting

	if openFlags.Has(OpenDirectory) {
		rightsBase &= DirectoryRights
	}
	if openFlags.Has(OpenCreate) {
		if !d.stat.RightsBase.Has(PathCreateFileRight) {
			return -1, ENOTCAPABLE
		}
	}
	if openFlags.Has(OpenTruncate) {
		if !d.stat.RightsBase.Has(PathFileStatSetSizeRight) {
			return -1, ENOTCAPABLE
		}
	}

	newFile, errno := d.file.PathOpen(ctx, lookupFlags, path, openFlags, rightsBase, rightsInheriting, fdFlags)
	if errno != ESUCCESS {
		return -1, errno
	}

	fileType := RegularFileType
	if openFlags.Has(OpenDirectory) {
		fileType = DirectoryType
	}

	newFD := t.Register(newFile, FDStat{
		FileType:         fileType,
		Flags:            fdFlags,
		RightsBase:       rightsBase,
		RightsInheriting: rightsInheriting,
	})
	return newFD, ESUCCESS
}

func (t *FileTable[T]) PathReadLink(ctx context.Context, fd FD, path string, buffer []byte) (int, Errno) {
	d, errno := t.lookupFD(fd, PathReadLinkRight)
	if errno != ESUCCESS {
		return 0, errno
	}
	return d.file.PathReadLink(ctx, path, buffer)
}

func (t *FileTable[T]) PathRemoveDirectory(ctx context.Context, fd FD, path string) Errno {
	d, errno := t.lookupFD(fd, PathRemoveDirectoryRight)
	if errno != ESUCCESS {
		return errno
	}
	return d.file.PathRemoveDirectory(ctx, path)
}

func (t *FileTable[T]) PathRename(ctx context.Context, fd FD, oldPath string, newFD FD, newPath string) Errno {
	oldDir, errno := t.lookupFD(fd, PathRenameSourceRight)
	if errno != ESUCCESS {
		return errno
	}
	newDir, errno := t.lookupFD(newFD, PathRenameTargetRight)
	if errno != ESUCCESS {
		return errno
	}
	return oldDir.file.PathRename(ctx, oldPath, newDir.file, newPath)
}

func (t *FileTable[T]) PathSymlink(ctx context.Context, oldPath string, fd FD, newPath string) Errno {
	d, errno := t.lookupFD(fd, PathSymlinkRight)
	if errno != ESUCCESS {
		return errno
	}
	return d.file.PathSymlink(ctx, oldPath, newPath)
}

func (t *FileTable[T]) PathUnlinkFile(ctx context.Context, fd FD, path string) Errno {
	d, errno := t.lookupFD(fd, PathUnlinkFileRight)
	if errno != ESUCCESS {
		return errno
	}
	return d.file.PathUnlinkFile(ctx, path)
}

// SizesGet is a helper function used to implement the ArgsSizesGet and
// EnvironSizesGet methods of the System interface. Given a list of values
// it returns the count and byte size of their representation in the ABI.
func SizesGet(values []string) (count, size int) {
	for _, value := range values {
		size += len(value) + 1
	}
	return len(values), size
}
