package fslist

import (
	"fmt"
)

type FSList interface {
	Add(AddData) error
	Close() error
	Delete(AddData) error
	Fetch(ReadOptions) <-chan AddData
	Flush() error
	Len() int
	Pending() bool
}

type ReadOptions struct {
	Limit      int
	DirsOnly   bool
	Prefix     string
	CurrentDir string
}

type Mode = string

const (
	ModeSQL    Mode = "sql"
	ModePebble Mode = "pebble"
)

func Open(mode Mode) (FSList, error) {
	// do a thing here
	switch mode {
	case ModeSQL:
		return OpenSQL()
	case ModePebble:
		return OpenPebble()
	}

	return nil, fmt.Errorf("Unknown mode: %v", mode)
}

func New(mode Mode) (FSList, error) {
	// do a thing here
	switch mode {
	case ModeSQL:
		return NewSQL()
	case ModePebble:
		return NewPebble()
	}

	return nil, fmt.Errorf("Unknown mode: %v", mode)
}
