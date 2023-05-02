package wasi

import "fmt"

// RIFlags are flags provided to SockRecv.
type RIFlags uint16

const (
	// RecvPeek indicates that SockRecv should return the message without
	// removing it from the socket's receive queue.
	RecvPeek RIFlags = 1 << iota

	// RecvWaitAll indicates that on byte-stream sockets, SockRecv should block
	// until the full amount of data can be returned.
	RecvWaitAll
)

// Has is true if the flag is set.
func (flags RIFlags) Has(f RIFlags) bool {
	return (flags & f) == f
}

var riflagsStrings = [...]string{
	"RecvPeek",
	"RecvWaitAll",
}

func (flags RIFlags) String() (s string) {
	if flags == 0 {
		return "RIFlags(0)"
	}
	for i, name := range riflagsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("RIFlags(%d)", flags)
	}
	return
}

// ROFlags are flags returned by SockRecv.
type ROFlags uint16

const (
	// RecvDataTruncated indicates that message data has been truncated.
	RecvDataTruncated ROFlags = 1 << iota
)

// Has is true if the flag is set.
func (flags ROFlags) Has(f ROFlags) bool {
	return (flags & f) == f
}

func (flags ROFlags) String() string {
	switch flags {
	case RecvDataTruncated:
		return "RecvDataTruncated"
	default:
		return fmt.Sprintf("ROFlags(%d)", flags)
	}
}

// SIFlags are flags provided to SockSend.
//
// As there are currently no flags defined, it must be set to zero.
type SIFlags uint16

// Has is true if the flag is set.
func (flags SIFlags) Has(f SIFlags) bool {
	return (flags & f) == f
}

func (flags SIFlags) String() string {
	return fmt.Sprintf("SIFlags(%d)", flags)
}

// SDFlags are flags provided to SockShutdown which indicate which channels
// on a socket to shut down.
type SDFlags uint16

const (
	// ShutdownRD disables further receive operations.
	ShutdownRD SDFlags = 1 << iota

	// ShutdownWR disables further send operations.
	ShutdownWR
)

// Has is true if the flag is set.
func (flags SDFlags) Has(f SDFlags) bool {
	return (flags & f) == f
}

var sdflagsStrings = [...]string{
	"ShutdownRD",
	"ShutdownWR",
}

func (flags SDFlags) String() (s string) {
	if flags == 0 {
		return "SDFlags(0)"
	}
	for i, name := range sdflagsStrings {
		if !flags.Has(1 << i) {
			continue
		}
		if len(s) > 0 {
			s += "|"
		}
		s += name
	}
	if len(s) == 0 {
		return fmt.Sprintf("SDFlags(%d)", flags)
	}
	return
}
