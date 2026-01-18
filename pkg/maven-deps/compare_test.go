package mavendeps

import (
	"testing"
)

// Test data
var (
	depA100   = MvnDependency{Group: "com.example", Name: "lib-a", Version: "1.0.0"}
	depA200   = MvnDependency{Group: "com.example", Name: "lib-a", Version: "2.0.0"}
	depA050   = MvnDependency{Group: "com.example", Name: "lib-a", Version: "0.5.0"}
	depASNAP  = MvnDependency{Group: "com.example", Name: "lib-a", Version: "SNAPSHOT"}
	depALocal = MvnDependency{Group: "com.example", Name: "lib-a", Version: "local"}

	depB100 = MvnDependency{Group: "com.example", Name: "lib-b", Version: "1.0.0"}
	depB200 = MvnDependency{Group: "com.example", Name: "lib-b", Version: "2.0.0"}

	depC100 = MvnDependency{Group: "com.example", Name: "lib-c", Version: "1.0.0"}
)

func TestMvnDependency_String(t *testing.T) {
	tests := []struct {
		name string
		dep  MvnDependency
		want string
	}{
		{
			name: "basic dependency",
			dep:  depA100,
			want: "com.example:lib-a:1.0.0",
		},
		{
			name: "snapshot version",
			dep:  depASNAP,
			want: "com.example:lib-a:SNAPSHOT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dep.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMvnDependency_Equals(t *testing.T) {
	tests := []struct {
		name string
		a    MvnDependency
		b    MvnDependency
		want bool
	}{
		{
			name: "equal dependencies",
			a:    depA100,
			b:    MvnDependency{Group: "com.example", Name: "lib-a", Version: "1.0.0"},
			want: true,
		},
		{
			name: "different versions",
			a:    depA100,
			b:    depA200,
			want: false,
		},
		{
			name: "different names",
			a:    depA100,
			b:    depB100,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equals(tt.b); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMvnDependency_HasSameCoordinates(t *testing.T) {
	tests := []struct {
		name string
		a    MvnDependency
		b    MvnDependency
		want bool
	}{
		{
			name: "same coordinates, same version",
			a:    depA100,
			b:    depA100,
			want: true,
		},
		{
			name: "same coordinates, different versions",
			a:    depA100,
			b:    depA200,
			want: true,
		},
		{
			name: "different coordinates",
			a:    depA100,
			b:    depB100,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.HasSameCoordinates(tt.b); got != tt.want {
				t.Errorf("HasSameCoordinates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMvnDependency_HasSemanticVersion(t *testing.T) {
	tests := []struct {
		name string
		dep  MvnDependency
		want bool
	}{
		{
			name: "valid semver",
			dep:  depA100,
			want: true,
		},
		{
			name: "snapshot version",
			dep:  depASNAP,
			want: false,
		},
		{
			name: "local version",
			dep:  depALocal,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dep.HasSemanticVersion(); got != tt.want {
				t.Errorf("HasSemanticVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMvnDependency_CompareVersion(t *testing.T) {
	tests := []struct {
		name string
		a    MvnDependency
		b    MvnDependency
		want Comparison
	}{
		{
			name: "equal versions",
			a:    depA100,
			b:    depA100,
			want: Equals,
		},
		{
			name: "lower version",
			a:    depA100,
			b:    depA200,
			want: Lower,
		},
		{
			name: "higher version",
			a:    depA200,
			b:    depA100,
			want: Higher,
		},
		{
			name: "non-semantic first",
			a:    depASNAP,
			b:    depA100,
			want: Unknown,
		},
		{
			name: "non-semantic second",
			a:    depA100,
			b:    depASNAP,
			want: Unknown,
		},
		{
			name: "both non-semantic",
			a:    depASNAP,
			b:    depALocal,
			want: Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.CompareVersion(tt.b); got != tt.want {
				t.Errorf("CompareVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMvnDependency_ToDiff(t *testing.T) {
	dep := depA200
	prevVersion := "1.5.0"

	diff := dep.ToDiff(prevVersion)

	if diff.MvnDependency != dep {
		t.Errorf("ToDiff() MvnDependency = %v, want %v", diff.MvnDependency, dep)
	}
	if diff.PrevVersion != prevVersion {
		t.Errorf("ToDiff() PrevVersion = %v, want %v", diff.PrevVersion, prevVersion)
	}
}

func TestFindAddedDeps(t *testing.T) {
	tests := []struct {
		name string
		prev []MvnDependency
		curr []MvnDependency
		want []MvnDependency
	}{
		{
			name: "one added",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependency{depB100},
		},
		{
			name: "multiple added",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA100, depB100, depC100},
			want: []MvnDependency{depB100, depC100},
		},
		{
			name: "none added",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependency{},
		},
		{
			name: "version change not considered added",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA200},
			want: []MvnDependency{},
		},
		{
			name: "empty prev",
			prev: []MvnDependency{},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependency{depA100, depB100},
		},
		{
			name: "empty curr",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{},
			want: []MvnDependency{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindAddedDeps(tt.prev, tt.curr)
			if !equalDeps(got, tt.want) {
				t.Errorf("FindAddedDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindRemovedDeps(t *testing.T) {
	tests := []struct {
		name string
		prev []MvnDependency
		curr []MvnDependency
		want []MvnDependency
	}{
		{
			name: "one removed",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{depA100},
			want: []MvnDependency{depB100},
		},
		{
			name: "multiple removed",
			prev: []MvnDependency{depA100, depB100, depC100},
			curr: []MvnDependency{depA100},
			want: []MvnDependency{depB100, depC100},
		},
		{
			name: "none removed",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependency{},
		},
		{
			name: "version change not considered removed",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA200},
			want: []MvnDependency{},
		},
		{
			name: "empty prev",
			prev: []MvnDependency{},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependency{},
		},
		{
			name: "empty curr",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{},
			want: []MvnDependency{depA100, depB100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindRemovedDeps(tt.prev, tt.curr)
			if !equalDeps(got, tt.want) {
				t.Errorf("FindRemovedDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindUpgradedDeps(t *testing.T) {
	tests := []struct {
		name string
		prev []MvnDependency
		curr []MvnDependency
		want []MvnDependencyDiff
	}{
		{
			name: "one upgraded",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA200},
			want: []MvnDependencyDiff{
				{MvnDependency: depA200, PrevVersion: "1.0.0"},
			},
		},
		{
			name: "multiple upgraded",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{depA200, depB200},
			want: []MvnDependencyDiff{
				{MvnDependency: depA200, PrevVersion: "1.0.0"},
				{MvnDependency: depB200, PrevVersion: "1.0.0"},
			},
		},
		{
			name: "none upgraded",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependencyDiff{},
		},
		{
			name: "downgrade not considered upgrade",
			prev: []MvnDependency{depA200},
			curr: []MvnDependency{depA100},
			want: []MvnDependencyDiff{},
		},
		{
			name: "non-semantic versions ignored",
			prev: []MvnDependency{depASNAP},
			curr: []MvnDependency{depALocal},
			want: []MvnDependencyDiff{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindUpgradedDeps(tt.prev, tt.curr)
			if !equalDepsDiff(got, tt.want) {
				t.Errorf("FindUpgradedDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindDowngradedDeps(t *testing.T) {
	tests := []struct {
		name string
		prev []MvnDependency
		curr []MvnDependency
		want []MvnDependencyDiff
	}{
		{
			name: "one downgraded",
			prev: []MvnDependency{depA200},
			curr: []MvnDependency{depA100},
			want: []MvnDependencyDiff{
				{MvnDependency: depA100, PrevVersion: "2.0.0"},
			},
		},
		{
			name: "multiple downgraded",
			prev: []MvnDependency{depA200, depB200},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependencyDiff{
				{MvnDependency: depA100, PrevVersion: "2.0.0"},
				{MvnDependency: depB100, PrevVersion: "2.0.0"},
			},
		},
		{
			name: "none downgraded",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependencyDiff{},
		},
		{
			name: "upgrade not considered downgrade",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA200},
			want: []MvnDependencyDiff{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindDowngradedDeps(tt.prev, tt.curr)
			if !equalDepsDiff(got, tt.want) {
				t.Errorf("FindDowngradedDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindChangedDeps(t *testing.T) {
	tests := []struct {
		name string
		prev []MvnDependency
		curr []MvnDependency
		want []MvnDependencyDiff
	}{
		{
			name: "non-semantic version change",
			prev: []MvnDependency{depASNAP},
			curr: []MvnDependency{depALocal},
			want: []MvnDependencyDiff{
				{MvnDependency: depALocal, PrevVersion: "SNAPSHOT"},
			},
		},
		{
			name: "semantic to non-semantic",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depASNAP},
			want: []MvnDependencyDiff{
				{MvnDependency: depASNAP, PrevVersion: "1.0.0"},
			},
		},
		{
			name: "non-semantic to semantic",
			prev: []MvnDependency{depASNAP},
			curr: []MvnDependency{depA100},
			want: []MvnDependencyDiff{
				{MvnDependency: depA100, PrevVersion: "SNAPSHOT"},
			},
		},
		{
			name: "semantic version change not included",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA200},
			want: []MvnDependencyDiff{},
		},
		{
			name: "unchanged versions not included",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA100},
			want: []MvnDependencyDiff{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindChangedDeps(tt.prev, tt.curr)
			if !equalDepsDiff(got, tt.want) {
				t.Errorf("FindChangedDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindUnchangedDeps(t *testing.T) {
	tests := []struct {
		name string
		prev []MvnDependency
		curr []MvnDependency
		want []MvnDependency
	}{
		{
			name: "one unchanged",
			prev: []MvnDependency{depA100, depB100},
			curr: []MvnDependency{depA100},
			want: []MvnDependency{depA100},
		},
		{
			name: "multiple unchanged",
			prev: []MvnDependency{depA100, depB100, depC100},
			curr: []MvnDependency{depA100, depB100},
			want: []MvnDependency{depA100, depB100},
		},
		{
			name: "none unchanged",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depB100},
			want: []MvnDependency{},
		},
		{
			name: "version change not unchanged",
			prev: []MvnDependency{depA100},
			curr: []MvnDependency{depA200},
			want: []MvnDependency{},
		},
		{
			name: "empty lists",
			prev: []MvnDependency{},
			curr: []MvnDependency{},
			want: []MvnDependency{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindUnchangedDeps(tt.prev, tt.curr)
			if !equalDeps(got, tt.want) {
				t.Errorf("FindUnchangedDeps() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions

func equalDeps(a, b []MvnDependency) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equals(b[i]) {
			return false
		}
	}
	return true
}

func equalDepsDiff(a, b []MvnDependencyDiff) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].MvnDependency.Equals(b[i].MvnDependency) || a[i].PrevVersion != b[i].PrevVersion {
			return false
		}
	}
	return true
}
