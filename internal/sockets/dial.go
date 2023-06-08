package sockets

import "syscall"

const EINPROGRESS = syscall.EINPROGRESS

// Dial creates a socket and connects to the specified address.
func Dial(rawAddr string) (int, error) {
	addr, sa, fd, err := Socket(rawAddr)
	if err != nil {
		return -1, err
	}
	opt := addr.Query()
	noDelay := intopt(opt, "nodelay", 1)
	if err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, noDelay); err != nil {
		Close(fd)
		return -1, err
	}
	nonBlock := boolopt(opt, "nonblock", true)
	if err := syscall.SetNonblock(fd, nonBlock); err != nil {
		Close(fd)
		return -1, err
	}
	err = syscall.Connect(fd, sa)
	if err != nil && err != EINPROGRESS {
		Close(fd)
		return -1, err
	}
	return fd, err
}
