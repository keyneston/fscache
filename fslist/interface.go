package fslist

import (
	"bytes"
	"fmt"
	"time"
)

type AddData struct {
	Name      string
	UpdatedAt time.Time
	IsDir     bool
}

func (a AddData) key() []byte {
	key := bytes.NewBuffer(make([]byte, 0, len(a.Name)+5))

	if a.IsDir {
		key.WriteString(dirPrefix)
	} else {
		key.WriteString(filePrefix)
	}

	key.WriteString(a.Name)

	return key.Bytes()
}

type FSList interface {
	Pending() bool
	Add(AddData) error
	Delete(AddData) error
	Len() int
	Fetch(ReadOptions) <-chan AddData
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
