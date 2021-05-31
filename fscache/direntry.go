package fscache

import (
	"fmt"
	"os"
	"path/filepath"
)

func getDirEntry(path string) (os.DirEntry, error) {
	dir, file := filepath.Split(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.Name() == file {
			return e, nil
		}
	}

	return nil, fmt.Errorf("Unable to find direntry for %v", path)
}
