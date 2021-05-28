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
	ch := make(chan AddData, 1)

	go func() {
		defer close(ch)

		var count int
		var err error

		if !opts.DirsOnly {
			count, err = s.copySet(ch, filePrefix, opts, count)
			if err != nil {
				s.logger.WithError(err).Error()
			}
		}

		count, err = s.copySet(ch, dirPrefix, opts, count)
		if err != nil {
			s.logger.WithError(err).Error()
		}
	}()

	return ch
}

func (s *PebbleList) copySet(ch chan<- AddData, keyType string, opts ReadOptions, count int) (int, error) {
	lowerBound := []byte(fmt.Sprintf("%s%s", keyType, opts.Prefix))
	middleBound := lowerBound
	upperBound := calcUpperBound(fmt.Sprintf("%s%s", keyType, opts.Prefix))

	if opts.CurrentDir != "" && opts.CurrentDir != opts.Prefix {
		middleBound = []byte(fmt.Sprintf("%s%s", keyType, opts.CurrentDir))
	}

	var err error
	count, err = s.fetchRange(ch, middleBound, upperBound, opts, count)
	if err != nil {
		return count, err
	}

	return s.fetchRange(ch, lowerBound, middleBound, opts, count)
}

func (s *PebbleList) fetchRange(ch chan<- AddData, lower, upper []byte, opts ReadOptions, count int) (int, error) {
	iterOpts := &pebble.IterOptions{
		LowerBound: lower,
		UpperBound: upper,
	}

	s.logger.WithField("iterOpts", logrus.Fields{
		"LowerBound": string(iterOpts.LowerBound),
		"UpperBoudn": string(iterOpts.UpperBound),
	}).Debug("Iterating")

	iter := s.db.NewIter(iterOpts)
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		if opts.Limit > 0 && count >= opts.Limit {
			return count, nil
		}

		var data AddData
		if err := json.Unmarshal(iter.Value(), &data); err != nil {
			return count, err
		}

		ch <- data
		count++
	}

	return 0, nil
}

func (s *PebbleList) Flush() error {
	return s.db.Flush()
}
