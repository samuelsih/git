package protocol

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHookScripts_Setup(t *testing.T) {
	workdir, err := os.Getwd()
	if err != nil {
		t.Fatal("Cannot read workdir,", err)
	}

	dirname := filepath.Join(workdir, "test", "hooks")

	err = os.MkdirAll(dirname, 0755)
	if err != nil {
		t.Fatal("cannot create dir,", err)
	}

	defer os.RemoveAll("test")

	hookConfig := HookScripts{
		PreReceive:  "#!/bin/bash\necho 'Pre-receive hook'",
		Update:      "#!/bin/bash\necho 'Update hook'",
		PostReceive: "#!/bin/bash\necho 'Post-receive hook'",
	}

	err = hookConfig.createHooksDir("test")
	if err != nil {
		t.Fatal(err)
	}

	expectedScripts := map[string]string{
		"pre-receive":  "#!/bin/bash\necho 'Pre-receive hook'",
		"update":       "#!/bin/bash\necho 'Update hook'",
		"post-receive": "#!/bin/bash\necho 'Post-receive hook'",
	}

	for name, expectedScript := range expectedScripts {
		filePath := filepath.Join(workdir, "test", "hooks", name)
		actualScript, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Failed to read file %s: %v", filePath, err)
		}

		if string(actualScript) != expectedScript {
			t.Errorf("Mismatched content in file %s.\nExpected:\n%s\nActual:\n%s",
				filePath, expectedScript, actualScript)
		}
	}
}

