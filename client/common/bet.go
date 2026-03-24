package common

import (
	"fmt"
	"strconv"
	"time"
)

type Bet struct {
	Name      string
	LastName  string
	Document  uint32
	Birthdate time.Time
	Number    uint32
}

func FromCSVRecord(record []string) (*Bet, error) {
	if len(record) < 5 {
		return nil, fmt.Errorf("invalid record length: %d", len(record))
	}

	dni, err := strconv.ParseUint(record[2], 10, 32)
	if err != nil {
		return nil, err
	}

	birthdate, err := time.Parse("2006-01-02", record[3])
	if err != nil {
		return nil, err
	}

	number, err := strconv.ParseUint(record[4], 10, 32)
	if err != nil {
		return nil, err
	}

	return &Bet{
		Name:      record[0],
		LastName:  record[1],
		Document:  uint32(dni),
		Birthdate: birthdate,
		Number:    uint32(number),
	}, nil
}
