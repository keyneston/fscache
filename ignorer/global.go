package ignorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/keyneston/fscache/internal/shared"
	"github.com/monochromegane/go-gitignore"
)

type GlobalIgnore struct {
	matcher gitignore.IgnoreMatcher
}

func GlobalIgnoreList() io.Reader {
	watchIgnores := `
.git/
.svn/
.cvs/
node_modules/
Application Support/
.cache/
.DS_File
.Trash
.rustup/toolchain
.rustup/update-hashes
.rustup/tmp
.rustup/downloads
pkg/darwin_amd64/
pkg/darwin_arm/
`
	ignoredKeys := []string{"GOMODCACHE", "GOCACHE", "GOTOOLDIR", "GOROOT"}

	buf := bytes.NewBufferString(strings.TrimLeft(watchIgnores, " \t\n"))

	for _, key := range ignoredKeys {
		if val := os.Getenv(key); val != "" {
			fmt.Fprintln(buf, val)
		}
	}

	if gopath := os.Getenv("GOPATH"); gopath != "" {
		fmt.Fprintln(buf, filepath.Join(gopath, "pkg"))
	}

	goVars, err := getGoVars()
	if err != nil {
		shared.Logger().Error().Err(err).Msg("trying to get go vars")
		// If we get an error return what we have so far
		return buf
	}

	for _, k := range ignoredKeys {
		if v, ok := goVars[k]; ok {
			fmt.Fprintln(buf, v)
		}
	}

	if dir, err := os.UserHomeDir(); err == nil {
		fmt.Fprintln(buf, filepath.Join(dir, "Library"))
	}

	return buf
}

func getGoVars() (map[string]string, error) {
	vars := map[string]string{}

	path, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	cmd := exec.Command(path, "env", "-json")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(buf.Bytes(), &vars); err != nil {
		return nil, err
	}

	return vars, nil
}

func NewGlobalIgnore() GlobalIgnore {
	matcher := gitignore.NewGitIgnoreFromReader("/", GlobalIgnoreList())

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
