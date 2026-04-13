# Server CLI (LLM Agent Quick Reference)

Use these commands when the Lampa server is already running.

## Purpose
- Check control plane health.
- Configure static, sequence, or JS processors for proxy endpoints at runtime.
- Fetch latest in-memory proxy request/response logs.

## Defaults
- Control host: `localhost`
- Control port: `8081`
- Command namespace: `lampa server ...` and `lampa proxy ...`

---

## 1) Health check

```bash
lampa server ping
```

Custom control port:

```bash
lampa server ping --port 46899
```

Expected: JSON with `status: "ok"`.

---

## 2) Set proxy response/processor

Two equivalent forms are supported:

```bash
lampa server set ...
# or
lampa server proxy set ...
```

### Minimal static

```bash
lampa server set \
  --endpoint /hello \
  --response.body 'hello'
```

### Static JSON response

```bash
lampa server proxy set \
  --endpoint /example \
  --response.status 200 \
  --response.content json \
  --response.body '{"ok":true}'
```

### Static with custom headers

```bash
lampa server proxy set \
  --endpoint /example \
  --response.body 'ok' \
  --response.header 'X-Debug:1' \
  --response.header 'Cache-Control:no-store'
```

### Static Content-Type override
If `Content-Type` is passed explicitly, it overrides `--response.content` preset.

```bash
lampa server set \
  --endpoint /example \
  --response.content json \
  --response.header 'Content-Type:text/plain' \
  --response.body 'forced text'
```

### Sequence processor (`--kind seq`)

Use indexed flags with 1-based step suffixes:
- `--response.body-N` (required for each step)
- `--response.status-N` (optional, default `200`)
- `--response.content-N` (optional, default `text`)
- `--response.header-N` (optional, repeatable)

Minimal sequence:

```bash
lampa server set \
  --kind seq \
  --endpoint /flaky \
  --response.body-1 'temporary error' \
  --response.body-2 'recovered'
```

Full sequence with per-step status/content/headers:

```bash
lampa server set \
  --kind seq \
  --endpoint /flaky \
  --response.status-1 500 \
  --response.content-1 text \
  --response.body-1 'fail once' \
  --response.header-1 'X-Step:1' \
  --response.status-2 200 \
  --response.content-2 json \
  --response.body-2 '{"ok":true}' \
  --response.header-2 'Cache-Control:no-store'
```

Mixed defaults + explicit override:

```bash
lampa server set \
  --kind seq \
  --endpoint /hello \
  --response.body-1 'hello' \
  --response.status-2 201 \
  --response.body-2 'created'
```

Sequence constraints:
- step index must be numeric and start at `1`
- indexes must be contiguous (`1,2,3...`; no gaps)
- if step `N` is declared, `--response.body-N` must be provided
- content presets (`json|text|html|raw`) and Content-Type override behavior are the same as static mode, but applied per step

### JS processor

```bash
lampa server set \
  --kind js \
  --endpoint /dynamic \
  --script 'function handle(req){ return Response.json({ path: req.url, ok: true }); }'
```

From file:

```bash
lampa server set \
  --kind js \
  --endpoint /dynamic \
  --script-file ./handler.js
```

More JS request-processing script samples: [`docs/server/response-js.md`](./server/response-js.md).

---

## 3) Set default fallback proxy processor (passthrough)

Use this to forward any unmatched endpoint to an upstream server.

```bash
lampa server proxy set-default \
  --kind pass \
  --server http://localhost:3000
```

Custom control port:

```bash
lampa server proxy set-default \
  --port 46899 \
  --kind pass \
  --server http://localhost:3000
```

---

## 4) Get proxy logs

Fetch latest N entries (latest first):

```bash
lampa proxy logs get-all -n 20
```

Custom control port:

```bash
lampa proxy logs get-all --port 46899 -n 50
```

Control API used by the CLI:

```text
GET /api/v0/proxy/logs?n=N
```

Response includes:
- `status`
- `count` (returned entries)
- `total` (entries currently stored)
- `sizeBytes` (approximate retained memory size)
- `entries` (latest-first)

---

## Flags

Required:
- `--endpoint` (must start with `/`)
- for `--kind static`: `--response.body`
- for `--kind js`: `--script` or `--script-file`

Optional:
- `--port` (default `8081`)
- `-n` (for `lampa proxy logs get-all`, must be `>0`)
- `--kind` (`static|seq|js`, default `static`) for `set`/`proxy set`
- `--response.status` (default `200`, valid `100..599`, static only)
- `--response.content` (`json|text|html|raw`, default `text`, static only)
- `--response.header` (repeatable `Name:Value`, static only)
- for `--kind seq`, use indexed flags only:
  - `--response.body-N` (required per step)
  - `--response.status-N` (optional, default `200`)
  - `--response.content-N` (optional, default `text`)
  - `--response.header-N` (repeatable `Name:Value`)
  - `N` is 1-based and must not contain gaps
- `--script` (inline JS, js only)
- `--script-file` (path to JS file, js only)
- `--kind pass` for `proxy set-default`
- `--server <url>` for `proxy set-default` (required)

Content presets:
- `json` -> `application/json`
- `text` -> `text/plain`
- `html` -> `text/html`
- `raw` -> no auto content type

---

## Agent-safe usage pattern

1. Verify control API first:
```bash
lampa server ping
```
2. Apply endpoint response/processor:
```bash
lampa server proxy set --endpoint /x --response.body '...'
# or JS:
# lampa server set --kind js --endpoint /x --script 'function handle(req){ ... }'
# or set default passthrough:
# lampa server proxy set-default --kind pass --server http://localhost:3000
```
3. If needed, use `-v` to print full control response JSON:
```bash
lampa -v server proxy set --endpoint /x --response.body '...'
```

---

## Common errors (especially for `--kind seq`)
- `invalid kind`: check `--kind` value (`static|seq|js`).
- `missing response body`: in seq mode, each declared step needs `--response.body-N`.
- `invalid response header`: header must be `Name:Value`.
- `invalid status`: must be in `100..599` (per step in seq mode).
- sequence index errors: use numeric suffixes, start at `1`, and keep indexes contiguous.

## Notes
- Runtime config is **in-memory** (not persisted across restarts).
- Proxy logs are **in-memory** with a 30MB cap and oldest-first eviction.
- Re-setting same endpoint replaces previous config (upsert behavior).
