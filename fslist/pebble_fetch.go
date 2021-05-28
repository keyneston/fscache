package fslist

import (
	"encoding/json"

	"github.com/cockroachdb/pebble"
	"github.com/sirupsen/logrus"
)

type pebbleFetcher struct {
	db          *pebble.DB
	ignoreCache *IgnoreCache
	count       int
	ch          chan<- AddData
	opts        ReadOptions
	logger      *logrus.Logger
}

func (pf *pebbleFetcher) Fetch() (int, error) {
	defer close(pf.ch)

	lowerBound := []byte(pf.opts.Prefix)
	middleBound := lowerBound
	upperBound := calcUpperBound(pf.opts.Prefix)

	if pf.opts.CurrentDir != "" && pf.opts.CurrentDir != pf.opts.Prefix {
		middleBound = []byte(pf.opts.CurrentDir)
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

		ignore := pf.ignoreCache.Get(data.Name)
		if ignore != nil && ignore.Match(data.Name, data.IsDir) {
			iter.SeekGE(calcUpperBound(string(data.Name)))
			continue
		}

		if pf.opts.DirsOnly && !data.IsDir {
			continue
		}

		pf.ch <- data
		pf.count++
	}

	return 0, nil
}
