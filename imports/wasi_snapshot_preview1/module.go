package wasi_snapshot_preview1

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wazergo"
	. "github.com/stealthrocket/wazergo/types"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/sys"
)

const moduleName = "wasi_snapshot_preview1"

// HostModule is a wazero host module for WASI.
//
// The host module manages the interaction between the host and the
// guest WASM module. The host module does not implement WASI on its
// own, and instead calls out to an implementation of the wasi.System
// interface provided via the WithWASI host module Option. This design
// means that the implementation doesn't have to concern itself with ABI
// details nor access the guest's memory.
var HostModule wazergo.HostModule[*Module] = functions{
	// WASI preview 1.
	"args_get":                wazergo.F2((*Module).ArgsGet),
	"args_sizes_get":          wazergo.F2((*Module).ArgsSizesGet),
	"environ_get":             wazergo.F2((*Module).EnvironGet),
	"environ_sizes_get":       wazergo.F2((*Module).EnvironSizesGet),
	"clock_res_get":           wazergo.F2((*Module).ClockResGet),
	"clock_time_get":          wazergo.F3((*Module).ClockTimeGet),
	"fd_advise":               wazergo.F4((*Module).FDAdvise),
	"fd_allocate":             wazergo.F3((*Module).FDAllocate),
	"fd_close":                wazergo.F1((*Module).FDClose),
	"fd_datasync":             wazergo.F1((*Module).FDDataSync),
	"fd_fdstat_get":           wazergo.F2((*Module).FDStatGet),
	"fd_fdstat_set_flags":     wazergo.F2((*Module).FDStatSetFlags),
	"fd_fdstat_set_rights":    wazergo.F3((*Module).FDStatSetRights),
	"fd_filestat_get":         wazergo.F2((*Module).FDFileStatGet),
	"fd_filestat_set_size":    wazergo.F2((*Module).FDFileStatSetSize),
	"fd_filestat_set_times":   wazergo.F4((*Module).FDFileStatSetTimes),
	"fd_pread":                wazergo.F4((*Module).FDPread),
	"fd_prestat_get":          wazergo.F2((*Module).FDPreStatGet),
	"fd_prestat_dir_name":     wazergo.F2((*Module).FDPreStatDirName),
	"fd_pwrite":               wazergo.F4((*Module).FDPwrite),
	"fd_read":                 wazergo.F3((*Module).FDRead),
	"fd_readdir":              wazergo.F4((*Module).FDReadDir),
	"fd_renumber":             wazergo.F2((*Module).FDRenumber),
	"fd_seek":                 wazergo.F4((*Module).FDSeek),
	"fd_sync":                 wazergo.F1((*Module).FDSync),
	"fd_tell":                 wazergo.F2((*Module).FDTell),
	"fd_write":                wazergo.F3((*Module).FDWrite),
	"path_create_directory":   wazergo.F2((*Module).PathCreateDirectory),
	"path_filestat_get":       wazergo.F4((*Module).PathFileStatGet),
	"path_filestat_set_times": wazergo.F6((*Module).PathFileStatSetTimes),
	"path_link":               wazergo.F5((*Module).PathLink),
	"path_open":               wazergo.F8((*Module).PathOpen),
	"path_readlink":           wazergo.F4((*Module).PathReadLink),
	"path_remove_directory":   wazergo.F2((*Module).PathRemoveDirectory),
	"path_rename":             wazergo.F4((*Module).PathRename),
	"path_symlink":            wazergo.F3((*Module).PathSymlink),
	"path_unlink_file":        wazergo.F2((*Module).PathUnlinkFile),
	"poll_oneoff":             wazergo.F4((*Module).PollOneOff),
	"proc_exit":               procExitShape((*Module).ProcExit),
	"proc_raise":              wazergo.F1((*Module).ProcRaise),
	"sched_yield":             wazergo.F0((*Module).SchedYield),
	"random_get":              wazergo.F1((*Module).RandomGet),
	"sock_accept":             wazergo.F3((*Module).SockAccept),
	"sock_recv":               wazergo.F5((*Module).SockRecv),
	"sock_send":               wazergo.F4((*Module).SockSend),
	"sock_shutdown":           wazergo.F2((*Module).SockShutdown),

	// WasmEdge sockets extension.
	"sock_open":       wazergo.F3((*Module).SockOpen),
	"sock_bind":       wazergo.F3((*Module).SockBind),
	"sock_connect":    wazergo.F3((*Module).SockConnect),
	"sock_listen":     wazergo.F2((*Module).SockListen),
	"sock_getsockopt": wazergo.F5((*Module).SockGetOpt),
	"sock_setsockopt": wazergo.F5((*Module).SockSetOpt),
}

