package wasi_test

import (
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
