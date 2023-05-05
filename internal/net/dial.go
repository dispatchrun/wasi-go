package net

import "syscall"

// Dial creates a socket and connects to the specified address.
func Dial(rawAddr string) (int, error) {
	addr, sa, fd, err := Socket(rawAddr)
	if err != nil {
		return -1, err
	}
	if err := syscall.Connect(fd, sa); err != nil {
		syscall.Close(fd)
		return -1, err
	}
	opt := addr.Query()
	noDelay := intopt(opt, "nodelay", 1)
	if err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, noDelay); err != nil {
		syscall.Close(fd)
		return -1, err
	}
	nonBlock := boolopt(opt, "nonblock", true)
	if err := syscall.SetNonblock(fd, nonBlock); err != nil {
		syscall.Close(fd)
		return -1, err
	}
	return fd, nil
}