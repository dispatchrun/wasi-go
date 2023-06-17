package wasi_streams

import (
	"context"
	"fmt"
	"io"

	"github.com/tetratelabs/wazero"
)

const ModuleName = "streams"

type stream struct {
	reader io.Reader
	writer io.Writer
}

type streams struct {
	streams            map[uint32]stream
	stream_handle_base uint32
}

var Streams = &streams{
	make(map[uint32]stream),
	1,
}

func Instantiate(ctx context.Context, r wazero.Runtime) error {
	_, err := r.NewHostModuleBuilder(ModuleName).
		NewFunctionBuilder().WithFunc(streamReadFn).Export("read").
		NewFunctionBuilder().WithFunc(dropInputStreamFn).Export("drop-input-stream").
		NewFunctionBuilder().WithFunc(writeStreamFn).Export("write").
		Instantiate(ctx)
	return err
}

func (s *streams) NewInputStream(reader io.Reader) uint32 {
	s.stream_handle_base++
	s.streams[s.stream_handle_base] = stream{
		reader: reader,
		writer: nil,
	}
	return s.stream_handle_base
}

func (s *streams) DeleteStream(handle uint32) {
	delete(s.streams, handle)
}

func (s *streams) NewOutputStream(writer io.Writer) uint32 {
	s.stream_handle_base++
	s.streams[s.stream_handle_base] = stream{
		reader: nil,
		writer: writer,
	}
	return s.stream_handle_base
}

func (s *streams) Read(handle uint32, data []byte) (int, bool, error) {
	stream, found := s.streams[handle]
	if !found {
		return 0, false, fmt.Errorf("Stream not found", handle)
	}
	if stream.reader == nil {
		return 0, false, fmt.Errorf("Not a readable stream")
	}

	n, err := stream.reader.Read(data)
	if err == io.EOF {
		return n, true, nil
	}
	return n, false, err
}

func (s *streams) Write(handle uint32, data []byte) (int, error) {
	stream, found := s.streams[handle]
	if !found {
		return 0, fmt.Errorf("Stream not found", handle)
	}
	if stream.writer == nil {
		return 0, fmt.Errorf("Not a writeable stream")
	}
	return stream.writer.Write(data)
}
