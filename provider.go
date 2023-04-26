package wasip1

// Provider provides an implementation of WASI preview 1.
type Provider interface {
	// ArgsGet reads command-line argument data.
	ArgsGet() ([]string, Errno)

	// EnvironGet reads environment variable data.
	//
	// Key/value pairs are expected to be joined with an '=' char.
	EnvironGet() ([]string, Errno)

	// ClockResGet returns the resolution of a clock.
	//
	// The function accepts the clock ID for which to return the resolution.
	//
	// Implementations are required to provide a non-zero value for supported
	// clocks. For unsupported clocks, EINVAL is returned.
	//
	// Note: This is similar to clock_getres in POSIX.
	ClockResGet(ClockID) (Timestamp, Errno)

	// ClockTimeGet returns the time value of a clock.
	//
	// The function accepts the clock ID for which to return the time. It
	// also accepts a precision which represents the maximum lag (exclusive)
	// that the returned time value may have, compared to its actual value.
	//
	// Note: This is similar to clock_gettime in POSIX.
	ClockTimeGet(id ClockID, precision Timestamp) (Timestamp, Errno)

	// FDAdvise provides file advisory information on a file descriptor
	//
	// Note: This is similar to posix_fadvise in POSIX.
	FDAdvise(fd FD, offset FileSize, length FileSize, advice Advice) Errno

	// FDAllocate forces the allocation of space in a file.
	//
	// Note: This is similar to posix_fallocate in POSIX.
	FDAllocate(fd FD, offset FileSize, length FileSize) Errno

	// FDClose closes a file descriptor.
	//
	// Note: This is similar to close in POSIX.
	FDClose(FD) Errno

	// FDDataSync synchronizes the data of a file to disk.
	//
	// Note: This is similar to fdatasync in POSIX.
	FDDataSync(FD) Errno

	// FDStatGet gets the attributes of a file descriptor.
	//
	// Note: This returns similar flags to fcntl(fd, F_GETFL) in POSIX, as
	// well as additional fields.
	FDStatGet(FD) (FDStat, Errno)

	// FDStatSetFlags adjusts the flags associated with a file descriptor.
	//
	// Note: This is similar to fcntl(fd, F_SETFL, flags) in POSIX.
	FDStatSetFlags(FD, FDFlags) Errno

	// FDStatSetRights adjusts the rights associated with a file descriptor.
	//
	// This can only be used to remove rights, and returns ENOTCAPABLE if
	// called in a way that would attempt to add rights.
	FDStatSetRights(fd FD, rightsBase, rightsInheriting Rights) Errno

	// FDFileStatGet returns the attributes of an open file.
	FDFileStatGet(FD) (FileStat, Errno)

	// FDFileStatSetSize adjusts the size of an open file.
	//
	// If this increases the file's size, the extra bytes are filled with
	// zeros.
	//
	// Note: This is similar to ftruncate in POSIX.
	FDFileStatSetSize(FD, FileSize) Errno

	// FDFileStatSetTimes adjusts the timestamps of an open file or directory.
	//
	// Note: This is similar to futimens in POSIX.
	FDFileStatSetTimes(fd FD, accessTime, modifyTime Timestamp, flags FSTFlags) Errno

	// FDPread reads from a file descriptor, without using and updating the
	// file descriptor's offset.
	//
	// On success, it returns the number of bytes read. On failure, it returns
	// an Errno.
	//
	// Note: This is similar to preadv in Linux (and other Unix-es).
	FDPread(fd FD, iovecs []IOVec, offset FileSize) (Size, Errno)

	// FDPreStatGet returns a description of the given pre-opened file
	// descriptor.
	FDPreStatGet(FD) (PreStat, Errno)

	// FDPreStatDirName returns a description of the given pre-opened file
	// descriptor.
	FDPreStatDirName(FD) (string, Errno)

	// FDPwrite writes from a file descriptor, without using and updating the
	// file descriptor's offset.
	//
	// On success, it returns the number of bytes read. On failure, it returns
	// an Errno.
	//
	// Note: This is similar to pwritev in Linux (and other Unix-es).
	//
	// Like Linux (and other Unix-es), any calls of pwrite (and other functions
	// to read or write) for a regular file by other threads in the WASI
	// process should not be interleaved while pwrite is executed.
	FDPwrite(fd FD, iovecs []IOVec, offset FileSize) (Size, Errno)

	// FDRead reads from a file descriptor.
	//
	// On success, it returns the number of bytes read. On failure, it returns
	// an Errno.
	//
	// Note: This is similar to readv in POSIX.
	FDRead(FD, []IOVec) (Size, Errno)

	// FDReadDir reads directory entries from a directory.
	//
	// At most cap(output) entries will be written. On success, the function
	// returns the number of entries written. On failure, it returns an Errno.
	FDReadDir(fd FD, output []DirEntry, cookie DirCookie) (Size, Errno)

	// FDRenumber atomically replaces a file descriptor by renumbering another
	// file descriptor. Due to the strong focus on thread safety, this
	// environment does not provide a mechanism to duplicate or renumber a file
	// descriptor to an arbitrary number, like dup2(). This would be prone to
	// race conditions, as an actual file descriptor with the same number could
	// be allocated by a different thread at the same time. This function
	// provides a way to atomically renumber file descriptors, which would
	// disappear if dup2() were to be removed entirely.
	FDRenumber(from, to FD) Errno

	// FDSeek moves the offset of a file descriptor.
	//
	// On success, this returns the new offset of the file descriptor, relative
	// to the start of the file. On failure, it returns an Errno.
	//
	// Note: This is similar to lseek in POSIX.
	FDSeek(FD, FileDelta, Whence) (FileSize, Errno)

	// FDSync synchronizes the data and metadata of a file to disk.
	//
	// Note: This is similar to fsync in POSIX.
	FDSync(FD) Errno

	// FDTell returns the current offset of a file descriptor.
	//
	// Note: This is similar to lseek(fd, 0, SEEK_CUR) in POSIX.
	FDTell(FD) (FileSize, Errno)

	// FDWrite write to a file descriptor.
	//
	// Note: This is similar to writev in POSIX.
	//
	// Like POSIX, any calls of write (and other functions to read or write)
	// for a regular file by other threads in the WASI process should not be
	// interleaved while write is executed.
	FDWrite(FD, []IOVec) (Size, Errno)

	// PathCreateDirectory create a directory.
	//
	// Note: This is similar to mkdirat in POSIX.
	PathCreateDirectory(FD, string) Errno

	// PathFileStatGet returns the attributes of a file or directory.
	//
	// Note: This is similar to stat in POSIX.
	PathFileStatGet(FD, LookupFlags, string) (FileStat, Errno)

	// PathFileStatSetTimes adjusts the timestamps of a file or directory.
	//
	// Note: This is similar to utimensat in POSIX.
	PathFileStatSetTimes(fd FD, accessTime, modifyTime Timestamp, flags FSTFlags) Errno

	// PathLink creates a hard link.
	//
	// Note: This is similar to linkat in POSIX.
	PathLink(oldFD FD, oldFlags LookupFlags, oldPath string, newFD FD, newPath string) Errno

	// PathOpen opens a file or directory.
	//
	// The returned file descriptor is not guaranteed to be the lowest-numbered
	// file descriptor not currently open; it is randomized to prevent
	// applications from depending on making assumptions about indexes, since
	// this is error-prone in multi-threaded contexts. The returned file
	// descriptor is guaranteed to be less than 2**31.
	//
	// Note: This is similar to openat in POSIX.
	PathOpen(fd FD, dirFlags LookupFlags, path string, openFlags OpenFlags, rightsBase, rightsInheriting Rights, fdFlags FDFlags) (FD, Errno)

	// PathReadLink reads the contents of a symbolic link.
	//
	// Note: This is similar to readlinkat in POSIX.
	PathReadLink(FD, string) (string, Size, Errno)

	// PathRemoveDirectory removes a directory.
	//
	// If the directory is not empty, ENOTEMPTY is returned.
	//
	// Note: This is similar to unlinkat(fd, path, AT_REMOVEDIR) in POSIX.
	PathRemoveDirectory(FD, string) Errno

	// PathRename renames a file or directory.
	//
	// Note: This is similar to renameat in POSIX.
	PathRename(fd FD, oldPath string, newFD FD, newPath string) Errno

	// PathSymlink creates a symbolic link.
	//
	// Note: This is similar to symlinkat in POSIX.
	PathSymlink(oldPath string, fd FD, newPath string) Errno

	// PathUnlinkFile unlinks a file.
	//
	// If the path refers to a directory, EISDIR is returned.
	//
	// Note: This is similar to unlinkat(fd, path, 0) in POSIX.
	PathUnlinkFile(FD, string) Errno

	// PollOneOff concurrently polls for the occurrence of a set of events.
	//
	// If len(subscriptions) == 0, EINVAL is returned.
	//
	// The function uses the provided []Event buffer to write output events.
	// If there is not enough space (len(subscriptions) > cap(events)) a new
	// buffer will be created and returned.
	PollOneOff(subscriptions []Subscription, events []Event) ([]Event, Errno)

	// ProcExit terminates the process normally.
	//
	// An exit code of 0 indicates successful termination of the program.
	// The meanings of other values is dependent on the environment.
	ProcExit(ExitCode) Errno

	// ProcRaise sends a signal to the process of the calling thread.
	//
	// Note: This is similar to raise in POSIX.
	ProcRaise(Signal) Errno

	// SchedYield temporarily yields execution of the calling thread.
	//
	// Note: This is similar to sched_yield in POSIX.
	SchedYield() Errno

	// RandomGet write high-quality random data into a buffer.
	//
	// This function blocks when the implementation is unable to immediately
	// provide sufficient high-quality random data. This function may execute
	// slowly, so when large mounts of random data are required, it's
	// advisable to use this function to seed a pseudo-random number generator,
	// rather than to provide the random data directly.
	RandomGet(b []byte) Errno

	// SockAccept accepts a new incoming connection.
	//
	// Note: This is similar to accept in POSIX.
	SockAccept(FD, FDFlags) (FD, Errno)

	// SockRecv receives a message from a socket.
	//
	// On success, this returns the number of bytes read along with
	// output flags. On failure, this returns an Errno.
	//
	// Note: This is similar to recv in POSIX, though it also supports reading
	// the data into multiple buffers in the manner of readv.
	SockRecv(FD, []IOVec, RIFlags) (Size, ROFlags, Errno)

	// SockSend sends a message on a socket.
	//
	// On success, this returns the number of bytes written. On failure, this
	// returns an Errno.
	//
	// Note: This is similar to send in POSIX, though it also supports
	// writing the data from multiple buffers in the manner of writev.
	SockSend(FD, []IOVec, SIFlags) (Size, Errno)

	// SockShutdown shuts down a socket's send and/or receive channels.
	//
	// Note: This is similar to shutdown in POSIX.
	SockShutdown(FD, SDFlags) Errno

	// Close closes the Provider.
	Close() error
}

// IOVec is a slice of bytes.
type IOVec []byte
