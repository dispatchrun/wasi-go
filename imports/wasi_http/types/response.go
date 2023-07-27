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

type Responses struct {
	responses      map[uint32]*Response
	baseResponseId uint32
	streams        *streams.Streams
	fields         *FieldsCollection
}

type OutResponses struct {
	responses      map[uint32]uint32
	baseResponseId uint32
}

func MakeOutresponses() *OutResponses {
	return &OutResponses{
		responses:      make(map[uint32]uint32),
		baseResponseId: 1,
	}
}

func (o *OutResponses) MakeOutparameter() uint32 {
	o.baseResponseId++
	return o.baseResponseId
}

func (o *OutResponses) GetResponseByOutparameter(out uint32) (uint32, bool) {
	r, ok := o.responses[out]
	return r, ok
}

func (r *Responses) GetResponse(handle uint32) (*Response, bool) {
	res, ok := r.responses[handle]
	return res, ok
}

func (r *Responses) DeleteResponse(handle uint32) {
	delete(r.responses, handle)
}

func (r *Responses) dropIncomingResponseFn(_ context.Context, mod api.Module, handle uint32) {
	delete(r.responses, handle)
}

func dropOutgoingResponseFn(_ context.Context, mod api.Module, handle uint32) {
	// pass
}

func (r *Responses) outgoingResponseWriteFn(ctx context.Context, mod api.Module, res, ptr uint32) {
	response, found := r.GetResponse(res)
	data := []byte{}
	if !found {
		// Error
		data = binary.LittleEndian.AppendUint32(data, 1)
		data = binary.LittleEndian.AppendUint32(data, 0)
	} else {
		writer := &bytes.Buffer{}
		stream := r.streams.NewOutputStream(writer)

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

func (r *Responses) newOutgoingResponseFn(_ context.Context, status, headers uint32) uint32 {
	res := &Response{&http.Response{}, headers, 0, nil}
	res.StatusCode = int(status)
	r.baseResponseId++
	r.responses[r.baseResponseId] = res
	return r.baseResponseId
}

func (o *OutResponses) setResponseOutparamFn(_ context.Context, mod api.Module, res, err, resOut, _msg_ptr, _msg_str uint32) uint32 {
	if err == 1 {
		// TODO: details here.
		return 1
	}
	o.responses[res] = resOut
	return 0
}

func (r *Responses) incomingResponseStatusFn(_ context.Context, mod api.Module, handle uint32) int32 {
	response, found := r.GetResponse(handle)
	if !found {
		log.Printf("Unknown handle: %v", handle)
		return 0
	}
	return int32(response.StatusCode)
}

func MakeResponses(s *streams.Streams, f *FieldsCollection) *Responses {
	return &Responses{map[uint32]*Response{}, 1, s, f}
}

func (r *Responses) MakeResponse(res *http.Response) uint32 {
	r.baseResponseId++
	r.responses[r.baseResponseId] = &Response{res, 0, 0, nil}
	return r.baseResponseId
}

func (r *Responses) incomingResponseHeadersFn(_ context.Context, mod api.Module, handle uint32) uint32 {
	res, found := r.GetResponse(handle)
	if !found {
		log.Printf("Unknown handle: %v", handle)
		return 0
	}
	if res.HeaderHandle == 0 {
		res.HeaderHandle = r.fields.MakeFields(Fields(res.Header))
	}
	return res.HeaderHandle
}

func (r *Responses) incomingResponseConsumeFn(_ context.Context, mod api.Module, handle, ptr uint32) {
	response, found := r.GetResponse(handle)
	le := binary.LittleEndian
	data := []byte{}
	if !found {
		// 0 == ok, 1 == is_err
		data = le.AppendUint32(data, 1)
	} else {
		// 0 == ok, 1 == is_err
		data = le.AppendUint32(data, 0)
		stream := r.streams.NewInputStream(response.Body)
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
