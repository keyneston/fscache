package shared

import (
	"os"
	"path/filepath"

	"github.com/nightlyone/lockfile"
	"github.com/slongfield/pyfmt"
)

type PID struct {
	location string
	lock     lockfile.Lockfile
}

func NewPID(template, root, cache string) (*PID, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	location, err := pyfmt.Fmt(template, map[string]interface{}{
		"cache":     filepath.Base(cache),
		"cachepath": cache,
		"root":      root,
		"home":      home,
	})
	if err != nil {
		return nil, err
	}

	lock, err := lockfile.New(location)
	if err != nil {
		return nil, err
	}

	p := &PID{
		location: location,
		lock:     lock,
	}
	return p, nil
}

func (p *PID) Acquire() (bool, error) {
	if err := p.lock.TryLock(); err != nil {
		if err.Error() == "Locked by other process" {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (p *PID) Release() error {
	return p.lock.Unlock()
}
