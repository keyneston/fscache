package fslist

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/require"
)

var testData = map[string]AddData{
	"/foo/bar":           AddData{Name: "/foo/bar", IsDir: true},
	"/foo/bar/baz":       AddData{Name: "/foo/bar/baz", IsDir: true},
	"/foo/bar/baz/1.txt": AddData{Name: "/foo/bar/baz/1.txt", IsDir: false},
	"/foo/bar/baz/2.txt": AddData{Name: "/foo/bar/baz/2.txt", IsDir: false},
	"/foo/bar/qaz":       AddData{Name: "/foo/bar/qaz", IsDir: false},
}

var __allTestData []AddData

func getTestData(names ...string) []AddData {
	res := []AddData{}

	for _, name := range names {
		res = append(res, testData[name])
	}
	return res
}

func getAllTestData() []AddData {
	if __allTestData != nil {
		return __allTestData
	}

	__allTestData = make([]AddData, 0, len(testData))
	for _, i := range testData {
		__allTestData = append(__allTestData, i)
	}

	sort.Sort(ByPath(__allTestData))

	return __allTestData
}

func TestPebble(t *testing.T) {
	type testCase struct {
		name     string
		testData []AddData
		input    ReadOptions
		expected []AddData
	}

	testCases := []testCase{
		{
			name:     "no_options",
			testData: getAllTestData(),
			expected: getAllTestData(),
			input:    ReadOptions{},
		},
		{
			name:     "specific item",
			testData: getAllTestData(),
			expected: getTestData("/foo/bar/qaz"),
			input:    ReadOptions{Prefix: "/foo/bar/qaz"},
		},
		{
			name:     "subtree",
			testData: getAllTestData(),
			expected: getTestData("/foo/bar/baz", "/foo/bar/baz/1.txt", "/foo/bar/baz/2.txt"),
			input:    ReadOptions{Prefix: "/foo/bar/baz"},
		},
		{
			name:     "duplicate adds",
			testData: append(getAllTestData(), getAllTestData()...),
			expected: getAllTestData(),
			input:    ReadOptions{},
		},
	}

	for _, c := range testCases {
		t.Run(fmt.Sprintf("test_pebble_fetch_%s", c.name), func(t *testing.T) {
			db, err := NewPebble()
			defer db.Close()
			defer os.RemoveAll(db.(*PebbleList).location)
			require.NoError(t, err)

			for _, d := range c.testData {
				db.Add(d)
			}

			res := []AddData{}
			for i := range db.Fetch(c.input) {
				res = append(res, i)
			}

			if diff := deep.Equal(c.expected, res); diff != nil {
				t.Errorf("db.Fetch(%#v) =\n%v", c.input, strings.Join(diff, "\n"))
			}
		})
	}
}
