package fscache

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/keyneston/fscachemonitor/fslist"
	"github.com/keyneston/fscachemonitor/internal/shared"
	"github.com/keyneston/fscachemonitor/watcher"
)

type FSCache struct {
	Filename string
	Root     string

	fileList fslist.FSList
	watcher  watcher.Watcher

	limit int

	closeCh   chan interface{}
	closeOnce sync.Once
}

func New(filename, root string, mode fslist.Mode) (*FSCache, error) {
	watcher, err := watcher.New(root)
	if err != nil {
		return nil, err
	}

	fs := &FSCache{
		Filename: filename,
		Root:     root,
		watcher:  watcher,
		limit:    40000,
	}

	fs.fileList, err = fslist.New(filename, mode)
	if err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *FSCache) Run() {
	fs.watcher.Start()
	fs.init()

	for {
		select {
		case events := <-fs.watcher.Stream():
			for _, e := range events {
				fs.handleEvent(e)
			}
		case <-fs.closeCh:
			return
		}
	}
}

func eventToAddData(e watcher.Event) fslist.AddData {
	return fslist.AddData{
		Name:      e.Path,
		IsDir:     e.Dir,
		UpdatedAt: time.Now(),
	}
}

func (fs *FSCache) handleEvent(e watcher.Event) {
	logger := shared.Logger()
	// TODO: find a better way:
	for _, seg := range strings.Split(e.Path, "/") {
		if skipFile(seg) {
			logger.Debugf("Skipping %q", e.Path)
			return
		}
	}

	switch e.Type {
	case watcher.EventTypeDelete:
		logger.Debugf("Removing %q", e.Path)
		if err := fs.fileList.Delete(eventToAddData(e)); err != nil {
			logger.Errorf("Error deleting file: %v", err)
		}
	case watcher.EventTypeAdd:
		logger.Debugf("Adding %q", e.Path)
		if err := fs.fileList.Add(eventToAddData(e)); err != nil {
			logger.Errorf("Error adding file: %v", err)
		}
	}
}

func (fs *FSCache) Close() {
	fs.closeOnce.Do(func() {
		fs.watcher.Stop()
		close(fs.closeCh)
	})
}

// init does the initial setup of walking
func (fs *FSCache) init() {
	filepath.WalkDir(fs.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			shared.Logger().Errorf("Error during init: %v", err)
			return nil
		}

		if skipFile(path) {
			shared.Logger().Debugf("Skipping %q", path)
			return filepath.SkipDir
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		fs.fileList.Add(fslist.AddData{
			Name:      path,
			UpdatedAt: info.ModTime(),
			IsDir:     d.IsDir(),
		})
		return nil
	})
}

func skipFile(path string) bool {
	switch filepath.Base(path) {
	case ".git", ".svn":
		return true
	default:
		return false
	}
}
