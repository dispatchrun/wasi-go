package wasi_streams

import (
	"context"
	"log"

	"github.com/tetratelabs/wazero/api"
)

func writeStreamFn(_ context.Context, mod api.Module, stream, ptr, l, result_ptr uint32) {
	data, ok := mod.Memory().Read(ptr, l)
	if !ok {
		log.Printf("Body read failed!\n")
		return
	}
	_, err := Streams.Write(stream, data)
	if err != nil {
		log.Printf("Failed to read: %v\n", err.Error())
	}
}
