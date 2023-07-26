package unix_test

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wasi-go/systems/unix"
	"github.com/stealthrocket/wasi-go/wasitest"
	"github.com/tetratelabs/wazero/sys"
	sysunix "golang.org/x/sys/unix"
)

func TestFS(t *testing.T) {
	f, err := os.Open("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	system := &unix.System{}
	rootFD := system.Preopen(unix.FD(f.Fd()), "/", wasi.FDStat{
		RightsBase:       wasi.AllRights,
		RightsInheriting: wasi.AllRights,
	})

	fsys := wasi.FS(context.Background(), system, rootFD)

	if err := fstest.TestFS(fsys,
		"empty",
		"message.txt",
		"tmp/one",
		"tmp/two",
		"tmp/three",
	); err != nil {
		t.Skipf("https://go-review.googlesource.com/c/go/+/503175:\n%v", err)
	}
}

func TestSystem(t *testing.T) {
	wasitest.TestSystem(t, makeSystem)
}

func TestWASIP1(t *testing.T) {
	files, _ := filepath.Glob("../testdata/*/*.wasm")
	wasitest.TestWASIP1(t, files, makeSystem)
}

func makeSystem(config wasitest.TestConfig) (wasi.System, error) {
	s := &unix.System{
		Args:    config.Args,
		Environ: config.Environ,
		Rand:    config.Rand,
		Yield: func(ctx context.Context) error {
			return nil
		},
		Exit: func(ctx context.Context, code int) error {
			panic(sys.NewExitError(uint32(code)))
		},
		Raise: func(ctx context.Context, code int) error {
			panic(sys.NewExitError(127 + uint32(code)))
		},
	}
	s.MaxOpenFiles = config.MaxOpenFiles
	s.MaxOpenDirs = config.MaxOpenDirs
	defer func() {
		if s != nil {
			s.Close(context.Background())
		}
	}()

	if now := config.Now; now != nil {
		epoch := now()
		s.Monotonic = func(context.Context) (uint64, error) {
			return uint64(now().Sub(epoch)), nil
		}
		s.MonotonicPrecision = time.Nanosecond
		s.Realtime = func(context.Context) (uint64, error) {
			return uint64(now().UnixNano()), nil
		}
		s.RealtimePrecision = time.Microsecond
	}

	stdin, err := pipe()
	if err != nil {
		return nil, err
	}
	if config.Stdin == nil {
		sysunix.Close(stdin[1])
	} else {
		f := os.NewFile(uintptr(stdin[1]), "/dev/stdin")
		go copyAndClose(f, config.Stdin)
	}
	s.Preopen(unix.FD(stdin[0]), "/dev/stdin", wasi.FDStat{
		FileType:   wasi.CharacterDeviceType,
		RightsBase: wasi.AllRights,
	})

	stdout, err := pipe()
	if err != nil {
		return nil, err
	}
	if config.Stdout == nil {
		sysunix.Close(stdout[0])
	} else {
		f := os.NewFile(uintptr(stdout[0]), "/dev/stdout")
		go copyAndClose(config.Stdout, f)
	}
	s.Preopen(unix.FD(stdout[1]), "/dev/stdout", wasi.FDStat{
		FileType:   wasi.CharacterDeviceType,
		RightsBase: wasi.AllRights,
	})

	stderr, err := pipe()
	if err != nil {
		return nil, err
	}
	if config.Stderr == nil {
		sysunix.Close(stderr[0])
	} else {
		f := os.NewFile(uintptr(stderr[0]), "/dev/stderr")
		go copyAndClose(config.Stderr, f)
	}
	s.Preopen(unix.FD(stderr[1]), "/dev/stderr", wasi.FDStat{
		FileType:   wasi.CharacterDeviceType,
		RightsBase: wasi.AllRights,
	})

	if config.RootFS != "" {
		rootFS, err := sysunix.Open(config.RootFS, sysunix.O_DIRECTORY, 0)
		if err != nil {
			return nil, err
		}
		s.Preopen(unix.FD(rootFS), "/", wasi.FDStat{
			FileType:         wasi.DirectoryType,
			RightsBase:       wasi.AllRights,
			RightsInheriting: wasi.AllRights,
		})
	}

	ret := s
	s = nil
	return ret, nil
}

func copyAndClose(w io.WriteCloser, r io.ReadCloser) {
	defer w.Close()
	defer r.Close()
	_, _ = io.Copy(w, r)
}

func pipe() (fds [2]int, err error) {
	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()
	if err := sysunix.Pipe(fds[:]); err != nil {
		return fds, err
	}
	sysunix.CloseOnExec(fds[0])
	sysunix.CloseOnExec(fds[1])
	return fds, nil
}

func TestSystemPollAndShutdown(t *testing.T) {
	testSystem(func(ctx context.Context, p *unix.System) {
		if err := p.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}

		// This call should block forever, unless async shutdown works, which is
		// what we are testing here.
		subscriptions := []wasi.Subscription{
			subscribeFDRead(0),
			subscribeFDRead(1),
		}
		events := make([]wasi.Event, len(subscriptions))

		n, errno := p.PollOneOff(ctx, subscriptions, events)
		if errno != wasi.ESUCCESS {
			t.Fatal(errno)
		}

		if !reflect.DeepEqual(subscriptions, []wasi.Subscription{
			subscribeFDRead(0),
			subscribeFDRead(1),
		}) {
			t.Error("poll_oneoff: altered subscriptions")
		}

		if !reflect.DeepEqual(events[:n], []wasi.Event{
			{UserData: 42, EventType: wasi.FDReadEvent, Errno: wasi.ECANCELED},
			{UserData: 43, EventType: wasi.FDReadEvent, Errno: wasi.ECANCELED},
		}) {
			t.Errorf("poll_oneoff: wrong events: %+v", events)
		}
	})
}

