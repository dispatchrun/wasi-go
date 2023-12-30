package types

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/common"
	"github.com/stealthrocket/wasi-go/imports/wasi_http/streams"
	"github.com/tetratelabs/wazero/api"
)

type Request struct {
	Method     string
	Path       string
	Query      string // not used in newer versions
	Scheme     string
	Authority  string
	Headers    uint32
	BodyBuffer *bytes.Buffer
}

func (r Request) Url() string {
	return fmt.Sprintf("%s://%s%s%s", r.Scheme, r.Authority, r.Path, r.Query)
}

type Requests struct {
	lock          sync.RWMutex
	requests      map[uint32]*Request
	requestIdBase uint32
	streams       *streams.Streams
	fields        *FieldsCollection
}

func MakeRequests(s *streams.Streams, f *FieldsCollection) *Requests {
	return &Requests{requests: map[uint32]*Request{}, requestIdBase: 1, streams: s, fields: f}
}

func (r *Requests) MakeRequest(req *http.Request) uint32 {
	request, id := r.newRequest()
	request.Method = req.Method
	// Fix this if port is missing.
	request.Authority = req.Host
	request.Path = req.URL.Path
	request.Headers = r.fields.MakeFields(Fields(req.Header))

	return id
}

func (r *Requests) newRequest() (*Request, uint32) {
	request := &Request{}
	requestIdBase := atomic.AddUint32(&r.requestIdBase, 1)
	r.lock.Lock()
	r.requests[requestIdBase] = request
	r.lock.Unlock()
	return request, requestIdBase
}

func (r *Requests) deleteRequest(handle uint32) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.requests, handle)
}

func (r *Requests) GetRequest(handle uint32) (*Request, bool) {
	r.lock.RLock()
	req, ok := r.requests[handle]
	r.lock.RUnlock()
	return req, ok
}

func (request *Request) MakeRequest(f *FieldsCollection) (*http.Response, error) {
	var body io.Reader = nil
	if request.BodyBuffer != nil {
		body = bytes.NewReader(request.BodyBuffer.Bytes())
	}
	r, err := http.NewRequest(request.Method, request.Url(), body)
	if err != nil {
		return nil, err
	}

	if fields, found := f.GetFields(request.Headers); found {
		r.Header = http.Header(fields)
	}

	return http.DefaultClient.Do(r)
}

func incomingRequestConsumeFn(ctx context.Context, mod api.Module, request, ptr uint32) {
	data := []byte{}
	// Unsupported for now.
	// error
	data = binary.LittleEndian.AppendUint32(data, 1)
	data = binary.LittleEndian.AppendUint32(data, 0)

	mod.Memory().Write(ptr, data)
}

func (r *Requests) incomingRequestHeadersFn(ctx context.Context, mod api.Module, request uint32) uint32 {
	req, ok := r.GetRequest(request)
	if !ok {
		return 0
	}
	return req.Headers
}

func (r *Requests) incomingRequestPathFn(ctx context.Context, mod api.Module, request, ptr uint32) {
	req, ok := r.GetRequest(request)
	if !ok {
		return
	}
	if err := common.WriteString(ctx, mod, ptr, req.Path); err != nil {
		panic(err.Error())
	}
}

func (r *Requests) incomingRequestPathFn_2023_10_18(ctx context.Context, mod api.Module, request, ptr uint32) {
	req, ok := r.GetRequest(request)
	if !ok {
		return
	}
	if err := common.WriteOptionalString(ctx, mod, ptr, req.Path); err != nil {
		panic(err.Error())
	}
}

func (r *Requests) incomingRequestAuthorityFn(ctx context.Context, mod api.Module, request, ptr uint32) {
	req, ok := r.GetRequest(request)
	if !ok {
		return
	}
	if err := common.WriteString(ctx, mod, ptr, req.Authority); err != nil {
		panic(err.Error())
	}
}

func (r *Requests) incomingRequestAuthorityFn_2023_10_18(ctx context.Context, mod api.Module, request, ptr uint32) {
	req, ok := r.GetRequest(request)
	if !ok {
		return
	}
	if err := common.WriteOptionalString(ctx, mod, ptr, req.Authority); err != nil {
		panic(err.Error())
	}
}

func (r *Requests) incomingRequestMethodFn(_ context.Context, mod api.Module, request, ptr uint32) {
	req, ok := r.GetRequest(request)
	if !ok {
		return
	}

	method := 0
	switch req.Method {
	case "GET":
		method = 0
	case "HEAD":
		method = 1
	case "POST":
		method = 2
	case "PUT":
		method = 3
	case "DELETE":
		method = 4
	case "CONNECT":
		method = 5
	case "OPTIONS":
		method = 6
	case "TRACE":
		method = 7
	case "PATCH":
		method = 8
	default:
		log.Fatalf("Unknown method: %s", req.Method)
	}

	data := []byte{}
	data = binary.LittleEndian.AppendUint32(data, uint32(method))
	mod.Memory().Write(ptr, data)
}

