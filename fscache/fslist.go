package fscache

import (
	"fmt"
	"io"
	"sort"
	"sync"
)

type empty struct{}

type FSList struct {
	mutex *sync.Mutex
	list  map[string]empty
}

func NewFSList() *FSList {
	return &FSList{
		mutex: &sync.Mutex{},
		list:  map[string]empty{},
	}
}

func (fs *FSList) Add(name string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	fs.list[name] = empty{}
}

func (fs *FSList) Delete(name string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	delete(fs.list, name)
}

func (fs *FSList) Write(w io.Writer) (int, error) {
	fs.mutex.Lock()

	l := make([]string, 0, len(fs.list))
	for k := range fs.list {
		l = append(l, k)
	}

	fs.mutex.Unlock()

	sort.Strings(l)

	sum := 0
	for _, file := range l {
		c, err := fmt.Println(w, file)
		sum += c
		if err != nil {
			return sum, err
		}
	}

	return sum, nil
}

func (fs *FSList) Len() int {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	return len(fs.list)
}
