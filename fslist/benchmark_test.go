package fslist

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkAdd(b *testing.B) {
	testData := []AddData{
		{Name: "/foo/bar/baz", IsDir: true, UpdatedAt: time.Now()},
	}

	for _, mode := range []Mode{
		ModeSQL,
		ModePebble,
	} {
		tmp, err := os.MkdirTemp("", "fscachemonitor-benchmark-*")
		if err != nil {
			b.Errorf("Error creating tmpdir: %v", err)
			continue
		}

		list, err := New(filepath.Join(tmp, "db"), mode)
		b.Run(fmt.Sprintf("%s_add", mode), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				list.Add(testData[0])
			}
		})

	}
}
