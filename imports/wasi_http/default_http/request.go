package default_http

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/types"
	"github.com/tetratelabs/wazero/api"
)

type Handler struct {
	req *types.Requests
	res *types.Responses
	f   *types.FieldsCollection
}

// Request handles HTTP serving. It's currently unimplemented
func requestFn(_ context.Context, mod api.Module, a, b, c, d, e, f, g, h, j, k, l, m, n, o uint32) int32 {
	return 0
}

func doError(mod api.Module, ptr uint32, msg string) {
	data := []byte{}
	data = binary.LittleEndian.AppendUint32(data, 1) // IsError == 1
	data = binary.LittleEndian.AppendUint32(data, 3) // Always "unexpected error" for now
	data = binary.LittleEndian.AppendUint32(data, 0) // TODO: pass string here.
	data = binary.LittleEndian.AppendUint32(data, 0) // as above

	if !mod.Memory().Write(ptr, data) {
		panic("Failed to write response!")
	}
}

// Handle handles HTTP client calls.
// The remaining parameters (b..h) are for the HTTP Options, currently unimplemented.
func (handler *Handler) handleFn(_ context.Context, mod api.Module, request, b, c, d, e, f, g, h uint32) uint32 {
	req, ok := handler.req.GetRequest(request)
	if !ok {
		log.Printf("Failed to get request: %v\n", request)
		return 0
	}
	r, err := req.MakeRequest(handler.f)
	if err != nil {
		log.Println(err.Error())
		return 0
	}
	return handler.res.MakeResponse(r)
}

// Handle handles HTTP client calls.
// The remaining parameters (b..h) are for the HTTP Options, currently unimplemented.
func (handler *Handler) handleFn_2023_10_18(_ context.Context, mod api.Module, request, b, c, d, e, f, g, h, ptr uint32) {
	req, ok := handler.req.GetRequest(request)
	if !ok {
		msg := fmt.Sprintf("Failed to get request: %v\n", request)
		log.Printf(msg)
		doError(mod, ptr, msg)
		return
	}
	r, err := req.MakeRequest(handler.f)
	if err != nil {
		log.Println(err.Error())
		doError(mod, ptr, err.Error())
		return
	}
	res := handler.res.MakeResponse(r)
	data := []byte{}

	data = binary.LittleEndian.AppendUint32(data, 0) // IsOk == 0
	data = binary.LittleEndian.AppendUint32(data, res)
	data = binary.LittleEndian.AppendUint32(data, 0) // Used for errors
	data = binary.LittleEndian.AppendUint32(data, 0) // Used for errors

	if !mod.Memory().Write(ptr, data) {
		panic("Failed to write response!")
	}
}
