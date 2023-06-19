package wasitest

import (
	"context"
	"os"
	"testing"

	"github.com/stealthrocket/wasi-go"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// TestSystem is a test suite which validates the behavior of wasi.System
// implementations.
func TestSystem(t *testing.T, makeSystem MakeSystem) {
	t.Run("proc", proc.runFunc(makeSystem))
	t.Run("poll", poll.runFunc(makeSystem))
}

type skip string

func (err skip) Error() string { return string(err) }

type newSystem func(TestConfig) wasi.System

type testFunc func(*testing.T, context.Context, newSystem)

type testSuite map[string]testFunc

func (tests testSuite) names() []string {
	names := maps.Keys(tests)
	slices.Sort(names)
	return names
}

func (tests testSuite) runFunc(makeSystem MakeSystem) func(*testing.T) {
	return func(t *testing.T) { tests.run(t, makeSystem) }
}

func (tests testSuite) run(t *testing.T, makeSystem MakeSystem) {
	for _, name := range tests.names() {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := testContext(t)
			defer cancel()

			tests[name](t, ctx, func(c TestConfig) wasi.System {
				devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
				if err != nil {
					t.Fatal(err)
				}
				defer devNull.Close()

				if c.Stdin == nil {
					c.Stdin = devNull
				}
				if c.Stdout == nil {
					c.Stdout = devNull
				}
				if c.Stderr == nil {
					c.Stderr = devNull
				}

				s, err := makeSystem(c)
				if err != nil {
					t.Fatalf("system initialization failed: %s", err)
				}
				t.Cleanup(func() {
					if err := s.Close(ctx); err != nil {
						t.Errorf("system closure failed: %s", err)
					}
				})
				return s
			})
		})
	}
}
