package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

const (
	OptPort            = "port"
	DefaultControlPort = 8081
)


func createPingCommand() *cli.Command {
	return &cli.Command{
		Name:  "ping",
		Usage: "check control server health",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  OptPort,
				Usage: "control server port",
				Value: DefaultControlPort,
			},
		},
		Action: CmdActionPing,
	}
}

func CmdActionPing(ctx context.Context, c *cli.Command) error {
	port := c.Int(OptPort)
	if err := validatePort(port); err != nil {
		return err
	}

	url := buildPingURL(port)
	client := &http.Client{Timeout: 5 * time.Second}

	response, err := pingControl(ctx, client, url)
	if err != nil {
		return err
	}

	formatted, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format ping response JSON: %w", err)
	}

	fmt.Printf(">> %s\n", url)
	fmt.Printf("<< %s\n", string(formatted))
	return nil
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", port)
	}
	return nil
}

func buildPingURL(port int) string {
	return fmt.Sprintf("http://localhost:%d/ping", port)
}

func pingControl(ctx context.Context, client *http.Client, url string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build ping request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call control server at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		body := strings.TrimSpace(string(payload))
		if body == "" {
			return nil, fmt.Errorf("control server returned HTTP %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("control server returned HTTP %d: %s", resp.StatusCode, body)
	}

	var data map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("invalid ping JSON response: %w", err)
	}

	status, _ := data["status"].(string)
	if status != "ok" {
		return nil, fmt.Errorf("unexpected ping status: %q", status)
	}

	return data, nil
}
