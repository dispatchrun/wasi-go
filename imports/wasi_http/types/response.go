package types

import (
	"context"
	"encoding/binary"
	"log"
	"net/http"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/streams"
	"github.com/tetratelabs/wazero/api"
)

type Response struct {
	*http.Response
	headerHandle uint32
}

type responses struct {
	responses      map[uint32]*Response
	baseResponseId uint32
}

var data = responses{make(map[uint32]*Response), 0}

func MakeResponse(res *http.Response) uint32 {
	data.baseResponseId++
	data.responses[data.baseResponseId] = &Response{res, 0}
	return data.baseResponseId
}

func GetResponse(handle uint32) (*Response, bool) {
	res, ok := data.responses[handle]
	return res, ok
}

func dropIncomingResponseFn(_ context.Context, mod api.Module, handle uint32) {
	delete(data.responses, handle)
}

func incomingResponseStatusFn(_ context.Context, mod api.Module, handle uint32) int32 {
	response, found := GetResponse(handle)
	if !found {
		log.Printf("Unknown handle: %v", handle)
		return 0
	}
	return int32(response.StatusCode)
}

func incomingResponseHeadersFn(_ context.Context, mod api.Module, handle uint32) uint32 {
	res, found := GetResponse(handle)
	if !found {
		log.Printf("Unknown handle: %v", handle)
		return 0
	}
	if res.headerHandle == 0 {
		res.headerHandle = MakeFields(Fields(res.Header))
	}
	return res.headerHandle
}

func incomingResponseConsumeFn(_ context.Context, mod api.Module, handle, ptr uint32) {
	response, found := GetResponse(handle)
	le := binary.LittleEndian
	data := []byte{}
	if !found {
		// 0 == ok, 1 == is_err
		data = le.AppendUint32(data, 1)
	} else {
		// 0 == ok, 1 == is_err
		data = le.AppendUint32(data, 0)
		stream := streams.Streams.NewInputStream(response.Body)
		// This is the stream number
		data = le.AppendUint32(data, stream)
	}
	mod.Memory().Write(ptr, data)
}

func futureResponseGetFn(_ context.Context, mod api.Module, handle, ptr uint32) {
	le := binary.LittleEndian
	data := []byte{}
	// 1 == is_some, 0 == none
	data = le.AppendUint32(data, 1)
	// 0 == ok, 1 == is_err, consistency ftw!
	data = le.AppendUint32(data, 0)
	// Copy the future into the actual
	data = le.AppendUint32(data, handle)
	mod.Memory().Write(ptr, data)
}
