package wasi

import (
	"context"
	"fmt"
	"io"
)

// Tracer wraps a System to log calls.
type Tracer struct {
	Writer io.Writer
	System
}

func (t *Tracer) Preopen(hostfd int, path string, fdstat FDStat) {
	t.printf("Preopen(%d, %q, ", hostfd, path)
	t.printFDStat(fdstat)
	t.printf(") => ")
	t.System.Preopen(hostfd, path, fdstat)
	t.printf("ok\n")
}

func (t *Tracer) Register(hostfd int, fdstat FDStat) FD {
	t.printf("Register(%d, ", hostfd)
	t.printFDStat(fdstat)
	t.printf(") => ")
	fd := t.System.Register(hostfd, fdstat)
	t.printf("%d\n", fd)
	return fd
}

func (t *Tracer) ArgsGet(ctx context.Context) ([]string, Errno) {
	t.printf("ArgsGet() => ")
	args, errno := t.System.ArgsGet(ctx)
	if errno == ESUCCESS {
		t.printf("%q", args)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return args, errno
}

func (t *Tracer) EnvironGet(ctx context.Context) ([]string, Errno) {
	t.printf("EnvironGet() => ")
	environ, errno := t.System.EnvironGet(ctx)
	if errno == ESUCCESS {
		t.printf("%q", environ)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return environ, errno
}

func (t *Tracer) ClockResGet(ctx context.Context, id ClockID) (Timestamp, Errno) {
	t.printf("ClockResGet(%d) => ", id)
	precision, errno := t.System.ClockResGet(ctx, id)
	if errno == ESUCCESS {
		t.printf("%d", precision)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return precision, errno
}

func (t *Tracer) ClockTimeGet(ctx context.Context, id ClockID, precision Timestamp) (Timestamp, Errno) {
	t.printf("ClockTimeGet(%d, %d) => ", id, precision)
	timestamp, errno := t.System.ClockTimeGet(ctx, id, precision)
	if errno == ESUCCESS {
		t.printf("%d", timestamp)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return timestamp, errno
}

func (t *Tracer) FDAdvise(ctx context.Context, fd FD, offset, length FileSize, advice Advice) Errno {
	t.printf("FDAdvise(%d, %d, %d, %s) => ", fd, offset, length, advice)
	errno := t.System.FDAdvise(ctx, fd, offset, length, advice)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDAllocate(ctx context.Context, fd FD, offset, length FileSize) Errno {
	t.printf("FDAllocate(%d, %d, %d) => ", fd, offset, length)
	errno := t.System.FDAllocate(ctx, fd, offset, length)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDClose(ctx context.Context, fd FD) Errno {
	t.printf("FDClose(%d) => ", fd)
	errno := t.System.FDClose(ctx, fd)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDDataSync(ctx context.Context, fd FD) Errno {
	t.printf("FDDataSync(%d) => ", fd)
	errno := t.System.FDDataSync(ctx, fd)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDStatGet(ctx context.Context, fd FD) (FDStat, Errno) {
	t.printf("FDStatGet(%d) => ", fd)
	fdstat, errno := t.System.FDStatGet(ctx, fd)
	if errno == ESUCCESS {
		t.printFDStat(fdstat)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return fdstat, errno
}

func (t *Tracer) FDStatSetFlags(ctx context.Context, fd FD, flags FDFlags) Errno {
	t.printf("FDStatSetFlags(%d, %s) => ", fd, flags)
	errno := t.System.FDStatSetFlags(ctx, fd, flags)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDFileStatGet(ctx context.Context, fd FD) (FileStat, Errno) {
	t.printf("FDFileStatGet(%d) => ", fd)
	filestat, errno := t.System.FDFileStatGet(ctx, fd)
	if errno == ESUCCESS {
		t.printFileStat(filestat)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return filestat, errno
}

func (t *Tracer) FDFileStatSetSize(ctx context.Context, fd FD, size FileSize) Errno {
	t.printf("FDFileStatSetSize(%d, %d) => ", fd, size)
	errno := t.System.FDFileStatSetSize(ctx, fd, size)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDFileStatSetTimes(ctx context.Context, fd FD, accessTime, modifyTime Timestamp, flags FSTFlags) Errno {
	t.printf("FDFileStatSetTimes(%d, %d, %d, %s) => ", fd, accessTime, modifyTime, flags)
	errno := t.System.FDFileStatSetTimes(ctx, fd, accessTime, modifyTime, flags)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDPread(ctx context.Context, fd FD, iovecs []IOVec, offset FileSize) (Size, Errno) {
	t.printf("FDPread(%d, ", fd)
	t.printIOVecsProto(iovecs)
	t.printf("%d) => ", offset)
	n, errno := t.System.FDRead(ctx, fd, iovecs)
	if errno == ESUCCESS {
		t.printf("[%d]byte: ", n)
		t.printIOVecs(iovecs, int(n))
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return n, errno
}

func (t *Tracer) FDPreStatGet(ctx context.Context, fd FD) (PreStat, Errno) {
	t.printf("FDPreStatGet(%d) => ", fd)
	prestat, errno := t.System.FDPreStatGet(ctx, fd)
	if errno == ESUCCESS {
		t.printf("{Type:%s,PreStatDir.NameLength:%d}", prestat.Type, prestat.PreStatDir.NameLength)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return prestat, errno
}

func (t *Tracer) FDPreStatDirName(ctx context.Context, fd FD) (string, Errno) {
	t.printf("FDPreStatDirName(%d) => ", fd)
	name, errno := t.System.FDPreStatDirName(ctx, fd)
	if errno == ESUCCESS {
		t.printf("%q", name)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return name, errno
}

func (t *Tracer) FDPwrite(ctx context.Context, fd FD, iovecs []IOVec, offset FileSize) (Size, Errno) {
	t.printf("FDPwrite(%d, ", fd)
	t.printIOVecs(iovecs, -1)
	t.printf(", %d) => ", offset)
	n, errno := t.System.FDPwrite(ctx, fd, iovecs, offset)
	if errno == ESUCCESS {
		t.printf("%d", n)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return n, errno
}

func (t *Tracer) FDRead(ctx context.Context, fd FD, iovecs []IOVec) (Size, Errno) {
	t.printf("FDRead(%d, ", fd)
	t.printIOVecsProto(iovecs)
	t.printf(") => ")
	n, errno := t.System.FDRead(ctx, fd, iovecs)
	if errno == ESUCCESS {
		t.printf("[%d]byte: ", n)
		t.printIOVecs(iovecs, int(n))
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return n, errno
}

func (t *Tracer) FDReadDir(ctx context.Context, fd FD, entries []DirEntry, cookie DirCookie, bufferSizeBytes int) (int, Errno) {
	t.printf("FDReadDir(%d, %d) => ", fd, cookie)
	n, errno := t.System.FDReadDir(ctx, fd, entries, cookie, bufferSizeBytes)
	if errno == ESUCCESS {
		t.printf("%d", n) // TODO: better output here
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return n, errno
}

func (t *Tracer) FDRenumber(ctx context.Context, from, to FD) Errno {
	t.printf("FDRenumber(%d, %d) => ", from, to)
	errno := t.System.FDRenumber(ctx, from, to)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDSeek(ctx context.Context, fd FD, offset FileDelta, whence Whence) (FileSize, Errno) {
	t.printf("FDSeek(%d, %d, %s) => ", fd, offset, whence)
	result, errno := t.System.FDSeek(ctx, fd, offset, whence)
	if errno == ESUCCESS {
		t.printf("%d", offset)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return result, errno
}

func (t *Tracer) FDSync(ctx context.Context, fd FD) Errno {
	t.printf("FDSync(%d) => ", fd)
	errno := t.System.FDSync(ctx, fd)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) FDTell(ctx context.Context, fd FD) (FileSize, Errno) {
	t.printf("FDTell(%d) => ", fd)
	fileSize, errno := t.System.FDTell(ctx, fd)
	if errno == ESUCCESS {
		t.printf("%d", fileSize)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return fileSize, errno
}

func (t *Tracer) FDWrite(ctx context.Context, fd FD, iovecs []IOVec) (Size, Errno) {
	t.printf("FDWrite(%d, ", fd)
	t.printIOVecs(iovecs, -1)
	t.printf(") => ")
	n, errno := t.System.FDWrite(ctx, fd, iovecs)
	if errno == ESUCCESS {
		t.printf("%d", n)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return n, errno
}

func (t *Tracer) PathCreateDirectory(ctx context.Context, fd FD, path string) Errno {
	t.printf("PathCreateDirectory(%d, %q) => ", fd, path)
	errno := t.System.PathCreateDirectory(ctx, fd, path)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) PathFileStatGet(ctx context.Context, fd FD, lookupFlags LookupFlags, path string) (FileStat, Errno) {
	t.printf("PathFileStatGet(%d, %s, %q) => ", fd, lookupFlags, path)
	filestat, errno := t.System.PathFileStatGet(ctx, fd, lookupFlags, path)
	if errno == ESUCCESS {
		t.printFileStat(filestat)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return filestat, errno
}

func (t *Tracer) PathFileStatSetTimes(ctx context.Context, fd FD, lookupFlags LookupFlags, path string, accessTime, modifyTime Timestamp, flags FSTFlags) Errno {
	t.printf("PathFileStatSetTimes(%d, %s, %q, %d, %d, %s) => ", fd, lookupFlags, path, accessTime, modifyTime, flags)
	errno := t.System.PathFileStatSetTimes(ctx, fd, lookupFlags, path, accessTime, modifyTime, flags)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) PathLink(ctx context.Context, oldFD FD, oldFlags LookupFlags, oldPath string, newFD FD, newPath string) Errno {
	t.printf("PathLink(%d, %s, %q, %d, %q) => ", oldFD, oldFlags, oldPath, newFD, newPath)
	errno := t.System.PathLink(ctx, oldFD, oldFlags, oldPath, newFD, newPath)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) PathOpen(ctx context.Context, fd FD, dirFlags LookupFlags, path string, openFlags OpenFlags, rightsBase, rightsInheriting Rights, fdFlags FDFlags) (FD, Errno) {
	t.printf("PathOpen(%d, %s, %q, %s, %s, %s, %s) => ", fd, dirFlags, path, openFlags, rightsBase, rightsInheriting, fdFlags)
	fd, errno := t.System.PathOpen(ctx, fd, dirFlags, path, openFlags, rightsBase, rightsInheriting, fdFlags)
	if errno == ESUCCESS {
		t.printf("%d", fd)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return fd, errno
}

func (t *Tracer) PathReadLink(ctx context.Context, fd FD, path string, buffer []byte) ([]byte, Errno) {
	t.printf("PathReadLink(%d, %q, [%d]byte) => ", fd, path, len(buffer))
	result, errno := t.System.PathReadLink(ctx, fd, path, buffer)
	if errno == ESUCCESS {
		t.printBytes(result)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return result, errno
}

func (t *Tracer) PathRemoveDirectory(ctx context.Context, fd FD, path string) Errno {
	t.printf("PathRemoveDirectory(%d, %q) => ", fd, path)
	errno := t.System.PathRemoveDirectory(ctx, fd, path)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) PathRename(ctx context.Context, fd FD, oldPath string, newFD FD, newPath string) Errno {
	t.printf("PathRename(%d, %q, %d, %q) => ", fd, oldPath, newFD, newPath)
	errno := t.System.PathRename(ctx, fd, oldPath, newFD, newPath)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) PathSymlink(ctx context.Context, oldPath string, fd FD, newPath string) Errno {
	t.printf("PathSymlink(%q, %d, %q) => ", oldPath, fd, newPath)
	errno := t.System.PathSymlink(ctx, oldPath, fd, newPath)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) PathUnlinkFile(ctx context.Context, fd FD, path string) Errno {
	t.printf("PathUnlinkFile(%d, %q) => ", fd, path)
	errno := t.System.PathUnlinkFile(ctx, fd, path)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) PollOneOff(ctx context.Context, subscriptions []Subscription, events []Event) (int, Errno) {
	t.printf("PollOneoff(")
	for i, s := range subscriptions {
		if i > 0 {
			t.printf(",")
		}
		t.printSubscription(s)
	}
	t.printf(") => ")
	n, errno := t.System.PollOneOff(ctx, subscriptions, events)
	switch {
	case errno == ESUCCESS && n == 0:
		t.printf("{}")
	case errno == ESUCCESS:
		for i, e := range events[:n] {
			if i > 0 {
				t.printf(",")
			}
			t.printEvent(e)
		}
	default:
		t.printErrno(errno)
	}
	t.printf("\n")
	return n, errno
}

func (t *Tracer) ProcExit(ctx context.Context, exitCode ExitCode) Errno {
	t.printf("ProcExit(%d) => ", exitCode)
	errno := t.System.ProcExit(ctx, exitCode)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) ProcRaise(ctx context.Context, signal Signal) Errno {
	t.printf("ProcRaise(%d) => ", signal)
	errno := t.System.ProcRaise(ctx, signal)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) SchedYield(ctx context.Context) Errno {
	t.printf("SchedYield() => ")
	errno := t.System.SchedYield(ctx)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) RandomGet(ctx context.Context, b []byte) Errno {
	t.printf("RandomGet([%d]byte) => ", len(b))
	errno := t.System.RandomGet(ctx, b)
	if errno == ESUCCESS {
		t.printBytes(b)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

func (t *Tracer) SockAccept(ctx context.Context, fd FD, flags FDFlags) (FD, Errno) {
	t.printf("SockAccept(%d, %s) => ", fd, flags)
	newfd, errno := t.System.SockAccept(ctx, fd, flags)
	if errno == ESUCCESS {
		t.printf("%d", newfd)
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return newfd, errno
}

func (t *Tracer) SockShutdown(ctx context.Context, fd FD, flags SDFlags) Errno {
	t.printf("SockShutdown(%d, %s) => ", fd, flags)
	errno := t.System.SockShutdown(ctx, fd, flags)
	if errno == ESUCCESS {
		t.printf("ok")
	} else {
		t.printErrno(errno)
	}
	t.printf("\n")
	return errno
}

// TODO: SockRecv(ctx context.Context, fd FD, iovecs []IOVec, flags RIFlags) (Size, ROFlags, Errno)
// TODO: SockSend(ctx context.Context, fd FD, iovecs []IOVec, flags SIFlags) (Size, Errno)

func (t *Tracer) Close(ctx context.Context) error {
	t.printf("Close() => ")
	err := t.System.Close(ctx)
	if err == nil {
		t.printf("ok\n")
	} else {
		t.printf("error (%s)\n", err)
	}
	return err
}

func (t *Tracer) printf(msg string, args ...interface{}) {
	fmt.Fprintf(t.Writer, msg, args...)
}

func (t *Tracer) printErrno(errno Errno) {
	t.printf("%s (%s)", errno.Name(), errno.Error())
}

func (t *Tracer) printSubscription(s Subscription) {
	t.printf("{EventType:%s,UserData:%#x,", s.EventType, s.UserData)
	if s.EventType == ClockEvent {
		c := s.GetClock()
		t.printf("ID:%s,", c.ID)
		if c.Flags != 0 {
			t.printf("Flags:%s,", c.Flags)
		}
		t.printf("Timeout:%d,Precision:%d}", c.Timeout, c.Precision)
	} else {
		fdrw := s.GetFDReadWrite()
		t.printf("FD:%d}", fdrw.FD)
	}
}

func (t *Tracer) printEvent(e Event) {
	t.printf("{EventType:%s,UserData:%#x", e.EventType, e.UserData)
	if e.Errno != 0 {
		t.printf(",Errno:%s", e.Errno.Name())
	}
	if e.EventType != ClockEvent {
		fdrw := e.FDReadWrite
		if fdrw.Flags != 0 {
			t.printf(",Flags:%s", fdrw.Flags)
		}
		t.printf(",NBytes:%d}", fdrw.NBytes)
	}
}

func (t *Tracer) printFDStat(s FDStat) {
	t.printf("{FileType:%s,Flags:%s,RightsBase:%s,RightsInheriting:%s})",
		s.FileType, s.Flags, s.RightsBase, s.RightsInheriting)
}

func (t *Tracer) printFileStat(s FileStat) {
	t.printf("%#v", s)
}

func (t *Tracer) printIOVecsProto(iovecs []IOVec) {
	t.printf("[%d]IOVec{", len(iovecs))
	for i, iovec := range iovecs {
		if i > 0 {
			t.printf(",")
		}
		t.printf("[%d]Byte", len(iovec))
	}
	t.printf("}")
}

func (t *Tracer) printIOVecs(iovecs []IOVec, size int) {
	t.printf("[%d]IOVec{", len(iovecs))
	for i, iovec := range iovecs {
		if i > 0 {
			t.printf(",")
		}
		switch {
		case size < 0:
			t.printBytes(iovec)
		case size > 0 && len(iovec) > size:
			t.printBytes(iovec[:size])
			size = 0
		case size > 0 && len(iovec) <= size:
			t.printBytes(iovec)
			size -= len(iovec)
		case size == 0:
			t.printf("[%d]Byte", len(iovec))
		}
	}
	t.printf("}")
}

const maxBytes = 32

func (t *Tracer) printBytes(b []byte) {
	t.printf("[%d]byte(\"", len(b))

	if len(b) > 0 {
		trunc := b
		if len(b) > maxBytes {
			trunc = trunc[:maxBytes]
		}
		for _, c := range trunc {
			if c < 32 || c >= 127 || c == '"' {
				t.printf("\\")
				switch {
				case c <= 9:
					t.printf("%d", c)
				case c == '"':
					t.printf("\"")
				case c == '\\':
					t.printf("\\")
				case c == '\r':
					t.printf("r")
				case c == '\n':
					t.printf("n")
				case c == '\t':
					t.printf("t")
				default:
					t.printf(`x%02x`, c)
				}
			} else {
				t.printf(`%c`, c)
			}
		}
	}
	t.printf("\"")
	if len(b) > maxBytes {
		t.printf("...")
	}
	t.printf(")")
}
