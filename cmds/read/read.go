package read

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/subcommands"
	"github.com/keyneston/fscachemonitor/fslist"
	"github.com/keyneston/fscachemonitor/internal/shared"
)

type Command struct {
	*shared.Config

	filename string
	sql      bool
	dirOnly  bool
	prefix   string

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

	f.StringVar(&c.filename, "c", "", "Cache file")
	f.StringVar(&c.filename, "cache", "", "Alias for -c")
	f.StringVar(&c.prefix, "p", "", "Prefix to limit paths returned")
	f.StringVar(&c.prefix, "prefix", "", "Alias for -p")
	f.IntVar(&c.limit, "n", 0, "Number of items to return. 0 for all")
	f.BoolVar(&c.dirOnly, "d", false, "Only return directories")
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if c.filename == "" {
		return shared.Exitf("Must specify file to read from")
	}

	list, err := fslist.Open(c.filename, fslist.ModeSQL)
	if err != nil {
		log.Fatalf("Error starting monitor: %v", err)
	}

	if err := list.Copy(os.Stdout, fslist.ReadOptions{
		Limit:    c.limit,
		DirsOnly: c.dirOnly,
		Prefix:   c.prefix,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
