package fscache

import (
	"bytes"
	"testing"

	"github.com/monochromegane/go-gitignore"
)

func TestCheckSkipPath(t *testing.T) {
	type testCase struct {
		path     string
		dir      bool
		expected bool
	}

	ignoreMatcher := gitignore.NewGitIgnoreFromReader("/", bytes.NewBufferString(watchIgnores))

	testCases := []testCase{
		{path: "/foo/bar/.git/baz", dir: false, expected: true},
		{path: "/foo/bar/.git/", dir: true, expected: true},
		{path: "/foo/bar/Library/Application Support/foo/bar", dir: false, expected: true},
		{path: "/foo/bar/Library/foo/bar", dir: false, expected: false},
	}

	for _, c := range testCases {
		out := checkSkipPath(ignoreMatcher, c.path, c.dir)

		if out != c.expected {
			t.Errorf("checkSkipPath(%q, dir:%v) = %v; want %v", c.path, c.dir, out, c.expected)
		}
	}
}
