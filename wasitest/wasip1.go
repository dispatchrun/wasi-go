package wasitest

import (
	"context"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	"github.com/stealthrocket/wazergo"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/sys"
)

// TestConfig carries the configuration used to create systems to run the test
// suites against.
type TestConfig struct {
	Args    []string
	Environ []string
	Stdin   io.ReadCloser
	Stdout  io.WriteCloser
	Stderr  io.WriteCloser
	Rand    io.Reader
	RootFS  string
	Now     func() time.Time
	// Limits, zero means none.
	MaxOpenFiles int
	MaxOpenDirs  int
}

// MakeSystem is a function used to create a system to run the test suites
// against.
//
// The test guarantees that the system will be closed when it isn't needed
// anymore.
type MakeSystem func(TestConfig) (wasi.System, error)

// TestWASIP1 is a generic test suite which runs the list of WebAssembly
// programs passed as file paths, creating a system and runtime to execute
// each of the test programs.
//
// Tests pass if the execution completed without trapping nor calling proc_exit
// with a non-zero exit code.
func TestWASIP1(t *testing.T, filePaths []string, makeSystem MakeSystem) {
	if len(filePaths) == 0 {
		t.Log("nothing to test")
	}

	for _, test := range filePaths {
		name := test
		for strings.HasPrefix(name, "../") {
			name = name[3:]
		}

		t.Run(name, func(t *testing.T) {
			bytecode, err := os.ReadFile(test)
			if err != nil {
				t.Fatal(err)
			}

			stdinR, stdinW, err := os.Pipe()
			if err != nil {
				t.Fatal("stdin:", err)
			}
			defer stdinR.Close()
			defer stdinW.Close()

			stdoutR, stdoutW, err := os.Pipe()
			if err != nil {
				t.Fatal("stdout:", err)
			}
			defer stdoutR.Close()
			defer stdoutW.Close()

			stderrR, stderrW, err := os.Pipe()
			if err != nil {
				t.Fatal("stderr:", err)
			}
			defer stderrR.Close()
			defer stderrW.Close()

			stdinW.Close() // nothing to read on stdin
			go io.Copy(os.Stdout, stdoutR)
			go io.Copy(os.Stderr, stderrR)

			system, err := makeSystem(TestConfig{
				Args: []string{
					filepath.Base(test),
				},
				Environ: []string{
					"PWD=" + filepath.Dir(test),
				},
				Stdin:  stdinR,
				Stdout: stdoutW,
				Stderr: stderrW,
				Rand:   rand.Reader,
				Now:    time.Now,
				RootFS: "/",
			})
			if err != nil {
				t.Fatal("system:", err)
			}
			ctx := context.Background()
			defer system.Close(ctx)

			runtime := wazero.NewRuntime(ctx)
			defer runtime.Close(ctx)

			ctx = wazergo.WithModuleInstance(ctx,
				wazergo.MustInstantiate(ctx, runtime,
					wasi_snapshot_preview1.NewHostModule(),
					wasi_snapshot_preview1.WithWASI(system),
				),
			)

			instance, err := runtime.Instantiate(ctx, bytecode)
			if err != nil {
				switch e := err.(type) {
				case *sys.ExitError:
					if exitCode := e.ExitCode(); exitCode != 0 {
						t.Error("exit code:", exitCode)
					}
				default:
					t.Error("instantiating wasm module instance:", err)
				}
			}
			if instance != nil {
				if err := instance.Close(ctx); err != nil {
					t.Error("closing wasm module instance:", err)
				}
			}
		})
	}
}
