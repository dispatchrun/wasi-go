//go:build unix

package imports

import (
	"context"
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
func (b *Builder) Instantiate(ctx context.Context, runtime wazero.Runtime) (ctx2 context.Context, err error) {
	name := defaultName
	if b.name != "" {
		name = b.name
	}

	stdin, stdout, stderr := b.stdin, b.stdout, b.stderr
	if !b.customStdio {
		stdin = syscall.Stdin
		stdout = syscall.Stdout
		stderr = syscall.Stderr
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

	var system wasi.System = &unix.System{
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
	defer func() {
		if err != nil {
			system.Close(context.Background())
		}
	}()

	if b.pathOpenSockets {
		system = &unix.PathOpenSockets{System: system}
	}

	if b.tracer != nil {
		system = &wasi.Tracer{Writer: b.tracer, System: system}
	}

	for _, stdio := range []struct {
		fd   int
		path string
	}{
		{stdin, "/dev/stdin"},
		{stdout, "/dev/stdout"},
		{stderr, "/dev/stderr"},
	} {
		rights := wasi.FileRights
		if descriptor.IsATTY(stdio.fd) {
			rights = wasi.TTYRights
		}
		stat := wasi.FDStat{
			FileType:   wasi.CharacterDeviceType,
			RightsBase: rights,
		}
		newfd, err := dup(stdio.fd)
		if err != nil {
			return ctx, fmt.Errorf("unable to duplicate %s fd %d: %w", stdio.path, stdio.fd, err)
		}
		if b.nonBlockingStdio {
			if err := syscall.SetNonblock(newfd, true); err != nil {
				return ctx, fmt.Errorf("unable to put %s in non-blocking mode: %w", stdio.path, err)
			}
			stat.Flags |= wasi.NonBlock
		}
		system.Preopen(newfd, stdio.path, stat)
	}

	for _, m := range b.mounts {
		fd, err := syscall.Open(m.dir, syscall.O_DIRECTORY, 0)
		if err != nil {
			return ctx, fmt.Errorf("unable to preopen directory %q: %w", m.dir, err)
		}
		rightsBase := wasi.DirectoryRights
		rightsInheriting := wasi.DirectoryRights | wasi.FileRights
		if m.mode == 'r' {
			rightsBase &^= wasi.WriteRights
			rightsInheriting &^= wasi.WriteRights
		}
		system.Preopen(fd, m.dir, wasi.FDStat{
			FileType:         wasi.DirectoryType,
			RightsBase:       rightsBase,
			RightsInheriting: rightsInheriting,
		})
	}

	for _, addr := range b.listens {
		fd, err := sockets.Listen(addr)
		if err != nil {
			return ctx, fmt.Errorf("unable to listen on %q: %w", addr, err)
		}
		system.Preopen(fd, addr, wasi.FDStat{
			FileType:         wasi.SocketStreamType,
			Flags:            wasi.NonBlock,
			RightsBase:       wasi.SockListenRights,
			RightsInheriting: wasi.SockConnectionRights,
		})
	}
	for _, addr := range b.dials {
		fd, err := sockets.Dial(addr)
		if err != nil && err != sockets.EINPROGRESS {
			return ctx, fmt.Errorf("unable to dial %q: %w", addr, err)
		}
		system.Preopen(fd, addr, wasi.FDStat{
			FileType:   wasi.SocketStreamType,
			Flags:      wasi.NonBlock,
			RightsBase: wasi.SockConnectionRights,
		})
	}

	var extensions []wasi_snapshot_preview1.Extension
	if b.socketsExtension != nil {
		extensions = append(extensions, *b.socketsExtension)
	}

	module := wazergo.MustInstantiate(ctx, runtime,
		wasi_snapshot_preview1.NewHostModule(extensions...),
		wasi_snapshot_preview1.WithWASI(system),
	)
	ctx = wazergo.WithModuleInstance(ctx, module)

	return ctx, nil
}

func dup(fd int) (int, error) {
	newfd, err := syscall.Dup(fd)
	if err != nil {
		return -1, err
	}
	syscall.CloseOnExec(newfd)
	return newfd, nil
}
