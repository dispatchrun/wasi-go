package wasitest

import (
	"context"
	"testing"

	"github.com/stealthrocket/wasi-go"
)

func testContext(t *testing.T) (context.Context, context.CancelFunc) {
	ctx, cancel := context.Background(), func() {}
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, deadline)
	}
	return ctx, cancel
}

func assertOK(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	if got != want {
		t.Helper()
		t.Fatalf("%T values mismatch\nwant = %+v\ngot  = %+v", want, want, got)
	}
}

func skipIfNotImplemented(t *testing.T, errno wasi.Errno) {
	if errno == wasi.ENOSYS {
		t.Helper()
		t.Skip("operation not implemented on this system")
	}
}
