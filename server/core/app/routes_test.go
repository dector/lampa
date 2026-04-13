package app

import "testing"

func TestComposeControlPath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		route    string
		want     string
	}{
		{name: "root base ping route", basePath: "/", route: RoutePing, want: "/ping"},
		{name: "empty base defaults to root", basePath: "", route: RoutePing, want: "/ping"},
		{name: "nested base ping route", basePath: "/control", route: RoutePing, want: "/control/ping"},
		{name: "nested base with trailing slash", basePath: "/control/", route: RoutePing, want: "/control/ping"},
		{name: "nested base proc_count route", basePath: "/control", route: RouteProcCount, want: "/control/api/v0/proc_count"},
		{name: "nested base proc_set route", basePath: "/control", route: RouteProcSet, want: "/control/api/v0/proc/set"},
		{name: "empty route returns clean base", basePath: "/control/", route: "", want: "/control"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComposeControlPath(tt.basePath, tt.route); got != tt.want {
				t.Fatalf("unexpected route path: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestControlRouteConstants(t *testing.T) {
	if got, want := RoutePing, "/ping"; got != want {
		t.Fatalf("unexpected RoutePing: got %q, want %q", got, want)
	}
	if got, want := RouteProcCount, "/api/v0/proc_count"; got != want {
		t.Fatalf("unexpected RouteProcCount: got %q, want %q", got, want)
	}
	if got, want := RouteProcSet, "/api/v0/proc/set"; got != want {
		t.Fatalf("unexpected RouteProcSet: got %q, want %q", got, want)
	}
}
