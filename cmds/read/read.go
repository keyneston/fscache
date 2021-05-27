package read

import (
	"context"
	"errors"
	"flag"
	"io"
	"os"

	"github.com/google/subcommands"
	"github.com/keyneston/fscachemonitor/internal/shared"
	"github.com/keyneston/fscachemonitor/proto"
	"github.com/sirupsen/logrus"
)

type Command struct {
	*shared.Config

	dirOnly bool
	prefix  string
	mode    string
	logger  *logrus.Logger

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

	f.StringVar(&c.prefix, "p", "", "Prefix to limit paths returned")
	f.StringVar(&c.prefix, "prefix", "", "Alias for -p")
	f.StringVar(&c.mode, "mode", "sql", "DB mode; experimental")
	f.IntVar(&c.limit, "n", 0, "Number of items to return. 0 for all")
	f.BoolVar(&c.dirOnly, "d", false, "Only return directories")
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	c.logger = shared.Logger().WithField("command", "read").Logger

	client, err := c.Client()
	if err != nil {
		return shared.Exitf("Error connecting to fscachemonitor: %v", err)
	}

	stream, err := client.GetFiles(context.Background(), &proto.ListRequest{
		Prefix: c.prefix,
		Limit:  int32(c.limit),
	})
	if err != nil {
		return shared.Exitf("Error fetching results: %v", err)
	}

	c.logger.Debugf("Got stream")
	for {
		file, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return shared.Exitf("Error fetching all results: %v", err)
		}

		if file == nil {
			continue
		}
		os.Stdout.WriteString(file.Name)
		os.Stdout.Write([]byte{'\n'})
	}

	c.logger.WithError(err).Debugf("Done")

	return subcommands.ExitSuccess
}
