package fslist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/rs/zerolog"
)

const (
	dirPrefix  = "dir:"
	filePrefix = "file:"
)

var _ FSList = &PebbleList{}

type PebbleList struct {
	db          *pebble.DB
	location    string
	ignoreCache *IgnoreCache

	logger *zerolog.Logger
}

func NewPebble() (FSList, error) {
	return openPebble()
}

func OpenPebble() (FSList, error) {
	return openPebble()
}

func openPebble() (FSList, error) {
	location, err := os.MkdirTemp("", "fscache-pebble-db-*")
	if err != nil {
		return nil, err
	}

	logger := shared.Logger().With().Str("database", location).Str("mode", "pebble").Logger()
	logger.Debug().Msg("opening pebble database")

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
		db:          db,
		ignoreCache: &IgnoreCache{},
		location:    location,
		logger:      &logger,
	}

	return s, nil
}

func (s *PebbleList) Close() error {
	return s.db.Close()
}

func (s *PebbleList) init() error {
	return nil
}

func (s *PebbleList) Pending() bool {
	return false
}

func (s *PebbleList) Add(data AddData) error {
	s.logger.Trace().Object("data", data).Msg("adding")

	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := s.db.Set(data.pebbleKey(), encoded, pebble.NoSync); err != nil {
		return err
	}

	if filepath.Base(data.Name) == ".gitignore" {
		if err := s.ignoreCache.Add(string(data.pebbleKey())); err != nil {
			return err
		}
	}

	return nil
}

func (s *PebbleList) Delete(data AddData) error {
	s.logger.Trace().Object("data", data).Msg("deleting")
	return s.db.Delete(data.pebbleKey(), pebble.NoSync)
}

func (s *PebbleList) Len() int {
	// TODO
	return 0
}

func (s *PebbleList) newPebbleFetcher(opts ReadOptions) (*pebbleFetcher, <-chan AddData) {
	ch := make(chan AddData, 1)

	l := s.logger.With().Str("module", "pebbleFetcher").Logger()
	return &pebbleFetcher{
		db:          s.db,
		ignoreCache: s.ignoreCache,
		logger:      &l,
		ch:          ch,
		opts:        opts,
		count:       0,
	}, ch
}

func (s *PebbleList) Fetch(opts ReadOptions) <-chan AddData {
	fetcher, ch := s.newPebbleFetcher(opts)
	go fetcher.Fetch()

	return ch
}

func (s *PebbleList) Flush() error {
	return s.db.Flush()
}
