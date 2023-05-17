package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

func printUsage() {
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
      Enable a sockets extension, either {none, auto, path_open,
      wasmedgev1, wasmedgev2}

   --pprof-addr <ADDR>
      Start a pprof server listening on the specified address

   --trace
      Enable logging of system calls (like strace)

   --non-blocking-stdio
      Enable non-blocking stdio

   -v, --version
      Print the version and exit

   -h, --help
      Show this usage information
`)
}

var (
	envs             stringList
	dirs             stringList
	listens          stringList
	dials            stringList
	socketExt        string
	pprofAddr        string
	trace            bool
	nonBlockingStdio bool
	version          bool
)

func main() {
	flagSet := flag.NewFlagSet("wasirun", flag.ExitOnError)
	flagSet.Usage = printUsage

	flagSet.Var(&envs, "env", "")
	flagSet.Var(&dirs, "dir", "")
	flagSet.Var(&listens, "listen", "")
	flagSet.Var(&dials, "dial", "")
	flagSet.StringVar(&socketExt, "sockets", "auto", "")
	flagSet.StringVar(&pprofAddr, "pprof-addr", "", "")
	flagSet.BoolVar(&trace, "trace", false, "")
	flagSet.BoolVar(&nonBlockingStdio, "non-blocking-stdio", false, "")
	flagSet.BoolVar(&version, "version", false, "")
	flagSet.BoolVar(&version, "v", false, "")
	flagSet.Parse(os.Args[1:])

	if version {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			fmt.Println("wasirun", info.Main.Version)
		} else {
			fmt.Println("wasirun", "devel")
		}
		os.Exit(0)
	}

	args := flagSet.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	if err := run(args[0], args[1:]); err != nil {
		if exitErr, ok := err.(*sys.ExitError); ok {
			os.Exit(int(exitErr.ExitCode()))
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(wasmFile string, args []string) error {
	wasmName := filepath.Base(wasmFile)
	wasmCode, err := os.ReadFile(wasmFile)
	if err != nil {
		return fmt.Errorf("could not read WASM file '%s': %w", wasmFile, err)
	}

	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	if pprofAddr != "" {
		go http.ListenAndServe(pprofAddr, nil)
	}

	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	wasmModule, err := runtime.CompileModule(ctx, wasmCode)
	if err != nil {
		return err
	}
	defer wasmModule.Close(ctx)

	builder := imports.NewBuilder().
		WithName(wasmName).
		WithArgs(args...).
		WithEnv(envs...).
		WithDirs(dirs...).
		WithListens(listens...).
		WithDials(dials...).
		WithNonBlockingStdio(nonBlockingStdio).
		WithSocketsExtension(socketExt, wasmModule).
		WithTracer(trace, os.Stderr)

	var system wasi.System
	ctx, system, err = builder.Instantiate(ctx, runtime)
	if err != nil {
		return err
	}
	defer system.Close(ctx)

	instance, err := runtime.InstantiateModule(ctx, wasmModule, wazero.NewModuleConfig())
	if err != nil {
		return err
	}
	return instance.Close(ctx)
}

type stringList []string

func (s stringList) String() string {
	return fmt.Sprintf("%v", []string(s))
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}
