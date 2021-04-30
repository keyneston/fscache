package fscache

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsevents"
	"github.com/sirupsen/logrus"
)

type FSCache struct {
	Filename string
	Root     string

	fileList *FSList
	watcher  *fsevents.EventStream

	limit int

	closeCh   chan interface{}
	closeOnce sync.Once
}

func New(filename, root string) (*FSCache, error) {
	return &FSCache{
		Filename: filename,
		Root:     root,
		watcher: &fsevents.EventStream{
			Paths:   []string{root},
			Latency: time.Second,
			Flags:   fsevents.WatchRoot | fsevents.FileEvents,
		},
		fileList: NewFSList(),
		limit:    40000,
	}, nil
}

func (fs *FSCache) Run() {
	fs.watcher.Start()
	fs.init()
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			fs.updateWritten()
		case events := <-fs.watcher.Events:
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

	f, err := os.OpenFile(fs.Filename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("Error: %v", err)
		return
	}
	defer f.Close()
	if _, err := fs.fileList.Write(f); err != nil {
		logrus.Errorf("Error: %v", err)
		return
	}
}

func (fs *FSCache) handleEvent(e fsevents.Event) {
	if skipFile(e.Path) {
		log.Printf("Skipping %q", e.Path)
		return
	}

	switch {
	case checkFlag(e.Flags, fsevents.ItemRemoved):
		logrus.Debugf("Removing %q", e.Path)
		fs.fileList.Delete(e.Path)
	case checkFlag(e.Flags, fsevents.ItemCreated):
		logrus.Debugf("Adding %q", e.Path)
		fs.fileList.Add(e.Path)
	}
}

func checkFlag(flags, needle fsevents.EventFlags) bool {
	return flags&needle == needle
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
