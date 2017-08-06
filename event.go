package event

import "time"

type Event struct {
	When time.Time
	What string // url
	Done bool
}
