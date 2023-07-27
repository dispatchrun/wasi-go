package wasi_http

import (
	"context"
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
}

func MakeWasiHTTP() *WasiHTTP {
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
	}
}

func (w *WasiHTTP) Instantiate(ctx context.Context, rt wazero.Runtime) error {
	if err := types.Instantiate(ctx, rt, w.s, w.r, w.rs, w.f, w.o); err != nil {
		return err
	}
	if err := streams.Instantiate(ctx, rt, w.s); err != nil {
		return err
	}
	if err := default_http.Instantiate(ctx, rt, w.r, w.rs, w.f); err != nil {
		return err
	}
	return nil
}

func DetectWasiHttp(module wazero.CompiledModule) bool {
	functions := module.ImportedFunctions()
	hasWasiHttp := false
	for _, f := range functions {
		moduleName, name, ok := f.Import()
		if !ok || moduleName != default_http.ModuleName {
			continue
		}
		switch name {
		case "handle":
			hasWasiHttp = true
		}
	}
	return hasWasiHttp
}

func (w *WasiHTTP) MakeHandler(m api.Module) http.Handler {
	return server.WasmServer{
		Module:    m,
		Requests:  w.r,
		Responses: w.rs,
		Fields:    w.f,
		OutParams: w.o,
	}
}

func (w *WasiHTTP) HandleHTTP(writer http.ResponseWriter, req *http.Request, m api.Module) {
	handler := w.MakeHandler(m)
	handler.ServeHTTP(writer, req)
}
