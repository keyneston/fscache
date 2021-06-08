package listignores

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/google/subcommands"
	"github.com/keyneston/fscache/fscache"
	"github.com/keyneston/fscache/internal/shared"
	"github.com/rs/zerolog"
)

type Command struct {
	*shared.Config
	logger zerolog.Logger
}

func (*Command) Name() string     { return "list-ignores" }
func (*Command) Synopsis() string { return "list global ignores" }
func (*Command) Usage() string {
	return `list-ignores:
`
}

func (c *Command) SetFlags(f *flag.FlagSet) {
	c.Config.SetFlags(f)
}

func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	io.Copy(os.Stdout, fscache.GlobalIgnoreList())
	fmt.Fprintln(os.Stdout, "")

	return subcommands.ExitSuccess
}
