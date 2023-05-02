package wasizero

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/stealthrocket/wasi"
	"github.com/stealthrocket/wazergo"
	. "github.com/stealthrocket/wazergo/types"
	"github.com/stealthrocket/wazergo/wasm"
	"github.com/tetratelabs/wazero/api"
)

const moduleName = "wasi_snapshot_preview1"

// HostModule is a wazero host module for WASI preview 1.
//
// The host module manages the interaction between the host and the
// guest WASM module. The host module does not implement WASI preview 1 on its
// own, and instead calls out to an implementation of the wasi.Provider
// interface, provided via the WithWASI host module Option. The host module is
// only responsible for (de)serializing inputs and outputs, and for interacting
// with the guest's memory.
var HostModule wazergo.HostModule[*Module] = functions{
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
}

// Option configures the host module.
type Option = wazergo.Option[*Module]

// WithWASI sets the WASI implementation.
func WithWASI(wasi wasi.Provider) Option {
	return wazergo.OptionFunc(func(m *Module) { m.WASI = wasi })
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
	WASI wasi.Provider

	iovecs []wasi.IOVec
	dirent []wasi.DirEntryName
}

func (m *Module) ArgsGet(ctx context.Context, argv Pointer[Uint32], buf Pointer[Uint8]) Errno {
	args, errno := m.WASI.ArgsGet()
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	return m.storeArgs(args, argv, buf)
}

func (m *Module) ArgsSizesGet(ctx context.Context, argc, bufLen Pointer[Int32]) Errno {
	args, errno := m.WASI.ArgsGet()
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	return m.countArgs(args, argc, bufLen)
}

func (m *Module) EnvironGet(ctx context.Context, envv Pointer[Uint32], buf Pointer[Uint8]) Errno {
	env, errno := m.WASI.EnvironGet()
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	return m.storeArgs(env, envv, buf)
}

