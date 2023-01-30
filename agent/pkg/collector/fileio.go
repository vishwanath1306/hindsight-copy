package collector

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
)

/*
A TraceToStore represents reports buffered in memory that need
to be written to disk.
*/
type TraceToStore struct {
	m        sync.Mutex
	trace_id uint64
	buffers  []*ReceivedBuffer // Group buffers by source agent
}

type ReceivedBuffer struct {
	trace_id     uint64
	source_agent string
	buffer       []byte
}

func initTraceToStore(trace_id uint64) *TraceToStore {
	var t TraceToStore
	t.trace_id = trace_id
	return &t
}

func (t *TraceToStore) AddBuffer(buf *ReceivedBuffer) {
	t.m.Lock()
	defer t.m.Unlock()

	t.buffers = append(t.buffers, buf)
}

func (t *TraceToStore) takeBuffers() []*ReceivedBuffer {
	t.m.Lock()
	defer t.m.Unlock()

	buffers := t.buffers
	t.buffers = nil
	return buffers
}

func (t *TraceToStore) AppendToFile(file *os.File) (err error) {
	buffers := t.takeBuffers()
	for _, buf := range buffers {
		err = buf.WriteToFile(file)
		if err != nil {
			return
		}
	}
	return nil
}

func doFileWrite(f *os.File, bs []byte) error {
	for len(bs) > 0 {
		written, err := f.Write(bs)
		if err != nil {
			return err
		}
		bs = bs[written:]
	}
	return nil
}

func writeFileLengthPrefixed(f *os.File, buf []byte) (err error) {
	sizebuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizebuf, uint32(len(buf)))
	err = doFileWrite(f, sizebuf)
	if err != nil {
		return
	}
	err = doFileWrite(f, buf)
	return
}

func doFileRead(f *os.File, dst []byte) error {
	for len(dst) > 0 {
		read, err := f.Read(dst)
		if err != nil {
			return err
		}
		if read == 0 {
			return fmt.Errorf("File ended before we could read %d more bytes", len(dst))
		}
		dst = dst[read:]
	}
	return nil
}

func readFileLengthPrefixed(f *os.File) (buf []byte, err error) {
	szbuf := make([]byte, 4)
	err = doFileRead(f, szbuf)
	if err != nil {
		return
	}
	sz := binary.LittleEndian.Uint32(szbuf)
	if sz > 1024*1024*100 {
		err = fmt.Errorf("Invalid read of size %d", sz)
		return
	}
	buf = make([]byte, sz)
	err = doFileRead(f, buf)
	return
}

func (b *ReceivedBuffer) WriteToFile(f *os.File) (err error) {
	/* Serialize as follows:
	  source_agent (length prefixed)
		buffer (length prefixed)
	*/
	err = writeFileLengthPrefixed(f, []byte(b.source_agent))
	if err != nil {
		return
	}
	err = writeFileLengthPrefixed(f, b.buffer)
	return
}

func ReadFromFile(f *os.File) (source_agent string, buffer []byte, err error) {
	var agentbs []byte
	agentbs, err = readFileLengthPrefixed(f)
	if err != nil {
		return
	}
	source_agent = string(agentbs)
	buffer, err = readFileLengthPrefixed(f)
	return
}
