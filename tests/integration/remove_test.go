package integration

import (
	"context"
	"io"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/keyneston/fscache/fslist"
	"github.com/keyneston/fscache/proto"
)

func TestRemoveFile(t *testing.T) {
	i := New(t, "integration-ignore")

	go i.cache.Run()
	defer i.CleanUp()

	fooTXT := i.createFile("foo.txt").done()
	barTXT := i.createFile("bar.txt").done()

	time.Sleep(1 * time.Second)

	i.require.NoError(os.Remove(fooTXT), "removing file")

	time.Sleep(2 * time.Second)

	stream, err := i.client.GetFiles(context.Background(), &proto.ListRequest{FilesOnly: true})
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
		{Name: barTXT, IsDir: false},
		//{Name: i.testDir, IsDir: true},
	}

	sort.Sort(fslist.ByPath(res))
	sort.Sort(fslist.ByPath(expected))

	i.assert.Len(res, len(expected))
	i.assert.ElementsMatch(expected, res)
}
