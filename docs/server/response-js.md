# JS request processor scripts (`--kind js`)

This page shows how JS processors work in `lampa server set --kind js` and provides runnable examples from simple to advanced.

## Quick start

Inline script:

```bash
lampa server set \
  --kind js \
  --endpoint /dynamic \
  --script 'function handle(req){ return Response.text("ok"); }'
```

Script from file:

```bash
lampa server set \
  --kind js \
  --endpoint /dynamic \
  --script-file ./handler.js
```

---

## `handle(req)` request shape

Your script must define:

```js
function handle(req) {
  // ...
}
```

`req` has this shape:

```js
{
  method: "GET" | "POST" | ...,
  url: "http://host/path?x=1" | "/path?x=1",
  headers: {
    "Header-Name": ["value1", "value2"]
  },
  body: "...raw request body as UTF-8 string..."
}
```

Notes:
- `headers` values are arrays (even when only one value exists).
- `body` is provided as a string.

---

## Response shape

`handle(req)` should return either:

1. **A response object**:

```js
{
  statusCode: 200,
  headers: {
    "Content-Type": "text/plain",
    "X-Trace": ["a", "b"]
  },
  body: "hello" // string or Uint8Array
}
```

2. **`null` / `undefined`** to mark request as unhandled (Lampa fallback handles it, usually `404 Not Found`).

Defaults when fields are omitted:
- `statusCode`: `200`
- `headers`: empty
- `body`: empty

If no `Content-Type` is set, server output defaults to `text/plain`.

---

## Built-in helper API (`Response`)

Lampa injects a `Response` helper:

- `Response.json(data, statusCode?, headers?)`
- `Response.text(data, statusCode?, headers?)`
- `Response.bytes(data, statusCode?, headers?)` (`data` can be `Uint8Array`)
- `Response.empty(statusCode?, headers?)`

All return a chainable builder with:
- `.status(code)`
- `.header(name, value)`
- `.contentType(value)`

Example:

```js
return Response
  .json({ ok: true })
  .status(201)
  .header("X-Debug", "1");
```

---

## Examples (simple → harder)

## 1) Minimal text response

```js
function handle(req) {
  return Response.text("hello from JS");
}
```

## 2) Echo request basics as JSON

```js
function handle(req) {
  return Response.json({
    method: req.method,
    url: req.url,
    hasBody: req.body.length > 0
  });
}
```

## 3) Method guard + custom status/headers

```js
function handle(req) {
  if (req.method !== "POST") {
    return Response
      .json({ error: "Method Not Allowed", expected: "POST" }, 405)
      .header("Allow", "POST");
  }

  return Response
    .json({ ok: true, received: req.body }, 200)
    .header("X-Processed-By", "quickjs");
}
```

## 4) Parse JSON body with validation

```js
function handle(req) {
  var payload;
  try {
    payload = req.body ? JSON.parse(req.body) : {};
  } catch (e) {
    return Response.json({ error: "Invalid JSON" }, 400);
  }

  if (!payload.name) {
    return Response.json({ error: "Field 'name' is required" }, 422);
  }

  return Response.json({
    ok: true,
    message: "Hello, " + payload.name + "!"
  }, 201);
}
```

## 5) Multi-route logic + CORS + binary response

```js
function handle(req) {
  // Basic CORS preflight
  if (req.method === "OPTIONS") {
    return Response.empty(204)
      .header("Access-Control-Allow-Origin", "*")
      .header("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
      .header("Access-Control-Allow-Headers", "Content-Type,Authorization");
  }

  // Route by URL path (works for absolute or relative URL strings)
  var path = req.url;
  var qIndex = path.indexOf("?");
  if (qIndex >= 0) path = path.slice(0, qIndex);

  var slashAfterHost = path.indexOf("/", path.indexOf("//") + 2);
  if (path.indexOf("http://") === 0 || path.indexOf("https://") === 0) {
    path = slashAfterHost >= 0 ? path.slice(slashAfterHost) : "/";
  }

  if (path === "/api/ping") {
    return Response.json({ ok: true, ts: Date.now() })
      .header("Access-Control-Allow-Origin", "*");
  }

  if (path === "/api/file") {
    var bytes = new Uint8Array([76, 65, 77, 80, 65, 10]); // "LAMPA\n"
    return Response
      .bytes(bytes, 200, {
        "Content-Type": "application/octet-stream",
        "Content-Disposition": "attachment; filename=sample.bin",
        "Access-Control-Allow-Origin": "*"
      });
  }

  return Response.json({ error: "Not Found", path: path }, 404)
    .header("Access-Control-Allow-Origin", "*");
}
```

---

## Troubleshooting tips

- Ensure `handle` exists and is a function.
- Return an object (or `Response.*(...)`) for handled requests.
- Return `null`/`undefined` only when you intentionally want fallback behavior.
- Keep `statusCode` numeric and `headers` values as string or string-array.
- For JSON APIs, prefer `Response.json(...)` to set `Content-Type` automatically.
