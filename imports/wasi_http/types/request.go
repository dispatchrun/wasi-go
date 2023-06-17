package types

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/tetratelabs/wazero/api"
)

type Request struct {
	Method    string
	Path      string
	Query     string
	Scheme    string
	Authority string
	Body      io.Reader
}

func (r Request) Url() string {
	return fmt.Sprintf("%s://%s%s%s", r.Scheme, r.Authority, r.Path, r.Query)
}

type requests struct {
	requests        map[uint32]*Request
	request_id_base uint32
}

var r = &requests{make(map[uint32]*Request), 1}

func (r *requests) newRequest() (*Request, uint32) {
	request := &Request{}
	r.request_id_base++
	r.requests[r.request_id_base] = request
	return request, r.request_id_base
}

func (r *requests) deleteRequest(handle uint32) {
	delete(r.requests, handle)
}

func GetRequest(handle uint32) (*Request, bool) {
	r, ok := r.requests[handle]
	return r, ok
}

func (request *Request) MakeRequest() (*http.Response, error) {
	r, err := http.NewRequest(request.Method, request.Url(), request.Body)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(r)
}

func (request *Request) RequestBody(body []byte) {
	request.Body = bytes.NewBuffer(body)
}

func newOutgoingRequestFn(_ context.Context, mod api.Module,
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
			request.Scheme = "https"
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

	return id
}

func dropOutgoingRequestFn(_ context.Context, mod api.Module, handle uint32) {
	r.deleteRequest(handle)
}

func outgoingRequestWriteFn(_ context.Context, mod api.Module, handle, ptr uint32) {

}
