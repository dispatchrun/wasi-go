package wasi_test

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/stealthrocket/wasi-go"
)

func TestErrno(t *testing.T) {
	for errno := wasi.Errno(0); errno < wasi.ENOTCAPABLE; errno++ {
		t.Run(errno.Name(), func(t *testing.T) {
			e1 := errno.Syscall()
			e2 := wasi.MakeErrno(e1)
			if e2 != errno {
				t.Errorf("conversion to syscall.Errno did not yield the same error code: want=%d got=%d", errno, e2)
			}
		})
	}
}

func TestMakeErrno(t *testing.T) {
	tests := []struct {
		error error
		errno wasi.Errno
	}{
		{nil, wasi.ESUCCESS},
		{syscall.EAGAIN, wasi.EAGAIN},
		{context.Canceled, wasi.ECANCELED},
		{context.DeadlineExceeded, wasi.ETIMEDOUT},
		{io.ErrUnexpectedEOF, wasi.EIO},
		{fs.ErrClosed, wasi.EIO},
		{net.ErrClosed, wasi.EIO},
		{syscall.EPERM, wasi.EPERM},
		{wasi.EAGAIN, wasi.EAGAIN},
		{os.ErrDeadlineExceeded, wasi.ETIMEDOUT},
	}

	for _, test := range tests {
		t.Run(fmt.Sprint(test.error), func(t *testing.T) {
			if errno := wasi.MakeErrno(test.error); errno != test.errno {
				t.Errorf("error mismatch: want=%d got=%d (%s)", test.errno, errno, errno)
			}
		})
	}
}
