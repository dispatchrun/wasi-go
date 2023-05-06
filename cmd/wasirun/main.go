package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	"github.com/stealthrocket/wasi-go/internal/sockets"
	"github.com/stealthrocket/wasi-go/systems/unix"
	"github.com/stealthrocket/wazergo"
	"github.com/tetratelabs/wazero"
)

const Version = "devel"

var (
	envs      stringList
	dirs      stringList
	listens   stringList
	dials     stringList
	socketExt string
	version   bool
	help      bool
	h         bool
)

func main() {
	flag.Var(&envs, "env", "Environment variable to pass to the WASM module.")
	flag.Var(&dirs, "dir", "Directory to pre-open.")
	flag.Var(&listens, "listen", "Socket to pre-open (and an address to listen on).")
	flag.Var(&dials, "dial", "Socket to pre-open (and an address to connect to).")
	flag.StringVar(&socketExt, "sockets", "", "Enable a sockets extension.")
	flag.BoolVar(&version, "version", false, "Print the version and exit.")
	flag.BoolVar(&help, "help", false, "Print usage information.")
	flag.BoolVar(&h, "h", false, "Print usage information.")
	flag.Parse()

	if version {
		fmt.Println("wasirun", Version)
		os.Exit(0)
	} else if h || help {
		showUsage()
		os.Exit(0)
	}

	if err := run(flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func showUsage() {
	fmt.Printf(`wasirun - Run a WebAssembly module

USAGE:
   wasirun [OPTIONS]... <MODULE> [--] [ARGS]...

ARGS:
   <MODULE>
      The path of the WebAssembly module to run

   [ARGS]...
      Arguments to pass to the module

OPTIONS:
   --dir <DIR>
      Grant access to the specified host directory

   --listen <ADDR>
      Grant access to a socket listening on the specified address

   --dial <ADDR>
      Grant access to a socket connected to the specified address

   --env <NAME=VAL>
      Pass an environment variable to the module

   --sockets <NAME>
      Enable a sockets extension with the specified name

   --version
      Print the version and exit

   -h, --help
      Show this usage information
`)
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: wasirun [OPTIONS]... <MODULE> [--] [ARGS]...")
	}

	wasmFile := args[0]
	wasmName := filepath.Base(wasmFile)
	wasmCode, err := os.ReadFile(wasmFile)
	if err != nil {
		return fmt.Errorf("could not read WASM file '%s': %w", wasmFile, err)
	}

	args = args[1:]
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	var system wasi.System = &unix.System{
		Args:               append([]string{wasmName}, args...),
		Environ:            envs,
		Realtime:           realtime,
		RealtimePrecision:  time.Microsecond,
		Monotonic:          monotonic,
		MonotonicPrecision: time.Nanosecond,
		Yield:              yield,
		Rand:               rand.Reader,
		Exit:               exit,
	}

	// Setup sockets extension.
	switch socketExt {
	case "path_open":
		system = &unix.PathOpenSockets{system}
	case "":
	default:
		return fmt.Errorf("unknown or unsupported socket extension: %s", socketExt)
	}

	// Preopen stdio.
	for _, stdio := range []struct {
		fd   int
		path string
	}{
		{syscall.Stdin, "/dev/stdin"},
		{syscall.Stdout, "/dev/stdout"},
		{syscall.Stderr, "/dev/stderr"},
	} {
		stat := wasi.FDStat{
			FileType:         wasi.CharacterDeviceType,
			RightsBase:       wasi.AllRights,
			RightsInheriting: wasi.AllRights,
		}
		system.Preopen(stdio.fd, stdio.path, stat)
	}

	// Preopen directories.
	for _, dir := range dirs {
		fd, err := syscall.Open(dir, syscall.O_DIRECTORY, 0)
		if err != nil {
			return err
		}
		defer syscall.Close(fd)

		system.Preopen(fd, dir, wasi.FDStat{
			FileType:         wasi.DirectoryType,
			RightsBase:       wasi.AllRights,
			RightsInheriting: wasi.AllRights,
		})
	}

	// Preopen sockets.
	for _, addr := range listens {
		fd, err := sockets.Listen(addr)
		if err != nil {
			return err
		}
		defer sockets.Close(fd)

		system.Preopen(fd, addr, wasi.FDStat{
			FileType:         wasi.SocketStreamType,
			Flags:            wasi.NonBlock,
			RightsBase:       wasi.SockListenRights,
			RightsInheriting: wasi.SockConnectionRights,
		})
	}
	for _, addr := range dials {
		fd, err := sockets.Dial(addr)
		if err != nil && err != sockets.EINPROGRESS {
			return err
		}
		defer sockets.Close(fd)

		system.Preopen(fd, addr, wasi.FDStat{
			FileType:   wasi.SocketStreamType,
			Flags:      wasi.NonBlock,
			RightsBase: wasi.SockConnectionRights,
		})
	}

	module := wazergo.MustInstantiate(ctx, runtime,
		wasi_snapshot_preview1.HostModule,
		wasi_snapshot_preview1.WithWASI(system),
	)
	ctx = wazergo.WithModuleInstance(ctx, module)

	instance, err := runtime.Instantiate(ctx, wasmCode)
	if err != nil {
		return err
	}
	return instance.Close(ctx)
}

var epoch = time.Now()

func realtime(context.Context) (uint64, error) {
	return uint64(time.Now().UnixNano()), nil
}

func monotonic(context.Context) (uint64, error) {
	return uint64(time.Since(epoch)), nil
}

func yield(ctx context.Context) error {
	runtime.Gosched()
	return nil
}

func exit(ctx context.Context, exitCode int) error {
	os.Exit(exitCode)
	return nil
}

type stringList []string

func (s stringList) String() string {
	return fmt.Sprintf("%v", []string(s))
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}
