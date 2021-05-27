package fslist

import (
	"bytes"
	"fmt"
	"io"
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
		key.WriteString("dir:")
	} else {
		key.WriteString("file:")
	}

	key.WriteString(a.Name)

	return key.Bytes()
}

type FSList interface {
	Pending() bool
	Add(AddData) error
	Delete(AddData) error
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
