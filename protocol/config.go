package protocol

import (
	"io/fs"
	"os"
	"path/filepath"
)

var (
	filePermission = 0755
)

// Config for server side git hooks
type HookScripts struct {
	PreReceive  string
	Update      string
	PostReceive string
}

// All config available
type Config struct {
	Hooks      HookScripts
	Dir        string
	GitPath    string
	AutoCreate bool
	AutoHooks  bool
	Auth       bool
}

// Configure hook scripts in the repo base directory.
//
// This setup will create directory in path parameters with the hooks.
func (c *HookScripts) createHooksDir(path string) error {
	basePath := filepath.Join(path, "hooks")
	scripts := map[string]string{
		"pre-receive":  c.PreReceive,
		"update":       c.Update,
		"post-receive": c.PostReceive,
	}

	hookFiles, err := os.ReadDir(basePath)
	if err == nil {
		for _, file := range hookFiles {
			dir := filepath.Join(basePath, file.Name())
			if err := os.Remove(dir); err != nil {
				return err
			}
		}
	}

	for name, script := range scripts {
		fullPath := filepath.Join(basePath, name)
		if script == "" {
			continue
		}

		if err := os.WriteFile(
			fullPath,
			[]byte(script),
			fs.FileMode(filePermission),
		); err != nil {
			return err
		}
	}

	return nil
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

	if c.AutoHooks {
		return c.setupHooks()
	}

	return nil
}

// Setup hooks for hook script. Return err if repo dir is not exist.
//
// This function will call createHooksDir in HookScripts.
func (c *Config) setupHooks() error {
	files, err := os.ReadDir(c.Dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		path := filepath.Join(c.Dir, file.Name())

		if err := c.Hooks.createHooksDir(path); err != nil {
			return err
		}
	}

	return nil
}
