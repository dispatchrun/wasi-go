package types

import (
	"context"
	"log"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/common"
	"github.com/tetratelabs/wazero/api"
)

func newFieldsFn(_ context.Context, mod api.Module, handle, ptr uint32) int32 {
	return 0
}

func allocateWriteString(ctx context.Context, m api.Module, s string) uint32 {
	ptr, err := common.Malloc(ctx, m, uint32(len(s)))
	if err != nil {
		log.Fatalf(err.Error())
	}
	m.Memory().Write(ptr, []byte(s))
	return ptr
}

func fieldsEntriesFn(ctx context.Context, mod api.Module, handle, out_ptr uint32) {
	headers := ResponseHeaders()
	l := uint32(len(headers))
	// 8 bytes per string/string
	ptr, err := common.Malloc(ctx, mod, l*16)
	if err != nil {
		log.Fatalf(err.Error())
	}
	data := []byte{}
	data = le.AppendUint32(data, ptr)
	data = le.AppendUint32(data, l)
	// write result
	mod.Memory().Write(out_ptr, data)

	// ok now allocate and write the strings.
	data = []byte{}
	for k, v := range headers {
		data = le.AppendUint32(data, allocateWriteString(ctx, mod, k))
		data = le.AppendUint32(data, uint32(len(k)))
		data = le.AppendUint32(data, allocateWriteString(ctx, mod, v[0]))
		data = le.AppendUint32(data, uint32(len(v[0])))
	}
	mod.Memory().Write(ptr, data)
}
