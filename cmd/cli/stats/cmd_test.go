package stats

import "testing"

func TestCreateCliCommand_IncludesExpectedSubcommands(t *testing.T) {
	cmd := CreateCliCommand()

	if cmd.Name != "stats" {
		t.Fatalf("unexpected command name: %s", cmd.Name)
	}

	want := map[string]bool{
		"collect": false,
		"compare": false,
	}
	for _, sub := range cmd.Commands {
		if _, ok := want[sub.Name]; ok {
			want[sub.Name] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Fatalf("expected stats subcommand %q", name)
		}
	}
}
