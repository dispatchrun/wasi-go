package wasip1

import (
	"encoding/binary"
	"reflect"
	"testing"
	"unsafe"
)

func TestSubscription(t *testing.T) {
	assertEqual(t, int(unsafe.Sizeof(Subscription{})), 48)
	assertEqual(t, int(unsafe.Offsetof(Subscription{}.UserData)), 0)
	assertEqual(t, int(unsafe.Sizeof(Subscription{}.UserData)), 8)
	assertEqual(t, int(unsafe.Offsetof(Subscription{}.EventType)), 8)
	assertEqual(t, int(unsafe.Offsetof(Subscription{}.variant)), 16)
	assertEqual(t, int(unsafe.Sizeof(Subscription{}.variant)), 32)

	assertEqual(t, int(unsafe.Sizeof(SubscriptionFDReadWrite{})), 4)
	assertEqual(t, int(unsafe.Offsetof(SubscriptionFDReadWrite{}.FD)), 0)
	assertEqual(t, int(unsafe.Sizeof(SubscriptionFDReadWrite{}.FD)), 4)

	assertEqual(t, int(unsafe.Sizeof(SubscriptionClock{})), 32)
	assertEqual(t, int(unsafe.Offsetof(SubscriptionClock{}.ID)), 0)
	assertEqual(t, int(unsafe.Offsetof(SubscriptionClock{}.Timeout)), 8)
	assertEqual(t, int(unsafe.Offsetof(SubscriptionClock{}.Precision)), 16)
	assertEqual(t, int(unsafe.Offsetof(SubscriptionClock{}.Flags)), 24)

	assertEqual(t, int(unsafe.Sizeof(Timestamp(0))), 8)

	assertEqual(t, int(unsafe.Sizeof(ClockID(0))), 4)
	assertEqual(t, Realtime, 0)
	assertEqual(t, Monotonic, 1)
	assertEqual(t, ProcessCPUTimeID, 2)
	assertEqual(t, ThreadCPUTimeID, 3)

	assertEqual(t, int(unsafe.Sizeof(ClockFlags(0))), 2)
	assertEqual(t, Abstime, 0x1)

	assertEqual(t, int(unsafe.Sizeof(Event{})), 32)
	assertEqual(t, int(unsafe.Offsetof(Event{}.UserData)), 0)
	assertEqual(t, int(unsafe.Sizeof(Event{}.UserData)), 8)
	assertEqual(t, int(unsafe.Offsetof(Event{}.Errno)), 8)
	assertEqual(t, int(unsafe.Sizeof(Event{}.Errno)), 2)
	assertEqual(t, int(unsafe.Offsetof(Event{}.EventType)), 10)
	assertEqual(t, int(unsafe.Offsetof(Event{}.FDReadWrite)), 16)

	assertEqual(t, int(unsafe.Sizeof(EventFDReadWrite{})), 16)
	assertEqual(t, int(unsafe.Offsetof(EventFDReadWrite{}.NBytes)), 0)
	assertEqual(t, int(unsafe.Sizeof(EventFDReadWrite{}.NBytes)), 8)
	assertEqual(t, int(unsafe.Offsetof(EventFDReadWrite{}.Flags)), 8)

	assertEqual(t, int(unsafe.Sizeof(EventType(0))), 1)
	assertEqual(t, Clock, 0)
	assertEqual(t, FDRead, 1)
	assertEqual(t, FDWrite, 2)

	assertEqual(t, int(unsafe.Sizeof(FDReadWriteFlags(0))), 2)
	assertEqual(t, Hangup, 0x1)
}

func TestSubscriptionFDReadWrite(t *testing.T) {
	actual := MakeSubscriptionFDReadWrite(0xFEEDF4CED00DCAFE, FDRead, SubscriptionFDReadWrite{int32(0xABCD)})

	expected := Subscription{
		UserData:  0xFEEDF4CED00DCAFE,
		EventType: FDRead,
		variant:   [32]byte{},
	}
	binary.LittleEndian.PutUint32(expected.variant[0:4], uint32(0xABCD))
	assertEqual(t, actual, expected)

	variant := actual.GetFDReadWrite()
	assertEqual(t, variant.FD, int32(0xABCD))
}

func TestSubscriptionClock(t *testing.T) {
	actual := MakeSubscriptionClock(0xFEEDF4CED00DCAFE, SubscriptionClock{
		ID:        Realtime,
		Timeout:   0xABCD,
		Precision: 0x5678,
		Flags:     Abstime,
	})

	expected := Subscription{
		UserData:  0xFEEDF4CED00DCAFE,
		EventType: Clock,
		variant:   [32]byte{},
	}
	binary.LittleEndian.PutUint32(expected.variant[0:4], uint32(Realtime))
	binary.LittleEndian.PutUint64(expected.variant[8:16], uint64(0xABCD))
	binary.LittleEndian.PutUint64(expected.variant[16:24], uint64(0x5678))
	binary.LittleEndian.PutUint16(expected.variant[24:26], uint16(Abstime))
	assertEqual(t, actual, expected)

	variant := actual.GetClock()
	assertEqual(t, variant.ID, Realtime)
	assertEqual(t, variant.Timeout, 0xABCD)
	assertEqual(t, variant.Precision, 0x5678)
	assertEqual(t, variant.Flags, Abstime)
}

func assertEqual[T any](t *testing.T, actual, expected T) {
	t.Helper()

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("%v != %v", actual, expected)
	}
}
