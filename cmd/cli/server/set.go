package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
	kindSeq    = "seq"
	kindJS     = "js"

	defaultProcessorKind = kindStatic
)

type procSetRequest struct {
	Kind     string                 `json:"kind"`
	Endpoint string                 `json:"endpoint"`
	Response *procSetStaticResponse `json:"response,omitempty"`
	Sequence []procSetSeqStep       `json:"sequence,omitempty"`
	JS       *procSetJSConfig       `json:"js,omitempty"`
}

type procSetSeqStep struct {
	Response procSetStaticResponse `json:"response"`
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

type setCommandInput struct {
	Port       int
	Kind       string
	Endpoint   string
	Status     int
	Content    string
	Body       string
	Headers    []string
	Script     string
	ScriptFile string
	SeqSteps   []seqStepInput
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
		Name:            name,
		Usage:           usage,
		SkipFlagParsing: true,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  OptPort,
				Usage: "control server port",
				Value: DefaultControlPort,
			},
			&cli.StringFlag{
				Name:  OptKind,
				Usage: "processor kind (static|seq|js)",
				Value: defaultProcessorKind,
			},
			&cli.StringFlag{
				Name:     OptEndpoint,
				Usage:    "endpoint path to override (must start with /)",
				Required: true,
			},
			&cli.IntFlag{
				Name:  OptResponseStatus,
				Usage: "response status code (static only; seq uses --response.status-N, N starts at 1 with no gaps)",
				Value: http.StatusOK,
			},
			&cli.StringFlag{
				Name:  OptResponseContent,
				Usage: "response content preset: json|text|html|raw (static only; seq uses --response.content-N)",
				Value: "text",
			},
			&cli.StringFlag{
				Name:  OptResponseBody,
				Usage: "response body string (required for --kind static; seq uses required --response.body-N per step)",
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
				Usage: "additional response header in Name:Value format (repeatable; seq uses --response.header-N)",
			},
		},
		Action: CmdActionSet,
	}
}

func CmdActionSet(ctx context.Context, c *cli.Command) error {
	input, err := parseSetCommandInput(c.Args().Slice())
	if err != nil {
		return err
	}

	payload, err := buildSetRequest(input)
	if err != nil {
		return err
	}

	url := buildProcSetURL(input.Port)
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
	rawArgv := c.Args().Slice()
	if len(rawArgv) == 0 {
		return buildSetRequest(setCommandInputFromParsedFlags(c))
	}

	input, err := parseSetCommandInput(rawArgv)
	if err != nil {
		return procSetRequest{}, err
	}
	return buildSetRequest(input)
}

func setCommandInputFromParsedFlags(c *cli.Command) setCommandInput {
	return setCommandInput{
		Port:       c.Int(OptPort),
		Kind:       strings.TrimSpace(c.String(OptKind)),
		Endpoint:   strings.TrimSpace(c.String(OptEndpoint)),
		Status:     c.Int(OptResponseStatus),
		Content:    strings.TrimSpace(c.String(OptResponseContent)),
		Body:       c.String(OptResponseBody),
		Headers:    c.StringSlice(OptResponseHeader),
		Script:     c.String(OptScript),
		ScriptFile: strings.TrimSpace(c.String(OptScriptFile)),
	}
}

func buildSetRequest(input setCommandInput) (procSetRequest, error) {
	if err := validatePort(input.Port); err != nil {
		return procSetRequest{}, err
	}

	if input.Kind != kindSeq && len(input.SeqSteps) > 0 {
		return procSetRequest{}, fmt.Errorf("indexed response flags require kind %q", kindSeq)
	}

	payload := procSetRequest{
		Kind:     input.Kind,
		Endpoint: input.Endpoint,
	}

	switch input.Kind {
	case kindStatic:
		if err := validateSetInput(kindStatic, input.Endpoint, input.Status, input.Body, ""); err != nil {
			return procSetRequest{}, err
		}
		headers, err := parseRawHeaders(input.Headers)
		if err != nil {
			return procSetRequest{}, err
		}

		presetContentType, err := contentPresetToContentType(input.Content)
		if err != nil {
			return procSetRequest{}, err
		}
		if headers.Get("Content-Type") == "" && presetContentType != "" {
			headers.Set("Content-Type", presetContentType)
		}

		contentType := strings.TrimSpace(headers.Get("Content-Type"))
		payload.Response = &procSetStaticResponse{
			Status:      input.Status,
			ContentType: contentType,
			Headers:     headers,
			Body:        input.Body,
		}
	case kindSeq:
		if err := validateSetInput(kindSeq, input.Endpoint, 0, "", ""); err != nil {
			return procSetRequest{}, err
		}
		if len(input.SeqSteps) == 0 {
			return procSetRequest{}, fmt.Errorf("missing indexed sequence flags: expected --%s-N", OptResponseBody)
		}

		payload.Sequence = make([]procSetSeqStep, 0, len(input.SeqSteps))
		for _, step := range input.SeqSteps {
			status := http.StatusOK
			if step.Status != nil {
				status = *step.Status
			}

			content := "text"
			if step.Content != nil {
				content = *step.Content
			}

			if err := validateSetInput(kindStatic, input.Endpoint, status, step.Body, ""); err != nil {
				return procSetRequest{}, fmt.Errorf("step %d: %w", step.Index, err)
			}

			headers, err := parseRawHeaders(step.Headers)
			if err != nil {
				return procSetRequest{}, fmt.Errorf("step %d: %w", step.Index, err)
			}

			presetContentType, err := contentPresetToContentType(content)
			if err != nil {
				return procSetRequest{}, fmt.Errorf("step %d: %w", step.Index, err)
			}
			if headers.Get("Content-Type") == "" && presetContentType != "" {
				headers.Set("Content-Type", presetContentType)
			}

			contentType := strings.TrimSpace(headers.Get("Content-Type"))
			payload.Sequence = append(payload.Sequence, procSetSeqStep{
				Response: procSetStaticResponse{
					Status:      status,
					ContentType: contentType,
					Headers:     headers,
					Body:        step.Body,
				},
			})
		}
	case kindJS:
		script, err := resolveScriptInput(input.Script, input.ScriptFile)
		if err != nil {
			return procSetRequest{}, err
		}
		if err := validateSetInput(kindJS, input.Endpoint, 0, "", script); err != nil {
			return procSetRequest{}, err
		}
		payload.JS = &procSetJSConfig{Script: script}
	default:
		return procSetRequest{}, fmt.Errorf("invalid kind %q: expected %q, %q or %q", input.Kind, kindStatic, kindSeq, kindJS)
	}

	return payload, nil
}

