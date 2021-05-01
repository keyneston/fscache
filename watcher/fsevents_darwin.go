// +build darwin

package watcher

import (
	"sync"
	"time"

	"github.com/fsnotify/fsevents"
)

func New(root string) (Watcher, error) {
	return &DarwinWatcher{
		eventStream: &fsevents.EventStream{
			Paths:   []string{root},
			Latency: time.Second,
			Flags:   fsevents.WatchRoot | fsevents.FileEvents,
		},
		closeCh:   make(chan bool),
		closeOnce: &sync.Once{},
		stream:    make(chan []Event, 10),
	}, nil
}

type DarwinWatcher struct {
	eventStream *fsevents.EventStream
	closeCh     chan bool
	closeOnce   *sync.Once
	stream      chan []Event
}

func (d *DarwinWatcher) Start() {
	d.eventStream.Start()

	go d.run()
}

func (d *DarwinWatcher) run() {
	for {
		select {
		case events := <-d.eventStream.Events:
			d.handleEvents(events)
		case <-d.closeCh:
			return
		}
	}
}

func checkFlag(flags, needle fsevents.EventFlags) bool {
	return flags&needle == needle
}

func (d *DarwinWatcher) handleEvents(events []fsevents.Event) {
	translated := []Event{}
	for _, e := range events {
		t := Event{Path: e.Path}
		switch {
		case checkFlag(e.Flags, fsevents.ItemRemoved):
			t.Type = EventTypeAdd
		case checkFlag(e.Flags, fsevents.ItemCreated):
			t.Type = EventTypeDelete
		}

		translated = append(translated, t)
	}

	d.stream <- translated
}

func (d *DarwinWatcher) Stop() {
	d.closeOnce.Do(func() {
		d.eventStream.Stop()
		close(d.closeCh)
	})
}

func (d *DarwinWatcher) Stream() <-chan []Event {
	return d.stream
}
