package common

import (
	"encoding/binary"
	"fmt"
	"time"
)

type MemoryWriter struct {
	buf []byte
	pos int
}

func NewMemoryWriter(buf_size int) *MemoryWriter {
	return &MemoryWriter{
		buf: make([]byte, buf_size),
		pos: 0,
	}
}

func SerializeHeader(w *MemoryWriter) {
	// AB = Agency Bet
	// 01 = protocol v1
	w.buf[w.pos] = 'A'
	w.buf[w.pos+1] = 'B'
	w.buf[w.pos+2] = 0
	w.buf[w.pos+3] = 1
	w.pos += 4
}

func SerializeInt(w *MemoryWriter, v uint32) {
	binary.BigEndian.PutUint32(w.buf[w.pos:], v)
	w.pos += 4
}

func SerializeString(w *MemoryWriter, s string) error {
	b := []byte(s)
	if len(b) > 255 {
		return fmt.Errorf("string %s is too long to serialize", s)
	}

	size := len(b)
	w.buf[w.pos] = byte(size)
	w.pos++

	copy(w.buf[w.pos:], b)
	w.pos += size

	return nil
}

func DaysSinceEpoch(t time.Time) int32 {
	// Unix timestamps use 1970-01-01 as Epoch
	return int32(t.UTC().Unix() / (24 * 60 * 60))
}

func SerializeDate(w *MemoryWriter, t time.Time) {
	SerializeInt(w, uint32(DaysSinceEpoch(t)))
}

func (b *Bet) Serialize() ([]byte, error) {
	const INT_SIZE = 4
	const HEADER_ID_SIZE = 4
	const HEADER_SIZE = HEADER_ID_SIZE + INT_SIZE
	payloadSize := 2 + len(b.Name) + len(b.LastName) + INT_SIZE*3
	totalSize := HEADER_SIZE + payloadSize

	w := NewMemoryWriter(totalSize)

	SerializeHeader(w)
	SerializeInt(w, uint32(payloadSize))
	if err := SerializeString(w, b.Name); err != nil {
		return nil, err
	}
	if err := SerializeString(w, b.LastName); err != nil {
		return nil, err
	}
	SerializeInt(w, b.Document)
	SerializeDate(w, b.Birthdate)
	SerializeInt(w, b.Number)

	return w.buf, nil
}
