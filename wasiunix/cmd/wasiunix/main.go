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
	"github.com/stealthrocket/wasi-go/wasiunix"
	"github.com/stealthrocket/wazergo"
	"github.com/tetratelabs/wazero"
)

const Version = "devel"

var (
	envs             strings
	dirs             strings
	nonBlockingStdio bool
	version          bool
	help             bool
	h                bool
)

func main() {
	flag.Var(&envs, "env", "Environment variables to pass to the WASM module.")
	flag.Var(&dirs, "dir", "Directories to pre-open.")
	flag.BoolVar(&nonBlockingStdio, "non-blocking-stdio", false, "Enable non-blocking stdio.")
	flag.BoolVar(&version, "version", false, "Print the version and exit.")
	flag.BoolVar(&help, "help", false, "Print usage information.")
	flag.BoolVar(&h, "h", false, "Print usage information.")
	flag.Parse()

	if version {
		fmt.Println("wasiunix", Version)
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
	fmt.Printf(`wasiunix - Run a WebAssembly module

USAGE:
   wasiunix [OPTIONS]... <MODULE> [--] [ARGS]...

ARGS:
   <MODULE>
      The path of the WebAssembly module to run

   [ARGS]...
      Arguments to pass to the module

OPTIONS:
   --dir <DIR>
      Grant access to the specified host directory		

   --env <NAME=VAL>
      Pass an environment variable to the module

  --non-blocking-stdio
      Enable non-blocking stdio

   --version
      Print the version and exit

   -h, --help
      Show this usage information
`)
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: wasiunix [OPTIONS]... <MODULE> [--] [ARGS]...")
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

	system := &wasiunix.System{
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
		if nonBlockingStdio {
			if err := syscall.SetNonblock(stdio.fd, true); err != nil {
				return err
			}
			stat.Flags |= wasi.NonBlock
		}
		system.Preopen(stdio.fd, stdio.path, stat)
	}

	for _, dir := range dirs {
		fd, err := syscall.Open(dir, syscall.O_DIRECTORY, 0)
		if err != nil {
			return err
		}
		system.Preopen(fd, dir, wasi.FDStat{
			FileType:         wasi.DirectoryType,
			RightsBase:       wasi.AllRights,
			RightsInheriting: wasi.AllRights,
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

type strings []string

func (s strings) String() string {
	return fmt.Sprintf("%v", []string(s))
}

func (s *strings) Set(value string) error {
	*s = append(*s, value)
	return nil
}
