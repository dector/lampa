# Server CLI (LLM Agent Quick Reference)

Use these commands when the Lampa server is already running.

## Purpose
- Check control plane health.
- Configure static or JS processors for proxy endpoints at runtime.

## Defaults
- Control host: `localhost`
- Control port: `8081`
- Command namespace: `lampa server ...`

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

## Flags

Required:
- `--endpoint` (must start with `/`)
- for `--kind static`: `--response.body`
- for `--kind js`: `--script` or `--script-file`

Optional:
- `--port` (default `8081`)
- `--kind` (`static|js`, default `static`) for `set`/`proxy set`
- `--response.status` (default `200`, valid `100..599`, static only)
- `--response.content` (`json|text|html|raw`, default `text`, static only)
- `--response.header` (repeatable `Name:Value`, static only)
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

## Notes
- Runtime config is **in-memory** (not persisted across restarts).
- Re-setting same endpoint replaces previous config (upsert behavior).
