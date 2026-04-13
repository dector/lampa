package proxy

import "testing"

func TestCreateCliCommand_IncludesLogsGetAll(t *testing.T) {
	cmd := CreateCliCommand()

	if got, want := cmd.Name, "proxy"; got != want {
		t.Fatalf("unexpected command name: got %q, want %q", got, want)
	}

	var logsCmdFound bool
	var getAllCmdFound bool
	for _, sub := range cmd.Commands {
		if sub.Name != "logs" {
			continue
		}
		logsCmdFound = true
		for _, logsSub := range sub.Commands {
			if logsSub.Name == "get-all" {
				getAllCmdFound = true
				break
			}
		}
	}

	if !logsCmdFound {
		t.Fatal("expected logs subcommand")
	}
	if !getAllCmdFound {
		t.Fatal("expected logs get-all subcommand")
	}
}
