package fslist

import (
	"path/filepath"
	"strings"

	"github.com/keyneston/fscache/internal/shared"
	"github.com/monochromegane/go-gitignore"
)

type IgnoreCache struct {
	cache map[string]gitignore.IgnoreMatcher
}

func (ic *IgnoreCache) Add(file string) error {
	if ic.cache == nil {
		ic.cache = make(map[string]gitignore.IgnoreMatcher)
	}

	superior := ic.findSuperior(file)
	shared.Logger().WithField("module", "IgnoreCache").WithField("file", file).WithField("superior", superior).Trace()

	matcher, err := gitignore.NewGitIgnore(file, ic.findSuperior(file)...)
	if err != nil {
		return err
	}

	ic.cache[filepath.Dir(file)] = matcher
	return nil
}

// Get finds the closest gitignore file. If no git ignore files exist above the
// input, then it returns nil.
func (ic *IgnoreCache) Get(file string) gitignore.IgnoreMatcher {
	segments := strings.Split(file, "/")

	for i := len(segments); i > 0; i-- {
		if matcher, ok := ic.cache[strings.Join(segments[0:i], "/")]; ok {
			return matcher
		}
	}

	return nil
}

func (ic *IgnoreCache) findSuperior(file string) []string {
	res := []string{}

	segments := strings.SplitAfter(filepath.Clean(file), "/")
	for i := range segments {
		name := filepath.Join(segments[0:i]...)
		if _, ok := ic.cache[name]; ok {
			res = append(res, name)
		}
	}

	return res
}
