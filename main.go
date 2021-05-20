package main

import (
	"context"
	"flag"
	"os"

	"github.com/google/subcommands"
	"github.com/keyneston/fscachemonitor/cmds/read"
	"github.com/keyneston/fscachemonitor/cmds/run"
	"github.com/keyneston/fscachemonitor/cmds/stop"
	"github.com/keyneston/fscachemonitor/internal/shared"
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

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
