package globals

import "testing"

func TestParseCIEnv(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    bool
		wantErr bool
	}{
		{name: "empty", value: "", want: false},
		{name: "whitespace empty", value: "  ", want: false},
		{name: "y", value: "y", want: true},
		{name: "yes", value: "yes", want: true},
		{name: "one", value: "1", want: true},
		{name: "true", value: "true", want: true},
		{name: "uppercase true", value: "TRUE", want: true},
		{name: "mixed case yes with spaces", value: " Yes ", want: true},
		{name: "n", value: "n", want: false},
		{name: "no", value: "no", want: false},
		{name: "zero", value: "0", want: false},
		{name: "false", value: "false", want: false},
		{name: "uppercase false", value: "FALSE", want: false},
		{name: "invalid", value: "maybe", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCIEnv(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlobalsInitParsesCIEnv(t *testing.T) {
	t.Setenv("CI", "yes")

	var g Globals
	if err := g.Init(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !g.UsePlainOutput {
		t.Fatalf("UsePlainOutput should be true")
	}
	if !g.UseOnlyStdout {
		t.Fatalf("UseOnlyStdout should be true")
	}
}

func TestGlobalsInitRejectsInvalidCIEnv(t *testing.T) {
	t.Setenv("CI", "maybe")

	var g Globals
	if err := g.Init(); err == nil {
		t.Fatalf("expected error")
	}
}
