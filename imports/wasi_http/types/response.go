package types

import (
	"bytes"
	"context"
	"encoding/binary"
	"log"
	"net/http"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/streams"
	"github.com/tetratelabs/wazero/api"
)

type Response struct {
	*http.Response
	HeaderHandle uint32
	streamHandle uint32
	Buffer       *bytes.Buffer
}

type responses struct {
	responses      map[uint32]*Response
	baseResponseId uint32
}

type outResponses struct {
	responses      map[uint32]uint32
	baseResponseId uint32
}

var data = responses{make(map[uint32]*Response), 0}
var outData = outResponses{make(map[uint32]uint32), 0}

func MakeResponse(res *http.Response) uint32 {
	r := &Response{res, 0, 0, nil}
	data.baseResponseId++
	data.responses[data.baseResponseId] = r
	return data.baseResponseId
}

func MakeOutparameter() uint32 {
	outData.baseResponseId++
	return outData.baseResponseId
}

func GetResponseByOutparameter(out uint32) (uint32, bool) {
	r, ok := outData.responses[out]
	return r, ok
}

func GetResponse(handle uint32) (*Response, bool) {
	res, ok := data.responses[handle]
	return res, ok
}

func DeleteResponse(handle uint32) {
	delete(data.responses, handle)
}

func dropIncomingResponseFn(_ context.Context, mod api.Module, handle uint32) {
	delete(data.responses, handle)
}

func dropOutgoingResponseFn(_ context.Context, mod api.Module, handle uint32) {
	// pass
}

func outgoingResponseWriteFn(ctx context.Context, mod api.Module, res, ptr uint32) {
	response, found := GetResponse(res)
	data := []byte{}
	if !found {
		// Error
		data = binary.LittleEndian.AppendUint32(data, 1)
		data = binary.LittleEndian.AppendUint32(data, 0)
	} else {
		writer := &bytes.Buffer{}
		stream := streams.Streams.NewOutputStream(writer)

		response.streamHandle = stream
		response.Buffer = writer
		// 0 == no error
		data = binary.LittleEndian.AppendUint32(data, 0)
		data = binary.LittleEndian.AppendUint32(data, stream)
	}
	if !mod.Memory().Write(ptr, data) {
		panic("Failed to write data!")
	}
}

func newOutgoingResponseFn(_ context.Context, status, headers uint32) uint32 {
	data.baseResponseId++
	r := &Response{&http.Response{}, headers, 0, nil}
	r.StatusCode = int(status)
	data.responses[data.baseResponseId] = r
	return data.baseResponseId
}

func setResponseOutparamFn(_ context.Context, mod api.Module, res, err, resOut, _msg_ptr, _msg_str uint32) uint32 {
	if err == 1 {
		// TODO: details here.
		return 1
	}
	outData.responses[res] = resOut
	return 0
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
	if res.HeaderHandle == 0 {
		res.HeaderHandle = MakeFields(Fields(res.Header))
	}
	return res.HeaderHandle
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
