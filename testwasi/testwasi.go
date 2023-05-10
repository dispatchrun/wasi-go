package testwasi

import (
	"context"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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
	Stdin   *os.File
	Stdout  *os.File
	Stderr  *os.File
	RootFS  *os.File
	Rand    io.Reader
	Now     func() time.Time
}

// MakeSystem is a function used to create a system to run the test suites
// against.
//
// The function returns the system and a callback that will be invoked after
// completing a test to tear down resources allocated by the system.
type MakeSystem func(TestConfig) (wasi.System, func(), error)

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

			root, err := syscall.Open("/", syscall.O_DIRECTORY, 0)
			if err != nil {
				t.Fatal("root:", err)
			}
			rootFile := os.NewFile(uintptr(root), "/")
			defer rootFile.Close()

			stdinW.Close() // nothing to read on stdin
			go io.Copy(os.Stdout, stdoutR)
			go io.Copy(os.Stderr, stderrR)

			system, teardown, err := makeSystem(TestConfig{
				Args: []string{
					filepath.Base(test),
				},
				Environ: []string{
					"PWD=" + filepath.Dir(test),
				},
				Stdin:  stdinR,
				Stdout: stdoutW,
				Stderr: stderrW,
				RootFS: rootFile,
				Rand:   rand.Reader,
				Now:    time.Now,
			})
			if err != nil {
				t.Fatal("system:", err)
			}
			defer teardown()
			ctx := context.Background()

			runtime := wazero.NewRuntime(ctx)
			defer runtime.Close(ctx)

			ctx = wazergo.WithModuleInstance(ctx,
				wazergo.MustInstantiate(ctx, runtime,
					wasi_snapshot_preview1.NewHostModule(wasi_snapshot_preview1.Base),
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
