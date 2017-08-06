package database

import (
	"database/sql"

	. "github.com/voutasaurus/event"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewDB(url string) (*DB, error) {
	conn, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return &DB{conn}, nil
}

func (db *DB) AddEvent(u *User, e *Event) error {
	// TODO: implement
	return nil
}

func (db *DB) GetEvents() ([]*Event, error) {
	// TODO: implement
	return nil, nil
}
