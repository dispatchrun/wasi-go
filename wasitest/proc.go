package wasitest

import (
	"context"
	"testing"
	"time"

	"github.com/stealthrocket/wasi-go"
	"github.com/tetratelabs/wazero/sys"
)

var proc = testSuite{
	"ProcExit panics with a value of type sys.ExitError": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		s := newSystem(TestConfig{})

		defer func() {
			switch v := recover().(type) {
			case nil:
				t.Error("proc_exit must not return")
			case *sys.ExitError:
				if exitCode := v.ExitCode(); exitCode != 42 {
					t.Errorf("exit error contains the wrong exit code: %d", exitCode)
				}
			default:
				t.Errorf("proc_exit panicked with a value of the wrong type: %T", v)
			}
		}()

		s.ProcExit(ctx, 42)
	},

	"ProcRaise panics with a value of type sys.ExitError": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		s := newSystem(TestConfig{})

		defer func() {
			switch v := recover().(type) {
			case nil:
				t.Error("proc_raise must not return")
			case *sys.ExitError:
				if exitCode := v.ExitCode(); exitCode != 127+42 {
					t.Errorf("exit error contains the wrong exit code: %d", exitCode)
				}
			default:
				t.Errorf("proc_raise panicked with a value of the wrong type: %T", v)
			}
		}()

		s.ProcRaise(ctx, 42)
	},

	"SchedYield does nothing": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		s := newSystem(TestConfig{})
		assertEqual(t, s.SchedYield(ctx), wasi.ESUCCESS)
	},

	"ArgsSizesGet returns zero when there are no arguments": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		s := newSystem(TestConfig{})
		count, bytes, errno := s.ArgsSizesGet(ctx)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, count, 0)
		assertEqual(t, bytes, 0)
	},

	"ArgsSizesGet returns the number of arguments and their size in bytes": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		args := []string{
			"hello",
			"world",
		}
		s := newSystem(TestConfig{
			Args: args,
		})
		gotCount, gotBytes, errno := s.ArgsSizesGet(ctx)
		wantCount, wantBytes := wasi.SizesGet(args)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, gotCount, wantCount)
		assertEqual(t, gotBytes, wantBytes)
	},

	"EnvironSizesGet returns zero when there are no environment variables": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		s := newSystem(TestConfig{})
		count, bytes, errno := s.EnvironSizesGet(ctx)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, count, 0)
		assertEqual(t, bytes, 0)
	},

	"EnvironSizesGet returns the number of environment variables and their size in bytes": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		environ := []string{
			"hello",
			"world",
		}
		s := newSystem(TestConfig{
			Environ: environ,
		})
		gotCount, gotBytes, errno := s.EnvironSizesGet(ctx)
		wantCount, wantBytes := wasi.SizesGet(environ)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, gotCount, wantCount)
		assertEqual(t, gotBytes, wantBytes)
	},

	"ClockResGet with an invalid clock id returns EINVAL": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		s := newSystem(TestConfig{
			Now: time.Now,
		})
		_, errno := s.ClockResGet(ctx, 42)
		assertEqual(t, errno, wasi.EINVAL)
	},

	"ClockTimeGet with an invalid clock id returns EINVAL": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		s := newSystem(TestConfig{
			Now: time.Now,
		})
		_, errno := s.ClockTimeGet(ctx, 42, 0)
		assertEqual(t, errno, wasi.EINVAL)
	},
}