// Option configures the host module.
type Option = wazergo.Option[*Module]

// WithWASI sets the WASI implementation.
func WithWASI(system wasi.System) Option {
	return wazergo.OptionFunc(func(m *Module) {
		m.WASI = system
		if s, ok := system.(wasi.SocketsExtension); ok {
			m.Sockets = s
		}
	})
}

type functions wazergo.Functions[*Module]

func (f functions) Name() string {
	return moduleName
}

func (f functions) Functions() wazergo.Functions[*Module] {
	return (wazergo.Functions[*Module])(f)
}

func (f functions) Instantiate(ctx context.Context, opts ...Option) (*Module, error) {
	mod := &Module{}
	wazergo.Configure(mod, opts...)
	if mod.WASI == nil {
		return nil, fmt.Errorf("WASI implementation not provided")
	}
	return mod, nil
}

type Module struct {
	WASI    wasi.System
	Sockets wasi.SocketsExtension

	iovecs []wasi.IOVec
	dirent []wasi.DirEntry
}

func (m *Module) ArgsGet(ctx context.Context, argv Pointer[Uint32], buf Pointer[Uint8]) Errno {
	args, errno := m.WASI.ArgsGet(ctx)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	return m.storeArgs(args, argv, buf)
}

func (m *Module) ArgsSizesGet(ctx context.Context, argc, bufLen Pointer[Int32]) Errno {
	args, errno := m.WASI.ArgsGet(ctx)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	return m.countArgs(args, argc, bufLen)
}

func (m *Module) EnvironGet(ctx context.Context, envv Pointer[Uint32], buf Pointer[Uint8]) Errno {
	env, errno := m.WASI.EnvironGet(ctx)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	return m.storeArgs(env, envv, buf)
}

func (m *Module) EnvironSizesGet(ctx context.Context, envc, bufLen Pointer[Int32]) Errno {
	env, errno := m.WASI.EnvironGet(ctx)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	return m.countArgs(env, envc, bufLen)
}

func (m *Module) storeArgs(args []string, argv Pointer[Uint32], buf Pointer[Uint8]) Errno {
	offset := buf.Offset()
	memory := buf.Memory()
	for i, arg := range args {
		length := uint32(len(arg) + 1)
		b, ok := memory.Read(offset, length)
		if !ok {
			return Errno(wasi.EFAULT)
		}
		copy(b, arg)
		b[len(arg)] = 0
		argv.Index(i).Store(Uint32(offset))
		offset += length
	}
	return Errno(wasi.ESUCCESS)
}

