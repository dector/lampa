package quickjsexec

import (
	nethttp "net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"

	corehttp "github.com/dector/lampa/server/core/http"
)

func TestToRequestValue(t *testing.T) {
	request := corehttp.HttpRequest{
		Method: "POST",
		Url: url.URL{
			Scheme:   "https",
			Host:     "example.com",
			Path:     "/api/items",
			RawQuery: "a=1",
		},
		Headers: nethttp.Header{
			"X-Test": {"v1", "v2"},
		},
		Body: []byte("hello"),
	}

	got := ToRequestValue(request)
	want := map[string]any{
		"method": "POST",
		"url":    "https://example.com/api/items?a=1",
		"headers": map[string][]string{
			"X-Test": {"v1", "v2"},
		},
		"body": "hello",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("ToRequestValue() mismatch (-want +got):\n%s", diff)
	}
}
