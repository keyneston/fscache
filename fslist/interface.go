package fslist

import (
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
	Write() error
	Copy(io.Writer, ReadOptions) error
}

type ReadOptions struct {
	Limit    int
	DirsOnly bool
}

type Mode = string

const (
	ModeSQL  Mode = "sql"
	ModeList      = "list"
)

func Open(path string, mode Mode) (FSList, error) {
	// do a thing here
	switch mode {
	case ModeSQL:
		return OpenSQL(path)
	case ModeList:
		return OpenList(path)
	}

	return nil, nil
}
