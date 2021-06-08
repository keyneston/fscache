package fscache

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monochromegane/go-gitignore"
)

type GlobalIgnore struct {
	matcher gitignore.IgnoreMatcher
}

func GlobalIgnoreList() io.Reader {
	watchIgnores := `
.git/
.svn/
Application Support/
.cache/
.DS_File
`
	buf := bytes.NewBufferString(strings.TrimSpace(watchIgnores))

	for _, key := range []string{"GOMODCACHE", "GOCACHE", "GOTOOLDIR", "GOROOT"} {
		if val := os.Getenv(key); val != "" {
			fmt.Println(buf, val)
		}
	}

	return buf
}

func NewGlobalIgnore(root string) GlobalIgnore {
	matcher := gitignore.NewGitIgnoreFromReader(root, GlobalIgnoreList())

	return GlobalIgnore{
		matcher: matcher,
	}
}

func (g GlobalIgnore) Match(path string, dir bool) bool {
	// TODO: find a better way:
	segments := strings.Split(path, "/")
	for i := range segments {
		if i != 0 {
			dir = true
		}

		path = strings.Join(segments[0:i], "/")
		if g.matcher.Match(path, dir) {
			return true
		}
	}

	return false
}
