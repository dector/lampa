package app

import (
	"fmt"
	"net/http"

	"github.com/dector/lampa/server/core/logstore"
)

// NewWebUIIndexHandler returns the Web UI landing page handler.
func NewWebUIIndexHandler(cfg ServerConfig, logs logstore.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		trafficCount := 0
		trafficSize := int64(0)
		if logs != nil {
			trafficCount = logs.Count()
			trafficSize = logs.SizeBytes()
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Lampa Web UI</title>
  <style>
    :root { color-scheme: dark; }
    body {
      background: #121212;
      color: #e7e2dc;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      line-height: 1.5;
      margin: 0;
      padding: 2rem;
    }
    main {
      background: #1a1a1a;
      border: 1px solid #2a2a2a;
      box-shadow: none;
      margin: 0 auto;
      max-width: 44rem;
      padding: 1.5rem;
    }
    h1 { margin: 0 0 1rem; }
    dl { display: grid; grid-template-columns: max-content 1fr; gap: .75rem 1rem; margin: 1.5rem 0 0; }
    dt { color: #a8a29e; font-weight: 700; }
    dd { margin: 0; }
    .status-row { align-items: center; display: flex; gap: .5rem; }
    .pill {
      align-items: center;
      background: #052e16;
      border: 1px solid #16a34a;
      color: #86efac;
      display: inline-flex;
      font-size: .75rem;
      font-weight: 700;
      gap: .3rem;
      padding: .15rem .45rem;
    }
    .pill::before {
      background: #22c55e;
      content: "";
      height: .4rem;
      width: .4rem;
    }
  </style>
</head>
<body>
  <main>
    <h1>Lampa Web UI</h1>
    <div class="status-row">Server: <span class="pill">running</span></div>
    <dl>
      <dt>Proxy port</dt><dd>%d</dd>
      <dt>Control port</dt><dd>%d</dd>
      <dt>Traffic count</dt><dd>%d</dd>
      <dt>Traffic size bytes</dt><dd>%d</dd>
    </dl>
  </main>
</body>
</html>
`, cfg.ProxyPort, cfg.ControlPort, trafficCount, trafficSize)
	}
}

// NewWebUIHealthHandler returns the Web UI health-check handler.
func NewWebUIHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
