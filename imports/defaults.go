package imports

import (
	"context"
	"crypto/rand"
	"runtime"
	"time"

	"github.com/tetratelabs/wazero/sys"
)

const (
	defaultName               = "wasirun-wasm-module"
	defaultRealtimePrecision  = time.Microsecond
	defaultMonotonicPrecision = time.Nanosecond
)

var defaultRand = rand.Reader

var epoch = time.Now()

func defaultRealtime(ctx context.Context) (uint64, error) {
	return uint64(time.Now().UnixNano()), nil
}

func defaultMonotonic(ctx context.Context) (uint64, error) {
	return uint64(time.Since(epoch)), nil
}

func defaultYield(ctx context.Context) error {
	runtime.Gosched()
	return nil
}

var defaultRaise func(ctx context.Context, signal int) error = nil

func defaultExit(ctx context.Context, exitCode int) error {
	panic(sys.NewExitError(uint32(exitCode)))
}
