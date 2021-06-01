package watcher

import (
	"github.com/fsnotify/fsevents"
)

func checkBitFlag(flags, needle fsevents.EventFlags) bool {
	return flags&needle != 0
}
