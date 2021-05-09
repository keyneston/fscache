package fslist

import "time"

type AddData struct {
	Name      string
	UpdatedAt time.Time
}

type FSList interface {
	Pending() bool
	Add(AddData) error
	Delete(name string) error
	Len() int
	Write() error
}
