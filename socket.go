package wasip1

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

// ROFlags are flags returned by SockRecv.
type ROFlags uint16

const (
	// RecvDataTruncated indicates that message data has been truncated.
	RecvDataTruncated ROFlags = 1 << iota
)

// SIFlags are flags provided to SockSend.
//
// As there are currently no flags defined, it must be set to zero.
type SIFlags uint16

// SDFlags are flags provided to SockShutdown which indicate which channels
// on a socket to shut down.
type SDFlags uint16

const (
	// ShutdownRD disables further receive operations.
	ShutdownRD SDFlags = 1 << iota

	// ShutdownWR disables further send operations.
	ShutdownWR
)
