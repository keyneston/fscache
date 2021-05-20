package read

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"github.com/keyneston/fscachemonitor/fslist"
	"github.com/keyneston/fscachemonitor/internal/shared"
)

type Command struct {
	*shared.Config

	dirOnly bool
	prefix  string

	limit int
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
	f.StringVar(&c.prefix, "prefix", "", "Alias for -p")
	f.IntVar(&c.limit, "n", 0, "Number of items to return. 0 for all")
	f.BoolVar(&c.dirOnly, "d", false, "Only return directories")
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	logger := shared.Logger()

	cache, err := c.CacheLocation()
	if err != nil {
		return shared.Exitf("Error getting database file: %v", err)
	}

	logger.Debugf("About to open")
	list, err := fslist.Open(cache, fslist.ModeSQL)
	if err != nil {
		return shared.Exitf("Error opening database: %v", err)
	}

	logger.Debugf("About to copy")
	if err := list.Copy(os.Stdout, fslist.ReadOptions{
		Limit:    c.limit,
		DirsOnly: c.dirOnly,
		Prefix:   c.prefix,
	}); err != nil {
		return shared.Exitf("Error reading database: %v", err)
	}
	logger.Debugf("Finished copying")

	return subcommands.ExitSuccess
}
