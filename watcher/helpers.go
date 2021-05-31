package watcher

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsevents"
)

func checkBitFlag(flags, needle fsevents.EventFlags) bool {
	return flags&needle != 0
}

var flagMappings = map[string]fsevents.EventFlags{
	"MustScanSubDirs":   fsevents.MustScanSubDirs,
	"UserDropped":       fsevents.UserDropped,
	"KernelDropped":     fsevents.KernelDropped,
	"EventIDsWrapped":   fsevents.EventIDsWrapped,
	"HistoryDone":       fsevents.HistoryDone,
	"RootChanged":       fsevents.RootChanged,
	"Mount":             fsevents.Mount,
	"Unmount":           fsevents.Unmount,
	"ItemCreated":       fsevents.ItemCreated,
	"ItemRemoved":       fsevents.ItemRemoved,
	"ItemInodeMetaMod":  fsevents.ItemInodeMetaMod,
	"ItemRenamed":       fsevents.ItemRenamed,
	"ItemModified":      fsevents.ItemModified,
	"ItemFinderInfoMod": fsevents.ItemFinderInfoMod,
	"ItemChangeOwner":   fsevents.ItemChangeOwner,
	"ItemXattrMod":      fsevents.ItemXattrMod,
	"ItemIsFile":        fsevents.ItemIsFile,
	"ItemIsDir":         fsevents.ItemIsDir,
	"ItemIsSymlink":     fsevents.ItemIsSymlink,
}

func flagsToString(flags fsevents.EventFlags) string {
	flagStrs := []string{}

	for k, v := range flagMappings {
		if checkBitFlag(v, flags) {
			flagStrs = append(flagStrs, k)
		}
	}

	return fmt.Sprintf("fsevents.EventFlags{%s}", strings.Join(flagStrs, ", "))
}
