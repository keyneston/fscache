package fscache

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/keyneston/fscachemonitor/fslist"
	"github.com/keyneston/fscachemonitor/watcher"

	"github.com/sirupsen/logrus"
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

func New(filename, root string, sql bool) (*FSCache, error) {
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

	if sql {
		fs.fileList, err = fslist.NewSQL(filename)
	} else {
		fs.fileList, err = fslist.NewList(filename)
	}
	if err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *FSCache) Run() {
	fs.watcher.Start()
	fs.init()
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			fs.updateWritten()
		case events := <-fs.watcher.Stream():
			for _, e := range events {
				fs.handleEvent(e)
			}
		case <-fs.closeCh:
			return
		}
	}
}

func (fs *FSCache) updateWritten() {
	if !fs.fileList.Pending() {
		return
	}
	logrus.Debugf("Pending updates, writing new cache")

	if err := fs.fileList.Write(); err != nil {
		logrus.Errorf("Error: %v", err)
		return
	}
}

func (fs *FSCache) handleEvent(e watcher.Event) {
	if skipFile(e.Path) {
		log.Printf("Skipping %q", e.Path)
		return
	}

	switch e.Type {
	case watcher.EventTypeAdd:
		logrus.Debugf("Removing %q", e.Path)
		if err := fs.fileList.Delete(e.Path); err != nil {
			logrus.Errorf("Error adding file: %v", err)
		}
	case watcher.EventTypeDelete:
		logrus.Debugf("Adding %q", e.Path)
		if err := fs.fileList.Add(e.Path); err != nil {
			logrus.Errorf("Error adding file: %v", err)
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
			log.Printf("Error during init: %v", err)
		}

		if skipFile(path) {
			log.Printf("Skipping %q", path)
			return filepath.SkipDir
		}

		fs.fileList.Add(path)
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
