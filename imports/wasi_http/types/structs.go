package types

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/stealthrocket/wasi-go/imports/wasi_http/common"
	"github.com/tetratelabs/wazero/api"
)

type Fields map[string][]string
type fieldsCollection struct {
	fields       map[uint32]Fields
	baseFieldsId uint32
}

var f = fieldsCollection{make(map[uint32]Fields), 0}

func GetFields(handle uint32) (Fields, bool) {
	fields, found := f.fields[handle]
	return fields, found
}

func newFieldsFn(_ context.Context, mod api.Module, ptr, len uint32) uint32 {
	data, ok := mod.Memory().Read(ptr, len*16)
	if !ok {
		fmt.Println("Error reading fields.")
		return 0
	}
	fields := make(Fields)
	for i := uint32(0); i < len; i++ {
		key_ptr := binary.LittleEndian.Uint32(data[i*16 : i*16+4])
		key_len := binary.LittleEndian.Uint32(data[i*16+4 : i*16+8])
		key, ok := common.ReadString(mod, key_ptr, key_len)
		if !ok {
			fmt.Println("Error reading key")
			return 0
		}
		val_ptr := binary.LittleEndian.Uint32(data[i*16+8 : i*16+12])
		val_len := binary.LittleEndian.Uint32(data[i*16+12 : i*16+16])
		val, ok := common.ReadString(mod, val_ptr, val_len)
		if !ok {
			fmt.Println("Error reading value")
			return 0
		}
		if _, found := fields[key]; !found {
			fields[key] = []string{}
		}
		fields[key] = append(fields[key], val)
	}
	return MakeFields(fields)
}

func MakeFields(fields Fields) uint32 {
	f.baseFieldsId++
	f.fields[f.baseFieldsId] = fields
	return f.baseFieldsId
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
	headers, found := GetFields(handle)
	if !found {
		return
	}
	l := uint32(len(headers))
	// 8 bytes per string/string
	ptr, err := common.Malloc(ctx, mod, l*16)
	if err != nil {
		log.Fatalf(err.Error())
	}

	le := binary.LittleEndian
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
