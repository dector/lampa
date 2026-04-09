package http

import nethttp "net/http"

// HttpRequest contains the most common and important request metadata.
type HttpRequest struct {
	Method        string              `json:"method"`
	Scheme        string              `json:"scheme"`
	Host          string              `json:"host"`
	Path          string              `json:"path"`
	RawQuery      string              `json:"rawQuery"`
	Query         map[string][]string `json:"query"`
	Headers       map[string]string   `json:"headers"`
	RemoteAddr    string              `json:"remoteAddr"`
	UserAgent     string              `json:"userAgent"`
	Referer       string              `json:"referer"`
	ContentType   string              `json:"contentType"`
	ContentLength int64               `json:"contentLength"`
	Protocol      string              `json:"protocol"`
}

// NewHttpRequest builds a concise request description from http.Request.
func NewHttpRequest(r *nethttp.Request) HttpRequest {
	return HttpRequest{
		Method:        r.Method,
		Scheme:        requestScheme(r),
		Host:          r.Host,
		Path:          r.URL.Path,
		RawQuery:      r.URL.RawQuery,
		Query:         r.URL.Query(),
		Headers:       firstHeaderValues(r.Header),
		RemoteAddr:    r.RemoteAddr,
		UserAgent:     r.UserAgent(),
		Referer:       r.Referer(),
		ContentType:   r.Header.Get("Content-Type"),
		ContentLength: r.ContentLength,
		Protocol:      r.Proto,
	}
}
