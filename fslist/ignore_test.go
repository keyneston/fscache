package fslist

import (
	"testing"

	"github.com/monochromegane/go-gitignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMatcher struct {
	name string
}

func (fakeMatcher) Match(string, bool) bool {
	return true
}

func TestIgnoreCache_Get(t *testing.T) {
	ic := IgnoreCache{
		cache: map[string]gitignore.IgnoreMatcher{},
	}

	dirs := []string{"/foo/bar/baz", "/foo/bar/baz/qaz"}
	for _, d := range dirs {
		ic.cache[d] = fakeMatcher{name: d}
	}

	type testCase struct {
		input    string
		expected string
	}
	testCases := []testCase{
		{input: "/foo/bar/baz", expected: "/foo/bar/baz"},
		{input: "/foo/bar/baz/foo/foo", expected: "/foo/bar/baz"},
		{input: "/foo/bar/baz/qaz/foo", expected: "/foo/bar/baz/qaz"},
		{input: "/", expected: ""},
	}

	for _, c := range testCases {
		matcher := ic.Get(c.input)

		if c.expected == "" {
			require.Nil(t, matcher)
			continue
		}

		require.NotNil(t, matcher)
		fake := matcher.(fakeMatcher)
		assert.Equal(t, fake.name, c.expected)

	}
}
