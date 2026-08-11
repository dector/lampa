package main

import "testing"

func TestCreateCliCommand_StatsReplacesRootCollectAndCompare(t *testing.T) {
	cmd := CreateCliCommand()

	foundStats := false
	for _, sub := range cmd.Commands {
		switch sub.Name {
		case "stats":
			foundStats = true
		case "collect", "compare":
			t.Fatalf("unexpected root command %q", sub.Name)
		}
	}

	if !foundStats {
		t.Fatal("expected root stats command")
	}
}
