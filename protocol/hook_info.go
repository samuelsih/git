package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const ZeroSHA = "0000000000000000000000000000000000000000"

var (
	errInvalidHookInput = errors.New("invalid hook input")
)

type HookInfo struct {
	Action   string
	RepoName string
	RepoPath string
	OldRev   string
	NewRev   string
	Ref      string
	RefType  string
	RefName  string
}

func ReadHookInput(input io.Reader) (HookInfo, error) {
	r := bufio.NewReader(input)
	line, _, err := r.ReadLine()
	if err != nil {
		return HookInfo{}, err
	}

	chunks := strings.Split(string(line), " ")
	if len(chunks) != 3 {
		return HookInfo{}, errInvalidHookInput
	}

	refchunks := strings.Split(chunks[2], "/")

	dir, _ := os.Getwd()
	info := HookInfo{
		RepoName: filepath.Base(dir),
		RepoPath: dir,
		OldRev:   chunks[0],
		NewRev:   chunks[1],
		Ref:      chunks[2],
		RefType:  refchunks[1],
		RefName:  refchunks[2],
	}

	info.Action = parseHookAction(info)

	return info, nil
}

func parseHookAction(h HookInfo) string {
	action, context := "push", "branch"

	if h.RefType == "tags" {
		context = "tag"
	}

	if h.OldRev == ZeroSHA && h.NewRev != ZeroSHA {
		return fmt.Sprintf("%s.create", context)
	}

	if h.OldRev != ZeroSHA && h.NewRev == ZeroSHA {
		return fmt.Sprintf("%s.delete", context)
	}

	return fmt.Sprintf("%s.%s", context, action)

}
