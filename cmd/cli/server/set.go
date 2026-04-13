package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	. "github.com/dector/lampa/internal/globals"
	"github.com/urfave/cli/v3"
)

const (
	OptKind            = "kind"
	OptEndpoint        = "endpoint"
	OptResponseStatus  = "response.status"
	OptResponseContent = "response.content"
	OptResponseBody    = "response.body"
	OptResponseHeader  = "response.header"
	OptScript          = "script"
	OptScriptFile      = "script-file"

	kindStatic = "static"
	kindJS     = "js"

	defaultProcessorKind = kindStatic
)

type procSetRequest struct {
	Kind     string                 `json:"kind"`
	Endpoint string                 `json:"endpoint"`
	Response *procSetStaticResponse `json:"response,omitempty"`
	JS       *procSetJSConfig       `json:"js,omitempty"`
}

type procSetStaticResponse struct {
	Status      int         `json:"status"`
	ContentType string      `json:"contentType,omitempty"`
	Headers     http.Header `json:"headers,omitempty"`
	Body        string      `json:"body"`
}

type procSetJSConfig struct {
	Script string `json:"script"`
}

func createSetCommand() *cli.Command {
	return newSetCommand("set", "set endpoint processor")
}

func createProxyCommand() *cli.Command {
	return &cli.Command{
		Name:  "proxy",
		Usage: "proxy processor operations",
		Commands: []*cli.Command{
			newSetCommand("set", "set endpoint processor"),
			createSetDefaultCommand(),
		},
	}
}

func newSetCommand(name string, usage string) *cli.Command {
	return &cli.Command{
		Name:  name,
		Usage: usage,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  OptPort,
				Usage: "control server port",
				Value: DefaultControlPort,
			},
			&cli.StringFlag{
				Name:  OptKind,
				Usage: "processor kind (static|js)",
				Value: defaultProcessorKind,
			},
			&cli.StringFlag{
				Name:     OptEndpoint,
				Usage:    "endpoint path to override (must start with /)",
				Required: true,
			},
			&cli.IntFlag{
				Name:  OptResponseStatus,
				Usage: "response status code",
				Value: http.StatusOK,
			},
			&cli.StringFlag{
				Name:  OptResponseContent,
				Usage: "response content preset: json|text|html|raw",
				Value: "text",
			},
			&cli.StringFlag{
				Name:  OptResponseBody,
				Usage: "response body string (required for --kind static)",
			},
			&cli.StringFlag{
				Name:  OptScript,
				Usage: "inline JS script for --kind js",
			},
			&cli.StringFlag{
				Name:  OptScriptFile,
				Usage: "path to JS file for --kind js",
			},
			&cli.StringSliceFlag{
				Name:  OptResponseHeader,
				Usage: "additional response header in Name:Value format (repeatable)",
			},
		},
		Action: CmdActionSet,
	}
}

func CmdActionSet(ctx context.Context, c *cli.Command) error {
	payload, err := buildSetRequestFromCommand(c)
	if err != nil {
		return err
	}

	url := buildProcSetURL(c.Int(OptPort))
	client := &http.Client{Timeout: 5 * time.Second}

	response, err := setControl(ctx, client, url, payload)
	if err != nil {
		return err
	}

	fmt.Printf("ok: set %s (%s)\n", payload.Endpoint, payload.Kind)
	if G.Verbosity >= VerbosityInfo {
		formatted, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format set response JSON: %w", err)
		}
		fmt.Printf("<< %s\n", string(formatted))
	}
	return nil
}

