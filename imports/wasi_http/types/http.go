package types

import (
	"context"
	"fmt"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/common"
	"github.com/stealthrocket/wasi-go/imports/wasi_http/streams"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const ModuleName = "types"

func logFn(ctx context.Context, mod api.Module, ptr, len uint32) {
	str, _ := common.ReadString(mod, ptr, len)
	fmt.Print(str)
}

func Instantiate(ctx context.Context, rt wazero.Runtime, s *streams.Streams, r *Requests, rs *Responses, f *FieldsCollection, o *OutResponses) error {
	_, err := rt.NewHostModuleBuilder(ModuleName).
		NewFunctionBuilder().WithFunc(r.newOutgoingRequestFn).Export("new-outgoing-request").
		NewFunctionBuilder().WithFunc(f.newFieldsFn).Export("new-fields").
		NewFunctionBuilder().WithFunc(f.dropFieldsFn).Export("drop-fields").
		NewFunctionBuilder().WithFunc(f.fieldsEntriesFn).Export("fields-entries").
		NewFunctionBuilder().WithFunc(r.dropOutgoingRequestFn).Export("drop-outgoing-request").
		NewFunctionBuilder().WithFunc(r.outgoingRequestWriteFn).Export("outgoing-request-write").
		NewFunctionBuilder().WithFunc(rs.dropIncomingResponseFn).Export("drop-incoming-response").
		NewFunctionBuilder().WithFunc(rs.incomingResponseStatusFn).Export("incoming-response-status").
		NewFunctionBuilder().WithFunc(rs.incomingResponseHeadersFn).Export("incoming-response-headers").
		NewFunctionBuilder().WithFunc(rs.incomingResponseConsumeFn).Export("incoming-response-consume").
		NewFunctionBuilder().WithFunc(futureResponseGetFn).Export("future-incoming-response-get").
		NewFunctionBuilder().WithFunc(r.incomingRequestMethodFn).Export("incoming-request-method").
		NewFunctionBuilder().WithFunc(r.incomingRequestPathFn).Export("incoming-request-path").
		NewFunctionBuilder().WithFunc(r.incomingRequestAuthorityFn).Export("incoming-request-authority").
		NewFunctionBuilder().WithFunc(r.incomingRequestHeadersFn).Export("incoming-request-headers").
		NewFunctionBuilder().WithFunc(incomingRequestConsumeFn).Export("incoming-request-consume").
		NewFunctionBuilder().WithFunc(r.dropIncomingRequestFn).Export("drop-incoming-request").
		NewFunctionBuilder().WithFunc(o.setResponseOutparamFn).Export("set-response-outparam").
		NewFunctionBuilder().WithFunc(rs.newOutgoingResponseFn).Export("new-outgoing-response").
		NewFunctionBuilder().WithFunc(rs.outgoingResponseWriteFn).Export("outgoing-response-write").
		NewFunctionBuilder().WithFunc(dropOutgoingResponseFn).Export("drop-outgoing-response").
		NewFunctionBuilder().WithFunc(logFn).Export("log-it").
		Instantiate(ctx)
	return err
}
