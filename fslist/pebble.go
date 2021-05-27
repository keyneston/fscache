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
	return openPebble(location)
}

func OpenPebble(location string) (FSList, error) {
	return openPebble(location)
}

//TODO: handle closing the DB connection

func openPebble(location string) (FSList, error) {
	logger := shared.Logger().WithField("database", location).WithField("mode", "pebble")
	logger.Debugf("opening pebble database")

	if location == "" {
		return nil, fmt.Errorf("Must supply a location for the database")
	}

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
	s.logger.WithField("data", data).Debugf("Adding")

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

func (s *PebbleList) Copy(w io.Writer, opts ReadOptions) error {
	iterOpts := &pebble.IterOptions{
		LowerBound: []byte(fmt.Sprintf("dir:%s", opts.Prefix)),
		UpperBound: calcUpperBound(fmt.Sprintf("dir:%s", opts.Prefix)),
	}
	s.logger.WithField("iterOpts", logrus.Fields{
		"LowerBound": string(iterOpts.LowerBound),
		"UpperBoudn": string(iterOpts.UpperBound),
	}).Debug("Iterating")

	iter := s.db.NewIter(iterOpts)
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		w.Write(iter.Key()[4:])
		w.Write([]byte{'\n'})
	}

	return nil
}
