package stop

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/keyneston/fscache/proto"
	"github.com/rs/zerolog"
)

type Command struct {
	*shared.Config
	logger zerolog.Logger

	restart bool
}

func (*Command) Name() string     { return "stop" }
func (*Command) Synopsis() string { return "Stop running fscache" }
func (*Command) Usage() string {
	return `stop:
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	c.Config.SetFlags(f)
	f.BoolVar(&c.restart, "r", false, "restart")
	f.BoolVar(&c.restart, "restart", false, "restart")
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c.logger = shared.Logger().With().Str("command", "shutdown").Logger()

	client, err := c.Client()
	if err != nil {
		return shared.Exitf("Error connecting to fscache: %v", err)
	}

	if _, err := client.Shutdown(context.Background(), &proto.ShutdownRequest{Restart: c.restart}); err != nil {
		return shared.Exitf("Error shutting down: %v", err)
	}

	return subcommands.ExitSuccess
}
