package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPingControl(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantErrPart string
	}{
		{
			name: "success response",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			},
		},
		{
			name: "non-200 status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				_, _ = w.Write([]byte("upstream failed"))
			},
			wantErrPart: "HTTP 502",
		},
		{
			name: "invalid json",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`not-json`))
			},
			wantErrPart: "invalid ping JSON response",
		},
		{
			name: "json without status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"foo":"bar"}`))
			},
			wantErrPart: `unexpected ping status: ""`,
		},
		{
			name: "wrong status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"nope"}`))
			},
			wantErrPart: `unexpected ping status: "nope"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/ping" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				tt.handler(w, r)
			}))
			defer ts.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			resp, err := pingControl(ctx, ts.Client(), ts.URL+"/ping")
			if tt.wantErrPart == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if got, _ := resp["status"].(string); got != "ok" {
					t.Fatalf("expected status ok in response, got %q", got)
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

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{name: "valid min", port: 1, wantErr: false},
		{name: "valid max", port: 65535, wantErr: false},
		{name: "zero", port: 0, wantErr: true},
		{name: "negative", port: -1, wantErr: true},
		{name: "too large", port: 65536, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePort(tt.port)
			if tt.wantErr && err == nil {
				t.Fatalf("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestBuildPingURL(t *testing.T) {
	got := buildPingURL(8081)
	want := "http://localhost:8081/ping"
	if got != want {
		t.Fatalf("buildPingURL() = %q, want %q", got, want)
	}
}
