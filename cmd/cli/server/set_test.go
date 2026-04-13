package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v3"
)

func TestContentPresetToContentType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "json", input: "json", want: "application/json"},
		{name: "text", input: "text", want: "text/plain"},
		{name: "html", input: "html", want: "text/html"},
		{name: "raw", input: "raw", want: ""},
		{name: "empty", input: "", want: ""},
		{name: "invalid", input: "xml", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := contentPresetToContentType(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected content type: got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseRawHeaders(t *testing.T) {
	headers, err := parseRawHeaders([]string{"X-Test:1", "X-Test:2", "Content-Type: application/json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	xTestValues := headers.Values("X-Test")
	if len(xTestValues) != 2 {
		t.Fatalf("unexpected X-Test values: %#v", xTestValues)
	}
	if xTestValues[0] != "1" || xTestValues[1] != "2" {
		t.Fatalf("unexpected X-Test values order: %#v", xTestValues)
	}
	if got, want := headers.Get("Content-Type"), "application/json"; got != want {
		t.Fatalf("unexpected content type header: got %q, want %q", got, want)
	}

	_, err = parseRawHeaders([]string{"NoColon"})
	if err == nil {
		t.Fatal("expected error for malformed header, got nil")
	}
}

func TestBuildSetRequestFromCommand(t *testing.T) {
	cmd := newSetCommand("set", "set endpoint processor")
	args := []string{
		"set",
		"--endpoint", "/example",
		"--response.status", "201",
		"--response.content", "json",
		"--response.body", `{"ok":true}`,
		"--response.header", "X-Test:1",
	}

	if err := cmd.Run(context.Background(), args); err == nil {
		// action fails because no server is running, but flags are parsed and validation path is covered.
	}

	payload, err := buildSetRequestFromCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := payload.Kind, "static"; got != want {
		t.Fatalf("unexpected kind: got %q, want %q", got, want)
	}
	if got, want := payload.Endpoint, "/example"; got != want {
		t.Fatalf("unexpected endpoint: got %q, want %q", got, want)
	}
	if payload.Response == nil {
		t.Fatal("expected static response payload")
	}
	if got, want := payload.Response.Status, 201; got != want {
		t.Fatalf("unexpected response status: got %d, want %d", got, want)
	}
	if got, want := payload.Response.ContentType, "application/json"; got != want {
		t.Fatalf("unexpected response contentType: got %q, want %q", got, want)
	}
	if got, want := payload.Response.Headers.Get("X-Test"), "1"; got != want {
		t.Fatalf("unexpected response header X-Test: got %q, want %q", got, want)
	}
	if got, want := payload.Response.Body, `{"ok":true}`; got != want {
		t.Fatalf("unexpected response body: got %q, want %q", got, want)
	}
}

func TestBuildSetRequestFromCommand_ContentTypeOverride(t *testing.T) {
	cmd := newSetCommand("set", "set endpoint processor")
	args := []string{
		"set",
		"--endpoint", "/example",
		"--response.content", "json",
		"--response.header", "Content-Type:text/custom",
		"--response.body", "ok",
	}
	_ = cmd.Run(context.Background(), args)

	payload, err := buildSetRequestFromCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.Response == nil {
		t.Fatal("expected static response payload")
	}
	if got, want := payload.Response.ContentType, "text/custom"; got != want {
		t.Fatalf("unexpected response contentType: got %q, want %q", got, want)
	}
}

func TestBuildSetRequestFromCommand_SeqRequest(t *testing.T) {
	cmd := newSetCommand("set", "set endpoint processor")
	args := []string{
		"set",
		"--kind", "seq",
		"--endpoint", "/flaky",
		"--response.status-1", "500",
		"--response.content-1", "text",
		"--response.body-1", "fail once",
		"--response.header-1", "X-Step:1",
		"--response.content-2", "json",
		"--response.body-2", `{"ok":true}`,
		"--response.header-2", "Content-Type:text/custom",
	}
	_ = cmd.Run(context.Background(), args)

	payload, err := buildSetRequestFromCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := payload.Kind, "seq"; got != want {
		t.Fatalf("unexpected kind: got %q, want %q", got, want)
	}
	if got, want := payload.Endpoint, "/flaky"; got != want {
		t.Fatalf("unexpected endpoint: got %q, want %q", got, want)
	}
	if payload.Response != nil {
		t.Fatal("expected static response payload to be omitted for seq kind")
	}
	if len(payload.Sequence) != 2 {
		t.Fatalf("unexpected sequence length: got %d, want 2", len(payload.Sequence))
	}

	step1 := payload.Sequence[0].Response
	if got, want := step1.Status, 500; got != want {
		t.Fatalf("unexpected step1 status: got %d, want %d", got, want)
	}
	if got, want := step1.ContentType, "text/plain"; got != want {
		t.Fatalf("unexpected step1 content type: got %q, want %q", got, want)
	}
	if got, want := step1.Headers.Get("X-Step"), "1"; got != want {
		t.Fatalf("unexpected step1 header X-Step: got %q, want %q", got, want)
	}

	step2 := payload.Sequence[1].Response
	if got, want := step2.Status, 200; got != want {
		t.Fatalf("unexpected step2 status default: got %d, want %d", got, want)
	}
	if got, want := step2.ContentType, "text/custom"; got != want {
		t.Fatalf("unexpected step2 content type: got %q, want %q", got, want)
	}
	if got, want := step2.Headers.Get("Content-Type"), "text/custom"; got != want {
		t.Fatalf("unexpected step2 content-type header override: got %q, want %q", got, want)
	}
}

func TestBuildSetRequestFromCommand_SeqValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantErrPart string
	}{
		{
			name: "missing body per step",
			args: []string{
				"set",
				"--kind", "seq",
				"--endpoint", "/flaky",
				"--response.body-1", "ok",
				"--response.status-2", "500",
			},
			wantErrPart: "step 2: missing response body",
		},
		{
			name: "invalid status per step",
			args: []string{
				"set",
				"--kind", "seq",
				"--endpoint", "/flaky",
				"--response.body-1", "ok",
				"--response.status-2", "99",
				"--response.body-2", "bad",
			},
			wantErrPart: "step 2: invalid status 99",
		},
		{
			name: "invalid content per step",
			args: []string{
				"set",
				"--kind", "seq",
				"--endpoint", "/flaky",
				"--response.body-1", "ok",
				"--response.content-1", "xml",
			},
			wantErrPart: "step 1: invalid response content",
		},
		{
			name: "invalid header per step",
			args: []string{
				"set",
				"--kind", "seq",
				"--endpoint", "/flaky",
				"--response.body-1", "ok",
				"--response.header-1", "NoColon",
			},
			wantErrPart: "step 1: invalid response header",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSetCommand("set", "set endpoint processor")
			_ = cmd.Run(context.Background(), tt.args)

			_, err := buildSetRequestFromCommand(cmd)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErrPart)
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErrPart, err.Error())
			}
		})
	}
}

