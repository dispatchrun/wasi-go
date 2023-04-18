package wasip1

import (
	"unsafe"
)

// Subscription is a subscription to an event.
type Subscription struct {
	// UserData is a user-provided value that is attached to the subscription
	// in the implementation and returned through Event.UserData.
	UserData uint64

	// EventType is the type of the event to which to subscribe.
	EventType EventType
	_         [7]byte

	// Variant is the contents of the event.
	//
	// It's a union field; either SubscriptionFDReadWrite or SubscriptionClock.
	// Use the Set and Get functions to access and mutate the variant.
	variant [32]byte
}

// MakeSubscriptionFDReadWrite makes a Subscription of type FDRead or FDWrite.
func MakeSubscriptionFDReadWrite(userData uint64, eventType EventType, fdrw SubscriptionFDReadWrite) Subscription {
	s := Subscription{UserData: userData, EventType: eventType}
	s.SetFDReadWrite(fdrw)
	return s
}

// MakeSubscriptionClock makes a Subscription of type Clock.
func MakeSubscriptionClock(userData uint64, c SubscriptionClock) Subscription {
	s := Subscription{UserData: userData, EventType: Clock}
	s.SetClock(c)
	return s
}

// SetFDReadWrite sets the subscription variant to a SubscriptionFDReadWrite.
func (s *Subscription) SetFDReadWrite(fdrw SubscriptionFDReadWrite) {
	variant := (*SubscriptionFDReadWrite)(unsafe.Pointer(&s.variant))
	*variant = fdrw
}

// GetFDReadWrite gets the embedded SubscriptionFDReadWrite.
func (s *Subscription) GetFDReadWrite() SubscriptionFDReadWrite {
	return *(*SubscriptionFDReadWrite)(unsafe.Pointer(&s.variant))
}

// SetClock sets the subscription variant to a SubscriptionClock.
func (s *Subscription) SetClock(c SubscriptionClock) {
	variant := (*SubscriptionClock)(unsafe.Pointer(&s.variant))
	*variant = c
}

// GetClock gets the embedded SubscriptionClock.
func (s *Subscription) GetClock() SubscriptionClock {
	return *(*SubscriptionClock)(unsafe.Pointer(&s.variant))
}

// SubscriptionFDReadWrite is the contents of a subscription when type is type
// is FDRead or FDWrite.
type SubscriptionFDReadWrite struct {
	// FD is the file descriptor on which to wait for it to become ready for
	// reading or writing.
	FD int32
}

// SubscriptionClock is the contents of a subscription when type is Clock.
type SubscriptionClock struct {
	// ID is the clock against which to compare the timestamp.
	ID ClockID

	// Timeout is the absolute or relative timestamp.
	Timeout Timestamp

	// Precision is the amount of time that the implementation may wait
	// additionally to coalesce with other events.
	Precision Timestamp

	// Flags specify whether the timeout is absolute or relative.
	Flags SubscriptionClockFlags
}

// Timestamp is a timestamp in nanoseconds.
type Timestamp uint64

// ClockID is an identifier for clocks.
type ClockID uint32

const (
	// Realtime is the clock measuring real time. Time value zero corresponds
	// with 1970-01-01T00:00:00Z.
	Realtime ClockID = iota

	// Monotonic is the store-wide monotonic clock, which is defined as a clock
	// measuring real time, whose value cannot be adjusted and which cannot
	// have negative clock jumps. The epoch of this clock is undefined. The
	// absolute time value of this clock therefore has no meaning.
	Monotonic

	// ProcessCPUTimeID is the CPU-time clock associated with the current
	// process.
	ProcessCPUTimeID

	// ThreadCPUTimeID is the CPU-time clock associated with the current
	// thread.
	ThreadCPUTimeID
)

// SubscriptionClockFlags are flags determining how to interpret the timestamp
// provided in SubscriptionClock.Timeout.
type SubscriptionClockFlags uint16

const (
	// SubscriptionClockAbstime is a flag indicatating that the timestam
	// provided in SubscriptionClock.Timeout is an absolute timestamp of
	// clock SubscriptionClock.ID. If unset, treat the timestamp provided
	// in SubscriptionClock.Timeout as relative to the current time value
	// of clock SubscriptionClock.ID.
	SubscriptionClockAbstime SubscriptionClockFlags = 1 << iota
)

// Event is an event that occurred.
type Event struct {
	// UserData is the user-provided value that got attached to
	// Subscription.UserData.
	UserData uint64

	// Errno is, if non-zero, an error that occurred while processing the
	// subscription request.
	Errno uint16

	// EventType is the type of event that occurred.
	EventType EventType

	// FDReadWrite is the contents of the event, if it is a FDRead or FDWrite.
	// Clock events ignore this field.
	FDReadWrite EventFDReadWrite
}

// EventFDReadWrite is the contents of an event when event type is FDRead or
// FDWrite.
type EventFDReadWrite struct {
	// NBytes is the number of bytes available for reading or writing.
	NBytes uint64

	// Flags is the state of the file descriptor.
	Flags EventRWFlags
}

// EventType is a type of a subscription to an event, or its occurrence.
type EventType uint8

const (
	// Clock is an event type that indicates that the time value of clock
	// SubscriptionClock.ID has reached timestamp SubscriptionClock.Timeout.
	Clock EventType = iota

	// FDRead is an event type that indicates that the file descriptor
	// SubscriptionFDReadWrite.FD has data available for reading.
	FDRead

	// FDWrite is an event type that indicates that the file descriptor
	// SubscriptionFDReadWrite.FD has data available for writing.
	FDWrite
)

// EventRWFlags is the state of the file descriptor subscribed to with FDRead
// or FDWrite.
type EventRWFlags uint16

// Has checks whether a flag is set.
func (flags EventRWFlags) Has(f EventRWFlags) bool {
	return (flags & f) != 0
}

const (
	// FDReadWriteHangup is a flag that indicates that the peer of this socket
	// has closed or disconnected.
	FDReadWriteHangup EventRWFlags = 1 << iota
)
