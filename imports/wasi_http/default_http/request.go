package default_http

import (
	"context"
	"log"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/types"
	"github.com/tetratelabs/wazero/api"
)

//var request = newHostFunc("request", requestFn,
//	[]api.ValueType{i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32, i32},
//	"a", "b", "c", "d", "e", "f", "g", "h", "j", "k", "l", "m", "n", "o")

func requestFn(_ context.Context, mod api.Module, a, b, c, d, e, f, g, h, j, k, l, m, n, o uint32) int32 {
	return 0
}

//var handle = newHostFunc("handle", handleFn,
//	[]api.ValueType{i32, i32, i32, i32, i32, i32, i32, i32},
//	"a", "b", "c", "d", "e", "f", "g", "h")

func handleFn(_ context.Context, mod api.Module, request, b, c, d, e, f, g, h uint32) int32 {
	req, ok := types.GetRequest(request)
	if !ok {
		log.Printf("Failed to get request: %v", request)
		return 0
	}
	r, err := req.MakeRequest()
	if err != nil {
		log.Printf(err.Error())
		return 0
	}
	types.SetResponse(r)
	return 1
}
