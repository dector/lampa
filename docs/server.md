# Server CLI (LLM Agent Quick Reference)

Use these commands when the Lampa server is already running.

## Purpose
- Check control plane health.
- Configure static responses for proxy endpoints at runtime.

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

## 2) Set static proxy response

Two equivalent forms are supported:

```bash
lampa server set ...
# or
lampa server proxy set ...
```

### Minimal

```bash
lampa server set \
  --endpoint /hello \
  --response.body 'hello'
```

### JSON response

```bash
lampa server proxy set \
  --endpoint /example \
  --response.status 200 \
  --response.content json \
  --response.body '{"ok":true}'
```

### With custom headers

```bash
lampa server proxy set \
  --endpoint /example \
  --response.body 'ok' \
  --response.header 'X-Debug:1' \
  --response.header 'Cache-Control:no-store'
```

### Content-Type override
If `Content-Type` is passed explicitly, it overrides `--response.content` preset.

```bash
lampa server set \
  --endpoint /example \
  --response.content json \
  --response.header 'Content-Type:text/plain' \
  --response.body 'forced text'
```

---

## Flags

Required:
- `--endpoint` (must start with `/`)
- `--response.body`

Optional:
- `--port` (default `8081`)
- `--kind` (currently only `static`)
- `--response.status` (default `200`, valid `100..599`)
- `--response.content` (`json|text|html|raw`, default `text`)
- `--response.header` (repeatable `Name:Value`)

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
2. Apply endpoint response:
```bash
lampa server proxy set --endpoint /x --response.body '...'
```
3. If needed, use `-v` to print full control response JSON:
```bash
lampa -v server proxy set --endpoint /x --response.body '...'
```

---

## Notes
- Runtime config is **in-memory** (not persisted across restarts).
- Re-setting same endpoint replaces previous config (upsert behavior).
