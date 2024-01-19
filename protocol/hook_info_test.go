package protocol

import (
	"strings"
	"testing"
)

func TestReadHookInput(t *testing.T) {
	input := "e285100b636ac67fa28d85685072158edaa01685 a3d33576d686e7dc1d90ec4b1a6e94e760a893b2 refs/heads/master\n"
	info, err := ReadHookInput(strings.NewReader(input))

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if info.OldRev != "e285100b636ac67fa28d85685072158edaa01685" {
		t.Errorf("Expected OldRev to be %s, but got %s", "e285100b636ac67fa28d85685072158edaa01685", info.OldRev)
	}

	if info.NewRev != "a3d33576d686e7dc1d90ec4b1a6e94e760a893b2" {
		t.Errorf("Expected NewRev to be %s, but got %s", "a3d33576d686e7dc1d90ec4b1a6e94e760a893b2", info.NewRev)
	}

	if info.Ref != "refs/heads/master" {
		t.Errorf("Expected Ref to be %s, but got %s", "refs/heads/master", info.Ref)
	}

	if info.RefType != "heads" {
		t.Errorf("Expected RefType to be %s, but got %s", "heads", info.RefType)
	}

	if info.RefName != "master" {
		t.Errorf("Expected RefName to be %s, but got %s", "master", info.RefName)
	}

}

func TestHookAction(t *testing.T) {
	examples := map[string]HookInfo{
		"branch.create": {
			OldRev:  "0000000000000000000000000000000000000000",
			NewRev:  "e285100b636ac67fa28d85685072158edaa01685",
			RefType: "heads",
		},
		"branch.delete": {
			OldRev:  "e285100b636ac67fa28d85685072158edaa01685",
			NewRev:  "0000000000000000000000000000000000000000",
			RefType: "heads",
		},
		"branch.push": {
			OldRev:  "e285100b636ac67fa28d85685072158edaa01685",
			NewRev:  "a3d33576d686e7dc1d90ec4b1a6e94e760a893b2",
			RefType: "heads",
		},
		"tag.create": {
			OldRev:  "0000000000000000000000000000000000000000",
			NewRev:  "e285100b636ac67fa28d85685072158edaa01685",
			RefType: "tags",
		},
		"tag.delete": {
			OldRev:  "e285100b636ac67fa28d85685072158edaa01685",
			NewRev:  "0000000000000000000000000000000000000000",
			RefType: "tags",
		},
	}

	for expected, hook := range examples {
		if expected != parseHookAction(hook) {
			t.Errorf("Expected %v, got %v", expected, parseHookAction(hook))
		}
	}
}
