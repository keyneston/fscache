package shared

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/google/subcommands"
)

type Config struct {
	PIDFile string

	cache string

	globalOnce sync.Once
}

func (c *Config) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&debug, "debug", false, "Enable verbose debug logging")
	f.StringVar(&c.PIDFile, "pid", "{home}/.cache/{cache}.pid", "Which PID file to use")
	f.StringVar(&c.cache, "cache", "", "Manually set cache, defaults to ~/.cache/fscache.sqlite")

	c.globalOnce.Do(func() {
		flag.BoolVar(&debug, "debug", false, "Enable verbose debug logging")
		flag.StringVar(&c.PIDFile, "pid", "{home}/.cache/{cache}.pid", "Which PID file to use")
		flag.StringVar(&c.cache, "cache", "", "Manually set cache, defaults to ~/.cache/fscache.sqlite")
	})
}

func (c *Config) RegisterGlobal() {
	// In order to keep all flag declarations in one place they are wrapped up
	// in the SetFlags function. IN order to register them globally we just
	// cheat and set a non-existent flag set.
	c.SetFlags(flag.NewFlagSet("", flag.ContinueOnError))
}

var DefaultCacheLocation = "${HOME}/.cache/fscache.sqlite"

func (c *Config) CacheLocation() (string, error) {
	if c.cache != "" {
		return c.cache, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return strings.Replace(DefaultCacheLocation, "${HOME}", home, -1), nil
}

func Exitf(format string, vars ...interface{}) subcommands.ExitStatus {
	if len(format) == 0 || format[len(format)-1] != '\n' {
		format = fmt.Sprintf("%s\n", format)
	}
	fmt.Fprintf(os.Stderr, format, vars...)
	return subcommands.ExitFailure
}
