package wasiunix_test

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stealthrocket/wasi"
	"github.com/stealthrocket/wasi/wasiunix"
)

func TestProviderShutdown(t *testing.T) {
	ctx := context.Background()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	p := wasiunix.Provider{
		Realtime:  realtime,
		Monotonic: monotonic,
	}
	defer p.Close(ctx)
	p.Preopen(int(r.Fd()), "fd0", wasi.FDStat{RightsBase: wasi.AllRights})
	p.Preopen(int(w.Fd()), "fd1", wasi.FDStat{RightsBase: wasi.AllRights})

	go func() {
		time.Sleep(100 * time.Millisecond)
		if err := p.Shutdown(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	// This call should block forever, unless async shutdown works, which is
	// what we are testing here.
	subscriptions := []wasi.Subscription{
		subscribeFDRead(0),
		subscribeFDRead(1),
	}
	events, errno := p.PollOneOff(ctx, subscriptions, nil)
	if errno != wasi.ESUCCESS {
		t.Fatal(errno)
	}
	if !reflect.DeepEqual(events, []wasi.Event{
		{UserData: 0, EventType: wasi.FDReadEvent, Errno: wasi.ECANCELED},
		{UserData: 1, EventType: wasi.FDReadEvent, Errno: wasi.ECANCELED},
	}) {
		t.Errorf("wrong events: %+v", events)
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
		wasi.UserData(fd),
		wasi.FDReadEvent,
		wasi.SubscriptionFDReadWrite{FD: fd},
	)
}

func subscribeFDWrite(fd wasi.FD) wasi.Subscription {
	return wasi.MakeSubscriptionFDReadWrite(
		wasi.UserData(fd),
		wasi.FDWriteEvent,
		wasi.SubscriptionFDReadWrite{FD: fd},
	)
}
