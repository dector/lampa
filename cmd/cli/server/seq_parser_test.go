package server

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseSeqStepInputsFromRawArgv_HappyPath(t *testing.T) {
	argv := []string{
		"set",
		"--kind", "seq",
		"--endpoint", "/flaky",
		"--response.body-1", "fail once",
		"--response.status-1", "500",
		"--response.content-1", "text",
		"--response.header-1", "X-Step:1",
		"--response.body-2=recovered",
		"--response.header-2", "Cache-Control:no-store",
	}

	steps, err := parseSeqStepInputsFromRawArgv(argv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 2 {
		t.Fatalf("unexpected steps count: got %d, want %d", len(steps), 2)
	}

	if got, want := steps[0].Index, 1; got != want {
		t.Fatalf("unexpected step[0] index: got %d, want %d", got, want)
	}
	if got, want := steps[0].Body, "fail once"; got != want {
		t.Fatalf("unexpected step[0] body: got %q, want %q", got, want)
	}
	if steps[0].Status == nil || *steps[0].Status != 500 {
		t.Fatalf("unexpected step[0] status: %#v", steps[0].Status)
	}
	if steps[0].Content == nil || *steps[0].Content != "text" {
		t.Fatalf("unexpected step[0] content: %#v", steps[0].Content)
	}
	if !reflect.DeepEqual(steps[0].Headers, []string{"X-Step:1"}) {
		t.Fatalf("unexpected step[0] headers: %#v", steps[0].Headers)
	}

	if got, want := steps[1].Index, 2; got != want {
		t.Fatalf("unexpected step[1] index: got %d, want %d", got, want)
	}
	if got, want := steps[1].Body, "recovered"; got != want {
		t.Fatalf("unexpected step[1] body: got %q, want %q", got, want)
	}
	if steps[1].Status != nil {
		t.Fatalf("expected step[1] status to be nil, got %#v", steps[1].Status)
	}
	if steps[1].Content != nil {
		t.Fatalf("expected step[1] content to be nil, got %#v", steps[1].Content)
	}
	if !reflect.DeepEqual(steps[1].Headers, []string{"Cache-Control:no-store"}) {
		t.Fatalf("unexpected step[1] headers: %#v", steps[1].Headers)
	}
}

func TestParseSeqStepInputsFromRawArgv_Failures(t *testing.T) {
	tests := []struct {
		name        string
		argv        []string
		wantErrPart string
	}{
		{
			name: "gap in indexes",
			argv: []string{"set", "--response.body-1", "one", "--response.body-3", "three"},
			wantErrPart: "sequence index gap",
		},
		{
			name: "invalid index zero",
			argv: []string{"set", "--response.body-0", "zero"},
			wantErrPart: "expected positive numeric suffix",
		},
		{
			name: "invalid non numeric index",
			argv: []string{"set", "--response.status-foo", "200"},
			wantErrPart: "expected positive numeric suffix",
		},
		{
			name: "duplicate body declaration",
			argv: []string{"set", "--response.body-1", "a", "--response.body-1", "b"},
			wantErrPart: "duplicate declaration",
		},
		{
			name: "missing body for declared step",
			argv: []string{"set", "--response.status-1", "200"},
			wantErrPart: "missing response body",
		},
		{
			name: "invalid status value",
			argv: []string{"set", "--response.body-1", "ok", "--response.status-1", "bad"},
			wantErrPart: "invalid value",
		},
		{
			name: "missing value",
			argv: []string{"set", "--response.body-1"},
			wantErrPart: "missing value",
		},
		{
			name: "indexed suffix is required",
			argv: []string{"set", "--response.body", "ok"},
			wantErrPart: "expected -N suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseSeqStepInputsFromRawArgv(tt.argv)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErrPart)
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErrPart, err.Error())
			}
		})
	}
}
