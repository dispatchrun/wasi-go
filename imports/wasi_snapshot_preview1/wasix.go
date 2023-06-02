package wasi_snapshot_preview1

import (
	"context"

	"github.com/stealthrocket/wasi-go"
	"github.com/stealthrocket/wazergo"
)

var Wasix = Extension{
	"sock_status":             wazergo.F2((*Module).WasixSockStatus),
	"sock_addr_local":         wazergo.F2((*Module).WasixSockAddrLocal),
	"sock_addr_peer":          wazergo.F2((*Module).WasixSockAddrPeer),
	"sock_open":               wazergo.F4((*Module).WasixSockOpen),
	"sock_set_opt_flag":       wazergo.F3((*Module).WasixSockSetOptFlag),
	"sock_get_opt_flag":       wazergo.F3((*Module).WasixSockGetOptFlag),
	"sock_set_opt_time":       wazergo.F3((*Module).WasixSockSetOptTime),
	"sock_get_opt_time":       wazergo.F3((*Module).WasixSockGetOptTime),
	"sock_set_opt_size":       wazergo.F3((*Module).WasixSockSetOptSize),
	"sock_get_opt_size":       wazergo.F3((*Module).WasixSockGetOptSize),
	"sock_join_multicast_v4":  wazergo.F3((*Module).WasixSockJoinMulticastV4),
	"sock_leave_multicast_v4": wazergo.F3((*Module).WasixSockLeaveMulticastV4),
	"sock_join_multicast_v6":  wazergo.F3((*Module).WasixSockJoinMulticastV6),
	"sock_leave_multicast_v6": wazergo.F3((*Module).WasixSockLeaveMulticastV6),
	"sock_bind":               wazergo.F2((*Module).WasixSockBind),
	"sock_listen":             wazergo.F2((*Module).WasixSockListen),
	"sock_accept_v2":          wazergo.F4((*Module).WasixSockAcceptV2),
	"sock_connect":            wazergo.F2((*Module).WasixSockConnect),
	"sock_recv_from":          wazergo.F7((*Module).WasixSockRecvFrom),
	"sock_send_to":            wazergo.F6((*Module).WasixSockSendTo),
	"sock_send_file":          wazergo.F5((*Module).WasixSockSendFile),
	"resolve":                 wazergo.F6((*Module).WasixResolve),
}

func (m *Module) WasixSockStatus(ctx context.Context, fd Int32, sockStatus Pointer[Uint8]) errno {
	return Errno(wasi.ENOSYS)
}

func (m *Module) WasixSockAddrLocal(ctx context.Context, fd Int32, addrPort Pointer[wasixAddrPort]) errno {
	return Errno(wasi.ENOSYS)
}

func (m *Module) WasixSockAddrPeer(ctx context.Context, fd Int32, addrPort Pointer[wasixAddrPort]) errno {
	return Errno(wasi.ENOSYS)
}

func (m *Module) WasixSockOpen(ctx context.Context, family Int32, sockType Int32, protocol Int32, fd Pointer[Int32]) errno {
	return Errno(wasi.ENOSYS)
}

type wasixAddrPort []byte
