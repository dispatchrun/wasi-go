package server

import (
	"context"
	"net/http"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/types"
	"github.com/tetratelabs/wazero/api"
)

type WasmServer struct {
	Module api.Module
}

func (w WasmServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	fn := w.Module.ExportedFunction("HTTP#handle")
	if fn == nil {
		res.WriteHeader(500)
		res.Write([]byte("Handler not found"))
		return
	}

	id := types.MakeRequest(req)
	out := types.MakeOutparameter()

	_, err := fn.Call(context.TODO(), uint64(id), uint64(out))
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
		return
	}
	responseId, found := types.GetResponseByOutparameter(out)
	if !found {
		res.WriteHeader(500)
		res.Write([]byte("Couldn't find outparameter mapping"))
	}
	r, found := types.GetResponse(responseId)
	if !found || r == nil {
		res.WriteHeader(500)
		res.Write([]byte("Couldn't find response"))
	}
	if headers, found := types.GetFields(r.HeaderHandle); found {
		for key, value := range headers {
			for ix := range value {
				res.Header().Add(key, value[ix])
			}
		}
		types.DeleteFields(r.HeaderHandle)
	}
	res.WriteHeader(r.StatusCode)
	data := r.Buffer.Bytes()
	res.Write(data)

	types.DeleteResponse(responseId)
}
