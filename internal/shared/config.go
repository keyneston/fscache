package shared

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/subcommands"
	"github.com/keyneston/fscache/proto"
	"google.golang.org/grpc"
)

type Config struct {
	PIDFile string

	Socket string

	globalOnce sync.Once
}

var DefaultSocketLocation = "${HOME}/.cache/fscache.socket"

func (c *Config) SetFlags(f *flag.FlagSet) {
	if f != nil {
		f.BoolVar(&debug, "debug", false, "Enable verbose debug logging")
		f.StringVar(&level, "log-level", "error", "Log level. Options: panic, fatal, error, warn, info, debug, trace")
		f.StringVar(&c.PIDFile, "pid", "{home}/.cache/{cache}.pid", "Which PID file to use")
		f.StringVar(&c.Socket, "socket", "", "Where to place the communications socket, defaults to ~/.cache/fscache.socket")
	}

	c.globalOnce.Do(func() {
		flag.BoolVar(&debug, "debug", false, "Enable verbose debug logging")
		flag.StringVar(&level, "log-level", "error", "Log level. Options: panic, fatal, error, warn, info, debug, trace")
		flag.StringVar(&c.PIDFile, "pid", "{home}/.cache/{cache}.pid", "Which PID file to use")
		flag.StringVar(&c.Socket, "socket", "", "Where to place the communications socket, defaults to ~/.cache/fscache.socket")
	})
}

func (c *Config) RegisterGlobal() {
	// In order to keep all flag declarations in one place they are wrapped up
	// in the SetFlags function. IN order to register them globally we just
	// cheat and set a non-existent flag set.
	c.SetFlags(nil)
}

func (c *Config) Client() (proto.FSCacheClient, error) {
	socket, err := c.SocketLocation()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(fmt.Sprintf("unix:%s", socket), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return proto.NewFSCacheClient(conn), nil
}

func (c *Config) SocketLocation() (string, error) {
	if c.Socket != "" {
		return c.Socket, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return strings.Replace(DefaultSocketLocation, "${HOME}", home, -1), nil
}

func Exitf(format string, vars ...interface{}) subcommands.ExitStatus {
	if len(format) == 0 || format[len(format)-1] != '\n' {
		format = fmt.Sprintf("%s\n", format)
	}
	fmt.Fprintf(os.Stderr, format, vars...)
	return subcommands.ExitFailure
}
