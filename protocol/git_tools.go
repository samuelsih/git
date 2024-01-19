package protocol

import (
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// initRepo will init the repo based on the git path given.
// It will also create hooks directory if auto hooks is enabled.
func initRepo(name string, config *Config) error {
	fullPath := path.Join(config.Dir, name)

	if err := exec.Command(config.GitPath, "init", "--bare", fullPath).Run(); err != nil {
		return err
	}

	return nil
}

// repoExists checks if repo exists
func repoExists(p string) bool {
	_, err := os.Stat(path.Join(p, "objects"))
	return err == nil
}

// gitCommand will execute the git command and return the output.
// This will throw error on windows, so run this on linux/wsl.
func gitCommand(name string, args ...string) (*exec.Cmd, io.ReadCloser) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Env = os.Environ()

	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	return cmd, r
}
