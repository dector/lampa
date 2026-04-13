package processor

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	corehttp "github.com/dector/lampa/server/core/http"
	"github.com/dector/lampa/server/core/utils/optional"
)

const defaultPassthroughTimeout = 5 * time.Second

// PassthroughReqProcessor forwards incoming request to upstream server.
type PassthroughReqProcessor struct {
	Server string
	Client *http.Client
}

func (p PassthroughReqProcessor) Process(request corehttp.HttpRequest) optional.Optional[corehttp.HttpResponse] {
	targetURL, err := composePassthroughURL(p.Server, request.Url.Path, request.Url.RawQuery)
	if err != nil {
		return optional.Some(newBadGatewayResponse())
	}

	req, err := http.NewRequest(request.Method, targetURL, bytes.NewReader(request.Body))
	if err != nil {
		return optional.Some(newBadGatewayResponse())
	}
	req.Header = request.Headers.Clone()

	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: defaultPassthroughTimeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		return optional.Some(newBadGatewayResponse())
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return optional.Some(newBadGatewayResponse())
	}

	return optional.Some(corehttp.HttpResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       responseBody,
	})
}

func composePassthroughURL(server, requestPath, rawQuery string) (string, error) {
	base := strings.TrimSpace(server)
	if base == "" {
		return "", fmt.Errorf("empty server")
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse server: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid server")
	}

	requestPath = strings.TrimSpace(requestPath)
	if requestPath == "" {
		requestPath = "/"
	}
	if strings.HasPrefix(requestPath, "/") {
		parsed.Path = path.Join(parsed.Path, requestPath)
	} else {
		parsed.Path = path.Join(parsed.Path, "/"+requestPath)
	}
	parsed.RawQuery = rawQuery

	return parsed.String(), nil
}

func newBadGatewayResponse() corehttp.HttpResponse {
	return corehttp.HttpResponse{
		StatusCode: http.StatusBadGateway,
		Headers: http.Header{
			"Content-Type": []string{"text/plain"},
		},
		Body: []byte("Bad Gateway"),
	}
}
