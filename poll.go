package wasi

import (
	"fmt"
	"unsafe"
)

// Subscription is a subscription to an event.
type Subscription struct {
	// UserData is a user-provided value that is attached to the subscription
	// in the implementation and returned through Event.UserData.
	UserData UserData

	// EventType is the type of the event to subscribe to.
	EventType EventType
	_         [7]byte

	// Variant is the contents of the subscription.
	//
	// It's a union field; either SubscriptionFDReadWrite or SubscriptionClock.
	// Use the Set and Get functions to access and mutate the variant.
	variant [32]byte
}

// UserData is a user-provided value that may be attached to objects that is
// retained when extracted from the implementation.
type UserData uint64

// MakeSubscriptionFDReadWrite makes a Subscription for FDReadEvent or
// FDWriteEvent events.
func MakeSubscriptionFDReadWrite(userData UserData, eventType EventType, fdrw SubscriptionFDReadWrite) Subscription {
	s := Subscription{UserData: userData, EventType: eventType}
	s.SetFDReadWrite(fdrw)
	return s
}

// MakeSubscriptionClock makes a Subscription for ClockEvent events.
func MakeSubscriptionClock(userData UserData, c SubscriptionClock) Subscription {
	s := Subscription{UserData: userData, EventType: ClockEvent}
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

// SubscriptionFDReadWrite is the contents of a subscription when event type
// is FDReadEvent or FDWriteEvent.
type SubscriptionFDReadWrite struct {
	// FD is the file descriptor to wait on.
	FD FD
}

// SubscriptionClock is the contents of a subscription when event type is
// ClockEvent.
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

// SubscriptionClockFlags are flags determining how to interpret the timestamp
// provided in SubscriptionClock.Timeout.
type SubscriptionClockFlags uint16

const (
	// Abstime is a flag indicating that the timestamp provided in
	// SubscriptionClock.Timeout is an absolute timestamp of clock
	// SubscriptionClock.ID. If unset, treat the timestamp provided in
	// SubscriptionClock.Timeout as relative to the current time value of clock
	// SubscriptionClock.ID.
	Abstime SubscriptionClockFlags = 1 << iota
)

// Has is true if the flag is set.
func (flags SubscriptionClockFlags) Has(f SubscriptionClockFlags) bool {
	return (flags & f) == f
}

func (flags SubscriptionClockFlags) String() string {
	switch flags {
	case Abstime:
		return "Abstime"
	default:
		return fmt.Sprintf("SubscriptionClockFlags(%d)", flags)
	}
}

// Event is an event that occurred.
type Event struct {
	// UserData is the user-provided value that got attached to
	// Subscription.UserData.
	UserData UserData

	// Errno is an error that occurred while processing the subscription
	// request.
	Errno Errno

	// EventType is the type of event that occurred.
	EventType EventType

	// FDReadWrite is the contents of the event, if it is a FDReadEvent or
	// FDWriteEvent. ClockEvent events ignore this field.
	FDReadWrite EventFDReadWrite
}

// EventFDReadWrite is the contents of an event when event type is FDReadEvent
// or FDWriteEvent.
type EventFDReadWrite struct {
	// NBytes is the number of bytes available for reading or writing.
	NBytes FileSize

	// Flags is the state of the file descriptor.
	Flags EventFDReadWriteFlags
}

// EventType is a type of a subscription to an event, or its occurrence.
type EventType uint8

const (
	// ClockEvent is an event type that indicates that the time value of clock
	// SubscriptionClock.ID has reached timestamp SubscriptionClock.Timeout.
	ClockEvent EventType = iota

	// FDReadEvent is an event type that indicates that the file descriptor
	// SubscriptionFDReadWrite.FD has data available for reading.
	FDReadEvent

	// FDWriteEvent is an event type that indicates that the file descriptor
	// SubscriptionFDReadWrite.FD has data available for writing.
	FDWriteEvent
)

func (e EventType) String() string {
	switch e {
	case ClockEvent:
		return "ClockEvent"
	case FDReadEvent:
		return "FDReadEvent"
	case FDWriteEvent:
		return "FDWriteEvent"
	default:
		return fmt.Sprintf("EventType(%d)", e)
	}
}

// EventFDReadWriteFlags is the state of the file descriptor subscribed to with
// FDReadEvent or FDWriteEvent.
type EventFDReadWriteFlags uint16

const (
	// Hangup is a flag that indicates that the peer of this socket
	// has closed or disconnected.
	Hangup EventFDReadWriteFlags = 1 << iota
)

// Has is true if the flag is set.
func (flags EventFDReadWriteFlags) Has(f EventFDReadWriteFlags) bool {
	return (flags & f) == f
}

func (flags EventFDReadWriteFlags) String() string {
	switch flags {
	case Hangup:
		return "Hangup"
	default:
		return fmt.Sprintf("EventFDReadWriteFlags(%d)", flags)
	}
}
