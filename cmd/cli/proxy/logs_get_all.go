package proxy

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
	optPort            = "port"
	optN               = "n"
	defaultControlPort = 8081
)

func createGetAllCommand() *cli.Command {
	return &cli.Command{
		Name:  "get-all",
		Usage: "get latest proxy request/response logs",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  optPort,
				Usage: "control server port",
				Value: defaultControlPort,
			},
			&cli.IntFlag{
				Name:  optN,
				Usage: "number of latest log entries to fetch",
				Value: 10,
			},
		},
		Action: CmdActionGetAll,
	}
}

func CmdActionGetAll(ctx context.Context, c *cli.Command) error {
	port := c.Int(optPort)
	if err := validatePort(port); err != nil {
		return err
	}

	n := c.Int(optN)
	if err := validateN(n); err != nil {
		return err
	}

	requestURL := buildProxyLogsURL(port, n)
	client := &http.Client{Timeout: 5 * time.Second}

	response, err := getProxyLogs(ctx, client, requestURL)
	if err != nil {
		return err
	}

	formatted, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format proxy logs JSON: %w", err)
	}

	fmt.Printf(">> %s\n", requestURL)
	fmt.Printf("<< %s\n", string(formatted))
	return nil
}

func validatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", port)
	}
	return nil
}

func validateN(n int) error {
	if n <= 0 {
		return fmt.Errorf("invalid n %d: must be greater than 0", n)
	}
	return nil
}

func buildProxyLogsURL(port int, n int) string {
	return fmt.Sprintf("http://localhost:%d/api/v0/proxy/logs?n=%d", port, n)
}

func getProxyLogs(ctx context.Context, client *http.Client, requestURL string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build proxy logs request: %w", err)
	}

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
		return nil, fmt.Errorf("invalid proxy logs JSON response: %w", err)
	}

	status, _ := response["status"].(string)
	if status != "ok" {
		return nil, fmt.Errorf("unexpected proxy logs status: %q", status)
	}

	return response, nil
}
