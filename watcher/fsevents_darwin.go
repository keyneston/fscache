// +build darwin

package watcher

import (
	"sync"
	"time"

	"github.com/fsnotify/fsevents"
	"github.com/keyneston/fscache/internal/shared"
)

func New(root string) (Watcher, error) {
	return &DarwinWatcher{
		eventStream: &fsevents.EventStream{
			Paths:   []string{root},
			Latency: time.Second,
			Flags:   fsevents.FileEvents,
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

func (d *DarwinWatcher) handleEvents(events []fsevents.Event) {
	logger := shared.Logger().WithField("module", "fsevents_darwin").Logger

	translated := []Event{}
	for _, e := range events {
		t := Event{Path: e.Path}

		switch {
		case checkBitFlag(e.Flags, fsevents.ItemRemoved):
			t.Type = EventTypeDelete
		case checkBitFlag(e.Flags, fsevents.ItemCreated):
			t.Type = EventTypeAdd
		case checkBitFlag(e.Flags, fsevents.ItemRenamed):
			t.Type = EventTypeDelete
		case checkBitFlag(e.Flags, fsevents.MustScanSubDirs):
			logger.WithField("event", e).Warn("MustScanSubDirs, skipping")
			continue
		}

		if t.Type == EventUnknown {
			logger.WithField("event", e).WithField("flags", flagsToString(e.Flags)).Warn("Unknown event type, skipping")
			continue
		}

		t.Dir = checkBitFlag(e.Flags, fsevents.ItemIsDir)
		translated = append(translated, t)
	}

	if len(translated) != 0 {
		d.stream <- translated
	}
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
