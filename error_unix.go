package wasi

import (
	"context"
	"errors"
	"fmt"
	"syscall"
)

func makeErrno(err error) Errno {
	if err == nil {
		return ESUCCESS
	}
	if err == syscall.EAGAIN {
		return EAGAIN
	}
	return makeErrnoSlow(err)
}

func makeErrnoSlow(err error) Errno {
	if err == context.Canceled {
		return ECANCELED
	}
	var sysErrno syscall.Errno
	if errors.As(err, &sysErrno) {
		if sysErrno == 0 {
			return ESUCCESS
		}
		return syscallErrnoToWASI(sysErrno)
	}
	var timeout interface{ Timeout() bool }
	if errors.As(err, &timeout) {
		if timeout.Timeout() {
			return ETIMEDOUT
		}
	}
	panic(fmt.Errorf("unexpected error: %v", err))
}
