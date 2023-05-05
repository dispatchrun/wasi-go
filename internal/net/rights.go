package net

import "github.com/stealthrocket/wasi-go"

const (
	// ListenRights are rights that should be attached to listener sockets.
	ListenRights = wasi.SockAcceptRight | wasi.PollFDReadWriteRight | wasi.FDFileStatGetRight | wasi.FDStatSetFlagsRight

	// ConnectionRights are rights that should be attached to
	// connection sockets.
	ConnectionRights = wasi.FDReadRight | wasi.FDWriteRight | wasi.PollFDReadWriteRight | wasi.SockShutdownRight | wasi.FDFileStatGetRight | wasi.FDStatSetFlagsRight
)
