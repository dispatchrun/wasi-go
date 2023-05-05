package sockets

import "syscall"

// Listen creates a socket that listens on the specified address.
func Listen(rawAddr string) (int, error) {
	addr, sa, fd, err := Socket(rawAddr)
	if err != nil {
		return -1, err
	}
	if err := syscall.Bind(fd, sa); err != nil {
		syscall.Close(fd)
		return -1, err
	}
	opt := addr.Query()
	nonBlock := boolopt(opt, "nonblock", true)
	if err := syscall.SetNonblock(fd, nonBlock); err != nil {
		syscall.Close(fd)
		return -1, err
	}
	backlog := intopt(opt, "backlog", 128)
	if err := syscall.Listen(fd, backlog); err != nil {
		syscall.Close(fd)
		return -1, err
	}
	return fd, nil
}
