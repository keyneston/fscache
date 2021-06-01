package watcher

import "github.com/fsnotify/fsevents"

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

func flagsToStrings(flags fsevents.EventFlags) []string {
	flagStrs := []string{}

	for k, v := range flagMappings {
		if checkBitFlag(v, flags) {
			flagStrs = append(flagStrs, k)
		}
	}

	return flagStrs
}
