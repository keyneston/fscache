package fslist

import (
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/sirupsen/logrus"
)

type pebbleFetcher struct {
	db     *pebble.DB
	count  int
	ch     chan<- AddData
	opts   ReadOptions
	logger *logrus.Logger
}

func (pf *pebbleFetcher) Fetch() (int, error) {
	defer close(pf.ch)

	var err error

	if !pf.opts.DirsOnly {
		pf.count, err = pf.copySet(filePrefix)
		if err != nil {
			pf.logger.WithError(err).Error()
		}
	}

	pf.count, err = pf.copySet(dirPrefix)
	if err != nil {
		pf.logger.WithError(err).Error()
	}

	return pf.count, nil
}

func (pf *pebbleFetcher) copySet(keyType string) (int, error) {
	lowerBound := []byte(fmt.Sprintf("%s%s", keyType, pf.opts.Prefix))
	middleBound := lowerBound
	upperBound := calcUpperBound(fmt.Sprintf("%s%s", keyType, pf.opts.Prefix))

	if pf.opts.CurrentDir != "" && pf.opts.CurrentDir != pf.opts.Prefix {
		middleBound = []byte(fmt.Sprintf("%s%s", keyType, pf.opts.CurrentDir))
	}

	var err error
	pf.count, err = pf.fetchRange(middleBound, upperBound)
	if err != nil {
		return pf.count, err
	}

	return pf.fetchRange(lowerBound, middleBound)
}

func (pf *pebbleFetcher) fetchRange(lower, upper []byte) (int, error) {
	iterOpts := &pebble.IterOptions{
		LowerBound: lower,
		UpperBound: upper,
	}

	pf.logger.WithField("iterOpts", logrus.Fields{
		"LowerBound": string(iterOpts.LowerBound),
		"UpperBoudn": string(iterOpts.UpperBound),
	}).Debug("Iterating")

	iter := pf.db.NewIter(iterOpts)
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		if pf.opts.Limit > 0 && pf.count >= pf.opts.Limit {
			return pf.count, nil
		}

		var data AddData
		if err := json.Unmarshal(iter.Value(), &data); err != nil {
			return pf.count, err
		}

		pf.ch <- data
		pf.count++
	}

	return 0, nil
}
