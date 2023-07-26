package streams

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
	streams          map[uint32]stream
	streamHandleBase uint32
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
	s.streamHandleBase++
	s.streams[s.streamHandleBase] = stream{
		reader: reader,
		writer: nil,
	}
	return s.streamHandleBase
}

func (s *streams) DeleteStream(handle uint32) {
	delete(s.streams, handle)
}

func (s *streams) NewOutputStream(writer io.Writer) uint32 {
	s.streamHandleBase++
	s.streams[s.streamHandleBase] = stream{
		reader: nil,
		writer: writer,
	}
	return s.streamHandleBase
}

func (s *streams) Read(handle uint32, data []byte) (int, bool, error) {
	stream, found := s.streams[handle]
	if !found {
		return 0, false, fmt.Errorf("stream not found: %d", handle)
	}
	if stream.reader == nil {
		return 0, false, fmt.Errorf("not a readable stream: %d", handle)
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
		return 0, fmt.Errorf("stream not found: %d", handle)
	}
	if stream.writer == nil {
		return 0, fmt.Errorf("not a writeable stream: %d", handle)
	}
	return stream.writer.Write(data)
}
