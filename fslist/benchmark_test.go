package fslist

import (
	"fmt"
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
		b.Run(fmt.Sprintf("%s_add", mode), func(b *testing.B) {
			list, err := New(mode)
			if err != nil {
				b.Fatalf("Error creating fslist: %v", err)
				return
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				list.Add(testData[0])
			}
		})

	}
}
