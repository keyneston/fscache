package integration

import (
	"context"
	"io"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/keyneston/fscache/fslist"
	"github.com/keyneston/fscache/proto"
)

func TestIgnoreEndToEnd(t *testing.T) {
	i := New(t, "integration-ignore")

	go i.cache.Run()
	defer i.CleanUp()

	i.createFile(".gitignore").with("*.ignored").done()
	i.createFile("foo.ignored").done()
	i.createFile("bar.ignored").done()
	i.createFile("bar.not").done()

	time.Sleep(2 * time.Second)

	stream, err := i.client.GetFiles(context.Background(), &proto.ListRequest{})
	i.require.NoError(err, "Error getting files")

	res := []fslist.AddData{}
	for {
		files, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Error receiving files: %v", err)
		}

		for _, f := range files.Files {
			res = append(res, fslist.AddDataFromProtoFile(f))
		}
	}

	expected := []fslist.AddData{
		{Name: filepath.Join(i.testDir, ".gitignore"), IsDir: false},
		{Name: filepath.Join(i.testDir, "bar.not-ignored"), IsDir: false},
		{Name: i.testDir, IsDir: true},
	}

	sort.Sort(fslist.ByPath(res))
	sort.Sort(fslist.ByPath(expected))

	i.assert.Len(res, 3)
	i.assert.ElementsMatch(expected, res)
}
