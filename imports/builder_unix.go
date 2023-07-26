//go:build unix

package imports

import (
	"context"
	"errors"
	"fmt"
	"syscall"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	"github.com/stealthrocket/wasi-go/internal/descriptor"
	"github.com/stealthrocket/wasi-go/internal/sockets"
	"github.com/stealthrocket/wasi-go/systems/unix"
	"github.com/stealthrocket/wazergo"
	"github.com/tetratelabs/wazero"
)

// Instantiate compiles and instantiates the WASI module and binds it to
// the specified context.
func (b *Builder) Instantiate(ctx context.Context, runtime wazero.Runtime) (ctxret context.Context, sys wasi.System, err error) {
	if len(b.errors) > 0 {
		return ctx, nil, errors.Join(b.errors...)
	}

	name := defaultName
	if b.name != "" {
		name = b.name
	}

	stdin, stdout, stderr := -1, -1, -1
	if b.customStdio {
		stdin, stdout, stderr = b.stdin, b.stdout, b.stderr
	}

	realtime := defaultRealtime
	if b.realtime != nil {
		realtime = b.realtime
	}
	realtimePrecision := defaultRealtimePrecision
	if b.realtimePrecision > 0 {
		realtimePrecision = b.realtimePrecision
	}
	monotonic := defaultMonotonic
	if b.monotonic != nil {
		monotonic = b.monotonic
	}
	monotonicPrecision := defaultMonotonicPrecision
	if b.monotonicPrecision > 0 {
		monotonicPrecision = b.monotonicPrecision
	}

	yield := defaultYield
	if b.yield != nil {
		yield = b.yield
	}
	raise := defaultRaise
	if b.raise != nil {
		raise = b.raise
	}
	exit := defaultExit
	if b.exit != nil {
		exit = b.exit
	}
	rand := defaultRand
	if b.rand != nil {
		rand = b.rand
	}

	unixSystem := &unix.System{
		Args:               append([]string{name}, b.args...),
		Environ:            b.env,
		Realtime:           realtime,
		RealtimePrecision:  realtimePrecision,
		Monotonic:          monotonic,
		MonotonicPrecision: monotonicPrecision,
		Yield:              yield,
		Raise:              raise,
		Rand:               rand,
		Exit:               exit,
	}
	unixSystem.MaxOpenFiles = b.maxOpenFiles
	unixSystem.MaxOpenDirs = b.maxOpenDirs

	system := wasi.System(unixSystem)
	defer func() {
		if system != nil {
			system.Close(context.Background())
		}
	}()

	if b.pathOpenSockets {
		system = &unix.PathOpenSockets{System: unixSystem}
	}
	if b.tracer != nil {
		system = wasi.Trace(b.tracer, system, b.tracerOptions...)
	}
	for _, wrap := range b.wrappers {
		system = wrap(system)
	}

	for fd, stdio := range []struct {
		fd   int
		open int
		path string
	}{
		{stdin, syscall.O_RDONLY, "/dev/stdin"},
		{stdout, syscall.O_WRONLY, "/dev/stdout"},
		{stderr, syscall.O_WRONLY, "/dev/stderr"},
	} {
		var err error
		if stdio.fd < 0 {
			stdio.fd, err = syscall.Open(stdio.path, stdio.open, 0)
			// Some systems may not allow opening stdio files on /dev, fallback
			// duplicating the process file descriptors which comes with the
			// limitation that setting the file descriptors to non-blocking will
			// also impact the behavior of stdio streams on the host.
			//
			// See: https://github.com/gitpod-io/gitpod/issues/17551
			if errors.Is(err, syscall.EACCES) {
				stdio.fd, err = dup(fd)
			}
		} else {
			stdio.fd, err = dup(stdio.fd)
		}
		if err != nil {
			return ctx, nil, fmt.Errorf("unable to open %s: %w", stdio.path, err)
		}
		rights := wasi.FileRights
		if descriptor.IsATTY(stdio.fd) {
			rights = wasi.TTYRights
		}
		stat := wasi.FDStat{
			FileType:   wasi.CharacterDeviceType,
			RightsBase: rights,
		}
		if b.nonBlockingStdio {
			if err := syscall.SetNonblock(stdio.fd, true); err != nil {
				return ctx, nil, fmt.Errorf("unable to put %s in non-blocking mode: %w", stdio.path, err)
			}
			stat.Flags |= wasi.NonBlock
		}
		unixSystem.Preopen(unix.FD(stdio.fd), stdio.path, stat)
	}

	for _, m := range b.mounts {
		fd, err := syscall.Open(m.dir, syscall.O_DIRECTORY, 0)
		if err != nil {
			return ctx, nil, fmt.Errorf("unable to preopen directory %q: %w", m.dir, err)
		}
		rightsBase := wasi.DirectoryRights
		rightsInheriting := wasi.DirectoryRights | wasi.FileRights
		if m.mode == 'r' {
			rightsBase &^= wasi.WriteRights
			rightsInheriting &^= wasi.WriteRights
		}
		unixSystem.Preopen(unix.FD(fd), m.dir, wasi.FDStat{
			FileType:         wasi.DirectoryType,
			RightsBase:       rightsBase,
			RightsInheriting: rightsInheriting,
		})
	}

	for _, addr := range b.listens {
		fd, err := sockets.Listen(addr)
		if err != nil {
			return ctx, nil, fmt.Errorf("unable to listen on %q: %w", addr, err)
		}
		unixSystem.Preopen(unix.FD(fd), addr, wasi.FDStat{
			FileType:         wasi.SocketStreamType,
			Flags:            wasi.NonBlock,
			RightsBase:       wasi.SockListenRights,
			RightsInheriting: wasi.SockConnectionRights,
		})
	}
	for _, addr := range b.dials {
		fd, err := sockets.Dial(addr)
		if err != nil && err != sockets.EINPROGRESS {
			return ctx, nil, fmt.Errorf("unable to dial %q: %w", addr, err)
		}
		unixSystem.Preopen(unix.FD(fd), addr, wasi.FDStat{
			FileType:   wasi.SocketStreamType,
			Flags:      wasi.NonBlock,
			RightsBase: wasi.SockConnectionRights,
		})
	}

	var extensions []wasi_snapshot_preview1.Extension
	if b.socketsExtension != nil {
		extensions = append(extensions, *b.socketsExtension)
	}

	hostModule := wasi_snapshot_preview1.NewHostModule(extensions...)

	instance := wazergo.MustInstantiate(ctx, runtime,
		wazergo.Decorate(hostModule, b.decorators...),
		wasi_snapshot_preview1.WithWASI(system),
	)

	ctx = wazergo.WithModuleInstance(ctx, instance)
	sys = system
	system = nil
	return ctx, sys, nil
}

func dup(fd int) (int, error) {
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()

	newfd, err := syscall.Dup(fd)
	if err != nil {
		return -1, err
	}
	syscall.CloseOnExec(newfd)
	return newfd, nil
}
