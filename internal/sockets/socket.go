package sockets

import (
	"fmt"
	"net"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
)

// Socket prepares a socket for the specified address.
func Socket(rawAddr string) (u *url.URL, sa syscall.Sockaddr, fd int, err error) {
	if !strings.Contains(rawAddr, "://") {
		rawAddr = "tcp://" + rawAddr
	}
	u, err = url.Parse(rawAddr)
	if err != nil {
		return nil, nil, -1, fmt.Errorf("bad address '%s': %w", rawAddr, err)
	}
	family, sa, err := socketAddress(u.Scheme, u.Host)
	if err != nil {
		return nil, nil, -1, err
	}
	opt := u.Query()
	fd, err = syscall.Socket(family, syscall.SOCK_STREAM, 0)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			Close(fd)
			fd = -1
		}
	}()
	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, intopt(opt, "reuseaddr", 1)); err != nil {
		return
	}
	return u, sa, fd, err
}

// Close closes a file descriptor created with Socket, Listen or Dial.
func Close(fd int) error {
	if fd < 0 {
		return syscall.EBADF
	}
	err := syscall.Close(fd)
	if err != nil {
		if err == syscall.EBADF {
			println("DEBUG: close", fd, "=> EBADF")
			debug.PrintStack()
		}
	}
	return err
}

func socketAddress(network, addr string) (int, syscall.Sockaddr, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return -1, nil, fmt.Errorf("unsupported --Listen network: %v", network)
	}
	host, portstr, err := net.SplitHostPort(addr)
	if err != nil {
		return 0, nil, err
	}
	port, err := net.LookupPort(network, portstr)
	if err != nil {
		return 0, nil, err
	}
	var ips []net.IP
	if host == "" && network == "tcp6" {
		ips = []net.IP{net.IPv6zero}
	} else if host == "" {
		ips = []net.IP{net.IPv4zero}
	} else {
		ips, err = net.LookupIP(host)
		if err != nil {
			return 0, nil, err
		}
	}
	if network == "tcp" || network == "tcp4" {
		for _, ip := range ips {
			if ipv4 := ip.To4(); ipv4 != nil {
				return syscall.AF_INET, &syscall.SockaddrInet4{
					Port: port,
					Addr: ([4]byte)(ipv4),
				}, nil
			}
		}
	} else if network == "tcp" || network == "tcp6" {
		for _, ip := range ips {
			if ipv6 := ip.To16(); ipv6 != nil {
				return syscall.AF_INET6, &syscall.SockaddrInet6{
					Port: port,
					Addr: ([16]byte)(ipv6),
				}, nil
			}
		}
	}
	return 0, nil, fmt.Errorf("no IPs for network %s and host: %s", network, addr)
}

func intopt(q url.Values, key string, defaultValue int) int {
	values, ok := q[key]
	if !ok || len(values) == 0 {
		return defaultValue
	}
	n, err := strconv.Atoi(values[0])
	if err != nil {
		return defaultValue
	}
	return n
}

func boolopt(q url.Values, key string, defaultValue bool) bool {
	values, ok := q[key]
	if !ok || len(values) == 0 {
		return defaultValue
	}
	switch values[0] {
	case "true", "t", "1", "yes":
		return true
	case "false", "f", "0", "no":
		return false
	default:
		return defaultValue
	}
}
