package wasitest

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stealthrocket/wasi-go"
)

var poll = testSuite{
	"with no subscriptions returns EINVAL": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})
		numEvents, errno := sys.PollOneOff(ctx, nil, nil)
		assertEqual(t, errno, wasi.EINVAL)
		assertEqual(t, numEvents, 0)
	},

	"an unknown file number sets the event to EBADF": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{})

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(42, wasi.FDReadEvent, wasi.SubscriptionFDReadWrite{FD: 1234}),
		}
		evs := make([]wasi.Event, len(subs))

		numEvents, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, numEvents, 1)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  42,
			Errno:     wasi.EBADF,
			EventType: wasi.FDReadEvent,
		})
	},

	"read from stdin": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		stdinR, stdinW := io.Pipe()
		defer stdinW.Close()
		defer stdinR.Close()

		sys := newSystem(TestConfig{
			Stdin: stdinR,
		})

		errno := sys.FDStatSetFlags(ctx, 0, wasi.NonBlock)
		assertEqual(t, errno, wasi.ESUCCESS)

		buffer := make([]byte, 32)
		n, errno := sys.FDRead(ctx, 0, []wasi.IOVec{buffer})
		assertEqual(t, n, ^wasi.Size(0))
		assertEqual(t, errno, wasi.EAGAIN)

		go func() {
			n, err := io.WriteString(stdinW, "Hello, World!")
			assertOK(t, err)
			assertEqual(t, n, 13)
		}()

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionFDReadWrite(42, wasi.FDReadEvent, wasi.SubscriptionFDReadWrite{FD: 0}),
		}
		evs := make([]wasi.Event, len(subs))

		numEvents, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, numEvents, 1)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, evs[0], wasi.Event{
			UserData:  42,
			EventType: wasi.FDReadEvent,
		})

		n, errno = sys.FDRead(ctx, 0, []wasi.IOVec{buffer})
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, n, 13)
		assertEqual(t, string(buffer[:n]), "Hello, World!")
	},

	"write to stdout": func(t *testing.T, ctx context.Context, newSystem newSystem) {
		stdoutR, stdoutW := io.Pipe()
		defer stdoutR.Close()
		defer stdoutW.Close()

		ch := make(chan []byte)
		go func() {
			b, err := io.ReadAll(stdoutR)
			assertOK(t, err)
			ch <- b
		}()

		sys := newSystem(TestConfig{
			Stdout: stdoutW,
		})

		errno := sys.FDStatSetFlags(ctx, 1, wasi.NonBlock)
		assertEqual(t, errno, wasi.ESUCCESS)

		n, errno := sys.FDWrite(ctx, 1, []wasi.IOVec{[]byte("Hello, World!")})
		if errno == wasi.ESUCCESS {
			assertEqual(t, errno, wasi.ESUCCESS)
			assertEqual(t, n, 13)
		} else {
			assertEqual(t, n, ^wasi.Size(0))
			assertEqual(t, errno, wasi.EAGAIN)

			subs := []wasi.Subscription{
				wasi.MakeSubscriptionFDReadWrite(42, wasi.FDWriteEvent, wasi.SubscriptionFDReadWrite{FD: 1}),
			}
			evs := make([]wasi.Event, len(subs))

			numEvents, errno := sys.PollOneOff(ctx, subs, evs)
			assertEqual(t, numEvents, 1)
			assertEqual(t, errno, wasi.ESUCCESS)
			assertEqual(t, evs[0], wasi.Event{
				UserData:  42,
				EventType: wasi.FDWriteEvent,
			})

			n, errno = sys.FDWrite(ctx, 1, []wasi.IOVec{[]byte("Hello, World!")})
			assertEqual(t, errno, wasi.ESUCCESS)
			assertEqual(t, n, 13)
		}

		assertEqual(t, sys.FDClose(ctx, 1), wasi.ESUCCESS)
		assertEqual(t, string(<-ch), "Hello, World!")
	},

	"monotonic clock with timeout in the future":   testPollTimeout(wasi.Monotonic, futureTimeout),
	"realtime clock with timeout in the future":    testPollTimeout(wasi.Realtime, futureTimeout),
	"process CPU clock with timeout in the future": testPollTimeout(wasi.ProcessCPUTimeID, futureTimeout),
	"thread CPU clock with timeout in the future":  testPollTimeout(wasi.ThreadCPUTimeID, futureTimeout),

	"monotonic clock with timeout in the past":   testPollTimeout(wasi.Monotonic, pastTimeout),
	"realtime clock with timeout in the past":    testPollTimeout(wasi.Realtime, pastTimeout),
	"process CPU clock with timeout in the past": testPollTimeout(wasi.ProcessCPUTimeID, pastTimeout),
	"thread CPU clock with timeout in the past":  testPollTimeout(wasi.ThreadCPUTimeID, pastTimeout),

	"monotonic clock with deadline in the future":   testPollDeadline(wasi.Monotonic, futureTimeout),
	"realtime clock with deadline in the future":    testPollDeadline(wasi.Realtime, futureTimeout),
	"process CPU clock with deadline in the future": testPollDeadline(wasi.ProcessCPUTimeID, futureTimeout),
	"thread CPU clock with deadline in the future":  testPollDeadline(wasi.ThreadCPUTimeID, futureTimeout),

	"monotonic clock with deadline in the past":   testPollDeadline(wasi.Monotonic, pastTimeout),
	"realtime clock with deadline in the past":    testPollDeadline(wasi.Realtime, pastTimeout),
	"process CPU clock with deadline in the past": testPollDeadline(wasi.ProcessCPUTimeID, pastTimeout),
	"thread CPU clock with deadline in the past":  testPollDeadline(wasi.ThreadCPUTimeID, pastTimeout),
}

