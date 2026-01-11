# URL Shortener (Go) — Code Review

Date: 2026-01-10

This review focuses on Go + backend best practices for the current implementation.

## Quick architecture map

- Entry point: `cmd/main.go`
- HTTP API: `handlers/http.go` (Gin)
- Business logic: `service/hasher.go`
- Persistence: `model/sqlite-repo.go` (SQLite) + `model/model.go` (types + repo interface)

Request flow:

`POST /` → `handlers.URLHandler.PostURL` → `service.GetHash` → `repo.Create`

`GET /:key` → `handlers.URLHandler.GetURL` → `repo.Retrieve` → redirect

`DELETE /:key` → `handlers.URLHandler.DeleteURL` → `repo.Delete`

## What’s good already

- Clear separation of concerns: handler/service/repository split is a good learning-friendly structure.
- SQL uses placeholders (`?`), avoiding SQL injection via string concatenation.
- DB schema enforces `key` uniqueness — a valuable invariant.
- Simple curl-based smoke script (`test.sh`) is a great learning tool.

## High-impact improvements (recommended)

### 1) Key generation: collision handling + better entropy

Current behavior in `service.GetHash`:

- It appends a random string to the input and hashes it, then truncates to 6 hex chars.
- Because it’s random, the *same long URL* posted twice will almost always get *different* keys. That may be fine, but it’s a design choice.

Concerns:

- 6 hex chars means only `16^6 ≈ 16.7M` possible keys.
- Collisions become realistic as the dataset grows.
- The handler currently treats any insert error as a 500; collisions should instead trigger retries.

Recommendations:

- If you want random short codes: generate them directly using `crypto/rand` (preferred), and retry on collision.
- If you want deterministic mapping: hash the URL without randomness (or store a canonical mapping and return existing key).

Suggested behavior:

- On insert, if the error is a uniqueness violation, retry generating a new key up to N times.
- Only then return 500.

### 2) URL validation & normalization

Right now, you accept any string and later redirect to it.

Recommendations:

- Parse with `net/url`.
- Require scheme `http` or `https`.
- Optionally normalize (lowercase scheme/host; trim spaces).

This protects against surprising schemes like `javascript:` and improves data quality.

### 3) HTTP semantics: status codes + handler flow correctness

- Redirect: currently uses `302 Found`. Many shorteners use `301 Moved Permanently` or `302` depending on intent.
  - Pick explicitly and document why.

- Delete: currently returns `202 Accepted` on success, but deletion is synchronous.
  - Prefer `204 No Content` on success.

- Bug: in `DeleteURL`, if the repo returns an error you write a 500 but don’t `return`, so the handler may also write 404/202 after that.

### 4) Context/timeouts for DB calls

Repo uses `Exec` / `QueryRow` without context.

Recommendations:

- Thread request context through:
  - Use `ExecContext` and `QueryRowContext`.
  - Pass `c.Request.Context()` from Gin into repo calls.

This prevents stuck requests from lingering indefinitely and is common practice in Go backends.

### 5) Observability & error quality

- In `cmd/main.go`, `log.Fatal("Can't open SQLite DB")` discards the actual error.
  - Prefer `log.Fatalf("open db: %v", err)`.

- Consider consistent JSON error responses.
  - Even if you keep plain text for now, returning a structured error is friendlier to API clients.

### 6) Config + graceful shutdown

Hard-coded values are fine for learning, but typical backend practice is:

- Read port/DB path from env or flags.
- Use an `http.Server` with timeouts and handle SIGINT/SIGTERM with `Shutdown(ctx)` for graceful termination.

## Smaller Go-idiom improvements

- In `cmd/main.go`, `handlers := &handlers.URLHandler{...}` shadows the imported `handlers` package name.
  - Rename the variable to `h` or `urlHandler`.

- Consider renaming the DB column `key` → `short_key` (optional).
  - It’s a common word and can be confusing when integrating with other tools/DBs.

## Testing suggestions (incremental)

You already have `test.sh` for manual testing.

Next small step:

- Add unit tests using `net/http/httptest`:
  - POST validation (missing `url`)
  - GET returns 404 when missing
  - GET returns redirect when present

- Add repo tests using SQLite `:memory:`:
  - Create + retrieve + delete
  - Uniqueness behavior

## Suggested “minimal patch set” (if you want me to implement)

Without changing the public API (routes/JSON):

1. Fix `DeleteURL` to return immediately on repo error.
2. Add URL parsing + scheme validation in `PostURL`.
3. Retry on insert collision (unique key violation), generating new codes.
4. Rename the `handlers := ...` variable in `cmd/main.go` to avoid shadowing.
5. (Optional) Use request context + `ExecContext`/`QueryRowContext` for DB calls.

---

If you tell me whether you want **random** short codes (always new) or **deterministic** (same URL → same key), I can implement the right version cleanly.
