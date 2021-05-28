package read

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/subcommands"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/keyneston/fscache/proto"
	"github.com/sirupsen/logrus"
)

type Command struct {
	*shared.Config
	logger *logrus.Logger

	dirsOnly bool
	prefix   string
	mode     string
	root     bool

	limit     int
	batchSize int
}

func (*Command) Name() string     { return "read" }
func (*Command) Synopsis() string { return "read from cache and return entries" }
func (*Command) Usage() string {
	return `read:
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	c.Config.SetFlags(f)

	f.StringVar(&c.prefix, "p", "", "Prefix to limit paths returned")
	f.BoolVar(&c.root, "r", false, "Auto discover root")
	f.StringVar(&c.prefix, "prefix", "", "Alias for -p")
	f.StringVar(&c.mode, "mode", "sql", "DB mode; experimental")
	f.IntVar(&c.limit, "n", 0, "Number of items to return. 0 for all")
	f.IntVar(&c.batchSize, "b", 1000, "Number of items to return per batch")
	f.BoolVar(&c.dirsOnly, "d", false, "Only return directories")
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c.logger = shared.Logger().WithField("command", "read").Logger

	cwd, err := os.Getwd()
	if err != nil {
		return shared.Exitf("Error finding cwd: %v", err)
	}

	if c.root && c.prefix == "" {
		root, err := findRoot()
		if err != nil {
			return shared.Exitf("Error finding root: %v", err)
		}

		c.prefix = root
	}

	client, err := c.Client()
	if err != nil {
		return shared.Exitf("Error connecting to fscache: %v", err)
	}

	c.prefix = cleanPrefix(c.prefix)

	stream, err := client.GetFiles(context.Background(), &proto.ListRequest{
		Prefix:     c.prefix,
		Limit:      int32(c.limit),
		BatchSize:  int32(c.batchSize),
		DirsOnly:   c.dirsOnly,
		CurrentDir: cleanPrefix(cwd),
	})
	if err != nil {
		return shared.Exitf("Error fetching results: %v", err)
	}

	c.logger.Debugf("Got stream")
	for {
		files, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return shared.Exitf("Error fetching all results: %v", err)
		}

		if files == nil {
			continue
		}
		for _, file := range files.Files {
			name := file.Name
			name, err = filepath.Rel(cwd, name)
			if err != nil {
				return shared.Exitf("Error getting relative path: %v", err)
			}

			os.Stdout.WriteString(name)
			os.Stdout.Write([]byte{'\n'})
		}
	}

	c.logger.WithError(err).Debugf("Done")

	return subcommands.ExitSuccess
}

// cleanPrefix adds a trailing '/' to a prefix if it is set and it doesn't have
// one
func cleanPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}

	if prefix[len(prefix)-1] != '/' {
		prefix = fmt.Sprintf("%s/", prefix)
	}
	return prefix
}

var roots = map[string]bool{
	".git": true,
	".svn": true,
	".hg":  true,
}

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if dir == "/" {
			return dir, nil
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			return "", err
		}

		for _, f := range files {
			name := f.Name()

			// Only check dot files. These (should?!?) come before non dot
			// files, so abort as soon as we hit non dot files
			if name[0] != '.' {
				break
			}
			if _, ok := roots[name]; ok {
				return dir, nil
			}
		}

		dir = filepath.Dir(dir)
	}

	return "/", nil
}
