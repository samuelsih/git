package protocol

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/google/uuid"
)

const (
	BranchPushAction   = "branch.push"
	BranchCreateAction = "branch.create"
	BranchDeleteAction = "branch.delete"
	TagCreateAction    = "tag.create"
	TagDeleteAction    = "tag.delete"

	tempDirFilePermission = 0774
)

type Receiver struct {
	Debug       bool
	MasterOnly  bool
	TmpDir      string
	HandlerFunc func(HookInfo, string) error
}

func IsForcePush(hook HookInfo) (bool, error) {
	if hook.OldRev == ZeroSHA || hook.NewRev == ZeroSHA {
		return false, nil
	}

	out, err := exec.Command("git", "merge-base", hook.OldRev, hook.NewRev).CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("git merge base failed: %s", out)
	}

	base := strings.TrimSpace(string(out))

	// Non fast-forwarded, meaning force
	return base != hook.OldRev, nil
}

func (r *Receiver) Handle(reader io.Reader) error {
	hook, err := ReadHookInput(reader)
	if err != nil {
		return err
	}

	if r.MasterOnly && hook.Ref != "refs/heads/master" {
		return fmt.Errorf("cant push to non-master branch")
	}

	id := uuid.NewString()

	tmpDir := path.Join(r.TmpDir, id)
	if err := os.MkdirAll(tmpDir, tempDirFilePermission); err != nil {
		return err
	}

	// Cleanup temp directory unless we're in debug mode
	if !r.Debug {
		defer os.RemoveAll(tmpDir)
	}

	archiveCmd := fmt.Sprintf("git archive '%s' | tar -x -C '%s'", hook.NewRev, tmpDir)
	buff, err := exec.Command("bash", "-c", archiveCmd).CombinedOutput()
	if err != nil {
		if len(buff) > 0 && strings.Contains(string(buff), "Damaged tar archive") {
			return fmt.Errorf("Error: repository might be empty!")
		}

		return fmt.Errorf("cant archive repo: %s", buff)
	}

	if r.HandlerFunc != nil {
		return r.HandlerFunc(hook, tmpDir)
	}

	return nil
}
