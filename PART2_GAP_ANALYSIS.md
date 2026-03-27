# Assignment Part 2 — Feature Analysis

> What's already implemented vs what needs to be added.

---

## ✅ Already Implemented (Part 2 extras beyond baseline)

| Feature | Where | Notes |
|---|---|---|
| **JWT Authentication** | `internal/auth/jwt.go`, `middleware.go` | Custom HMAC-SHA256 — no external JWT lib |
| **Middleware pipeline** | `internal/server/middleware.go` | Logging, CORS, Auth — composable `Chain()` |
| **Database integration** | `internal/database/` | PostgreSQL with `pgxpool`, migrations, repos |
| **Unit Testing** | `internal/auth/auth_test.go` | 4 tests: bcrypt, persistence, JWT roundtrip |
| **Logging system** | `middleware.go` `LoggingMiddleware` | Method, path, latency per request |
| **File upload** | `handlers.go` `HandleUploadTrack` | Multipart, format validation, DB insert |

---

## ❌ Not Yet Implemented (recommended to add)

The following are explicitly listed in the Part 2 description and are **missing**:

### 1. 🚦 Rate Limiting
**What:** Limit requests per IP to prevent abuse (e.g., brute-force login attempts).
**Go approach:** Token bucket per IP using a `sync.Map` of counters + `time.Ticker` cleanup goroutine.
**Marks impact:** Directly mentioned in the spec, demonstrates concurrency + system design.

---

### 2. 👷 Worker Pool
**What:** A bounded pool of goroutines to process library scan jobs concurrently (instead of sequential loop).
**Go approach:** Channel-based worker pool — N workers reading from a `jobs chan Track` and writing results to a `results chan`.
**Marks impact:** Directly mentioned, demonstrates advanced goroutine patterns beyond basic `go func()`.

---

### 3. 🔤 Advanced String Manipulation
**What:** The spec specifically calls for *parsing, validation, transformation*.
**Currently missing:** No input validation beyond empty checks — no username format rules, no email validation, no string parsing beyond basic `strings.Split`.
**Suggested implementation:**
- Username validation: alphanumeric + underscore, 3–20 chars (`regexp`)
- Search query sanitization and normalization (trim, lowercase, multi-word tokenization)
- Track title/artist cleanup from filenames (strip underscores/hyphens, title-case)

---

### 4. 📊 Mathematical / Statistical Computations
**What:** The spec calls for *statistical calculations, numeric processing, or algorithmic operations*.
**Currently missing entirely.**
**Suggested implementation:**
- Library statistics endpoint: `GET /api/library/stats`
  - Total tracks, total duration, average track length
  - Most common genre, tracks per album distribution
  - These require numeric aggregation and basic statistical ops (mean, mode)

---

## Recommendation: Implement in this order

| Priority | Feature | Effort | Marks Value |
|---|---|---|---|
| 1 | **Rate limiting** middleware | Low (~50 lines) | High — directly listed |
| 2 | **String validation** (username, search) | Low (~30 lines) | High — directly listed |
| 3 | **Library stats** endpoint (math) | Medium (~60 lines) | High — directly listed |
| 4 | **Worker pool** for library scanning | Medium (~80 lines) | High — directly listed |

All four can be implemented in a single session. Want me to implement them?