func (m *Module) EnvironSizesGet(ctx context.Context, envc, bufLen Pointer[Int32]) Errno {
	env, errno := m.WASI.EnvironGet()
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
	value, errno := m.WASI.ClockResGet(wasi.ClockID(clockID))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	precision.Store(Uint64(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) ClockTimeGet(ctx context.Context, clockID Int32, precision Uint64, timestamp Pointer[Uint64]) Errno {
	value, errno := m.WASI.ClockTimeGet(wasi.ClockID(clockID), wasi.Timestamp(precision))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	timestamp.Store(Uint64(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDAdvise(ctx context.Context, fd Int32, offset, length Uint64, advice Int32) Errno {
	return Errno(m.WASI.FDAdvise(wasi.FD(fd), wasi.FileSize(offset), wasi.FileSize(length), wasi.Advice(advice)))
}

func (m *Module) FDAllocate(ctx context.Context, fd Int32, offset, length Uint64) Errno {
	return Errno(m.WASI.FDAllocate(wasi.FD(fd), wasi.FileSize(offset), wasi.FileSize(length)))
}

func (m *Module) FDClose(ctx context.Context, fd Int32) Errno {
	return Errno(m.WASI.FDClose(wasi.FD(fd)))
}

func (m *Module) FDDataSync(ctx context.Context, fd Int32) Errno {
	return Errno(m.WASI.FDDataSync(wasi.FD(fd)))
}

func (m *Module) FDStatGet(ctx context.Context, fd Int32, stat Pointer[FDStat]) Errno {
	value, errno := m.WASI.FDStatGet(wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	stat.Store(FDStat{value})
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDStatSetFlags(ctx context.Context, fd Int32, flags Uint32) Errno {
	return Errno(m.WASI.FDStatSetFlags(wasi.FD(fd), wasi.FDFlags(flags)))
}

func (m *Module) FDStatSetRights(ctx context.Context, fd Int32, rightsBase, rightsInheriting Uint64) Errno {
	return Errno(m.WASI.FDStatSetRights(wasi.FD(fd), wasi.Rights(rightsBase), wasi.Rights(rightsInheriting)))
}

func (m *Module) FDFileStatGet(ctx context.Context, fd Int32, stat Pointer[FileStat]) Errno {
	value, errno := m.WASI.FDFileStatGet(wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	stat.Store(FileStat{value})
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDFileStatSetSize(ctx context.Context, fd Int32, size Uint64) Errno {
	return Errno(m.WASI.FDFileStatSetSize(wasi.FD(fd), wasi.FileSize(size)))
}

func (m *Module) FDFileStatSetTimes(ctx context.Context, fd Int32, accessTime, modifyTime Uint64, flags Int32) Errno {
	return Errno(m.WASI.FDFileStatSetTimes(wasi.FD(fd), wasi.Timestamp(accessTime), wasi.Timestamp(modifyTime), wasi.FSTFlags(flags)))
}

func (m *Module) FDPread(ctx context.Context, fd Int32, iovecs List[IOVec], offset Uint64, nread Pointer[Int32]) Errno {
	memory := nread.Memory()
	data, ok := m.getIOVecs(memory, iovecs)
	if !ok {
		return Errno(wasi.EFAULT)
	}
	value, errno := m.WASI.FDPread(wasi.FD(fd), data, wasi.FileSize(offset))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nread.Store(Int32(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDPreStatGet(ctx context.Context, fd Int32, prestat Pointer[PreStat]) Errno {
	value, errno := m.WASI.FDPreStatGet(wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	prestat.Store(PreStat{value})
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDPreStatDirName(ctx context.Context, fd Int32, dirName Bytes) Errno {
	value, errno := m.WASI.FDPreStatDirName(wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	if len(value) != len(dirName) {
		return Errno(wasi.EINVAL)
	}
	copy(dirName, value)
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDPwrite(ctx context.Context, fd Int32, iovecs List[IOVec], offset Uint64, nwritten Pointer[Int32]) Errno {
	memory := nwritten.Memory()
	data, ok := m.getIOVecs(memory, iovecs)
	if !ok {
		return Errno(wasi.EFAULT)
	}
	value, errno := m.WASI.FDPwrite(wasi.FD(fd), data, wasi.FileSize(offset))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDRead(ctx context.Context, fd Int32, iovecs List[IOVec], nread Pointer[Int32]) Errno {
	memory := nread.Memory()
	data, ok := m.getIOVecs(memory, iovecs)
	if !ok {
		return Errno(wasi.EFAULT)
	}
	value, errno := m.WASI.FDRead(wasi.FD(fd), data)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nread.Store(Int32(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDReadDir(ctx context.Context, fd Int32, buf Bytes, cookie Uint64, nwritten Pointer[Int32]) Errno {
	var errno wasi.Errno
	m.dirent, _, errno = m.WASI.FDReadDir(wasi.FD(fd), m.dirent[:0], len(buf), wasi.DirCookie(cookie))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	const sizeofDirEnt = int(unsafe.Sizeof(wasi.DirEntry{}))
	var n int
	for i := range m.dirent {
		e := &m.dirent[i]
		copy(buf[n:], unsafe.Slice((*byte)(unsafe.Pointer(&e.Entry)), sizeofDirEnt))
		n += sizeofDirEnt
		copy(buf[n:], e.Name)
		n += len(e.Name)
	}
	nwritten.Store(Int32(n))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDRenumber(ctx context.Context, from, to Int32) Errno {
	return Errno(m.WASI.FDRenumber(wasi.FD(from), wasi.FD(to)))
}

func (m *Module) FDSeek(ctx context.Context, fd Int32, delta Int64, whence Int32, size Pointer[Uint64]) Errno {
	value, errno := m.WASI.FDSeek(wasi.FD(fd), wasi.FileDelta(delta), wasi.Whence(whence))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	size.Store(Uint64(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDSync(ctx context.Context, fd Int32) Errno {
	return Errno(m.WASI.FDSync(wasi.FD(fd)))
}

func (m *Module) FDTell(ctx context.Context, fd Int32, size Pointer[Uint64]) Errno {
	value, errno := m.WASI.FDTell(wasi.FD(fd))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	size.Store(Uint64(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) FDWrite(ctx context.Context, fd Int32, iovecs List[IOVec], nwritten Pointer[Int32]) Errno {
	memory := nwritten.Memory()
	data, ok := m.getIOVecs(memory, iovecs)
	if !ok {
		return Errno(wasi.EFAULT)
	}
	value, errno := m.WASI.FDWrite(wasi.FD(fd), data)
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) PathCreateDirectory(ctx context.Context, fd Int32, path String) Errno {
	return Errno(m.WASI.PathCreateDirectory(wasi.FD(fd), string(path)))
}

func (m *Module) PathFileStatGet(ctx context.Context, fd Int32, flags Int32, path String, stat Pointer[FileStat]) Errno {
	value, errno := m.WASI.PathFileStatGet(wasi.FD(fd), wasi.LookupFlags(flags), string(path))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	stat.Store(FileStat{value})
	return Errno(wasi.ESUCCESS)
}

func (m *Module) PathFileStatSetTimes(ctx context.Context, fd Int32, lookupFlags Int32, path String, accessTime, modifyTime Uint64, fstFlags Int32) Errno {
	return Errno(m.WASI.PathFileStatSetTimes(wasi.FD(fd), wasi.LookupFlags(lookupFlags), string(path), wasi.Timestamp(accessTime), wasi.Timestamp(modifyTime), wasi.FSTFlags(fstFlags)))
}

func (m *Module) PathLink(ctx context.Context, oldFD Int32, oldFlags Int32, oldPath Bytes, newFD Int32, newPath Bytes) Errno {
	return Errno(m.WASI.PathLink(wasi.FD(oldFD), wasi.LookupFlags(oldFlags), string(oldPath), wasi.FD(newFD), string(newPath)))
}

func (m *Module) PathOpen(ctx context.Context, fd Int32, dirFlags Int32, path String, openFlags Int32, rightsBase, rightsInheriting Uint64, fdFlags Int32, openfd Pointer[Int32]) Errno {
	value, errno := m.WASI.PathOpen(wasi.FD(fd), wasi.LookupFlags(dirFlags), string(path), wasi.OpenFlags(openFlags), wasi.Rights(rightsBase), wasi.Rights(rightsInheriting), wasi.FDFlags(fdFlags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	openfd.Store(Int32(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) PathReadLink(ctx context.Context, fd Int32, path String, buf Bytes, nwritten Pointer[Int32]) Errno {
	str, errno := m.WASI.PathReadLink(wasi.FD(fd), string(path))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	if len(str) > len(buf) {
		return Errno(wasi.ERANGE)
	}
	copy(buf, str)
	nwritten.Store(Int32(len(str)))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) PathRemoveDirectory(ctx context.Context, fd Int32, path String) Errno {
	return Errno(m.WASI.PathRemoveDirectory(wasi.FD(fd), string(path)))
}

func (m *Module) PathRename(ctx context.Context, oldFD Int32, oldPath String, newFD Int32, newPath String) Errno {
	return Errno(m.WASI.PathRename(wasi.FD(oldFD), string(oldPath), wasi.FD(newFD), string(newPath)))
}

func (m *Module) PathSymlink(ctx context.Context, oldPath String, fd Int32, newPath String) Errno {
	return Errno(m.WASI.PathSymlink(string(oldPath), wasi.FD(fd), string(newPath)))
}

func (m *Module) PathUnlinkFile(ctx context.Context, fd Int32, path String) Errno {
	return Errno(m.WASI.PathUnlinkFile(wasi.FD(fd), string(path)))
}

func (m *Module) PollOneOff(ctx context.Context, subscriptionsPtr Pointer[Subscription], eventsPtr Pointer[Event], nSubscriptions Int32, n Pointer[Int32]) Errno {
	if nSubscriptions <= 0 {
		return Errno(wasi.EINVAL)
	}
	memory := n.Memory()
	const sizeofSubscription = int(unsafe.Sizeof(wasi.Subscription{}))
	const sizeofEvent = int(unsafe.Sizeof(wasi.Event{}))
	count := int(nSubscriptions)
	b, ok := memory.Read(subscriptionsPtr.Offset(), uint32(count*sizeofSubscription))
	if !ok {
		return Errno(wasi.EFAULT)
	}
	subscriptions := unsafe.Slice((*wasi.Subscription)(unsafe.Pointer(&b[0])), count)
	b, ok = memory.Read(eventsPtr.Offset(), uint32(count*sizeofEvent))
	if !ok {
		return Errno(wasi.EFAULT)
	}
	events := unsafe.Slice((*wasi.Event)(unsafe.Pointer(&b[0])), count)

	var errno wasi.Errno
	events, errno = m.WASI.PollOneOff(subscriptions, events[:0])
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	n.Store(Int32(len(events)))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) ProcExit(ctx context.Context, mod api.Module, exitCode Int32) {
	// Give the provider a chance to exit.
	m.WASI.ProcExit(wasi.ExitCode(exitCode))

	// Ensure other callers see the exit code.
	_ = mod.CloseWithExitCode(ctx, uint32(exitCode))

	// Prevent any code from executing after this function. For example, LLVM
	// inserts unreachable instructions after calls to exit.
	// See: https://github.com/emscripten-core/emscripten/issues/12322
	panic(fmt.Errorf("module exited with code %d", exitCode))
}

func (m *Module) ProcRaise(ctx context.Context, signal Int32) Errno {
	return Errno(m.WASI.ProcRaise(wasi.Signal(signal)))
}

func (m *Module) SchedYield(ctx context.Context) Errno {
	return Errno(m.WASI.SchedYield())
}

func (m *Module) RandomGet(ctx context.Context, buf Bytes) Errno {
	return Errno(m.WASI.RandomGet(buf))
}

func (m *Module) SockAccept(ctx context.Context, fd Int32, flags Int32, connfd Pointer[Int32]) Errno {
	value, errno := m.WASI.SockAccept(wasi.FD(fd), wasi.FDFlags(flags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	connfd.Store(Int32(value))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) SockRecv(ctx context.Context, fd Int32, iovecs List[IOVec], iflags Int32, nread Pointer[Int32], oflags Pointer[Int32]) Errno {
	memory := nread.Memory()
	data, ok := m.getIOVecs(memory, iovecs)
	if !ok {
		return Errno(wasi.EFAULT)
	}
	size, roflags, errno := m.WASI.SockRecv(wasi.FD(fd), data, wasi.RIFlags(iflags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nread.Store(Int32(size))
	oflags.Store(Int32(roflags))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) SockSend(ctx context.Context, fd Int32, iovecs List[IOVec], flags Int32, nwritten Pointer[Int32]) Errno {
	memory := nwritten.Memory()
	data, ok := m.getIOVecs(memory, iovecs)
	if !ok {
		return Errno(wasi.EFAULT)
	}
	size, errno := m.WASI.SockSend(wasi.FD(fd), data, wasi.SIFlags(flags))
	if errno != wasi.ESUCCESS {
		return Errno(errno)
	}
	nwritten.Store(Int32(size))
	return Errno(wasi.ESUCCESS)
}

func (m *Module) SockShutdown(ctx context.Context, fd Int32, flags Int32) Errno {
	return Errno(m.WASI.SockShutdown(wasi.FD(fd), wasi.SDFlags(flags)))
}

func (m *Module) Close(ctx context.Context) error {
	return m.WASI.Close()
}

func (m *Module) getIOVecs(memory api.Memory, iovecs List[IOVec]) ([]wasi.IOVec, bool) {
	// TODO: can we remove the need for this conversion?
	count := iovecs.Len()
	m.iovecs = m.iovecs[:0]
	for i := 0; i < count; i++ {
		m.iovecs = append(m.iovecs, wasi.IOVec(iovecs.Index(i).Load()))
	}
	return m.iovecs, true
}

type FDStat struct{ wasi.FDStat }

func (FDStat) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (FDStat) LoadObject(memory api.Memory, object []byte) (f FDStat) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&f)), f.ObjectSize()), object)
	return
}

func (f FDStat) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&f)), f.ObjectSize()))
}

func (FDStat) ObjectSize() int {
	return int(unsafe.Sizeof(FDStat{}))
}

type FileStat struct{ wasi.FileStat }

func (FileStat) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (FileStat) LoadObject(memory api.Memory, object []byte) (f FileStat) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&f)), f.ObjectSize()), object)
	return
}

func (f FileStat) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&f)), f.ObjectSize()))
}

func (FileStat) ObjectSize() int {
	return int(unsafe.Sizeof(FileStat{}))
}

type PreStat struct{ wasi.PreStat }

func (PreStat) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (PreStat) LoadObject(memory api.Memory, object []byte) (p PreStat) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&p)), p.ObjectSize()), object)
	return
}

func (p PreStat) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&p)), p.ObjectSize()))
}

func (PreStat) ObjectSize() int {
	return int(unsafe.Sizeof(PreStat{}))
}

// https://github.com/WebAssembly/WASI/blob/main/phases/snapshot/docs.md#-iovec-record
type IOVec []byte

func (arg IOVec) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	Bytes(arg.LoadObject(memory, object)).Format(w)
}

func (arg IOVec) LoadObject(memory api.Memory, object []byte) IOVec {
	offset := binary.LittleEndian.Uint32(object[:4])
	length := binary.LittleEndian.Uint32(object[4:])
	return wasm.Read(memory, offset, length)
}

func (arg IOVec) StoreObject(memory api.Memory, object []byte) {
	panic("NOT IMPLEMENTED")
}

func (arg IOVec) ObjectSize() int {
	return 8
}

type Subscription struct{ wasi.Subscription }

func (Subscription) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (Subscription) LoadObject(memory api.Memory, object []byte) (s Subscription) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&s)), s.ObjectSize()), object)
	return
}

func (s Subscription) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&s)), s.ObjectSize()))
}

func (Subscription) ObjectSize() int {
	return int(unsafe.Sizeof(Subscription{}))
}

type Event struct{ wasi.Event }

func (Event) FormatObject(w io.Writer, memory api.Memory, object []byte) {
	panic("not implemented")
}

func (Event) LoadObject(memory api.Memory, object []byte) (e Event) {
	copy(unsafe.Slice((*byte)(unsafe.Pointer(&e)), e.ObjectSize()), object)
	return
}

func (e Event) StoreObject(memory api.Memory, object []byte) {
	copy(object, unsafe.Slice((*byte)(unsafe.Pointer(&e)), e.ObjectSize()))
}

func (Event) ObjectSize() int {
	return int(unsafe.Sizeof(Event{}))
}

// procExit is a bit different; it doesn't have a return value,
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
