# Assignment Part 2 — Feature Analysis

> Based on the assignment spec: each contributor implements **2 additional features**.

---

## ✅ Your 2 Features (Already Implemented)

### 1. Database Integration — PostgreSQL

**Where:** `internal/database/`

- `pgxpool` connection pool (10 max connections), auto-ping health check
- Idempotent schema migrations on startup (`users`, `tracks`, `playlists`, `playlist_tracks`)
- **Repository Pattern:** `UserRepo` + `TrackRepo` with parameterized queries (prevents SQL injection)
- `TrackRepo.Upsert()` — `ON CONFLICT DO UPDATE` for idempotent background sync
- `database.New()` runs migrations automatically — no manual SQL setup needed

**Why it improves the system:** Persistence across restarts, concurrent-safe queries via pgxpool, scalable to multiple users.

---

### 2. JWT Authentication (custom implementation)

**Where:** `internal/auth/jwt.go` · `internal/server/middleware.go`

- Pure Go **HMAC-SHA256** signing — no external JWT library (`crypto/hmac` + `crypto/sha256`)
- Claims: `sub` (username), `role`, `iat`, `exp` (expiry)
- `AuthMiddleware` validates every protected route — injects `X-User` and `X-Role` headers
- Token expiry check, signature verification, wrong-secret rejection
- Tested in `TestJWTRoundtrip` (4/4 PASS)

**Why it improves the system:** Stateless auth — no sessions to store, scales horizontally, role-based access is extensible.

---

## 🔲 Other Contributor's 2 Features (To Be Implemented)

Pick any 2 from below — all directly named in the Part 2 spec:

### Option A: Rate Limiting *(recommended)*
Limit requests per IP to prevent brute-force login attacks. Token bucket per client IP stored in a `sync.Map`, with a cleanup goroutine using `time.Ticker`.
- ~50 lines of code as a new middleware
- Fits cleanly alongside `LoggingMiddleware` and `CORSMiddleware`

### Option B: Worker Pool *(recommended)*
Replace the sequential library scan loop with a bounded goroutine pool.
- N workers reading from a `jobs chan` and writing results to a `results chan`
- Demonstrates advanced channel/goroutine patterns

### Option C: Advanced String Manipulation
- Username format validation: alphanumeric + underscore, 3–20 chars (regex)
- Search query normalization: trim, lowercase, multi-word tokenization
- Filename-to-title cleanup: strip underscores/hyphens, apply title-case

### Option D: Library Statistics (Mathematical Computations)
`GET /api/library/stats` endpoint returning:
- Total tracks, total duration (sum)
- Average track length (mean)
- Most common genre (mode)
- Tracks per album distribution

**Suggestion for other contributor:** **Rate Limiting + Worker Pool** — both are directly named in the spec and showcase distinct skills (middleware vs goroutine patterns).
