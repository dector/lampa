package generate

import "testing"

func TestExtractCompileSdkFromOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "valid output",
			output: "compileSdk=36",
			want:   "36",
		},
		{
			name:   "output with noise",
			output: "Building...\ncompileSdk=34\nDone",
			want:   "34",
		},
		{
			name:   "output with whitespace",
			output: "compileSdk=35  ",
			want:   "35",
		},
		{
			name:   "output with leading whitespace",
			output: "  compileSdk=33  ",
			want:   "33",
		},
		{
			name:   "empty output",
			output: "",
			want:   "",
		},
		{
			name:   "no match",
			output: "some gradle output",
			want:   "",
		},
		{
			name:   "multiple lines with match in middle",
			output: "Configuration on demand is an incubating feature.\ncompileSdk=35\nBUILD SUCCESSFUL",
			want:   "35",
		},
		{
			name:   "output with warnings before match",
			output: "Warning: Some deprecation warning\ncompileSdk=34\nTask completed",
			want:   "34",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCompileSdkFromOutput(tt.output)
			if got != tt.want {
				t.Errorf("extractCompileSdkFromOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
