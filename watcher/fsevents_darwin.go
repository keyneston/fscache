// +build darwin

package watcher

import (
	"os"
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
	logger := shared.Logger().With().Str("module", "fsevents_darwin").Logger()

	logger.Trace().Interface("events", events).Msg("got events")
	translated := []Event{}
	for _, e := range events {
		t := Event{Path: e.Path}

		switch {
		// ItemRenamed needs to be first, that way we can check if the file
		// exists or not and use that for determaining whether it was
		// deleted or created.
		case checkBitFlag(e.Flags, fsevents.ItemRenamed):
			logger.Trace().Interface("event", e).Strs("flags", flagsToStrings(e.Flags)).Msg("ItemRenamed")
			_, err := os.Stat(e.Path)
			if err != nil {
				t.Type = EventTypeDelete
			} else {
				t.Type = EventTypeAdd
			}
		case checkBitFlag(e.Flags, fsevents.ItemRemoved):
			logger.Trace().Interface("event", e).Strs("flags", flagsToStrings(e.Flags)).Msg("ItemRemoved")
			t.Type = EventTypeDelete
		case checkBitFlag(e.Flags, fsevents.ItemCreated):
			logger.Trace().Interface("event", e).Strs("flags", flagsToStrings(e.Flags)).Msg("ItemCreated")
			t.Type = EventTypeAdd
		case checkBitFlag(e.Flags, fsevents.MustScanSubDirs):
			logger.Warn().Interface("event", e).Msg("MustScanSubDirs, skipping")
			continue
		}

		if t.Type == EventUnknown {
			logger.Warn().Interface("event", e).Strs("flags", flagsToStrings(e.Flags)).Msg("Unknown event type, skipping")
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
