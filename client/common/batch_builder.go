package common

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

const (
	MaxBatchSize = 8 * 1024 // 8KB
	BatchHeader  = "BB\x00\x01"
)

type BatchBuilder struct {
	Reader      *csv.Reader
	MaxItems    int
	NextBetBuf  []byte
	ReadingLine int
	ReachedEOF  bool
}

func NewBatchBuilder(csvFile *os.File, maxItems int) *BatchBuilder {
	b := &BatchBuilder{
		Reader:      csv.NewReader(csvFile),
		MaxItems:    maxItems,
		ReadingLine: 0,
		ReachedEOF:  false,
	}

	b.PrepareNextBet()

	return b
}

func (b *BatchBuilder) PrepareNextBet() error {
	record, err := b.Reader.Read()
	b.ReadingLine++
	if err != nil {
		if err == io.EOF {
			b.ReachedEOF = true
		}
		b.NextBetBuf = nil
		return err
	}

	bet, err := FromCSVRecord(record)
	if err != nil {
		return fmt.Errorf("line %d: %w", b.ReadingLine, err)
	}

	buf, err := bet.Serialize()
	if err != nil {
		b.NextBetBuf = nil
		return fmt.Errorf("line %d: %w", b.ReadingLine, err)
	}

	b.NextBetBuf = buf
	return nil
}

func (b *BatchBuilder) BuildNext() ([]byte, int, error) {
	var batchBuf bytes.Buffer
	batchBuf.Write([]byte(BatchHeader))
	// Slot for final batch items num
	batchBuf.Write(make([]byte, 4))
	items := 0

	for !b.ReachedEOF && items < b.MaxItems &&
		batchBuf.Len()+len(b.NextBetBuf) < MaxBatchSize {

		batchBuf.Write(b.NextBetBuf)
		items++

		err := b.PrepareNextBet()
		if err != nil && !b.ReachedEOF {
			return nil, 0, err
		}
	}

	data := batchBuf.Bytes()
	binary.BigEndian.PutUint32(data[4:], uint32(items))
	return data, items, nil
}