func buildSetRequestFromCommand(c *cli.Command) (procSetRequest, error) {
	port := c.Int(OptPort)
	if err := validatePort(port); err != nil {
		return procSetRequest{}, err
	}

	kind := strings.TrimSpace(c.String(OptKind))
	endpoint := strings.TrimSpace(c.String(OptEndpoint))
	status := c.Int(OptResponseStatus)
	content := strings.TrimSpace(c.String(OptResponseContent))
	body := c.String(OptResponseBody)
	rawHeaders := c.StringSlice(OptResponseHeader)
	scriptInline := c.String(OptScript)
	scriptFile := strings.TrimSpace(c.String(OptScriptFile))

	headers, err := parseRawHeaders(rawHeaders)
	if err != nil {
		return procSetRequest{}, err
	}

	script, err := resolveScriptInput(scriptInline, scriptFile)
	if err != nil {
		return procSetRequest{}, err
	}

	if err := validateSetInput(kind, endpoint, status, body, script); err != nil {
		return procSetRequest{}, err
	}

	payload := procSetRequest{
		Kind:     kind,
		Endpoint: endpoint,
	}

	switch kind {
	case kindStatic:
		presetContentType, err := contentPresetToContentType(content)
		if err != nil {
			return procSetRequest{}, err
		}
		if headers.Get("Content-Type") == "" && presetContentType != "" {
			headers.Set("Content-Type", presetContentType)
		}

		contentType := strings.TrimSpace(headers.Get("Content-Type"))
		payload.Response = &procSetStaticResponse{
			Status:      status,
			ContentType: contentType,
			Headers:     headers,
			Body:        body,
		}
	case kindJS:
		payload.JS = &procSetJSConfig{Script: script}
	}

	return payload, nil
}

func validateSetInput(kind, endpoint string, status int, body string, script string) error {
	kind = strings.TrimSpace(kind)
	if kind != kindStatic && kind != kindJS {
		return fmt.Errorf("invalid kind %q: expected %q or %q", kind, kindStatic, kindJS)
	}

	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" || !strings.HasPrefix(endpoint, "/") {
		return fmt.Errorf("invalid endpoint %q: must start with /", endpoint)
	}

	switch kind {
	case kindStatic:
		if status < 100 || status > 599 {
			return fmt.Errorf("invalid status %d: must be between 100 and 599", status)
		}
		if strings.TrimSpace(body) == "" {
			return fmt.Errorf("missing response body for kind %q", kindStatic)
		}
	case kindJS:
		if strings.TrimSpace(script) == "" {
			return fmt.Errorf("missing script for kind %q", kindJS)
		}
	}

	return nil
}

func resolveScriptInput(scriptInline string, scriptFile string) (string, error) {
	inline := strings.TrimSpace(scriptInline)
	file := strings.TrimSpace(scriptFile)

	if inline != "" && file != "" {
		return "", fmt.Errorf("%s and %s cannot be used together", OptScript, OptScriptFile)
	}

	if file == "" {
		return scriptInline, nil
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to read %s %q: %w", OptScriptFile, file, err)
	}
	return string(content), nil
}

func contentPresetToContentType(content string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(content)) {
	case "", "raw":
		return "", nil
	case "json":
		return "application/json", nil
	case "text":
		return "text/plain", nil
	case "html":
		return "text/html", nil
	default:
		return "", fmt.Errorf("invalid response content %q: expected json|text|html|raw", content)
	}
}

func parseRawHeaders(rawHeaders []string) (http.Header, error) {
	headers := make(http.Header)
	for _, raw := range rawHeaders {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid response header %q: expected Name:Value", raw)
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if name == "" {
			return nil, fmt.Errorf("invalid response header %q: empty header name", raw)
		}
		headers.Add(name, value)
	}
	return headers, nil
}

func buildProcSetURL(port int) string {
	return fmt.Sprintf("http://localhost:%d/api/v0/proc/set", port)
}

func setControl(ctx context.Context, client *http.Client, url string, payload procSetRequest) (map[string]any, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode set request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("failed to build set request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call control server at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		body := strings.TrimSpace(string(raw))
		if body == "" {
			return nil, fmt.Errorf("control server returned HTTP %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("control server returned HTTP %d: %s", resp.StatusCode, body)
	}

	var response map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("invalid set JSON response: %w", err)
	}

	status, _ := response["status"].(string)
	if status != "ok" {
		return nil, fmt.Errorf("unexpected set status: %q", status)
	}

	return response, nil
}
