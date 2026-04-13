package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBuildProxyLogsURL(t *testing.T) {
	got := buildProxyLogsURL(8081, 25)
	want := "http://localhost:8081/api/v0/proxy/logs?n=25"
	if got != want {
		t.Fatalf("buildProxyLogsURL() = %q, want %q", got, want)
	}
}

func TestValidateN(t *testing.T) {
	if err := validateN(1); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if err := validateN(0); err == nil {
		t.Fatal("expected validation error for n=0")
	}
}

func TestGetProxyLogs(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantErrPart string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got, want := r.Method, http.MethodGet; got != want {
					t.Fatalf("unexpected method: got %s, want %s", got, want)
				}
				if got, want := r.URL.Path, "/api/v0/proxy/logs"; got != want {
					t.Fatalf("unexpected path: got %s, want %s", got, want)
				}
				if got, want := r.URL.Query().Get("n"), "2"; got != want {
					t.Fatalf("unexpected n query: got %q, want %q", got, want)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok","count":1,"entries":[{"id":1}]}`))
			},
		},
		{
			name: "http error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"status":"error"}`))
			},
			wantErrPart: "HTTP 500",
		},
		{
			name: "invalid json",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("not-json"))
			},
			wantErrPart: "invalid proxy logs JSON response",
		},
		{
			name: "unexpected status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"error"}`))
			},
			wantErrPart: `unexpected proxy logs status: "error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer ts.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			response, err := getProxyLogs(ctx, ts.Client(), ts.URL+"/api/v0/proxy/logs?n=2")
			if tt.wantErrPart == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got, want := response["status"], "ok"; got != want {
					t.Fatalf("unexpected status: got %v, want %v", got, want)
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

func TestCmdActionGetAll_FlagsValidation(t *testing.T) {
	cmd := createGetAllCommand()
	err := cmd.Run(context.Background(), []string{"get-all", "--n", "0"})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}
