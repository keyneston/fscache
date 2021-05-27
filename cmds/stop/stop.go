package stop

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Command struct {
	*shared.Config
	logger *logrus.Logger
}

func (*Command) Name() string     { return "stop" }
func (*Command) Synopsis() string { return "Stop running fscache" }
func (*Command) Usage() string {
	return `stop:
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	c.Config.SetFlags(f)
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c.logger = shared.Logger().WithField("command", "shutdown").Logger

	client, err := c.Client()
	if err != nil {
		return shared.Exitf("Error connecting to fscache: %v", err)
	}

	if _, err := client.Shutdown(context.Background(), &emptypb.Empty{}); err != nil {
		return shared.Exitf("Error shutting down: %v", err)
	}

	return subcommands.ExitSuccess
}
