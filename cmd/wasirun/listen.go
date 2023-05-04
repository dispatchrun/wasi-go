package main

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/stealthrocket/wasi-go"
	"golang.org/x/sys/unix"
)

const (
	listenRights = wasi.SockAcceptRight | wasi.PollFDReadWriteRight | wasi.FDFileStatGetRight | wasi.FDStatSetFlagsRight
	connRights   = wasi.FDReadRight | wasi.FDWriteRight | wasi.PollFDReadWriteRight | wasi.SockShutdownRight | wasi.FDFileStatGetRight | wasi.FDStatSetFlagsRight
)

func listen(rawAddr string) (fd int, err error) {
	addr := rawAddr
	if !strings.Contains(rawAddr, "://") {
		addr = "tcp://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		return -1, fmt.Errorf("could not parse --listen address '%s': %w", rawAddr, err)
	}
	family, sa, err := socketAddress(u.Scheme, u.Host)
	if err != nil {
		return -1, err
	}
	opt := u.Query()
	fd, err = unix.Socket(family, unix.SOCK_STREAM, 0)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			unix.Close(fd)
			fd = -1
		}
	}()
	if err = unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, intopt(opt, "reuseaddr", 1)); err != nil {
		return
	}
	if err = unix.SetNonblock(fd, boolopt(opt, "nonblock", true)); err != nil {
		return
	}
	if err = unix.Bind(fd, sa); err != nil {
		return
	}
	err = unix.Listen(fd, intopt(opt, "backlog", 128))
	return
}

func socketAddress(network, addr string) (int, unix.Sockaddr, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		return -1, nil, fmt.Errorf("unsupported --listen network: %v", network)
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
				return unix.AF_INET, &unix.SockaddrInet4{
					Port: port,
					Addr: ([4]byte)(ipv4),
				}, nil
			}
		}
	} else if network == "tcp" || network == "tcp6" {
		for _, ip := range ips {
			if ipv6 := ip.To16(); ipv6 != nil {
				return unix.AF_INET6, &unix.SockaddrInet6{
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
