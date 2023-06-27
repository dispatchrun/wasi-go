package wasi

import "context"

// System is the WebAssembly System Interface (WASI).
type System interface {
	// ArgsSizesGet reads command-line argument data sizes.
	//
	// The implementation should return the number of args, and the number of
	// bytes required to hold the strings (including terminating null bytes).
	ArgsSizesGet(ctx context.Context) (argCount int, stringBytes int, errno Errno)

	// ArgsGet reads command-line argument data.
	ArgsGet(ctx context.Context) ([]string, Errno)

	// EnvironSizesGet reads environment variable data sizes.
	//
	// The implementation should return the number of env variables, and the
	// number of bytes required to hold the strings (including terminating
	// null bytes).
	EnvironSizesGet(ctx context.Context) (argCount int, stringBytes int, errno Errno)

	// EnvironGet reads environment variable data.
	//
	// Key/value pairs are expected to be joined with an '=' char.
	EnvironGet(ctx context.Context) ([]string, Errno)

	// ClockResGet returns the resolution of a clock.
	//
	// The function accepts the clock ID for which to return the resolution.
	//
	// Implementations are required to provide a non-zero value for supported
	// clocks. For unsupported clocks, EINVAL is returned.
	//
	// Note: This is similar to clock_getres in POSIX.
	ClockResGet(ctx context.Context, id ClockID) (Timestamp, Errno)

	// ClockTimeGet returns the time value of a clock.
	//
	// The function accepts the clock ID for which to return the time. It
	// also accepts a precision which represents the maximum lag (exclusive)
	// that the returned time value may have, compared to its actual value.
	//
	// Note: This is similar to clock_gettime in POSIX.
	ClockTimeGet(ctx context.Context, id ClockID, precision Timestamp) (Timestamp, Errno)

	// FDAdvise provides file advisory information on a file descriptor
	//
	// Note: This is similar to posix_fadvise in POSIX.
	FDAdvise(ctx context.Context, fd FD, offset FileSize, length FileSize, advice Advice) Errno

	// FDAllocate forces the allocation of space in a file.
	//
	// Note: This is similar to posix_fallocate in POSIX.
	FDAllocate(ctx context.Context, fd FD, offset FileSize, length FileSize) Errno

	// FDClose closes a file descriptor.
	//
	// Note: This is similar to close in POSIX.
	FDClose(ctx context.Context, fd FD) Errno

	// FDDataSync synchronizes the data of a file to disk.
	//
	// Note: This is similar to fdatasync in POSIX.
	FDDataSync(ctx context.Context, fd FD) Errno

	// FDStatGet gets the attributes of a file descriptor.
	//
	// Note: This returns similar flags to fcntl(fd, F_GETFL) in POSIX, as
	// well as additional fields.
	FDStatGet(ctx context.Context, fd FD) (FDStat, Errno)

	// FDStatSetFlags adjusts the flags associated with a file descriptor.
	//
	// Note: This is similar to fcntl(fd, F_SETFL, flags) in POSIX.
	FDStatSetFlags(ctx context.Context, fd FD, flags FDFlags) Errno

	// FDStatSetRights adjusts the rights associated with a file descriptor.
	//
	// This can only be used to remove rights, and returns ENOTCAPABLE if
	// called in a way that would attempt to add rights.
	FDStatSetRights(ctx context.Context, fd FD, rightsBase, rightsInheriting Rights) Errno

	// FDFileStatGet returns the attributes of an open file.
	FDFileStatGet(ctx context.Context, fd FD) (FileStat, Errno)

	// FDFileStatSetSize adjusts the size of an open file.
	//
	// If this increases the file's size, the extra bytes are filled with
	// zeros.
	//
	// Note: This is similar to ftruncate in POSIX.
	FDFileStatSetSize(ctx context.Context, fd FD, size FileSize) Errno

	// FDFileStatSetTimes adjusts the timestamps of an open file or directory.
	//
	// Note: This is similar to futimens in POSIX.
	FDFileStatSetTimes(ctx context.Context, fd FD, accessTime, modifyTime Timestamp, flags FSTFlags) Errno

	// FDPread reads from a file descriptor, without using and updating the
	// file descriptor's offset.
	//
	// On success, it returns the number of bytes read. On failure, it returns
	// an Errno.
	//
	// Note: This is similar to preadv in Linux (and other Unix-es).
	FDPread(ctx context.Context, fd FD, iovecs []IOVec, offset FileSize) (Size, Errno)

	// FDPreStatGet returns a description of the given pre-opened file
	// descriptor.
	FDPreStatGet(ctx context.Context, fd FD) (PreStat, Errno)

	// FDPreStatDirName returns a description of the given pre-opened file
	// descriptor.
	FDPreStatDirName(ctx context.Context, fd FD) (string, Errno)

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
	FDPwrite(ctx context.Context, fd FD, iovecs []IOVec, offset FileSize) (Size, Errno)

	// FDRead reads from a file descriptor.
	//
	// On success, it returns the number of bytes read. On failure, it returns
	// an Errno.
	//
	// Note: This is similar to readv in POSIX.
	FDRead(ctx context.Context, fd FD, iovecs []IOVec) (Size, Errno)

	// FDReadDir reads directory entries from a directory.
	//
	// The implementation must write entries to the provided entries buffer,
	// and return the number of entries written.
	//
	// The implementation must ensure that the entries fit into a buffer
	// with the specified size (bufferSizeBytes). It's ok if the final entry
	// only partially fits into such a buffer.
	FDReadDir(ctx context.Context, fd FD, entries []DirEntry, cookie DirCookie, bufferSizeBytes int) (int, Errno)

	// FDRenumber atomically replaces a file descriptor by renumbering another
	// file descriptor. Due to the strong focus on thread safety, this
	// environment does not provide a mechanism to duplicate or renumber a file
	// descriptor to an arbitrary number, like dup2(). This would be prone to
	// race conditions, as an actual file descriptor with the same number could
	// be allocated by a different thread at the same time. This function
	// provides a way to atomically renumber file descriptors, which would
	// disappear if dup2() were to be removed entirely.
	FDRenumber(ctx context.Context, from, to FD) Errno

	// FDSeek moves the offset of a file descriptor.
	//
	// On success, this returns the new offset of the file descriptor, relative
	// to the start of the file. On failure, it returns an Errno.
	//
	// Note: This is similar to lseek in POSIX.
	FDSeek(ctx context.Context, fd FD, offset FileDelta, whence Whence) (FileSize, Errno)

	// FDSync synchronizes the data and metadata of a file to disk.
	//
	// Note: This is similar to fsync in POSIX.
	FDSync(ctx context.Context, fd FD) Errno

	// FDTell returns the current offset of a file descriptor.
	//
	// Note: This is similar to lseek(fd, 0, SEEK_CUR) in POSIX.
	FDTell(ctx context.Context, fd FD) (FileSize, Errno)

	// FDWrite write to a file descriptor.
	//
	// Note: This is similar to writev in POSIX.
	//
	// Like POSIX, any calls of write (and other functions to read or write)
	// for a regular file by other threads in the WASI process should not be
	// interleaved while write is executed.
	FDWrite(ctx context.Context, fd FD, iovecs []IOVec) (Size, Errno)

	// PathCreateDirectory create a directory.
	//
	// Note: This is similar to mkdirat in POSIX.
	PathCreateDirectory(ctx context.Context, fd FD, path string) Errno

	// PathFileStatGet returns the attributes of a file or directory.
	//
	// Note: This is similar to stat in POSIX.
	PathFileStatGet(ctx context.Context, fd FD, lookupFlags LookupFlags, path string) (FileStat, Errno)

	// PathFileStatSetTimes adjusts the timestamps of a file or directory.
	//
	// Note: This is similar to utimensat in POSIX.
	PathFileStatSetTimes(ctx context.Context, fd FD, lookupFlags LookupFlags, path string, accessTime, modifyTime Timestamp, flags FSTFlags) Errno

	// PathLink creates a hard link.
	//
	// Note: This is similar to linkat in POSIX.
	PathLink(ctx context.Context, oldFD FD, oldFlags LookupFlags, oldPath string, newFD FD, newPath string) Errno

	// PathOpen opens a file or directory.
	//
	// The returned file descriptor is not guaranteed to be the lowest-numbered
	// file descriptor not currently open; it is randomized to prevent
	// applications from depending on making assumptions about indexes, since
	// this is error-prone in multi-threaded contexts. The returned file
	// descriptor is guaranteed to be less than 2**31.
	//
	// Note: This is similar to openat in POSIX.
	PathOpen(ctx context.Context, fd FD, dirFlags LookupFlags, path string, openFlags OpenFlags, rightsBase, rightsInheriting Rights, fdFlags FDFlags) (FD, Errno)

	// PathReadLink reads the contents of a symbolic link.
	//
	// The implementation must read the path into the specified buffer and
	// returns the number of bytes written. If the buffer is not large enough
	// to hold the contents of the symbolic link, the implementation must
	// return ERANGE.
	//
	// Note: This is similar to readlinkat in POSIX.
	PathReadLink(ctx context.Context, fd FD, path string, buffer []byte) (int, Errno)

	// PathRemoveDirectory removes a directory.
	//
	// If the directory is not empty, ENOTEMPTY is returned.
	//
	// Note: This is similar to unlinkat(fd, path, AT_REMOVEDIR) in POSIX.
	PathRemoveDirectory(ctx context.Context, fd FD, path string) Errno

	// PathRename renames a file or directory.
	//
	// Note: This is similar to renameat in POSIX.
	PathRename(ctx context.Context, fd FD, oldPath string, newFD FD, newPath string) Errno

	// PathSymlink creates a symbolic link.
	//
	// Note: This is similar to symlinkat in POSIX.
	PathSymlink(ctx context.Context, oldPath string, fd FD, newPath string) Errno

	// PathUnlinkFile unlinks a file.
	//
	// If the path refers to a directory, EISDIR is returned.
	//
	// Note: This is similar to unlinkat(fd, path, 0) in POSIX.
	PathUnlinkFile(ctx context.Context, fd FD, path string) Errno

	// PollOneOff concurrently polls for the occurrence of a set of events.
	//
	// If len(subscriptions) == 0, EINVAL is returned.
	//
	// The function writes events to the provided []Event buffer, expecting
	// len(events)>=len(subscriptions).
	PollOneOff(ctx context.Context, subscriptions []Subscription, events []Event) (int, Errno)

	// ProcExit terminates the process normally.
	//
	// An exit code of 0 indicates successful termination of the program.
	// The meanings of other values is dependent on the environment.
	ProcExit(ctx context.Context, exitCode ExitCode) Errno

	// ProcRaise sends a signal to the process of the calling thread.
	//
	// Note: This is similar to raise in POSIX.
	ProcRaise(ctx context.Context, signal Signal) Errno

	// SchedYield temporarily yields execution of the calling thread.
	//
	// Note: This is similar to sched_yield in POSIX.
	SchedYield(ctx context.Context) Errno

	// RandomGet write high-quality random data into a buffer.
	//
	// This function blocks when the implementation is unable to immediately
	// provide sufficient high-quality random data. This function may execute
	// slowly, so when large mounts of random data are required, it's
	// advisable to use this function to seed a pseudo-random number generator,
	// rather than to provide the random data directly.
	RandomGet(ctx context.Context, b []byte) Errno

	// SockOpen opens a socket.
	//
	// Note: This is similar to socket in POSIX.
	SockOpen(ctx context.Context, family ProtocolFamily, socketType SocketType, protocol Protocol, rightsBase, rightsInheriting Rights) (FD, Errno)

	// SockBind binds a socket to an address.
	//
	// The method returns the address that the socket has been bound to, which
	// may differ from the one passed as argument. For example, in cases where
	// the caller used an address with port 0, and the system is responsible for
	// selecting a free port to bind the socket to.
	//
	// The implementation must not retain the socket address.
	//
	// Note: This is similar to bind in POSIX.
	SockBind(ctx context.Context, fd FD, addr SocketAddress) (SocketAddress, Errno)

	// SockConnect connects a socket to an address, returning the local socket
	// address that the connection was made from.
	//
	// The implementation must not retain the socket address.
	//
	// Note: This is similar to connect in POSIX.
	SockConnect(ctx context.Context, fd FD, addr SocketAddress) (SocketAddress, Errno)

	// SockListen allows the socket to accept connections with SockAccept.
	//
	// Note: This is similar to listen in POSIX.
	SockListen(ctx context.Context, fd FD, backlog int) Errno

	// SockAccept accepts a new incoming connection.
	//
	// The method returns a pair of socket addresses where the first one is the
	// local server address that accepted the connection, and the second is the
	// peer address that the connection was established from.
	//
	// Although the method returns the address of the connecting entity, WASI
	// preview 1 does not currently support passing the address to the calling
	// WebAssembly module via the "sock_accept" host function call. This
	// address is only used by implementations and wrappers of the System
	// interface, and is discarded before returning control to the WebAssembly
	// module.
	//
	// Note: This is similar to accept in POSIX.
	SockAccept(ctx context.Context, fd FD, flags FDFlags) (newfd FD, peer, addr SocketAddress, err Errno)

	// SockRecv receives a message from a socket.
	//
	// On success, this returns the number of bytes read along with
	// output flags. On failure, this returns an Errno.
	//
	// Note: This is similar to recv in POSIX, though it also supports reading
	// the data into multiple buffers in the manner of readv.
	SockRecv(ctx context.Context, fd FD, iovecs []IOVec, flags RIFlags) (Size, ROFlags, Errno)

	// SockSend sends a message on a socket.
	//
	// On success, this returns the number of bytes written. On failure, this
	// returns an Errno.
	//
	// Note: This is similar to send in POSIX, though it also supports
	// writing the data from multiple buffers in the manner of writev.
	SockSend(ctx context.Context, fd FD, iovecs []IOVec, flags SIFlags) (Size, Errno)

	// SockSendTo sends a message on a socket.
	//
	// It's similar to SockSend, but accepts an additional SocketAddress.
	//
	// Note: This is similar to sendto in POSIX, though it also supports
	// writing the data from multiple buffers in the manner of writev.
	SockSendTo(ctx context.Context, fd FD, iovecs []IOVec, flags SIFlags, addr SocketAddress) (Size, Errno)

	// SockRecvFrom receives a message from a socket.
	//
	// It's similar to SockRecv, but returns an additional SocketAddress.
	//
	// Note: This is similar to recvfrom in POSIX, though it also supports reading
	// the data into multiple buffers in the manner of readv.
	SockRecvFrom(ctx context.Context, fd FD, iovecs []IOVec, flags RIFlags) (Size, ROFlags, SocketAddress, Errno)

	// SockGetOpt gets a socket option.
	//
	// Note: This is similar to getsockopt in POSIX.
	SockGetOpt(ctx context.Context, fd FD, option SocketOption) (SocketOptionValue, Errno)

	// SockSetOpt sets a socket option.
	//
	// Note: This is similar to setsockopt in POSIX.
	SockSetOpt(ctx context.Context, fd FD, option SocketOption, value SocketOptionValue) Errno

	// SockLocalAddress gets the local address of the socket.
	//
	// The returned address is only valid until the next call on this
	// interface. Assume that any method may invalidate the address.
	//
	// Note: This is similar to getsockname in POSIX.
	SockLocalAddress(ctx context.Context, fd FD) (SocketAddress, Errno)

	// SockRemoteAddress gets the address of the peer when the socket is a
	// connection.
	//
	// The returned address is only valid until the next call on this
	// interface. Assume that any method may invalidate the address.
	//
	// Note: This is similar to getpeername in POSIX.
	SockRemoteAddress(ctx context.Context, fd FD) (SocketAddress, Errno)

	// SockAddressInfo get a list of IP addresses and port numbers for a
	// host name and service.
	//
	// The function populates the AddressInfo.Address fields of the provided
	// results slice, and returns a count indicating how many results were
	// written.
	//
	// The returned addresses are only valid until the next call on this
	// interface. Assume that any method may invalidate the addresses.
	//
	// Note: This is similar to getaddrinfo in POSIX.
	SockAddressInfo(ctx context.Context, name, service string, hints AddressInfo, results []AddressInfo) (int, Errno)

	// SockShutdown shuts down a socket's send and/or receive channels.
	//
	// Note: This is similar to shutdown in POSIX.
	SockShutdown(ctx context.Context, fd FD, flags SDFlags) Errno

	// Close closes the System.
	Close(ctx context.Context) error
}