func (m *Module) countArgs(args []string, argc, bufLen Pointer[Int32]) Errno {
	argc.Store(Int32(len(args)))
	var size int
	for _, arg := range args {
		size += len(arg) + 1 // include null terminator
	}
	bufLen.Store(Int32(size))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) ClockResGet(ctx context.Context, clockID Int32, precision Pointer[Uint64]) Errno {
	result, errno := m.WASI.ClockResGet(ctx, wasi.ClockID(clockID))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	precision.Store(Uint64(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) ClockTimeGet(ctx context.Context, clockID Int32, precision Uint64, timestamp Pointer[Uint64]) Errno {
	result, errno := m.WASI.ClockTimeGet(ctx, wasi.ClockID(clockID), wasi.Timestamp(precision))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	timestamp.Store(Uint64(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDAdvise(ctx context.Context, fd Int32, offset, length Uint64, advice Int32) Errno {
	return Errno(m.WASI.FDAdvise(ctx, wasi.FD(fd), wasi.FileSize(offset), wasi.FileSize(length), wasi.Advice(advice)))
}

func (m *Module) FDAllocate(ctx context.Context, fd Int32, offset, length Uint64) Errno {
	return Errno(m.WASI.FDAllocate(ctx, wasi.FD(fd), wasi.FileSize(offset), wasi.FileSize(length)))
}

func (m *Module) FDClose(ctx context.Context, fd Int32) Errno {
	return Errno(m.WASI.FDClose(ctx, wasi.FD(fd)))
}

func (m *Module) FDDataSync(ctx context.Context, fd Int32) Errno {
	return Errno(m.WASI.FDDataSync(ctx, wasi.FD(fd)))
}

func (m *Module) FDStatGet(ctx context.Context, fd Int32, stat Pointer[wasi.FDStat]) Errno {
	result, errno := m.WASI.FDStatGet(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	stat.Store(result)
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDStatSetFlags(ctx context.Context, fd Int32, flags Uint32) Errno {
	return Errno(m.WASI.FDStatSetFlags(ctx, wasi.FD(fd), wasi.FDFlags(flags)))
}

func (m *Module) FDStatSetRights(ctx context.Context, fd Int32, rightsBase, rightsInheriting Uint64) Errno {
	return Errno(m.WASI.FDStatSetRights(ctx, wasi.FD(fd), wasi.Rights(rightsBase), wasi.Rights(rightsInheriting)))
}

func (m *Module) FDFileStatGet(ctx context.Context, fd Int32, stat Pointer[wasi.FileStat]) Errno {
	result, errno := m.WASI.FDFileStatGet(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	stat.Store(result)
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDFileStatSetSize(ctx context.Context, fd Int32, size Uint64) Errno {
	return Errno(m.WASI.FDFileStatSetSize(ctx, wasi.FD(fd), wasi.FileSize(size)))
}

func (m *Module) FDFileStatSetTimes(ctx context.Context, fd Int32, accessTime, modifyTime Uint64, flags Int32) Errno {
	return Errno(m.WASI.FDFileStatSetTimes(ctx, wasi.FD(fd), wasi.Timestamp(accessTime), wasi.Timestamp(modifyTime), wasi.FSTFlags(flags)))
}

func (m *Module) FDPread(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], offset Uint64, nread Pointer[Int32]) Errno {
	m.iovecs = iovecs.Append(m.iovecs[:0])
	result, errno := m.WASI.FDPread(ctx, wasi.FD(fd), m.iovecs, wasi.FileSize(offset))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nread.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDPreStatGet(ctx context.Context, fd Int32, prestat Pointer[wasi.PreStat]) Errno {
	result, errno := m.WASI.FDPreStatGet(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	prestat.Store(result)
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDPreStatDirName(ctx context.Context, fd Int32, dirName Bytes) Errno {
	result, errno := m.WASI.FDPreStatDirName(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	if len(result) != len(dirName) {
		return Errno(wasi.EINVAL)
	}
	copy(dirName, result)
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDPwrite(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], offset Uint64, nwritten Pointer[Int32]) Errno {
	m.iovecs = iovecs.Append(m.iovecs[:0])
	result, errno := m.WASI.FDPwrite(ctx, wasi.FD(fd), m.iovecs, wasi.FileSize(offset))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDRead(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], nread Pointer[Int32]) Errno {
	m.iovecs = iovecs.Append(m.iovecs[:0])
	result, errno := m.WASI.FDRead(ctx, wasi.FD(fd), m.iovecs)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nread.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDReadDir(ctx context.Context, fd Int32, buf Bytes, cookie Uint64, nwritten Pointer[Int32]) Errno {
	if len(m.dirent) == 0 {
		m.dirent = make([]wasi.DirEntry, 1024)
	}

	var dirent [wasi.SizeOfDirent]byte
	var numBytes int

	for numBytes < len(buf) {
		n, errno := m.WASI.FDReadDir(ctx, wasi.FD(fd), m.dirent, wasi.DirCookie(cookie), len(buf)-numBytes)
		if errno != wasi.ESUCCESS {
			return Errno(errno)
		}
		if n == 0 {
			break
		}
		for _, d := range m.dirent[:n] {
			binary.LittleEndian.PutUint64(dirent[0:], uint64(d.Next))
			binary.LittleEndian.PutUint64(dirent[8:], uint64(d.INode))
			binary.LittleEndian.PutUint32(dirent[16:], uint32(len(d.Name)))
			binary.LittleEndian.PutUint32(dirent[20:], uint32(d.Type))

			numBytes += copy(buf[numBytes:], dirent[:])
			numBytes += copy(buf[numBytes:], d.Name)

			cookie = Uint64(d.Next)
		}
	}

	nwritten.Store(Int32(numBytes))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDRenumber(ctx context.Context, from, to Int32) Errno {
	return Errno(m.WASI.FDRenumber(ctx, wasi.FD(from), wasi.FD(to)))
}

func (m *Module) FDSeek(ctx context.Context, fd Int32, delta Int64, whence Int32, size Pointer[Uint64]) Errno {
	result, errno := m.WASI.FDSeek(ctx, wasi.FD(fd), wasi.FileDelta(delta), wasi.Whence(whence))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	size.Store(Uint64(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDSync(ctx context.Context, fd Int32) Errno {
	return Errno(m.WASI.FDSync(ctx, wasi.FD(fd)))
}

func (m *Module) FDTell(ctx context.Context, fd Int32, size Pointer[Uint64]) Errno {
	result, errno := m.WASI.FDTell(ctx, wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	size.Store(Uint64(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDWrite(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], nwritten Pointer[Int32]) Errno {
	m.iovecs = iovecs.Append(m.iovecs[:0])
	result, errno := m.WASI.FDWrite(ctx, wasi.FD(fd), m.iovecs)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) PathCreateDirectory(ctx context.Context, fd Int32, path String) Errno {
	return Errno(m.WASI.PathCreateDirectory(ctx, wasi.FD(fd), string(path)))
}

func (m *Module) PathFileStatGet(ctx context.Context, fd Int32, flags Int32, path String, stat Pointer[wasi.FileStat]) Errno {
	result, errno := m.WASI.PathFileStatGet(ctx, wasi.FD(fd), wasi.LookupFlags(flags), string(path))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	stat.Store(result)
	return Errno(wasi.ESUCCESS)
}

func (m *Module) PathFileStatSetTimes(ctx context.Context, fd Int32, lookupFlags Int32, path String, accessTime, modifyTime Uint64, fstFlags Int32) Errno {
	return Errno(m.WASI.PathFileStatSetTimes(ctx, wasi.FD(fd), wasi.LookupFlags(lookupFlags), string(path), wasi.Timestamp(accessTime), wasi.Timestamp(modifyTime), wasi.FSTFlags(fstFlags)))
}

func (m *Module) PathLink(ctx context.Context, oldFD Int32, oldFlags Int32, oldPath Bytes, newFD Int32, newPath Bytes) Errno {
	return Errno(m.WASI.PathLink(ctx, wasi.FD(oldFD), wasi.LookupFlags(oldFlags), string(oldPath), wasi.FD(newFD), string(newPath)))
}

func (m *Module) PathOpen(ctx context.Context, fd Int32, dirFlags Int32, path String, openFlags Int32, rightsBase, rightsInheriting Uint64, fdFlags Int32, openfd Pointer[Int32]) Errno {
	result, errno := m.WASI.PathOpen(ctx, wasi.FD(fd), wasi.LookupFlags(dirFlags), string(path), wasi.OpenFlags(openFlags), wasi.Rights(rightsBase), wasi.Rights(rightsInheriting), wasi.FDFlags(fdFlags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	openfd.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) PathReadLink(ctx context.Context, fd Int32, path String, buf Bytes, nwritten Pointer[Int32]) Errno {
	result, errno := m.WASI.PathReadLink(ctx, wasi.FD(fd), string(path), buf)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(len(result)))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) PathRemoveDirectory(ctx context.Context, fd Int32, path String) Errno {
	return Errno(m.WASI.PathRemoveDirectory(ctx, wasi.FD(fd), string(path)))
}

func (m *Module) PathRename(ctx context.Context, oldFD Int32, oldPath String, newFD Int32, newPath String) Errno {
	return Errno(m.WASI.PathRename(ctx, wasi.FD(oldFD), string(oldPath), wasi.FD(newFD), string(newPath)))
}

func (m *Module) PathSymlink(ctx context.Context, oldPath String, fd Int32, newPath String) Errno {
	return Errno(m.WASI.PathSymlink(ctx, string(oldPath), wasi.FD(fd), string(newPath)))
}

func (m *Module) PathUnlinkFile(ctx context.Context, fd Int32, path String) Errno {
	return Errno(m.WASI.PathUnlinkFile(ctx, wasi.FD(fd), string(path)))
}

func (m *Module) PollOneOff(ctx context.Context, in Pointer[wasi.Subscription], out Pointer[wasi.Event], nSubscriptions Int32, nEvents Pointer[Int32]) Errno {
	if nSubscriptions <= 0 {
		return Errno(wasi.EINVAL)
	}
	n, errno := m.WASI.PollOneOff(ctx,
		in.UnsafeSlice(int(nSubscriptions)),
		out.UnsafeSlice(int(nSubscriptions)),
	)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nEvents.Store(Int32(n))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) ProcExit(ctx context.Context, mod api.Module, exitCode Int32) {
	// Give the implementation a chance to exit.
	m.WASI.ProcExit(ctx, wasi.ExitCode(exitCode))

	// Ensure other callers see the exit code.
	_ = mod.CloseWithExitCode(ctx, uint32(exitCode))

	// Prevent any code from executing after this function. For example, LLVM
	// inserts unreachable instructions after calls to exit.
	// See: https://github.com/emscripten-core/emscripten/issues/12322
	panic(sys.NewExitError(uint32(exitCode)))
}

func (m *Module) ProcRaise(ctx context.Context, signal Int32) Errno {
	return Errno(m.WASI.ProcRaise(ctx, wasi.Signal(signal)))
}

func (m *Module) SchedYield(ctx context.Context) Errno {
	return Errno(m.WASI.SchedYield(ctx))
}

func (m *Module) RandomGet(ctx context.Context, buf Bytes) Errno {
	return Errno(m.WASI.RandomGet(ctx, buf))
}

func (m *Module) SockAccept(ctx context.Context, fd Int32, flags Int32, connfd Pointer[Int32]) Errno {
	result, errno := m.WASI.SockAccept(ctx, wasi.FD(fd), wasi.FDFlags(flags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	connfd.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) SockRecv(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], iflags Int32, nread Pointer[Int32], oflags Pointer[Int32]) Errno {
	m.iovecs = iovecs.Append(m.iovecs[:0])
	size, roflags, errno := m.WASI.SockRecv(ctx, wasi.FD(fd), m.iovecs, wasi.RIFlags(iflags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nread.Store(Int32(size))
	oflags.Store(Int32(roflags))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) SockSend(ctx context.Context, fd Int32, iovecs List[wasi.IOVec], flags Int32, nwritten Pointer[Int32]) Errno {
	m.iovecs = iovecs.Append(m.iovecs[:0])
	size, errno := m.WASI.SockSend(ctx, wasi.FD(fd), m.iovecs, wasi.SIFlags(flags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(size))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) SockShutdown(ctx context.Context, fd Int32, flags Int32) Errno {
	return Errno(m.WASI.SockShutdown(ctx, wasi.FD(fd), wasi.SDFlags(flags)))
}

func (m *Module) SockOpen(ctx context.Context, family Int32, sockType Int32, openfd Pointer[Int32]) Errno {
	if m.Sockets == nil {
		return Errno(wasi.ENOSYS)
	}
	result, errno := m.Sockets.SockOpen(wasi.ProtocolFamily(family), wasi.SocketType(sockType))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	openfd.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) SockBind(ctx context.Context, fd Int32, addr Pointer[wasi.SocketAddress], port Uint32) Errno {
	if m.Sockets == nil {
		return Errno(wasi.ENOSYS)
	}
	return Errno(m.Sockets.SockBind(wasi.FD(fd), addr.Load(), wasi.Port(port)))
}

func (m *Module) SockConnect(ctx context.Context, fd Int32, addr Pointer[wasi.SocketAddress], port Uint32) Errno {
	if m.Sockets == nil {
		return Errno(wasi.ENOSYS)
	}
	return Errno(m.Sockets.SockConnect(wasi.FD(fd), addr.Load(), wasi.Port(port)))
}

func (m *Module) SockListen(ctx context.Context, fd Int32, backlog Int32) Errno {
	if m.Sockets == nil {
		return Errno(wasi.ENOSYS)
	}
	return Errno(m.Sockets.SockListen(wasi.FD(fd), int(backlog)))
}

func (m *Module) SockSetOpt(ctx context.Context, fd Int32, level Int32, option Int32, value Pointer[Int32], valueLen Int32) Errno {
	if m.Sockets == nil {
		return Errno(wasi.ENOSYS)
	}
	if valueLen != 4 {
		// Only int options are supported for now.
		return Errno(wasi.EINVAL)
	}
	return Errno(m.Sockets.SockSetOptInt(wasi.FD(fd), wasi.SocketOptionLevel(level), wasi.SocketOption(option), int(value.Load())))
}

func (m *Module) SockGetOpt(ctx context.Context, fd Int32, level Int32, option Int32, value Pointer[Int32], valueLen Int32) Errno {
	if m.Sockets == nil {
		return Errno(wasi.ENOSYS)
	}
	if valueLen != 4 {
		// Only int options are supported for now.
		return Errno(wasi.EINVAL)
	}
	result, errno := m.Sockets.SockGetOptInt(wasi.FD(fd), wasi.SocketOptionLevel(level), wasi.SocketOption(option))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	value.Store(Int32(result))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) Close(ctx context.Context) error {
	return m.WASI.Close(ctx)
}

// procExit is a bit different; it doesn't have a return result,
// and needs access to api.Module.
func procExitShape[T any, P Param[P]](fn func(T, context.Context, api.Module, P)) wazergo.Function[T] {
	var arg P
	return wazergo.Function[T]{
		Params: []Value{arg},
		Func: func(this T, ctx context.Context, module api.Module, stack []uint64) {
			var arg P
			var memory = module.Memory()
			fn(this, ctx, module, arg.LoadValue(memory, stack))
		},
	}
}
