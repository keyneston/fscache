package run

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/google/subcommands"
	daemon "github.com/sevlyar/go-daemon"

	"github.com/keyneston/fscache/fscache"
	"github.com/keyneston/fscache/fslist"
	"github.com/keyneston/fscache/internal/shared"
)

type Command struct {
	*shared.Config

	root      string
	mode      string
	daemonize bool
}

func (*Command) Name() string     { return "run" }
func (*Command) Synopsis() string { return "Run fscache" }
func (*Command) Usage() string {
	return `run:
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	c.Config.SetFlags(f)

	f.StringVar(&c.root, "r", "", "Root directory to monitor")
	f.StringVar(&c.root, "root", "", "Alias for -r")
	f.StringVar(&c.mode, "mode", "pebble", "DB mode; experimental")
	f.BoolVar(&c.daemonize, "daemonize", false, "Launch as a daemon")
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var err error

	if c.root == "" {
		c.root, err = os.UserHomeDir()
		if err != nil {
			return shared.Exitf("Unable to get root location: %v", err)
		}
	}

	socketLoc, err := c.SocketLocation()
	if err != nil {
		return shared.Exitf("Unable to get socket location: %v", err)
	}

	if c.daemonize {
		daemonCtx := daemon.Context{}
		child, _ := daemonCtx.Reborn()
		if child != nil {
			return subcommands.ExitSuccess
		}
	}

	pid, err := shared.NewPID(c.PIDFile, c.root, socketLoc)
	if err != nil {
		return shared.Exitf("Error creating pid file: %v", err)
	}

	if ok, err := pid.Acquire(); err != nil {
		return shared.Exitf("Error starting monitor: %v", err)
	} else if !ok {
		fmt.Fprintf(os.Stdout, "fscache is already running\n")
		return subcommands.ExitSuccess
	}

	// Cleanup old socket (if one exists)
	if _, err := os.Stat(socketLoc); err == nil {
		if err := os.Remove(socketLoc); err != nil {
			return shared.Exitf("Error cleaning old socket: %v", err)
		}
	}

	fs, err := fscache.New(socketLoc, c.root, fslist.Mode(c.mode))
	if err != nil {
		return shared.Exitf("Error starting monitor: %v", err)
	}

	if shouldRestart := fs.Run(); shouldRestart {
		// Restart will exec and cause no return value if restart is successful
		return shared.Exitf("Error restarting: %v", restart())

	}

	return subcommands.ExitSuccess
}

func restart() error {
	bin, err := exec.LookPath(os.Args[0])
	if err != nil {
		return fmt.Errorf("error locating %q: %w", os.Args[0], err)
	}

	shared.Logger().Debug().Str("bin", bin).Strs("args", os.Args).Msg("restarting")
	return syscall.Exec(bin, os.Args, os.Environ())
}
