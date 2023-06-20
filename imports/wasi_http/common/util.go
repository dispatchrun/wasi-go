package common

import (
	"context"
	"log"

	"github.com/tetratelabs/wazero/api"
)

func Malloc(ctx context.Context, m api.Module, size uint32) (uint32, error) {
	malloc := m.ExportedFunction("cabi_realloc")
	result, err := malloc.Call(ctx, 0, 0, 4, uint64(size))
	if err != nil {
		log.Fatalf(err.Error())
	}
	return uint32(result[0]), err
}

func ReadString(mod api.Module, ptr, len uint32) (string, bool) {
	data, ok := mod.Memory().Read(ptr, len)
	if !ok {
		return "", false
	}
	return string(data), true
}
