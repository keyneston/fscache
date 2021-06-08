package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	listignores "github.com/keyneston/fscache/cmds/list-ignores"
	"github.com/keyneston/fscache/cmds/read"
	"github.com/keyneston/fscache/cmds/run"
	"github.com/keyneston/fscache/cmds/stop"
	"github.com/keyneston/fscache/internal/shared"
)

func main() {
	sharedConf := &shared.Config{}
	sharedConf.RegisterGlobal()

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&run.Command{Config: sharedConf}, "")
	subcommands.Register(&read.Command{Config: sharedConf}, "")
	subcommands.Register(&stop.Command{Config: sharedConf}, "")
	subcommands.Register(&listignores.Command{Config: sharedConf}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
