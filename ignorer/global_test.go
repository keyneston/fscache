package ignorer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalIgnore(t *testing.T) {
	type testCase struct {
		path     string
		dir      bool
		expected bool
	}

	ignoreMatcher := NewGlobalIgnore()

	testCases := []testCase{
		{path: "/foo/bar/.git/baz", dir: false, expected: true},
		{path: "/foo/bar/.git/", dir: true, expected: true},
		{path: "/foo/bar/Library/Application Support/foo/bar", dir: false, expected: true},
		{path: "/Users/Alice/.Trash/Library/foo/bar", dir: false, expected: true},
	}

	for _, c := range testCases {
		match := ignoreMatcher.Match(c.path, c.dir)
		assert.Equal(t, c.expected, match, "GlobalIgnorer.Match(%q, %v)", c.path, c.dir)
	}
}
