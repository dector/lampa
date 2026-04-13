package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	. "github.com/dector/lampa/internal/globals"
	"github.com/urfave/cli/v3"
)

const (
	OptServer = "server"

	defaultProcessorKindPass = "pass"
)

type procSetDefaultRequest struct {
	Kind   string `json:"kind"`
	Server string `json:"server"`
}

func createSetDefaultCommand() *cli.Command {
	return &cli.Command{
		Name:  "set-default",
		Usage: "set default/fallback processor",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  OptPort,
				Usage: "control server port",
				Value: DefaultControlPort,
			},
			&cli.StringFlag{
				Name:  OptKind,
				Usage: "processor kind (pass)",
				Value: defaultProcessorKindPass,
			},
			&cli.StringFlag{
				Name:     OptServer,
				Usage:    "upstream server URL for passthrough processor",
				Required: true,
			},
		},
		Action: CmdActionSetDefault,
	}
}

func CmdActionSetDefault(ctx context.Context, c *cli.Command) error {
	payload, err := buildSetDefaultRequestFromCommand(c)
	if err != nil {
		return err
	}

	requestURL := buildProcSetDefaultURL(c.Int(OptPort))
	client := &http.Client{Timeout: 5 * time.Second}

	response, err := setDefaultControl(ctx, client, requestURL, payload)
	if err != nil {
		return err
	}

	fmt.Printf("ok: set default (%s) -> %s\n", payload.Kind, payload.Server)
	if G.Verbosity >= VerbosityInfo {
		formatted, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format set-default response JSON: %w", err)
		}
		fmt.Printf("<< %s\n", string(formatted))
	}
	return nil
}

func buildSetDefaultRequestFromCommand(c *cli.Command) (procSetDefaultRequest, error) {
	port := c.Int(OptPort)
	if err := validatePort(port); err != nil {
		return procSetDefaultRequest{}, err
	}

	kind := strings.TrimSpace(c.String(OptKind))
	server := strings.TrimSpace(c.String(OptServer))
	if err := validateSetDefaultInput(kind, server); err != nil {
		return procSetDefaultRequest{}, err
	}

	return procSetDefaultRequest{Kind: kind, Server: server}, nil
}

func validateSetDefaultInput(kind string, server string) error {
	if strings.TrimSpace(kind) != defaultProcessorKindPass {
		return fmt.Errorf("invalid kind %q: expected %q", kind, defaultProcessorKindPass)
	}

	parsedServer, err := url.Parse(strings.TrimSpace(server))
	if err != nil || parsedServer.Scheme == "" || parsedServer.Host == "" {
		return fmt.Errorf("invalid server %q: expected full URL with scheme and host", server)
	}

	return nil
}

func buildProcSetDefaultURL(port int) string {
	return fmt.Sprintf("http://localhost:%d/api/v0/proc/default/set", port)
}

func setDefaultControl(ctx context.Context, client *http.Client, requestURL string, payload procSetDefaultRequest) (map[string]any, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to encode set-default request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("failed to build set-default request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call control server at %s: %w", requestURL, err)
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
		return nil, fmt.Errorf("invalid set-default JSON response: %w", err)
	}

	status, _ := response["status"].(string)
	if status != "ok" {
		return nil, fmt.Errorf("unexpected set-default status: %q", status)
	}

	return response, nil
}
