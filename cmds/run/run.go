package run

import (
	"context"
	"flag"
	"log"

	"github.com/google/subcommands"
	"github.com/keyneston/fscachemonitor/fscache"
	"github.com/keyneston/fscachemonitor/internal/shared"
)

type Command struct {
	*shared.Config

	root     string
	filename string
	sql      bool
}

func (*Command) Name() string     { return "run" }
func (*Command) Synopsis() string { return "Run fscachemonitor" }
func (*Command) Usage() string {
	return `run:
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	c.Config.SetFlags(f)

	f.StringVar(&c.root, "r", "", "Root directory to monitor")
	f.StringVar(&c.root, "root", "", "Alias for -r")
	f.StringVar(&c.filename, "c", "", "File to output cache to")
	f.StringVar(&c.filename, "cache", "", "Alias for -c")
	f.BoolVar(&c.sql, "s", false, "Use SQLite3 backed file")
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if c.root == "" {
		return shared.Exitf("Must specify root to watch")
	}

	if c.filename == "" {
		return shared.Exitf("Must specify file to output cache to")
	}

	fs, err := fscache.New(c.filename, c.root, c.sql)
	if err != nil {
		log.Fatalf("Error starting monitor: %v", err)
	}
	fs.Run()

	return subcommands.ExitSuccess
}
