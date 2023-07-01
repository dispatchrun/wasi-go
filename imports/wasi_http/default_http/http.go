package default_http

import (
	"context"

	"github.com/tetratelabs/wazero"
)

const ModuleName = "default-outgoing-HTTP"

func Instantiate(ctx context.Context, r wazero.Runtime) error {
	_, err := r.NewHostModuleBuilder(ModuleName).
		NewFunctionBuilder().WithFunc(requestFn).Export("request").
		NewFunctionBuilder().WithFunc(handleFn).Export("handle").
		Instantiate(ctx)
	return err
}
