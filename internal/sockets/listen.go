package sockets

import "syscall"

// Listen creates a socket that listens on the specified address.
func Listen(rawAddr string) (int, error) {
	addr, sa, fd, err := Socket(rawAddr)
	if err != nil {
		return -1, err
	}
	opt := addr.Query()
	reuseAddr := intopt(opt, "reuseaddr", 1)
	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, reuseAddr); err != nil {
		Close(fd)
		return -1, err
	}
	if err := syscall.Bind(fd, sa); err != nil {
		Close(fd)
		return -1, err
	}
	nonBlock := boolopt(opt, "nonblock", true)
	if err := syscall.SetNonblock(fd, nonBlock); err != nil {
		Close(fd)
		return -1, err
	}
	backlog := intopt(opt, "backlog", 128)
	if err := syscall.Listen(fd, backlog); err != nil {
		Close(fd)
		return -1, err
	}
	return fd, nil
}
