package protocol

import (
	"io/fs"
	"os"
)

var (
	filePermission = 0755
)

// All config available
type Config struct {
	Dir        string
	GitPath    string
	AutoCreate bool
	Auth       bool
}

// Base setup for the config.
//
// This will create the directory that contains repositories if not exist.
func (c *Config) Setup() error {
	if _, err := os.Stat(c.Dir); err != nil {
		if err = os.Mkdir(c.Dir, fs.FileMode(filePermission)); err != nil {
			return err
		}
	}

	return nil
}
