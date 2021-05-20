package fscache

import (
	"log"
	"os"
	"path/filepath"
	"strings"
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

func New(filename, root string) (*FSCache, error) {
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

	fs.fileList, err = fslist.NewSQL(filename)
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

func (fs *FSCache) handleEvent(e watcher.Event) {
	// TODO: find a better way:
	for _, seg := range strings.Split(e.Path, "/") {
		if skipFile(seg) {
			logrus.Debugf("Skipping %q", e.Path)
			return
		}
	}

	switch e.Type {
	case watcher.EventTypeDelete:
		logrus.Debugf("Removing %q", e.Path)
		if err := fs.fileList.Delete(e.Path); err != nil {
			logrus.Errorf("Error deleting file: %v", err)
		}
	case watcher.EventTypeAdd:
		logrus.Debugf("Adding %q", e.Path)
		if err := fs.fileList.Add(fslist.AddData{
			Name:      e.Path,
			UpdatedAt: time.Now(),
		}); err != nil {
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
			logrus.Debugf("Skipping %q", path)
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
