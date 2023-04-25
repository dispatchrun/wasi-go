package wasip1

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
