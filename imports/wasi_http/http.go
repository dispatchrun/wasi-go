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

func Instantiate(ctx context.Context, rt wazero.Runtime) error {
	if err := types.Instantiate(ctx, rt); err != nil {
		return err
	}
	if err := streams.Instantiate(ctx, rt); err != nil {
		return err
	}
	if err := default_http.Instantiate(ctx, rt); err != nil {
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

func HandleHTTP(w http.ResponseWriter, r *http.Request, m api.Module) {
	handler := server.WasmServer{Module: m}
	handler.ServeHTTP(w, r)
}
