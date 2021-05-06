package fslist

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"
)

type empty struct{}

var _ FSList = &List{}

type List struct {
	mutex    *sync.Mutex
	list     map[string]empty
	pending  int64
	filename string
}

func NewList(filename string) (*List, error) {
	return &List{
		mutex:    &sync.Mutex{},
		list:     map[string]empty{},
		filename: filename,
	}, nil
}

func (fs *List) Pending() bool {
	return atomic.LoadInt64(&fs.pending) == 1
}

func (fs *List) Add(name string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.setPending()

	fs.list[name] = empty{}

	return nil
}

func (fs *List) setPending() {
	atomic.StoreInt64(&fs.pending, 1)
}

func (fs *List) clearPending() {
	atomic.StoreInt64(&fs.pending, 0)
}

func (fs *List) Delete(name string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.setPending()
	delete(fs.list, name)

	return nil
}

func (fs *List) Write() error {
	f, err := os.OpenFile(fs.filename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	fs.mutex.Lock()
	fs.clearPending()

	l := make([]string, 0, len(fs.list))
	for k := range fs.list {
		l = append(l, k)
	}

	fs.mutex.Unlock()
	sort.Strings(l)

	for _, file := range l {
		_, err := fmt.Fprintln(f, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *List) Len() int {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	return len(fs.list)
}
