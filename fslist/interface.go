package fslist

import (
	"fmt"
	"io"
	"time"
)

type AddData struct {
	Name      string
	UpdatedAt time.Time
	IsDir     bool
}

type FSList interface {
	Pending() bool
	Add(AddData) error
	Delete(name string) error
	Len() int
	Copy(io.Writer, ReadOptions) error
}

type ReadOptions struct {
	Limit    int
	DirsOnly bool
	Prefix   string
}

type Mode = string

const (
	ModeSQL    Mode = "sql"
	ModePebble Mode = "pebble"
)

func Open(path string, mode Mode) (FSList, error) {
	// do a thing here
	switch mode {
	case ModeSQL:
		return OpenSQL(path)
	case ModePebble:
		return OpenPebble(path)
	}

	return nil, fmt.Errorf("Unknown mode: %v", mode)
}

func New(path string, mode Mode) (FSList, error) {
	// do a thing here
	switch mode {
	case ModeSQL:
		return NewSQL(path)
	case ModePebble:
		return NewPebble(path)
	}

	return nil, fmt.Errorf("Unknown mode: %v", mode)
}
