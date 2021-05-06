package fslist

import (
	"database/sql"
	"fmt"
)

var _ FSList = &SQList{}

func NewSQL(location string) (FSList, error) {
	db, err := sql.Open("sqlite3", location)
	if err != nil {
		return nil, fmt.Errorf("Error creating SQList: %w", err)
	}

	return &SQList{
		db: db,
	}, nil
}

type SQList struct {
	db *sql.DB
}

func (s *SQList) Pending() bool {
	return false
}

func (s *SQList) Add(name string) error {

	return nil
}

func (s *SQList) Delete(name string) error {

	return nil
}

func (s *SQList) Len() int {

	return 0
}

func (s *SQList) Write() error {
	return nil
}
