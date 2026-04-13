package proxy

import "testing"

func TestCreateCliCommand_IncludesExpectedSubcommands(t *testing.T) {
	cmd := CreateCliCommand()

	if got, want := cmd.Name, "proxy"; got != want {
		t.Fatalf("unexpected command name: got %q, want %q", got, want)
	}

	expected := map[string]bool{
		"ping":        false,
		"set":         false,
		"set-default": false,
		"logs":        false,
	}
	var getAllCmdFound bool

	for _, sub := range cmd.Commands {
		if _, ok := expected[sub.Name]; ok {
			expected[sub.Name] = true
		}
		if sub.Name != "logs" {
			continue
		}
		for _, logsSub := range sub.Commands {
			if logsSub.Name == "get-all" {
				getAllCmdFound = true
				break
			}
		}
	}

	for name, found := range expected {
		if !found {
			t.Fatalf("expected %s subcommand", name)
		}
	}
	if !getAllCmdFound {
		t.Fatal("expected logs get-all subcommand")
	}
}
