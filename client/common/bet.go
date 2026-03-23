package common

import "time"

type Bet struct {
	Name      string
	LastName  string
	Document  uint32
	Birthdate time.Time
	Number    uint32
}
