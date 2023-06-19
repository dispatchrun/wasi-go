package wasitest

import (
	"context"
	"testing"
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
