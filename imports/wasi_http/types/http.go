package types

import (
	"context"

	"github.com/tetratelabs/wazero"
)

const ModuleName = "types"

func Instantiate(ctx context.Context, r wazero.Runtime) error {
	_, err := r.NewHostModuleBuilder(ModuleName).
		NewFunctionBuilder().WithFunc(newOutgoingRequestFn).Export("new-outgoing-request").
		NewFunctionBuilder().WithFunc(newFieldsFn).Export("new-fields").
		NewFunctionBuilder().WithFunc(fieldsEntriesFn).Export("fields-entries").
		NewFunctionBuilder().WithFunc(dropOutgoingRequestFn).Export("drop-outgoing-request").
		NewFunctionBuilder().WithFunc(outgoingRequestWriteFn).Export("outgoing-request-write").
		NewFunctionBuilder().WithFunc(dropIncomingResponseFn).Export("drop-incoming-response").
		NewFunctionBuilder().WithFunc(incomingResponseStatusFn).Export("incoming-response-status").
		NewFunctionBuilder().WithFunc(incomingResponseHeadersFn).Export("incoming-response-headers").
		NewFunctionBuilder().WithFunc(incomingResponseConsumeFn).Export("incoming-response-consume").
		NewFunctionBuilder().WithFunc(futureResponseGetFn).Export("future-incoming-response-get").
		Instantiate(ctx)
	return err
}
