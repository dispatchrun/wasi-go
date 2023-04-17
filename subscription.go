package wasip1

import "encoding/binary"

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

// SetFDReadWrite sets the subscription variant to a SubscriptionFDReadWrite.
func (s *Subscription) SetFDReadWrite(fdrw SubscriptionFDReadWrite) {
	s.variant = [32]byte{}
	binary.LittleEndian.PutUint32(s.variant[:], uint32(fdrw.FD))
}

// GetFDReadWrite gets the embedded SubscriptionFDReadWrite.
func (s *Subscription) GetFDReadWrite() SubscriptionFDReadWrite {
	return SubscriptionFDReadWrite{
		FD: int32(binary.LittleEndian.Uint32(s.variant[:])),
	}
}

// SubscriptionFDReadWrite is the contents of a subscription when type is type
// is FDRead or FDWrite.
type SubscriptionFDReadWrite struct {
	// FD is the file descriptor on which to wait for it to become ready for
	// reading or writing.
	FD int32
}

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
	_         [5]byte

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
	_     [6]byte
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
