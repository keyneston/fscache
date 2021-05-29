package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/keyneston/fscache/fscache"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/keyneston/fscache/proto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	shared.SetLevel(logrus.TraceLevel)
	shared.SetPrettyLogging()
}

type integration struct {
	t       *testing.T
	require *require.Assertions
	assert  *assert.Assertions
	cache   *fscache.FSCache
	client  proto.FSCacheClient

	name      string
	testDir   string
	socketLoc string
}

func New(t *testing.T, name string) *integration {
	require := require.New(t)
	assert := assert.New(t)

	tmp, err := os.MkdirTemp("", fmt.Sprintf("%s-*", name))
	require.NoError(err, "Error creating workdir")

	socketLoc := filepath.Join(tmp, "socket")
	testDir := filepath.Join(tmp, "test")
	cache, err := fscache.New(
		socketLoc,
		testDir,
		"pebble",
	)
	require.NoError(err, "Error creating fscache")

	client, err := (&shared.Config{Socket: socketLoc}).Client()
	require.NoError(err, "Error creating client")

	return &integration{
		t:         t,
		name:      name,
		require:   require,
		assert:    assert,
		cache:     cache,
		testDir:   testDir,
		socketLoc: socketLoc,
		client:    client,
	}
}

func (i *integration) createFile(pathSegments ...string) createFile {
	return createFile{
		path: filepath.Join(append([]string{i.testDir}, pathSegments...)...),
		t:    i.t,
	}
}

type createFile struct {
	t        *testing.T
	path     string
	contents []string
}

func (c createFile) with(lines ...string) createFile {
	c.contents = append(c.contents, lines...)
	return c
}

func (c createFile) done() {
	dir := filepath.Dir(c.path)
	err := os.MkdirAll(dir, 0755)
	require.NoError(c.t, err, "Unable to create dir: %q", dir)

	f, err := os.OpenFile(c.path, os.O_CREATE|os.O_RDWR, 0644)
	require.NoError(c.t, err, "Unable to create file %q", c.path)
	defer f.Close()

	for _, l := range c.contents {
		fmt.Fprintln(f, l)
	}
}
