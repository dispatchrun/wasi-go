package streams

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/tetratelabs/wazero"
)

const ModuleName = "streams"

type Stream struct {
	reader io.Reader
	writer io.Writer
}

type Streams struct {
	lock             sync.RWMutex
	streams          map[uint32]Stream
	streamHandleBase uint32
}

func MakeStreams() *Streams {
	return &Streams{
		streams:          make(map[uint32]Stream),
		streamHandleBase: 1,
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
	return s.newStream(reader, nil)
}

func (s *Streams) NewOutputStream(writer io.Writer) uint32 {
	return s.newStream(nil, writer)
}

func (s *Streams) newStream(reader io.Reader, writer io.Writer) uint32 {
	streamHandleBase := atomic.AddUint32(&s.streamHandleBase, 1)
	s.lock.Lock()
	s.streams[streamHandleBase] = Stream{
		reader: reader,
		writer: writer,
	}
	s.lock.Unlock()
	return streamHandleBase
}

func (s *Streams) DeleteStream(handle uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.streams, handle)
}

func (s *Streams) GetStream(handle uint32) (stream Stream, found bool) {
	s.lock.RLock()
	stream, found = s.streams[handle]
	s.lock.RUnlock()
	return
}

func (s *Streams) Read(handle uint32, data []byte) (int, bool, error) {
	stream, found := s.GetStream(handle)
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
	stream, found := s.GetStream(handle)
	if !found {
		return 0, fmt.Errorf("stream not found: %d", handle)
	}
	if stream.writer == nil {
		return 0, fmt.Errorf("not a writeable stream: %d", handle)
	}
	return stream.writer.Write(data)
}
