package watcher

import (
	"testing"

	"github.com/fsnotify/fsevents"
	"github.com/stretchr/testify/assert"
)

func TestCheckBitFlag(t *testing.T) {
	type testCase struct {
		name     string
		flags    fsevents.EventFlags
		needle   fsevents.EventFlags
		expected bool
	}

	testCases := []testCase{
		{"dir_created", fsevents.ItemCreated | fsevents.ItemIsDir, fsevents.ItemCreated, true},
		{"dir_removed", fsevents.ItemRemoved | fsevents.ItemIsDir, fsevents.ItemRemoved, true},
		{"dir_created_not_removed", fsevents.ItemCreated | fsevents.ItemIsDir, fsevents.ItemRemoved, false},
		{"dir_removed_not_created ", fsevents.ItemRemoved | fsevents.ItemIsDir, fsevents.ItemCreated, false},
	}

	for _, c := range testCases {
		out := checkBitFlag(c.flags, c.needle)
		assert.Equal(t, c.expected, out, c.name)
	}
}