func TestSystemPollBadFileDescriptor(t *testing.T) {
	testSystem(func(ctx context.Context, p *unix.System) {
		subscriptions := []wasi.Subscription{
			subscribeFDRead(0),
			// Subscribe to a file descriptor which is not registered in the
			// system. This must not fail the poll_oneoff call and instead
			// report an error on the
			subscribeFDRead(42),
		}
		events := make([]wasi.Event, len(subscriptions))

		n, err := p.PollOneOff(ctx, subscriptions, events)
		if err != wasi.ESUCCESS {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(subscriptions, []wasi.Subscription{
			subscribeFDRead(0),
			subscribeFDRead(42),
		}) {
			t.Error("poll_oneoff: altered subscriptions")
		}

		if n != 1 {
			t.Errorf("poll_oneoff: wrong number of events: %d", n)
		} else if !reflect.DeepEqual(events[0], wasi.Event{
			UserData:  84,
			EventType: wasi.FDReadEvent,
			Errno:     wasi.EBADF,
		}) {
			t.Errorf("poll_oneoff: wrong event (0): %+v", events[0])
		}
	})
}

func TestSystemPollMissingMonotonicClock(t *testing.T) {
	testSystem(func(ctx context.Context, p *unix.System) {
		p.Monotonic = nil

		subscriptions := []wasi.Subscription{
			subscribeFDRead(0),
			subscribeTimeout(1 * time.Second),
		}
		events := make([]wasi.Event, len(subscriptions))

		n, err := p.PollOneOff(ctx, subscriptions, events)
		if err != wasi.ESUCCESS {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(subscriptions, []wasi.Subscription{
			subscribeFDRead(0),
			subscribeTimeout(1 * time.Second),
		}) {
			t.Error("poll_oneoff: altered subscriptions")
		}

		if n != 1 {
			t.Errorf("poll_oneoff: wrong number of events: %d", n)
		} else if !reflect.DeepEqual(events[0], wasi.Event{
			UserData:  42,
			EventType: wasi.ClockEvent,
			Errno:     wasi.ENOTSUP,
		}) {
			t.Errorf("poll_oneoff: wrong event (0): %+v", events[0])
		}
	})
}

func TestSockAddressInfo(t *testing.T) {
	testSystem(func(ctx context.Context, s *unix.System) {
		results := make([]wasi.AddressInfo, 64)
		tcp4Hint := wasi.AddressInfo{Family: wasi.InetFamily, SocketType: wasi.StreamSocket, Protocol: wasi.TCPProtocol}

		// Lookup :http. It's probably 80, but let's be sure.
		httpPort, err := net.LookupPort("tcp", "http")
		if err != nil {
			t.Fatal(err)
		}

		// Test name resolution (example.com => $IP) and port resolution (http => 80).
		n, errno := s.SockAddressInfo(ctx, "example.com", "http", tcp4Hint, results)
		if n <= 0 || errno != wasi.ESUCCESS {
			t.Fatalf("SockAddressInfo => %d, %s", n, errno)
		}
		var port int
		switch a := results[0].Address.(type) {
		case *wasi.Inet4Address:
			port = a.Port
		case *wasi.Inet6Address:
			port = a.Port
		default:
			t.Fatalf("unexpected address: %#v", a)
		}
		if port != httpPort {
			t.Fatalf("unexpected port: got %d, expect %d", port, httpPort)
		}

		// Test AI_NUMERICHOST and AI_NUMERICSERV.
		numericHint := tcp4Hint
		numericHint.Flags |= wasi.NumericHost | wasi.NumericService
		n, errno = s.SockAddressInfo(ctx, "1.2.3.4", "56", numericHint, results)
		if n != 1 || errno != wasi.ESUCCESS {
			t.Fatalf("SockAddressInfo => %d, %s", n, errno)
		}
		if ipv4, ok := results[0].Address.(*wasi.Inet4Address); !ok {
			t.Fatalf("unexpected result: %#v", results[n])
		} else if host := ipv4.String(); host != "1.2.3.4:56" {
			t.Fatalf("unexpected result: %s", host)
		}

		// Test AI_PASSIVE.
		passiveHint := tcp4Hint
		passiveHint.Flags |= wasi.Passive
		n, errno = s.SockAddressInfo(ctx, "", "80", passiveHint, results)
		if n != 1 || errno != wasi.ESUCCESS {
			t.Fatalf("SockAddressInfo => %d, %s", n, errno)
		}
		if ipv4, ok := results[0].Address.(*wasi.Inet4Address); !ok {
			t.Fatalf("unexpected result: %#v", results[n])
		} else if host := ipv4.String(); host != "0.0.0.0:80" {
			t.Fatalf("unexpected result: %s", host)
		}
	})
}

func testSystem(f func(context.Context, *unix.System)) {
	ctx := context.Background()

	p := newSystem()
	defer p.Close(ctx)

	fds, err := pipe()
	if err != nil {
		panic(err)
	}
	p.Preopen(unix.FD(fds[0]), "fd0", wasi.FDStat{RightsBase: wasi.AllRights})
	p.Preopen(unix.FD(fds[1]), "fd1", wasi.FDStat{RightsBase: wasi.AllRights})

	f(ctx, p)
}

func newSystem() *unix.System {
	return &unix.System{
		Realtime:           realtime,
		RealtimePrecision:  time.Microsecond,
		Monotonic:          monotonic,
		MonotonicPrecision: time.Nanosecond,
	}
}

var epoch = time.Now()

func realtime(context.Context) (uint64, error) {
	return uint64(time.Now().UnixNano()), nil
}

func monotonic(context.Context) (uint64, error) {
	return uint64(time.Since(epoch)), nil
}

func subscribeFDRead(fd wasi.FD) wasi.Subscription {
	return wasi.MakeSubscriptionFDReadWrite(
		wasi.UserData(42+fd),
		wasi.FDReadEvent,
		wasi.SubscriptionFDReadWrite{FD: fd},
	)
}

func subscribeTimeout(timeout time.Duration) wasi.Subscription {
	return wasi.MakeSubscriptionClock(
		wasi.UserData(42),
		wasi.SubscriptionClock{
			ID:        wasi.Monotonic,
			Timeout:   wasi.Timestamp(timeout),
			Precision: wasi.Timestamp(time.Nanosecond),
		},
	)
}
