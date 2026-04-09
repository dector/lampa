package quickjsexec

import corehttp "github.com/dector/lampa/server/core/http"

// ToRequestValue maps HttpRequest to a JavaScript-friendly value.
//
// Body is passed as a UTF-8 string.
func ToRequestValue(request corehttp.HttpRequest) map[string]any {
	headers := make(map[string][]string, len(request.Headers))
	for name, values := range request.Headers {
		items := make([]string, 0, len(values))
		items = append(items, values...)
		headers[name] = items
	}

	return map[string]any{
		"method":  request.Method,
		"url":     request.Url.String(),
		"headers": headers,
		"body":    string(request.Body),
	}
}
