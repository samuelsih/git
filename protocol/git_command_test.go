package protocol

import "testing"

func TestGitCommand(t *testing.T) {
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		testcases := map[string]GitCommand{
			"git-upload-pack 'hello.git'":        {"git-upload-pack", "hello.git", "git-upload-pack 'hello.git'"},
			"git upload-pack 'hello.git'":        {"git upload-pack", "hello.git", "git upload-pack 'hello.git'"},
			"git-upload-pack '/hello.git'":       {"git-upload-pack", "hello.git", "git-upload-pack 'hello.git'"},
			"git-upload-pack '/hello/world.git'": {"git-upload-pack", "hello/world.git", "git-upload-pack 'hello.git'"},
			"git-receive-pack 'hello.git'":       {"git-receive-pack", "hello.git", "git-receive-pack 'hello.git'"},
			"git receive-pack 'hello.git'":       {"git receive-pack", "hello.git", "git receive-pack 'hello.git'"},
			"git-upload-archive 'hello.git'":     {"git-upload-archive", "hello.git", "git-upload-archive 'hello.git'"},
			"git upload-archive 'hello.git'":     {"git upload-archive", "hello.git", "git upload-archive 'hello.git'"},
		}

		for args, expected := range testcases {
			cmd, err := ParseGitCommand(args)
			if err != nil {
				t.Errorf("Error: %#v", err)
			}

			if expected.Command != cmd.Command {
				t.Errorf("Command: Expected %s, got %s", expected.Command, cmd.Command)
			}

			if expected.Repo != cmd.Repo {
				t.Errorf("Repo: Expected %s, got %s", expected.Repo, cmd.Repo)
			}
		}
	})

	t.Run("should error", func(t *testing.T) {
		cmd, err := ParseGitCommand("git i-dont-know")
		if err == nil {
			t.Error("Should throw error, got nil for i-dont-know")
		}

		if cmd != (GitCommand{}) {
			t.Error("cmd should return empty value")
		}
	})
}
