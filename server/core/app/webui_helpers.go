package app

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/dector/lampa/server/core/logstore"
)

const webUIRequestsLimit = 100

type webUIRequestLog struct {
	ID            int64
	Href          string
	Time          string
	Status        string
	Method        string
	Path          string
	RequestBytes  int
	ResponseBytes int
	StatusClass   string
	MethodClass   string
}

type webUITrafficDetail struct {
	ID                int64
	Time              string
	Duration          string
	Status            string
	Method            string
	URL               string
	Path              string
	RequestHeaders    []webUIHeader
	RequestBody       string
	RequestTruncated  bool
	ResponseHeaders   []webUIHeader
	ResponseBody      string
	ResponseTruncated bool
	Error             string
}

type webUIHeader struct {
	Name  string
	Value string
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
			ID:            entry.ID,
			Href:          formatWebUITrafficDetailHref(entry.ID),
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

func webUITrafficDetailSnapshot(logs logstore.Store, id int64) (webUITrafficDetail, bool) {
	if logs == nil {
		return webUITrafficDetail{}, false
	}

	entry, ok := logs.Get(id)
	if !ok {
		return webUITrafficDetail{}, false
	}

	return webUITrafficDetail{
		ID:                entry.ID,
		Time:              formatWebUITrafficDetailTime(entry.Timestamp),
		Duration:          formatWebUITrafficDuration(entry.DurationMs),
		Status:            formatWebUIRequestStatus(entry.Response.Status),
		Method:            entry.Request.Method,
		URL:               entry.Request.URL,
		Path:              formatWebUIRequestPath(entry.Request),
		RequestHeaders:    formatWebUIHeaders(entry.Request.Headers),
		RequestBody:       string(entry.Request.Body),
		RequestTruncated:  entry.RequestTruncated,
		ResponseHeaders:   formatWebUIHeaders(entry.Response.Headers),
		ResponseBody:      string(entry.Response.Body),
		ResponseTruncated: entry.ResponseTruncated,
		Error:             entry.Error,
	}, true
}

func formatWebUIRequestTime(t time.Time) string {
	if t.IsZero() {
		return "--:--:--"
	}
	return t.Local().Format("15:04:05")
}

func formatWebUITrafficDetailTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format(time.RFC3339)
}

func formatWebUITrafficDuration(durationMs int64) string {
	if durationMs <= 0 {
		return ""
	}
	return fmt.Sprintf("%dms", durationMs)
}

func formatWebUITrafficDetailHref(id int64) string {
	return fmt.Sprintf("/traffic/%d", id)
}

func formatWebUIHeaders(headers http.Header) []webUIHeader {
	if len(headers) == 0 {
		return nil
	}

	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]webUIHeader, 0, len(names))
	for _, name := range names {
		out = append(out, webUIHeader{Name: name, Value: strings.Join(headers.Values(name), ", ")})
	}
	return out
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
