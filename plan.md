# Request/Response In-Memory Logs — Implementation Plan

## Goal
Add runtime in-memory request/response logging to `server/core` with:
1. bounded memory usage (max **30 MB**),
2. eviction when limit is reached (keep most recent entries),
3. control endpoint to fetch latest **N** log entries,
4. CLI command: `lampa proxy logs get-all -n N`.

---

## Scope
### In scope
- Logging each proxied request/response pair in memory.
- Shared log storage available to both proxy and control servers.
- New control API to read latest N entries.
- New CLI command to fetch and print these entries.
- Unit/integration tests for buffer policy, endpoint contract, and CLI behavior.

### Out of scope (for this task)
- Persistent storage on disk.
- Filtering/searching logs by endpoint/status/etc.
- Real-time streaming (SSE/WebSocket).

---

## Proposed Design

## 1) Log model
Define a log entry struct in `server/core` (new package, e.g. `server/core/logstore`):
- `id` (monotonic sequence)
- `timestamp` (UTC)
- request fields: method, url/path, headers, body (not truncated)
- response fields: status, headers, body (not truncated)
- optional metadata: durationMs, error string
- `sizeBytes` (estimated in-memory size used for cap accounting)

### Size accounting strategy
- Use a deterministic approximation function per entry (all strings + body byte slices + metadata).
- Keep cap at `30 * 1024 * 1024` bytes.

### Eviction policy
- Append new entries.
- If total exceeds 30MB, evict from the **oldest** side until under cap.
- If a single entry is larger than cap:
  - either truncate request/response bodies before storing,
  - or drop entry with marker (final behavior decided during implementation; prefer truncation with flags).

---

## 2) Thread-safe in-memory buffer
Implement bounded buffer with:
- mutex protection,
- queue/slice of entries,
- running `totalBytes`,
- methods:
  - `Add(entry)`
  - `Latest(n int) []Entry`
  - `Count() int`
  - `SizeBytes() int64`

Ensure non-blocking behavior is reasonable (small critical sections).

---

## 3) Proxy integration
Integrate logging in request handler path (`server/core/app/handler.go`):
- capture incoming request (before processing),
- capture outgoing response (after processor result),
- push entry to shared log store.

Implementation detail:
- Extend handler/mux constructors to optionally accept log store dependency (similar to processor store sharing).
- Maintain backward compatibility for existing constructor signatures via wrappers/default nil behavior.

---

## 4) Control API endpoint
Add route constant + handler in `server/core/app`:
- route proposal: `GET /api/v0/proxy/logs`
- query param: `n` (latest N entries)
- default N (e.g. 10), max N guard (e.g. 1000)

Response proposal:
```json
{
  "status": "ok",
  "count": 20,
  "total": 128,
  "sizeBytes": 1048576,
  "entries": [ ...latest first... ]
}
```

Error handling:
- invalid `n` -> 400 with control error JSON.
- no log store configured -> 500.

---

## 5) CLI command
Implement top-level command chain required by request:
- `lampa proxy logs get-all -n N`

CLI behavior:
- call control endpoint `GET /api/v0/proxy/logs?n=N` on control port (default 8081).
- print JSON pretty output.
- support `--port` override.
- validate `-n` (>0).

Potential command organization:
- add new top-level `proxy` command in `cmd/cli`.
- keep existing `lampa server proxy ...` commands unchanged.
- optionally add compatibility alias later if needed.

---

## 6) Testing plan

### Core log buffer tests
- add under `server/core/logstore/*_test.go`:
  - stores entries under cap,
  - evicts oldest when exceeding cap,
  - handles oversized entry behavior,
  - returns latest N in correct order,
  - concurrency safety smoke test.

### Control handler tests
- add/extend `server/core/app/control_handler_test.go` and `mux_test.go`:
  - route is wired,
  - valid `n` response shape,
  - invalid `n` returns 400,
  - respects latest-first ordering.

### Proxy integration tests
- extend existing handler tests:
  - request processed + log entry appears,
  - logged fields match request/response basics.

### CLI tests
- add tests under `cmd/cli/proxy/*_test.go`:
  - URL creation,
  - argument validation,
  - success and non-200 control responses.

---

## Milestones

## Milestone 1 — Foundations (Data model + bounded store) ✅ Done
**Deliverables**
- New in-memory log store package.
- Entry model and size accounting.
- Eviction implementation with 30MB cap.
- Unit tests for cap/eviction/latest behavior.

**Exit criteria**
- All new store tests pass.

## Milestone 2 — Server integration (capture logs) ✅ Done
**Deliverables**
- Handler/mux wiring updated to accept shared log store.
- Request/response logging on proxy path.
- Integration tests for recording behavior.

**Exit criteria**
- Existing handler tests still pass.
- New integration tests verify entries are recorded.

## Milestone 3 — Control endpoint (read latest N) ✅ Done
**Deliverables**
- New control route + handler.
- Input validation for `n`.
- JSON response contract with metadata.
- Handler/mux tests.

**Exit criteria**
- Endpoint returns expected payload and error codes.

## Milestone 4 — CLI command (`lampa proxy logs get-all`) ✅ Done
**Deliverables**
- New top-level CLI command tree: `proxy logs get-all`.
- `-n` and `--port` flags.
- HTTP call + formatted output.
- CLI tests.

**Exit criteria**
- Command works against running control server.

## Milestone 5 — Docs + polish ✅ Done
**Deliverables**
- Update `README.md` and `docs/server.md` with logs API/CLI examples.
- Add notes on memory cap, eviction, and limitations.

**Exit criteria**
- Documentation reflects final API and CLI syntax.

## Milestone 6 — Final validation ✅ Done
**Deliverables**
- Run full test suite (root + `server/core`).
- Manual smoke run:
  - send proxy traffic,
  - fetch logs via control endpoint,
  - fetch logs via CLI.

**Exit criteria**
- No regressions, command and endpoint verified end-to-end.

---

## Risks / Decisions to lock early
- Exact serialization for headers/bodies in log entries.
- Oversized single-entry policy (truncate vs drop).
- Entry ordering in API response (latest-first recommended).
- Whether to include full body by default or safe truncation limit.

---

## Suggested implementation order
1. Milestone 1
2. Milestone 2
3. Milestone 3
4. Milestone 4
5. Milestone 5
6. Milestone 6
