package fslist

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cockroachdb/pebble"
	"github.com/keyneston/fscachemonitor/internal/shared"
	"github.com/sirupsen/logrus"
)

var _ FSList = &PebbleList{}

type PebbleList struct {
	db       *pebble.DB
	location string

	logger *logrus.Entry
}

func NewPebble(location string) (FSList, error) {
	if location == "" {
		return nil, fmt.Errorf("Must supply a location for the database")
	}

	shared.Logger().WithField("database", location).Debugf("new sqlite3 database")
	list, err := OpenPebble(location)
	if err != nil {
		return nil, err
	}
	s := list.(*PebbleList)

	return s, s.init()
}

func OpenPebble(location string) (FSList, error) {
	logger := shared.Logger().WithField("database", location).WithField("mode", "pebble")
	logger.Debugf("opening sqlite3 database")

	db, err := pebble.Open(location, &pebble.Options{})
	if err != nil {
		return nil, fmt.Errorf("Error creating PebbleList: %w", err)
	}

	s := &PebbleList{
		db:       db,
		location: location,
		logger:   logger,
	}

	return s, nil
}

func (s *PebbleList) init() error {
	return nil
}

func (s *PebbleList) Pending() bool {
	return false
}

func (s *PebbleList) Add(data AddData) error {
	s.logger.WithField("data", data).Debugf("Adding")

	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := s.db.Set([]byte(data.Name), encoded, pebble.NoSync); err != nil {
		return err
	}

	return nil
}

func (s *PebbleList) Delete(name string) error {
	return nil
}

func (s *PebbleList) Len() int {
	// TODO
	return 0
}

func (s *PebbleList) Copy(w io.Writer, opts ReadOptions) error {
	return nil
}
