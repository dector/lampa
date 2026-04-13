package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildSetDefaultRequestFromCommand(t *testing.T) {
	cmd := createSetDefaultCommand()
	_ = cmd.Run(context.Background(), []string{"set-default", "--kind", "pass", "--server", "http://localhost:9090"})

	payload, err := buildSetDefaultRequestFromCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := payload.Kind, "pass"; got != want {
		t.Fatalf("unexpected kind: got %q, want %q", got, want)
	}
	if got, want := payload.Server, "http://localhost:9090"; got != want {
		t.Fatalf("unexpected server: got %q, want %q", got, want)
	}
}

func TestValidateSetDefaultInput(t *testing.T) {
	if err := validateSetDefaultInput("pass", "http://localhost:9090"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := validateSetDefaultInput("static", "http://localhost:9090"); err == nil {
		t.Fatal("expected kind validation error")
	}
	if err := validateSetDefaultInput("pass", "localhost:9090"); err == nil {
		t.Fatal("expected server validation error")
	}
}

func TestBuildProcSetDefaultURL(t *testing.T) {
	got := buildProcSetDefaultURL(8081)
	want := "http://localhost:8081/api/v0/proc/default/set"
	if got != want {
		t.Fatalf("buildProcSetDefaultURL() = %q, want %q", got, want)
	}
}

func TestSetDefaultControl(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantErrPart string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got, want := r.Method, http.MethodPost; got != want {
					t.Fatalf("unexpected method: got %s, want %s", got, want)
				}
				if got, want := r.URL.Path, "/api/v0/proc/default/set"; got != want {
					t.Fatalf("unexpected path: got %s, want %s", got, want)
				}
				var payload procSetDefaultRequest
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatalf("decode request: %v", err)
				}
				if got, want := payload.Kind, "pass"; got != want {
					t.Fatalf("unexpected kind: got %q, want %q", got, want)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok","kind":"pass","server":"http://localhost:9090"}`))
			},
		},
		{
			name: "http error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"status":"error","error":"invalid kind"}`))
			},
			wantErrPart: "HTTP 400",
		},
		{
			name: "invalid response json",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("not-json"))
			},
			wantErrPart: "invalid set-default JSON response",
		},
		{
			name: "unexpected status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"error"}`))
			},
			wantErrPart: `unexpected set-default status: "error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer ts.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			response, err := setDefaultControl(ctx, ts.Client(), ts.URL+"/api/v0/proc/default/set", procSetDefaultRequest{
				Kind:   "pass",
				Server: "http://localhost:9090",
			})

			if tt.wantErrPart == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got, want := response["status"], "ok"; got != want {
					t.Fatalf("unexpected response status: got %v, want %v", got, want)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErrPart)
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErrPart, err.Error())
			}
		})
	}
}

func TestCreateProxyCommand_IncludesSetDefault(t *testing.T) {
	cmd := createProxyCommand()
	found := false
	for _, sub := range cmd.Commands {
		if sub.Name == "set-default" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected proxy command to include set-default subcommand")
	}
}