func TestValidateSetInput(t *testing.T) {
	if err := validateSetInput("static", "/ok", 200, "body", ""); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := validateSetInput("js", "/ok", 0, "", "function handle(req){return null;}"); err != nil {
		t.Fatalf("expected no error for js kind, got %v", err)
	}
	if err := validateSetInput("dynamic", "/ok", 200, "body", ""); err == nil {
		t.Fatal("expected kind validation error, got nil")
	}
	if err := validateSetInput("static", "bad", 200, "body", ""); err == nil {
		t.Fatal("expected endpoint validation error, got nil")
	}
	if err := validateSetInput("static", "/ok", 99, "body", ""); err == nil {
		t.Fatal("expected status validation error, got nil")
	}
	if err := validateSetInput("static", "/ok", 200, "", ""); err == nil {
		t.Fatal("expected static body validation error, got nil")
	}
	if err := validateSetInput("js", "/ok", 0, "", ""); err == nil {
		t.Fatal("expected js script validation error, got nil")
	}
}

func TestSetControl(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.HandlerFunc
		wantErrPart string
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("unexpected method: %s", r.Method)
				}
				if r.URL.Path != "/api/v0/proc/set" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				if got := r.Header.Get("Content-Type"); got != "application/json" {
					t.Fatalf("unexpected request content type: %q", got)
				}
				raw, _ := io.ReadAll(r.Body)
				if !bytes.Contains(raw, []byte(`"endpoint":"/example"`)) {
					t.Fatalf("unexpected request body: %s", string(raw))
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"ok","endpoint":"/example","kind":"static"}`))
			},
		},
		{
			name: "http error",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"status":"error","error":"invalid endpoint"}`))
			},
			wantErrPart: "HTTP 400",
		},
		{
			name: "invalid response json",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("not-json"))
			},
			wantErrPart: "invalid set JSON response",
		},
		{
			name: "unexpected status",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"error"}`))
			},
			wantErrPart: `unexpected set status: "error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(tt.handler))
			defer ts.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			response, err := setControl(ctx, ts.Client(), ts.URL+"/api/v0/proc/set", procSetRequest{
				Kind:     "static",
				Endpoint: "/example",
				Response: &procSetStaticResponse{Status: 200, Body: "ok"},
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

func TestBuildProcSetURL(t *testing.T) {
	got := buildProcSetURL(8081)
	want := "http://localhost:8081/api/v0/proc/set"
	if got != want {
		t.Fatalf("buildProcSetURL() = %q, want %q", got, want)
	}
}

func TestCmdActionSet_FlagsValidation(t *testing.T) {
	cmd := createSetCommand()
	err := cmd.Run(context.Background(), []string{"set", "--port", "0", "--endpoint", "/a", "--response.body", "x"})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestBuildSetRequestFromCommand_InvalidContentPreset(t *testing.T) {
	cmd := &cli.Command{}
	cmd.Flags = newSetCommand("set", "set endpoint processor").Flags
	_ = cmd.Run(context.Background(), []string{"set", "--endpoint", "/a", "--response.content", "yaml", "--response.body", "x"})

	_, err := buildSetRequestFromCommand(cmd)
	if err == nil {
		t.Fatal("expected invalid content error, got nil")
	}
}

func TestSetControl_RequestEncoding(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload procSetRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Endpoint != "/example" {
			t.Fatalf("unexpected endpoint: %q", payload.Endpoint)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := setControl(ctx, ts.Client(), ts.URL, procSetRequest{Kind: "static", Endpoint: "/example", Response: &procSetStaticResponse{Status: 200, Body: "ok"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildSetRequestFromCommand_JS_InlineScript(t *testing.T) {
	cmd := newSetCommand("set", "set endpoint processor")
	_ = cmd.Run(context.Background(), []string{
		"set",
		"--kind", "js",
		"--endpoint", "/js",
		"--script", "function handle(req){ return Response.json({ok:true}); }",
	})

	payload, err := buildSetRequestFromCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := payload.Kind, "js"; got != want {
		t.Fatalf("unexpected kind: got %q, want %q", got, want)
	}
	if payload.Response != nil {
		t.Fatal("expected static response to be omitted for js kind")
	}
	if payload.JS == nil || strings.TrimSpace(payload.JS.Script) == "" {
		t.Fatal("expected js script payload")
	}
}

func TestBuildSetRequestFromCommand_JS_ScriptFile(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "handler.js")
	script := "function handle(req){ return Response.text('ok'); }"
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}

	cmd := newSetCommand("set", "set endpoint processor")
	_ = cmd.Run(context.Background(), []string{
		"set",
		"--kind", "js",
		"--endpoint", "/js",
		"--script-file", scriptPath,
	})

	payload, err := buildSetRequestFromCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload.JS == nil {
		t.Fatal("expected js payload")
	}
	if got, want := payload.JS.Script, script; got != want {
		t.Fatalf("unexpected script payload: got %q, want %q", got, want)
	}
}

func TestBuildSetRequestFromCommand_JS_ScriptConflict(t *testing.T) {
	cmd := newSetCommand("set", "set endpoint processor")
	_ = cmd.Run(context.Background(), []string{
		"set",
		"--kind", "js",
		"--endpoint", "/js",
		"--script", "function handle(req){return null;}",
		"--script-file", "handler.js",
	})

	_, err := buildSetRequestFromCommand(cmd)
	if err == nil {
		t.Fatal("expected conflict validation error, got nil")
	}
}

func TestBuildSetRequestFromCommand_JS_MissingScript(t *testing.T) {
	cmd := newSetCommand("set", "set endpoint processor")
	_ = cmd.Run(context.Background(), []string{
		"set",
		"--kind", "js",
		"--endpoint", "/js",
	})

	_, err := buildSetRequestFromCommand(cmd)
	if err == nil {
		t.Fatal("expected missing script error, got nil")
	}
}
