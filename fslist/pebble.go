package fslist

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cockroachdb/pebble"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/sirupsen/logrus"
)

const (
	dirPrefix  = "dir:"
	filePrefix = "file:"
)

var _ FSList = &PebbleList{}

type PebbleList struct {
	db       *pebble.DB
	location string

	logger *logrus.Entry
}

func NewPebble() (FSList, error) {
	return openPebble()
}

func OpenPebble() (FSList, error) {
	return openPebble()
}

//TODO: handle closing the DB connection

func openPebble() (FSList, error) {
	location, err := os.MkdirTemp("", "fscache-pebble-db-*")
	if err != nil {
		return nil, err
	}

	logger := shared.Logger().WithField("database", location).WithField("mode", "pebble")
	logger.Debugf("opening pebble database")

	if location == "" {
		return nil, fmt.Errorf("Must supply a location for the database")
	}

	// location should be a socket, then we create the DB in a tmp dir each time.

	db, err := pebble.Open(location, &pebble.Options{
		DisableWAL: true, // Database is wiped away at start, so WAL is not needed.
	})
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
	s.logger.WithField("data", data).Tracef("Adding")

	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := s.db.Set(data.key(), encoded, pebble.NoSync); err != nil {
		return err
	}

	return nil
}

func (s *PebbleList) Delete(data AddData) error {
	s.logger.WithField("data", data).Tracef("Deleting")
	return s.db.Delete(data.key(), pebble.NoSync)
}

func (s *PebbleList) Len() int {
	// TODO
	return 0
}

func (s *PebbleList) newPebbleFetcher(opts ReadOptions) (*pebbleFetcher, <-chan AddData) {
	ch := make(chan AddData, 1)

	return &pebbleFetcher{
		db:     s.db,
		logger: s.logger.WithField("module", "pebbleFetcher").Logger,
		ch:     ch,
		opts:   opts,
		count:  0,
	}, ch
}

// calcUpperBound takes a string and converts its last character to one greater than it is. e.g. prefix => prefiy. That way it can match all all things that being with prefix but nothing else.
func calcUpperBound(prefix string) []byte {
	if len(prefix) == 0 {
		return []byte{}
	}

	p := []byte(prefix)

	p[len(p)-1] = p[len(p)-1] + 1
	return p
}

func (s *PebbleList) Fetch(opts ReadOptions) <-chan AddData {
	fetcher, ch := s.newPebbleFetcher(opts)
	go fetcher.Fetch()

	return ch
}

func (s *PebbleList) Flush() error {
	return s.db.Flush()
}
