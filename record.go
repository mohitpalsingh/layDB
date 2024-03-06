package laydb

import (
	"encoding/binary"
	"errors"
	"time"
)

/* Record is inspired from CaskDB */

var (
	// Error code for invalid entry
	ErrInvalidEntry = errors.New("Invalid Entry")
)

const (
	// keySize (4) + valueSize (4) + timestamp(4)
	headerSize = 12
)

type (
	// structure of record
	record struct {
		meta      *meta
		timestamp uint32
	}

	// structure of meta of the record
	meta struct {
		key       []byte
		value     []byte
		keySize   uint32
		valueSize uint32
	}
)

// to create a new entry
func newInternalLog(key []byte, value []byte, timestamp uint32) *record {
	return &record{
		timestamp: timestamp,
		meta: &meta{
			key:       key,
			value:     value,
			keySize:   uint32(len(key)),
			valueSize: uint32(len(value)),
		},
	}
}

// to create new record without value
func newRecord(key []byte) *record {
	return newInternalLog(key, nil, uint32(time.Now().UnixNano()))
}

// to create new record with value
func newRecordWithValue(key []byte, value []byte) *record {
	return newInternalLog(key, value, uint32(time.Now().Unix()))
}

// to get the size of the record
func (e *record) size() uint32 {
	return headerSize + e.meta.keySize + e.meta.valueSize
}

// encode returns the slice after the entry is encoded
// the entry stored format:
// |--------------------------------------------------|
// |	ks	|	vs	 |	timestamp	|	key	 |	value |
// |--------------------------------------------------|
// | uint32 | uint32 |    uint32    | []byte | []byte |
// |--------------------------------------------------|

func (e *record) encode() ([]byte, error) {
	if e == nil || e.meta.keySize == 0 {
		return nil, ErrInvalidEntry
	}

	ks := e.meta.keySize
	vs := e.meta.valueSize
	buf := make([]byte, e.size())

	binary.BigEndian.PutUint32(buf[:4], ks)
	binary.BigEndian.PutUint32(buf[4:8], vs)
	binary.BigEndian.PutUint32(buf[8:12], e.timestamp)
	copy(buf[headerSize:headerSize+ks], e.meta.key)
	if vs > 0 {
		copy(buf[headerSize+ks:headerSize+ks+vs], e.meta.value)
	}

	return buf, nil
}

// decode returns the record after decoding the entry
func decode(buf []byte) (*record, error) {
	ks := binary.BigEndian.Uint32(buf[:4])
	vs := binary.BigEndian.Uint32(buf[4:8])
	timestamp := binary.BigEndian.Uint32(buf[8:12])

	return &record{
		timestamp: timestamp,
		meta: &meta{
			keySize:   ks,
			valueSize: vs,
			key:       buf[headerSize : headerSize+ks],
			value:     buf[headerSize+ks : headerSize+ks+vs],
		},
	}, nil
}
