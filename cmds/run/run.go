package run

import (
	"context"
	"flag"
	"fmt"
	"os"

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
	f.BoolVar(&c.sql, "s", true, "Use SQLite3 backed file")
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var err error

	if c.root == "" {
		c.root, err = os.UserHomeDir()
		if err != nil {
			return shared.Exitf("Unable to get root location: %v", err)
		}
	}

	c.root, err = c.CacheLocation()
	if err != nil {
		return shared.Exitf("Unable to get cache location: %v", err)
	}

	pid, err := shared.NewPID(c.PIDFile, c.root, c.filename)
	if err != nil {
		return shared.Exitf("Error creating pid file: %v", err)
	}

	if ok, err := pid.Acquire(); err != nil {
		return shared.Exitf("Error starting monitor: %v", err)
	} else if !ok {
		fmt.Fprintf(os.Stdout, "fscachemonitor is already running\n")
		return subcommands.ExitSuccess
	}

	fs, err := fscache.New(c.filename, c.root)
	if err != nil {
		return shared.Exitf("Error starting monitor: %v", err)
	}
	fs.Run()

	return subcommands.ExitSuccess
}
