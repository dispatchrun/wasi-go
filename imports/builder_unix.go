//go:build unix

package imports

import (
	"context"
	"fmt"
	"syscall"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	"github.com/stealthrocket/wasi-go/internal/sockets"
	"github.com/stealthrocket/wasi-go/systems/unix"
	"github.com/stealthrocket/wazergo"
	"github.com/tetratelabs/wazero"
)

// Instantiate compiles and instantiates the WASI module and binds it to
// the specified context.
func (b *Builder) Instantiate(ctx context.Context, runtime wazero.Runtime) (context.Context, error) {
	name := defaultName
	if b.name != "" {
		name = b.name
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
		{syscall.Stdin, "/dev/stdin"},
		{syscall.Stdout, "/dev/stdout"},
		{syscall.Stderr, "/dev/stderr"},
	} {
		stat := wasi.FDStat{
			FileType:   wasi.CharacterDeviceType,
			RightsBase: wasi.AllRights & ^(wasi.FDSeekRight | wasi.FDTellRight),
		}
		if b.nonBlockingStdio {
			if err := syscall.SetNonblock(stdio.fd, true); err != nil {
				return ctx, fmt.Errorf("unable to put %s in non-blocking mode: %w", stdio.path, err)
			}
			stat.Flags |= wasi.NonBlock
		}
		system.Preopen(stdio.fd, stdio.path, stat)
	}

	for _, m := range b.mounts {
		fd, err := syscall.Open(m.dir, syscall.O_DIRECTORY, 0)
		if err != nil {
			return ctx, fmt.Errorf("unable to preopen directory %q: %w", m.dir, err)
		}

		rights := wasi.AllRights
		if m.mode == 'r' {
			rights &^= wasi.WriteRights
		}
		system.Preopen(fd, m.dir, wasi.FDStat{
			FileType:         wasi.DirectoryType,
			RightsBase:       rights,
			RightsInheriting: rights,
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
