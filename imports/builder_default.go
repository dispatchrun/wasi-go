//go:build !unix

package imports

import (
	"context"
	"fmt"
	"runtime"

	"github.com/stealthrocket/wasi-go"
	"github.com/tetratelabs/wazero"
)

func (b *Builder) Instantiate(ctx context.Context, _ wazero.Runtime) (ctxret context.Context, sys wasi.System, err error) {
	return ctx, nil, fmt.Errorf("wasi-go is not available on GOOS=%s", runtime.GOOS)
}
