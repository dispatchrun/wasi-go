package streams

import (
	"context"
	"fmt"
	"io"

	"github.com/tetratelabs/wazero"
)

const ModuleName = "streams"

type Stream struct {
	reader io.Reader
	writer io.Writer
}

type Streams struct {
	streams          map[uint32]Stream
	streamHandleBase uint32
}

func MakeStreams() *Streams {
	return &Streams{
		make(map[uint32]Stream),
		1,
	}
}

func Instantiate(ctx context.Context, r wazero.Runtime, s *Streams) error {
	_, err := r.NewHostModuleBuilder(ModuleName).
		NewFunctionBuilder().WithFunc(s.streamReadFn).Export("read").
		NewFunctionBuilder().WithFunc(s.dropInputStreamFn).Export("drop-input-stream").
		NewFunctionBuilder().WithFunc(s.writeStreamFn).Export("write").
		Instantiate(ctx)
	return err
}

func (s *Streams) NewInputStream(reader io.Reader) uint32 {
	s.streamHandleBase++
	s.streams[s.streamHandleBase] = Stream{
		reader: reader,
		writer: nil,
	}
	return s.streamHandleBase
}

func (s *Streams) DeleteStream(handle uint32) {
	delete(s.streams, handle)
}

func (s *Streams) NewOutputStream(writer io.Writer) uint32 {
	s.streamHandleBase++
	s.streams[s.streamHandleBase] = Stream{
		reader: nil,
		writer: writer,
	}
	return s.streamHandleBase
}

func (s *Streams) Read(handle uint32, data []byte) (int, bool, error) {
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

func (s *Streams) Write(handle uint32, data []byte) (int, error) {
	stream, found := s.streams[handle]
	if !found {
		return 0, fmt.Errorf("stream not found: %d", handle)
	}
	if stream.writer == nil {
		return 0, fmt.Errorf("not a writeable stream: %d", handle)
	}
	return stream.writer.Write(data)
}