func (r *Requests) newOutgoingRequestFn(_ context.Context, mod api.Module,
	method, method_ptr, method_len,
	path_ptr, path_len,
	query_ptr, query_len,
	scheme_is_some, scheme, scheme_ptr, scheme_len,
	authority_ptr, authority_len, header_handle uint32) uint32 {

	request, id := r.newRequest()

	switch method {
	case 0:
		request.Method = "GET"
	case 1:
		request.Method = "HEAD"
	case 2:
		request.Method = "POST"
	case 3:
		request.Method = "PUT"
	case 4:
		request.Method = "DELETE"
	case 5:
		request.Method = "CONNECT"
	case 6:
		request.Method = "OPTIONS"
	case 7:
		request.Method = "TRACE"
	case 8:
		request.Method = "PATCH"
	default:
		log.Fatalf("Unknown method: %d", method)
	}

	path, ok := mod.Memory().Read(uint32(path_ptr), uint32(path_len))
	if !ok {
		return 0
	}
	request.Path = string(path)

	query, ok := mod.Memory().Read(uint32(query_ptr), uint32(query_len))
	if !ok {
		return 0
	}
	request.Query = string(query)

	request.Scheme = "https"
	if scheme_is_some == 1 {
		if scheme == 0 {
			request.Scheme = "http"
		}
		if scheme == 2 {
			d, ok := mod.Memory().Read(uint32(scheme_ptr), uint32(scheme_len))
			if !ok {
				return 0
			}
			request.Scheme = string(d)
		}
	}

	authority, ok := mod.Memory().Read(uint32(authority_ptr), uint32(authority_len))
	if !ok {
		return 0
	}
	request.Authority = string(authority)

	request.Headers = header_handle

	return id
}

func (r *Requests) newOutgoingRequestFn_2023_10_18(_ context.Context, mod api.Module,
	method, method_ptr, method_len,
	path_is_some, path_ptr, path_len,
	scheme_is_some, scheme, scheme_ptr, scheme_len,
	authority_is_some, authority_ptr, authority_len, header_handle uint32) uint32 {

	request, id := r.newRequest()

	switch method {
	case 0:
		request.Method = "GET"
	case 1:
		request.Method = "HEAD"
	case 2:
		request.Method = "POST"
	case 3:
		request.Method = "PUT"
	case 4:
		request.Method = "DELETE"
	case 5:
		request.Method = "CONNECT"
	case 6:
		request.Method = "OPTIONS"
	case 7:
		request.Method = "TRACE"
	case 8:
		request.Method = "PATCH"
	default:
		log.Fatalf("Unknown method: %d", method)
	}

	path, ok := mod.Memory().Read(uint32(path_ptr), uint32(path_len))
	if !ok {
		return 0
	}
	request.Path = string(path)

	request.Scheme = "https"
	if scheme_is_some == 1 {
		if scheme == 0 {
			request.Scheme = "http"
		}
		if scheme == 2 {
			d, ok := mod.Memory().Read(uint32(scheme_ptr), uint32(scheme_len))
			if !ok {
				return 0
			}
			request.Scheme = string(d)
		}
	}

	authority, ok := mod.Memory().Read(uint32(authority_ptr), uint32(authority_len))
	if !ok {
		return 0
	}
	request.Authority = string(authority)

	request.Headers = header_handle

	return id
}

func (r *Requests) dropOutgoingRequestFn(_ context.Context, mod api.Module, handle uint32) {
	r.deleteRequest(handle)
}

func (r *Requests) dropIncomingRequestFn(_ context.Context, mod api.Module, handle uint32) {
	req, found := r.GetRequest(handle)
	if !found {
		return
	}
	r.fields.DeleteFields(req.Headers)
	// Delete body stream here
	r.deleteRequest(handle)
}

func (r *Requests) outgoingRequestWriteFn(_ context.Context, mod api.Module, handle, ptr uint32) {
	request, found := r.GetRequest(handle)
	if !found {
		fmt.Printf("Failed to find request: %d\n", handle)
		return
	}
	request.BodyBuffer = &bytes.Buffer{}
	stream := r.streams.NewOutputStream(request.BodyBuffer)

	data := []byte{}
	data = binary.LittleEndian.AppendUint32(data, 0)
	data = binary.LittleEndian.AppendUint32(data, stream)
	mod.Memory().Write(ptr, data)
}
