package server

import (
	"context"
	"net/http"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/types"
	"github.com/tetratelabs/wazero/api"
)

type WasmServer struct {
	Module api.Module
	f      *types.FieldsCollection
	r      *types.Requests
	rs     *types.Responses
	o      *types.OutResponses
}

func (w WasmServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	fn := w.Module.ExportedFunction("HTTP#handle")
	if fn == nil {
		res.WriteHeader(500)
		res.Write([]byte("Handler not found"))
		return
	}

	id := w.r.MakeRequest(req)
	out := w.o.MakeOutparameter()

	_, err := fn.Call(context.TODO(), uint64(id), uint64(out))
	if err != nil {
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
		return
	}
	responseId, found := w.o.GetResponseByOutparameter(out)
	if !found {
		res.WriteHeader(500)
		res.Write([]byte("Couldn't find outparameter mapping"))
	}
	r, found := w.rs.GetResponse(responseId)
	if !found || r == nil {
		res.WriteHeader(500)
		res.Write([]byte("Couldn't find response"))
	}
	if headers, found := w.f.GetFields(r.HeaderHandle); found {
		for key, value := range headers {
			for ix := range value {
				res.Header().Add(key, value[ix])
			}
		}
		w.f.DeleteFields(r.HeaderHandle)
	}
	res.WriteHeader(r.StatusCode)
	data := r.Buffer.Bytes()
	res.Write(data)

	w.rs.DeleteResponse(responseId)
}
