package types

import (
	"context"
	"fmt"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/common"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const ModuleName = "types"

func logFn(ctx context.Context, mod api.Module, ptr, len uint32) {
	str, _ := common.ReadString(mod, ptr, len)
	fmt.Print(str)
}

func Instantiate(ctx context.Context, r wazero.Runtime) error {
	_, err := r.NewHostModuleBuilder(ModuleName).
		NewFunctionBuilder().WithFunc(newOutgoingRequestFn).Export("new-outgoing-request").
		NewFunctionBuilder().WithFunc(newFieldsFn).Export("new-fields").
		NewFunctionBuilder().WithFunc(dropFieldsFn).Export("drop-fields").
		NewFunctionBuilder().WithFunc(fieldsEntriesFn).Export("fields-entries").
		NewFunctionBuilder().WithFunc(dropOutgoingRequestFn).Export("drop-outgoing-request").
		NewFunctionBuilder().WithFunc(outgoingRequestWriteFn).Export("outgoing-request-write").
		NewFunctionBuilder().WithFunc(dropIncomingResponseFn).Export("drop-incoming-response").
		NewFunctionBuilder().WithFunc(incomingResponseStatusFn).Export("incoming-response-status").
		NewFunctionBuilder().WithFunc(incomingResponseHeadersFn).Export("incoming-response-headers").
		NewFunctionBuilder().WithFunc(incomingResponseConsumeFn).Export("incoming-response-consume").
		NewFunctionBuilder().WithFunc(futureResponseGetFn).Export("future-incoming-response-get").
		NewFunctionBuilder().WithFunc(incomingRequestMethodFn).Export("incoming-request-method").
		NewFunctionBuilder().WithFunc(incomingRequestPathFn).Export("incoming-request-path").
		NewFunctionBuilder().WithFunc(incomingRequestAuthorityFn).Export("incoming-request-authority").
		NewFunctionBuilder().WithFunc(incomingRequestHeadersFn).Export("incoming-request-headers").
		NewFunctionBuilder().WithFunc(incomingRequestConsumeFn).Export("incoming-request-consume").
		NewFunctionBuilder().WithFunc(dropIncomingRequestFn).Export("drop-incoming-request").
		NewFunctionBuilder().WithFunc(setResponseOutparamFn).Export("set-response-outparam").
		NewFunctionBuilder().WithFunc(newOutgoingResponseFn).Export("new-outgoing-response").
		NewFunctionBuilder().WithFunc(outgoingResponseWriteFn).Export("outgoing-response-write").
		NewFunctionBuilder().WithFunc(dropOutgoingResponseFn).Export("drop-outgoing-response").
		NewFunctionBuilder().WithFunc(logFn).Export("log-it").
		Instantiate(ctx)
	return err
}
