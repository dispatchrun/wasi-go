package imports

import (
	"github.com/stealthrocket/wasi-go/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero"
)

// DetectExtensions detects extensions to WASI preview 1.
func DetectExtensions(module wazero.CompiledModule) (ext []wasi_snapshot_preview1.Extension) {
	if sockets := DetectSocketsExtension(module); sockets != nil {
		ext = append(ext, *sockets)
	}
	return
}

// DetectSocketsExtension determines the sockets extension in
// use by inspecting a WASM module's host imports.
//
// This function can detect WasmEdge v1 and WasmEdge v2.
func DetectSocketsExtension(module wazero.CompiledModule) *wasi_snapshot_preview1.Extension {
	functions := module.ImportedFunctions()
	hasWasmEdgeSockets := false
	sockAcceptParamCount := 0
	sockLocalAddrParamCount := 0
	sockPeerAddrParamCount := 0
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
		}
		switch name {
		case "sock_accept":
			sockAcceptParamCount = len(f.ParamTypes())
		case "sock_getlocaladdr":
			sockLocalAddrParamCount = len(f.ParamTypes())
		case "sock_getpeeraddr":
			sockPeerAddrParamCount = len(f.ParamTypes())
		}
	}
	if hasWasmEdgeSockets || sockAcceptParamCount == 2 {
		if sockAcceptParamCount == 2 ||
			sockLocalAddrParamCount == 4 ||
			sockPeerAddrParamCount == 4 {
			return &wasi_snapshot_preview1.WasmEdgeV1
		} else {
			return &wasi_snapshot_preview1.WasmEdgeV2
		}
	}
	return nil
}
