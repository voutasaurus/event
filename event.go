package event

import "time"

type Event struct {
	when time.Time
	what string // url
	done bool
}
