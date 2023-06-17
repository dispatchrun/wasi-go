package types

import (
	"context"
	"encoding/binary"
	"net/http"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/wasi_streams"
	"github.com/tetratelabs/wazero/api"
)

var le = binary.LittleEndian

var response *http.Response

func SetResponse(r *http.Response) {
	response = r
}

func ResponseHeaders() http.Header {
	return response.Header
}

func dropIncomingResponseFn(_ context.Context, mod api.Module, handle uint32) {
	// pass
}

func incomingResponseStatusFn(_ context.Context, mod api.Module, handle uint32) int32 {
	return int32(response.StatusCode)
}

func incomingResponseHeadersFn(_ context.Context, mod api.Module, handle uint32) int32 {
	return 1
}

func incomingResponseConsumeFn(_ context.Context, mod api.Module, handle, ptr uint32) {
	stream := wasi_streams.Streams.NewInputStream(response.Body)
	data := []byte{}
	// 0 == ok, 1 == is_err
	data = le.AppendUint32(data, 0)
	// This is the stream number
	data = le.AppendUint32(data, stream)
	mod.Memory().Write(ptr, data)
}

func futureResponseGetFn(_ context.Context, mod api.Module, handle, ptr uint32) {
	data := []byte{}
	// 1 == is_some, 0 == none
	data = le.AppendUint32(data, 1)
	// 0 == ok, 1 == is_err, consistency ftw!
	data = le.AppendUint32(data, 0)
	// Copy the future into the actual
	data = le.AppendUint32(data, handle)
	mod.Memory().Write(ptr, data)
}