const (
	futureTimeout = 10 * time.Millisecond
	pastTimeout   = -1 * time.Second // longer absolute value to notice if poll_oneoff waits or blocks
)

func testPollTimeout(clock wasi.ClockID, timeout time.Duration) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{
			Now: time.Now,
		})

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionClock(42, wasi.SubscriptionClock{
				ID:        clock,
				Timeout:   wasi.Timestamp(timeout),
				Precision: wasi.Timestamp(time.Millisecond),
			}),
		}
		evs := make([]wasi.Event, len(subs))
		now := time.Now()

		numEvents, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, numEvents, 1)
		if evs[0].Errno == wasi.ENOTSUP {
			t.Skip("clock not supported on this system")
		}
		if elapsed := time.Since(now); elapsed < timeout {
			t.Errorf("returned too early: %s < 10ms", elapsed)
		}
		assertEqual(t, evs[0], wasi.Event{
			UserData:  42,
			EventType: wasi.ClockEvent,
		})
	}
}

func testPollDeadline(clock wasi.ClockID, timeout time.Duration) testFunc {
	return func(t *testing.T, ctx context.Context, newSystem newSystem) {
		sys := newSystem(TestConfig{
			Now: time.Now,
		})

		timestamp, errno := sys.ClockTimeGet(ctx, clock, 1)
		switch errno {
		case wasi.ESUCCESS:
		case wasi.ENOTSUP:
			t.Skip("clock not supported on this system")
		default:
			t.Fatal("ClockTimeGet:", errno)
		}

		subs := []wasi.Subscription{
			wasi.MakeSubscriptionClock(42, wasi.SubscriptionClock{
				ID:        clock,
				Timeout:   timestamp + wasi.Timestamp(timeout),
				Precision: wasi.Timestamp(time.Millisecond),
				Flags:     wasi.Abstime,
			}),
		}
		evs := make([]wasi.Event, len(subs))
		now := time.Now()

		numEvents, errno := sys.PollOneOff(ctx, subs, evs)
		assertEqual(t, errno, wasi.ESUCCESS)
		assertEqual(t, numEvents, 1)
		if elapsed := time.Since(now); elapsed < timeout {
			t.Errorf("PollOneOff returned too early: %s < 10ms", elapsed)
		}
		assertEqual(t, evs[0], wasi.Event{
			UserData:  42,
			EventType: wasi.ClockEvent,
		})
	}
}
