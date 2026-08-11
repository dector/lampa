package app

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dector/lampa/server/core/logstore"
)

const webUIRequestsLimit = 100

type webUIRequestLog struct {
	Time          string
	Status        string
	Method        string
	Path          string
	RequestBytes  int
	ResponseBytes int
	StatusClass   string
	MethodClass   string
}

func webUIStatsSnapshot(logs logstore.Store) (int, int64) {
	if logs == nil {
		return 0, 0
	}
	return logs.Count(), logs.SizeBytes()
}

func webUIRequestsSnapshot(logs logstore.Store) ([]webUIRequestLog, int) {
	if logs == nil {
		return nil, 0
	}

	count := logs.Count()
	entries := logs.Latest(webUIRequestsLimit)
	requests := make([]webUIRequestLog, 0, len(entries))
	for _, entry := range entries {
		requests = append(requests, webUIRequestLog{
			Time:          formatWebUIRequestTime(entry.Timestamp),
			Status:        formatWebUIRequestStatus(entry.Response.Status),
			Method:        entry.Request.Method,
			Path:          formatWebUIRequestPath(entry.Request),
			RequestBytes:  len(entry.Request.Body),
			ResponseBytes: len(entry.Response.Body),
			StatusClass:   webUIStatusClass(entry.Response.Status),
			MethodClass:   webUIMethodClass(entry.Request.Method),
		})
	}

	hidden := count - len(entries)
	if hidden < 0 {
		hidden = 0
	}
	return requests, hidden
}

func formatWebUIRequestTime(t time.Time) string {
	if t.IsZero() {
		return "--:--:--"
	}
	return t.Local().Format("15:04:05")
}

func formatWebUIRequestStatus(status int) string {
	if status <= 0 {
		return "---"
	}
	return fmt.Sprintf("%03d", status)
}

func formatWebUIRequestPath(req logstore.Request) string {
	if req.Path != "" {
		return req.Path
	}
	if req.URL == "" {
		return "/"
	}
	parsed, err := url.Parse(req.URL)
	if err != nil || parsed.Path == "" {
		return req.URL
	}
	if parsed.RawQuery != "" {
		return parsed.Path + "?" + parsed.RawQuery
	}
	return parsed.Path
}

func webUIStatusClass(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "status status-ok"
	case status >= 300 && status < 400:
		return "status status-redirect"
	case status >= 400 && status < 500:
		return "status status-warn"
	case status >= 500:
		return "status status-error"
	default:
		return "status status-muted"
	}
}

func webUIMethodClass(method string) string {
	switch strings.ToUpper(method) {
	case "GET":
		return "method method-get"
	case "POST":
		return "method method-post"
	case "PUT":
		return "method method-put"
	case "PATCH":
		return "method method-patch"
	case "DELETE":
		return "method method-delete"
	default:
		return "method method-other"
	}
}