func parseSetCommandInput(rawArgv []string) (setCommandInput, error) {
	input := setCommandInput{
		Port:    DefaultControlPort,
		Kind:    defaultProcessorKind,
		Status:  http.StatusOK,
		Content: "text",
	}

	hasIndexedSeqFlags := false

	for i := 0; i < len(rawArgv); i++ {
		token := rawArgv[i]
		if !strings.HasPrefix(token, "--") {
			return setCommandInput{}, fmt.Errorf("unexpected argument %q", token)
		}

		if isIndexedSeqFlagToken(token) {
			hasIndexedSeqFlags = true
			if !strings.Contains(token, "=") {
				if i+1 >= len(rawArgv) {
					return setCommandInput{}, fmt.Errorf("missing value for %s", token)
				}
				i++
			}
			continue
		}

		name, inlineValue := splitLongFlagToken(token)
		value, err := parseFlagValue(name, inlineValue, rawArgv, &i)
		if err != nil {
			return setCommandInput{}, err
		}

		switch name {
		case OptPort:
			port, err := parseIntFlagValue(token, value)
			if err != nil {
				return setCommandInput{}, err
			}
			input.Port = port
		case OptKind:
			input.Kind = strings.TrimSpace(value)
		case OptEndpoint:
			input.Endpoint = strings.TrimSpace(value)
		case OptResponseStatus:
			status, err := parseIntFlagValue(token, value)
			if err != nil {
				return setCommandInput{}, err
			}
			input.Status = status
		case OptResponseContent:
			input.Content = strings.TrimSpace(value)
		case OptResponseBody:
			input.Body = value
		case OptResponseHeader:
			input.Headers = append(input.Headers, value)
		case OptScript:
			input.Script = value
		case OptScriptFile:
			input.ScriptFile = strings.TrimSpace(value)
		default:
			return setCommandInput{}, fmt.Errorf("unknown flag --%s", name)
		}
	}

	if hasIndexedSeqFlags {
		steps, err := parseSeqStepInputsFromRawArgv(rawArgv)
		if err != nil {
			return setCommandInput{}, err
		}
		input.SeqSteps = steps
	}

	return input, nil
}

func isIndexedSeqFlagToken(token string) bool {
	for _, candidate := range []string{OptResponseBody, OptResponseStatus, OptResponseContent, OptResponseHeader} {
		if strings.HasPrefix(token, "--"+candidate+"-") {
			return true
		}
	}
	return false
}

func splitLongFlagToken(token string) (name string, inlineValue string) {
	withoutPrefix := strings.TrimPrefix(token, "--")
	parts := strings.SplitN(withoutPrefix, "=", 2)
	name = parts[0]
	if len(parts) == 2 {
		inlineValue = parts[1]
	}
	return name, inlineValue
}

func parseFlagValue(flagName, inlineValue string, rawArgv []string, idx *int) (string, error) {
	if inlineValue != "" {
		return inlineValue, nil
	}

	if *idx+1 >= len(rawArgv) {
		return "", fmt.Errorf("missing value for --%s", flagName)
	}
	*idx++
	return rawArgv[*idx], nil
}

func parseIntFlagValue(token string, value string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("invalid integer value for %s: %q", token, value)
	}
	return parsed, nil
}

func validateSetInput(kind, endpoint string, status int, body string, script string) error {
	kind = strings.TrimSpace(kind)
	if kind != kindStatic && kind != kindSeq && kind != kindJS {
		return fmt.Errorf("invalid kind %q: expected %q, %q or %q", kind, kindStatic, kindSeq, kindJS)
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
