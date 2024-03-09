package laydb

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrCorrupt  = errors.New("log corrupt")
	ErrClosed   = errors.New("log closed")
	ErrNotFound = errors.New("not found")
	ErrEOF      = errors.New("end of file reached")
)

type Log struct {
	mu     sync.RWMutex
	path   string
	sfile  *os.File
	wbatch Batch

	closed  bool
	corrupt bool
}

type bpos struct {
	pos int
	end int
}

func open(path string) (*Log, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	l := &Log{path: path}

	if err := os.MkdirAll(path, os.ModeAppend); err != nil {
		return nil, err
	}

	if err := l.load(); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *Log) load() error {
	files, err := os.ReadDir(l.path)
	if len(files) == 0 {
		l.sfile, err = os.OpenFile(l.path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModeAppend)
		if err != nil {
			return err
		}
	} else {
		l.sfile, err = os.OpenFile(l.path, os.O_WRONLY, os.ModeAppend)
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Log) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		if l.corrupt {
			return ErrCorrupt
		}
		return ErrClosed
	}

	if err := l.sfile.Sync(); err != nil {
		return err
	}

	if err := l.sfile.Close(); err != nil {
		return err
	}

	l.closed = true

	if l.corrupt {
		return ErrCorrupt
	}
	return nil
}

func (l *Log) Write(data []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.corrupt {
		return ErrCorrupt
	} else if l.closed {
		return ErrClosed
	}

	l.wbatch.Clear()
	l.wbatch.Write(data)
	return l.writeBatch(&l.wbatch)
}

func (l *Log) appendEntry(dst []byte, data []byte) (out []byte, cpos bpos) {
	return appendBinaryEntry(dst, data)
}

func appendBinaryEntry(dst []byte, data []byte) (out []byte, cpos bpos) {
	pos := len(dst)
	dst = appendUvarint(dst, uint64(len(data)))
	dst = append(dst, data...)
	return dst, bpos{pos, len(dst)}
}

func appendUvarint(dst []byte, x uint64) []byte {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], x)
	dst = append(dst, buf[:n]...)
	return dst
}

type Batch struct {
	entries []batchEntry
	datas   []byte
}

type batchEntry struct {
	size int
}

func (b *Batch) Write(data []byte) {
	b.entries = append(b.entries, batchEntry{len(data)})
	b.datas = append(b.datas, data...)
}

func (b *Batch) Clear() {
	b.entries = b.entries[:0]
	b.datas = b.datas[:0]
}

func (l *Log) WriteBatch(b *Batch) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.corrupt {
		return ErrCorrupt
	} else if l.closed {
		return ErrClosed
	}
	if len(b.datas) == 0 {
		return nil
	}
	return l.writeBatch(b)
}

func (l *Log) writeBatch(b *Batch) error {
	datas := b.datas
	for i := 0; i < len(b.entries); i++ {
		// data := datas[:b.entries[i].size]
		datas = datas[b.entries[i].size:]
	}

	if _, err := l.sfile.Write(datas); err != nil {
		return err
	}

	if err := l.sfile.Sync(); err != nil {
		return err
	}

	b.Clear()
	return nil
}

func loadNextBinaryEntry(data []byte) (n int, err error) {
	size, n := binary.Uvarint(data)
	if n <= 0 {
		return 0, ErrCorrupt
	}
	if uint64(len(data)-n) < size {
		return 0, ErrCorrupt
	}

	return n + int(size), nil
}

func (l *Log) Read(index uint64) (data []byte, err error) {
	l.mu.RLock()
	defer l.mu.Unlock()

	if l.corrupt {
		return nil, ErrCorrupt
	} else if l.closed {
		return nil, ErrClosed
	}
	data, err = os.ReadFile(l.path)
	size, n := binary.Uvarint(data)
	if n <= 0 {
		return nil, ErrCorrupt
	}

	if uint64(len(data)-n) < size {
		return nil, ErrCorrupt
	}

	data = make([]byte, size)
	copy(data, data[n:])

	return data, nil
}

func (l *Log) Sync() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.corrupt {
		return ErrCorrupt
	} else if l.closed {
		return ErrClosed
	}

	return l.sfile.Sync()
}
