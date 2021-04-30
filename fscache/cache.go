package fscache

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type FSCache struct {
	Filename string
	Root     string

	fileList *FSList
	watcher  *fsnotify.Watcher
	watches  int64
	limit    int

	closeCh   chan interface{}
	closeOnce sync.Once
}

func New(filename, root string) (*FSCache, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FSCache{
		Filename: filename,
		Root:     root,
		watcher:  watcher,
		fileList: NewFSList(),
		limit:    40000,
	}, nil
}

func (fs *FSCache) Run() {
	fs.init()

	for {
		select {
		case event := <-fs.watcher.Events:
			fs.handleEvent(event)
		case err := <-fs.watcher.Errors:
			log.Printf("Error from watcher: %v", err)
		case <-fs.closeCh:
			return
		}
	}
}

func (fs *FSCache) handleEvent(e fsnotify.Event) {
	switch e.Op {
	case fsnotify.Create:
		fs.fileList.Add(e.Name)
	case fsnotify.Write:
		// pass
	case fsnotify.Remove:
		fs.fileList.Delete(e.Name)
	case fsnotify.Rename:
		fs.fileList.Delete(e.Name)
	case fsnotify.Chmod:
		// pass
	}
}

func (fs *FSCache) Close() {
	fs.closeOnce.Do(func() {
		fs.watcher.Close()
		close(fs.closeCh)
	})
}

// init does the initial setup of walking
func (fs *FSCache) init() {
	filepath.WalkDir(fs.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error during init (%d/%d): %v",
				atomic.LoadInt64(&fs.watches), fs.limit, err)
		}

		if skipFile(path) {
			log.Printf("Skipping %q", path)
			return filepath.SkipDir
		}

		fs.fileList.Add(path)
		if d.IsDir() {
			atomic.AddInt64(&fs.watches, 1)
			fs.watcher.Add(path)
		}

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
