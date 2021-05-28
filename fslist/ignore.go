package fslist

import "github.com/monochromegane/go-gitignore"

type IgnoreCache struct {
	cache map[string]gitignore.IgnoreMatcher
}

func (i *IgnoreCache) Add(file string) error {
	return nil
}

func (i *IgnoreCache) Get(file string) (gitignore.IgnoreMatcher, error) {
	return nil, nil
}
