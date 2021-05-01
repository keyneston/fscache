package main

import (
	"context"
	"flag"
	"log"
	"os"
	"syscall"

	"github.com/google/subcommands"
	"github.com/keyneston/fscachemonitor/cmds/run"
	"github.com/keyneston/fscachemonitor/internal/shared"
)

func main() {
	sharedConf := &shared.Config{}
	sharedConf.RegisterGlobal()

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&run.Command{Config: sharedConf}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))

	//if err := setLimits(); err != nil {
	//	log.Fatalf("Error updating limits: %v", err)
	//}

}

func setLimits() error {
	var limit syscall.Rlimit

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}

	limit.Cur = 9999999
	limit.Max = 9999999

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		return err
	}

	log.Printf("Limits: %v", limit)

	return nil
}
