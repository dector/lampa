package http

import nethttp "net/http"

func requestScheme(r *nethttp.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func firstHeaderValues(h nethttp.Header) map[string]string {
	headers := make(map[string]string, len(h))
	for key, values := range h {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}
