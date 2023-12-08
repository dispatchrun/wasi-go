package wasi_http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/default_http"
	"github.com/stealthrocket/wasi-go/imports/wasi_http/server"
	"github.com/stealthrocket/wasi-go/imports/wasi_http/streams"
	"github.com/stealthrocket/wasi-go/imports/wasi_http/types"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type WasiHTTP struct {
	s  *streams.Streams
	f  *types.FieldsCollection
	r  *types.Requests
	rs *types.Responses
	o  *types.OutResponses
	v  string
}

func MakeWasiHTTP(version string) *WasiHTTP {
	s := streams.MakeStreams()
	f := types.MakeFields()
	r := types.MakeRequests(s, f)
	rs := types.MakeResponses(s, f)
	o := types.MakeOutresponses()

	return &WasiHTTP{
		s:  s,
		f:  f,
		r:  r,
		rs: rs,
		o:  o,
		v:  version,
	}
}

func (w *WasiHTTP) Instantiate(ctx context.Context, rt wazero.Runtime) error {
	switch w.v {
	case "v1":
		return w.instantiateV1(ctx, rt)
	case "2023_10_18":
		return w.instantiate_2023_10_18(ctx, rt)
	default:
		return fmt.Errorf("unknown version: %v", w.v)
	}
}

func (w *WasiHTTP) instantiateV1(ctx context.Context, rt wazero.Runtime) error {
	if err := types.Instantiate_v1(ctx, rt, w.s, w.r, w.rs, w.f, w.o); err != nil {
		return err
	}
	if err := streams.Instantiate_v1(ctx, rt, w.s); err != nil {
		return err
	}
	if err := default_http.Instantiate(ctx, rt, w.r, w.rs, w.f, w.v); err != nil {
		return err
	}
	return nil
}

func (w *WasiHTTP) instantiate_2023_10_18(ctx context.Context, rt wazero.Runtime) error {
	if err := types.Instantiate_2023_10_18(ctx, rt, w.s, w.r, w.rs, w.f, w.o); err != nil {
		return err
	}
	if err := streams.Instantiate_2023_10_18(ctx, rt, w.s); err != nil {
		return err
	}
	if err := default_http.Instantiate(ctx, rt, w.r, w.rs, w.f, w.v); err != nil {
		return err
	}
	return nil
}

func DetectWasiHttp(module wazero.CompiledModule) (bool, string) {
	functions := module.ImportedFunctions()
	hasWasiHttp := false
	version := ""
	for _, f := range functions {
		moduleName, name, ok := f.Import()
		if !ok || (moduleName != default_http.ModuleName && moduleName != default_http.ModuleName_2023_10_18) {
			continue
		}
		switch name {
		case "handle":
			hasWasiHttp = true
			switch moduleName {
			case default_http.ModuleName:
				version = "v1"
			case default_http.ModuleName_2023_10_18:
				version = "2023_10_18"
			default:
				version = "unknown"
			}
		}
	}
	return hasWasiHttp, version
}

func (w *WasiHTTP) MakeHandler(ctx context.Context, m api.Module) http.Handler {
	fnName := ""
	switch w.v {
	case "v1":
		fnName = "HTTP#handle"
	case "2023_10_18":
		fnName = "exports_wasi_http_0_2_0_rc_2023_10_18_incoming_handler_handle"
	}
	return server.WasmServer{
		Ctx:       ctx,
		Module:    m,
		Requests:  w.r,
		Responses: w.rs,
		Fields:    w.f,
		OutParams: w.o,
		HandleFn:  fnName,
	}
}
