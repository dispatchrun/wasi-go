package common

import (
	"context"
	"encoding/binary"
	"fmt"
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

func writeStringToMemory(ctx context.Context, module api.Module, str string) (uint32, error) {
	data := []byte(str)
	strPtr, err := Malloc(ctx, module, uint32(len(data)))
	if err != nil {
		return 0, err
	}
	if !module.Memory().WriteString(strPtr, str) {
		return 0, fmt.Errorf("failed to write string")
	}
	return strPtr, nil
}

func WriteOptionalString(ctx context.Context, module api.Module, ptr uint32, str string) error {
	strPtr, err := writeStringToMemory(ctx, module, str)
	if err != nil {
		return err
	}
	data := []byte{}
	data = binary.LittleEndian.AppendUint32(data, 1) // is some
	data = binary.LittleEndian.AppendUint32(data, strPtr)
	data = binary.LittleEndian.AppendUint32(data, uint32(len(str)))
	if !module.Memory().Write(ptr, data) {
		return fmt.Errorf("failed to write struct")
	}
	return nil
}

func WriteString(ctx context.Context, module api.Module, ptr uint32, str string) error {
	strPtr, err := writeStringToMemory(ctx, module, str)
	if err != nil {
		return err
	}
	data := []byte{}
	data = binary.LittleEndian.AppendUint32(data, strPtr)
	data = binary.LittleEndian.AppendUint32(data, uint32(len(str)))
	if !module.Memory().Write(ptr, data) {
		return fmt.Errorf("failed to write struct")
	}
	return nil
}

func WriteUint32(ctx context.Context, mod api.Module, val uint32) (uint32, error) {
	ptr, err := Malloc(ctx, mod, 4)
	if err != nil {
		return 0, err
	}
	data := []byte{}
	data = binary.LittleEndian.AppendUint32(data, val)
	if !mod.Memory().Write(ptr, data) {
		return 0, fmt.Errorf("failed to write uint32")
	}
	return ptr, nil
}
