package imports

import (
	"github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero"
)

// DetectSocketsExtension determines the sockets extension in
// use by inspecting a WASM module's host imports.
//
// This function can detect WasmEdge v1 and WasmEdge v2.
func DetectSocketsExtension(module wazero.CompiledModule) *wasi_snapshot_preview1.Extension {
	functions := module.ImportedFunctions()
	hasWasmEdgeSockets := false
	sockAcceptParamCount := 0
	for _, f := range functions {
		moduleName, name, ok := f.Import()
		if !ok || moduleName != wasi_snapshot_preview1.HostModuleName {
			continue
		}
		switch name {
		case "sock_open", "sock_bind", "sock_connect", "sock_listen",
			"sock_send_to", "sock_recv_from", "sock_getsockopt", "sock_setsockopt",
			"sock_getlocaladdr", "sock_getpeeraddr", "sock_getaddrinfo":
			hasWasmEdgeSockets = true
		case "sock_accept":
			hasWasmEdgeSockets = true
			sockAcceptParamCount = len(f.ParamTypes())
		}
	}
	switch {
	case hasWasmEdgeSockets && sockAcceptParamCount == 2:
		return &wasi_snapshot_preview1.WasmEdgeV1
	case hasWasmEdgeSockets:
		return &wasi_snapshot_preview1.WasmEdgeV2
	default:
		return nil
	}
}
