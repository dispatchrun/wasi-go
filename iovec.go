package wasip1

import "sync"

// IOVec is a slice of bytes.
type IOVec []byte

var iovecPool sync.Pool // []IOVec

// GetIOVecs gets an []IOVec with at least the specified capacity.
func GetIOVecs(capacity int) []IOVec {
	if c := iovecPool.Get(); c != nil {
		if iovecs := c.([]IOVec); cap(iovecs) >= capacity {
			return iovecs[:0]
		}
	}
	return make([]IOVec, 0, capacity)
}

// PutIOVecs returns an []IOVec to the pool.
func PutIOVecs(iovecs []IOVec) {
	iovecPool.Put(iovecs[:0])
}
