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

	if pf.opts.Prefix != "" {
		return pf.fetchRangeWithPrefix()
	}

	return pf.fetchRange(nil, nil)
}

// fetchRangeWithPrefix calculates and executes the fetches necessary if we
// have a prefix. Ideally we will return items closest to the current working
// directory first, then return everything else.
//
// This is done by splitting the range into two sections:
// 1. current working directory => end of actual range
// 2. start of actual range => current working directory
func (pf *pebbleFetcher) fetchRangeWithPrefix() (int, error) {
	var lowerBound, middleBound, upperBound []byte

	if pf.opts.Prefix != "" {
		lowerBound = []byte(pf.opts.Prefix)
		middleBound = lowerBound
		upperBound = calcUpperBound(pf.opts.Prefix)

		if pf.opts.CurrentDir != "" && pf.opts.CurrentDir != pf.opts.Prefix {
			middleBound = []byte(pf.opts.CurrentDir)
		}
	}

	var err error
	pf.count, err = pf.fetchRange(middleBound, upperBound)
	if err != nil {
		return pf.count, err
	}

	return pf.fetchRange(lowerBound, middleBound)

}

// fetchRange does the heavy lifting of actually iterating from lower to upper
// and sending them onto the channel.
func (pf *pebbleFetcher) fetchRange(lower, upper []byte) (int, error) {
	var iterOpts *pebble.IterOptions
	if len(lower) > 0 || len(upper) > 0 {
		iterOpts = &pebble.IterOptions{
			LowerBound: lower,
			UpperBound: upper,
		}
	}

	pf.logger.WithField("bounds", logrus.Fields{
		"lower": lower,
		"upper": upper,
	}).WithField("module", "pebbleFetcher").Debug("Doing a fetchRange")

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
		pf.logger.WithField("file", data.Name).Tracef("Checking")

		ignore := pf.ignoreCache.Get(data.Name)
		if ignore != nil && ignore.Match(data.Name, data.IsDir) {
			pf.logger.WithField("file", data.Name).Tracef("Skipping")

			// This requires that the final character be a '/' otherwise the
			// skip will skip adjacent files:
			//
			// Without trailing /
			//
			// * file <- ignore this
			// * file.foo <- don't want to skip this
			// * file/foo <- want to ignore this
			//
			// With trailing /
			//
			// * file.foo <- don't want to skip this
			// * file/ <- ignore this
			// * file/foo <- want to ignore this
			//
			if data.IsDir {
				iter.SeekGE(calcUpperBound(string(data.Name)))
			}
			continue
		}

		if pf.opts.DirsOnly && !data.IsDir {
			pf.logger.WithField("file", data.Name).Tracef("Skipping non-dir")
			continue
		} else if pf.opts.FilesOnly && data.IsDir {
			pf.logger.WithField("file", data.Name).Tracef("Skipping non-file")
			continue
		}

		pf.ch <- data
		pf.count++
	}

	return 0, nil
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
