# Assignment Part 1 — Implementation Coverage

> **GTMPC: Go Terminal Music Player — Client-Server Architecture**
> Covers all Part 1 criteria with specific file and line references.

---

## ✅ 1. HTTP Server with Efficient Routing and Request Management *(4 marks)*

**File:** `internal/server/server.go` · `internal/server/handlers.go`

The server uses Go's standard `net/http` package with a **structured two-mux routing strategy**:

```
Public routes  →  /api/health, /api/auth/register, /api/auth/login
Protected routes → /api/library/tracks, /api/library/search,
                   /api/library/upload, /api/stream/{trackID}
```

**What makes the routing efficient:**

| Feature | Implementation |
|---|---|
| Separate public/protected mux | `server.go:27-42` — avoids auth overhead for public routes |
| Middleware chaining | `Chain(mux, LoggingMiddleware, CORSMiddleware)` |
| Route-level auth | Protected mux wrapped with `AuthMiddleware` |
| Correct HTTP methods | `POST` for auth, `GET` for library, `POST` for upload |
| Timeout configuration | `ReadTimeout: 15s`, `WriteTimeout: 60s`, `IdleTimeout: 120s` |
| File upload (multipart) | `HandleUploadTrack` — 50MB limit via `ParseMultipartForm` |
| Audio streaming | `http.ServeFile` in `HandleStreamTrack` — native HTTP Range support |

**API Endpoints:**

```
GET  /api/health               → Health check
POST /api/auth/register        → Register new user (bcrypt hashed)
POST /api/auth/login           → Login, returns JWT
GET  /api/library/tracks       → Paginated track list (JWT required)
GET  /api/library/search?q=    → Full-text search (JWT required)
POST /api/library/upload       → Upload audio file (JWT required)
GET  /api/stream/{trackID}     → Audio stream with Range support (JWT required)
GET  /                         → Serves Web UI (index.html)
GET  /static/*                 → Static assets (CSS, JS)
```

---

## ✅ 2. Robust Security — bcrypt Credential Handling *(4 marks)*

**Files:** `internal/auth/auth.go` · `internal/auth/db_service.go` · `internal/auth/jwt.go`

### bcrypt Password Hashing

```go
// internal/auth/auth.go
hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
```

- **Cost factor `bcrypt.DefaultCost` (10)** — computationally expensive, resists brute force
- Password hash is **never stored or returned in plaintext** — `PasswordHash` is the only stored credential
- **Plain text check in tests** — `auth_test.go:37` explicitly asserts `PasswordHash != plaintext`
- Duplicate usernames rejected with `ErrUserAlreadyExists`
- Empty credentials rejected with `ErrEmptyCredentials` before any DB operation

### JWT Authentication (No external library)

```go
// internal/auth/jwt.go — pure Go HMAC-SHA256 implementation
func GenerateToken(username, role string, secret []byte, ttl time.Duration) (string, error)
func ValidateToken(tokenStr string, secret []byte) (*Claims, error)
```

- **Algorithm:** HMAC-SHA256 — `crypto/hmac` + `crypto/sha256` (stdlib only)
- **Claims:** `sub` (username), `role`, `iat` (issued at), `exp` (expiry)
- **Token expiry validated** — `time.Now().Unix() > claims.ExpAt` returns `ErrInvalidToken`
- **Signature verification** — wrong secret returns `ErrInvalidToken` (proven in `TestJWTRoundtrip`)
- Secret loaded from `GTMPC_JWT_SECRET` env var, defaults with warning logged

---

## ✅ 3. Concurrency Techniques to Optimize Execution *(4 marks)*

**File:** `internal/server/server.go:62-95`

The `Start()` method uses a `sync.WaitGroup` to manage **two concurrent goroutines**:

```go
var wg sync.WaitGroup

// Goroutine 1: Background library scanner
wg.Add(1)
go func() {
    defer wg.Done()
    s.backgroundScanner(ctx, lib, trackRepo, scanPaths)
}()

// Goroutine 2: HTTP server
wg.Add(1)
go func() {
    defer wg.Done()
    s.httpServer.ListenAndServe()
}()
```

**Concurrency patterns used:**

| Pattern | Usage |
|---|---|
| `sync.WaitGroup` | Coordinates HTTP server + background scanner |
| `context.Context` cancellation | `<-ctx.Done()` triggers graceful shutdown |
| `chan error` | Propagates server startup errors to parent goroutine |
| `goroutine-per-request` | Standard `net/http` — each request served in its own goroutine |
| `time.NewTicker` | Background scanner runs **every 5 minutes** without blocking |
| Signal handling goroutine | `go func() { sig := <-sigCh; cancel() }()` in `main.go` |

**Graceful shutdown** uses a 10-second deadline context:
```go
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
return s.httpServer.Shutdown(shutdownCtx)
```

---

## ✅ 4. Goroutines for Asynchronous Workflows *(3 marks)*

**File:** `internal/server/server.go:97-146`

The `backgroundScanner` goroutine is a fully async workflow running independently of the HTTP server:

```
Startup
   │
   ├─► [Goroutine] HTTP Server  → handles requests concurrently
   │
   └─► [Goroutine] backgroundScanner
           │
           ├─► Initial scan: lib.Scan(ctx, paths)
           ├─► syncTracksToDatabase → UPSERT each track to PostgreSQL
           │
           └─► ticker loop (every 5 min):
                   ├─► Rescan all music directories
                   └─► Re-sync updated tracks to PostgreSQL
```

The background scanner uses context propagation to stop cleanly on `SIGINT`/`SIGTERM` — it doesn't block the main HTTP serving goroutine at all.

```go
case <-ctx.Done():
    log.Println("[SCANNER] Background scanner stopped")
    return
case <-ticker.C:
    lib.Scan(ctx, paths)
    syncTracksToDatabase(ctx, lib, trackRepo)
```

---

## ✅ 5. Functional UI Interacting with Backend Endpoints *(3 marks)*

**Files:** `web/index.html` · `web/style.css` · `web/app.js`

The Web UI is a single-page application served directly by the Go server at `http://localhost:8080`.

### UI Features

| Feature | Implementation |
|---|---|
| **Login / Register** | Tab-switched forms, POST to `/api/auth/login` and `/api/auth/register` |
| **Auto-login on register** | Immediately calls login after registration |
| **Session persistence** | JWT stored in `localStorage`, re-used on page reload |
| **Library browser** | Fetches `GET /api/library/tracks`, renders sortable table |
| **Search** | Instant client-side filtering across title, artist, album |
| **Upload** | File picker → `POST /api/library/upload` multipart → library refreshes |
| **Audio player** | Fetch stream with JWT → Blob URL → `<audio>` element |
| **Transport controls** | Play, Pause, Previous, Next |
| **Seek bar** | Draggable, filled with progress gradient |
| **Volume control** | Custom range input |
| **Auto-next** | Plays next track on `ended` event |
| **Keyboard shortcuts** | Space (play/pause), N/P (next/prev), ←/→ (seek ±5s) |

### How the UI Talks to the Server

```
Browser                              Go Server (localhost:8080)
   │                                          │
   ├─POST /api/auth/login ──────────────────► HandleLogin
   │◄── { token, user } ────────────────────┤
   │                                          │
   ├─GET /api/library/tracks                  │
   │    Authorization: Bearer <token> ──────► HandleGetTracks (JWT validated by AuthMiddleware)
   │◄── [{ id, title, artist, ... }] ───────┤
   │                                          │
   ├─GET /api/stream/{id}                     │
   │    Authorization: Bearer <token> ──────► HandleStreamTrack → http.ServeFile (Range support)
   │◄── Audio bytes ─────────────────────────┤
```

---

## ✅ 6. Well-Organized, Readable, Modular Codebase *(2 marks)*

### Package Structure

```
gtmpc/
├── api/           # Shared types (Track, User, request/response structs)
├── cmd/
│   ├── player/    # TUI client binary entry point
│   └── server/    # HTTP server binary entry point (main.go)
├── internal/
│   ├── auth/      # bcrypt auth service, JWT, unit tests
│   ├── config/    # Config loading (XDG, env override)
│   ├── database/  # PostgreSQL: pool, migrations, UserRepo, TrackRepo
│   ├── library/   # Music file scanning
│   └── server/    # HTTP handlers, middleware, routing
├── pkg/           # Public reusable packages
├── web/           # Static Web UI (HTML, CSS, JS)
└── .env           # Local environment config (gitignored)
```

### Design Principles Applied

| Principle | Evidence |
|---|---|
| **Repository Pattern** | `UserRepo`, `TrackRepo` decouple DB from business logic |
| **Dependency Injection** | `Handlers` struct accepts `DBService`, `TrackRepo` — no global state |
| **Interface segregation** | `auth.Service` interface lets handlers be tested without a DB |
| **Idempotent operations** | `UPSERT ON CONFLICT DO UPDATE` — re-scanning never duplicates |
| **Graceful degradation** | `.env` missing → falls back to system env vars with log warning |
| **Separation of concerns** | Auth, DB, Library, Server, Config each in their own package |
| **Middleware composition** | `Chain(h, LoggingMiddleware, CORSMiddleware)` — composable |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    Client Layer                          │
│  ┌────────────────────┐    ┌─────────────────────────┐  │
│  │  Web Browser UI    │    │  TUI Player (cmd/player)│  │
│  │  web/index.html    │    │  (local, beep audio)    │  │
│  └────────┬───────────┘    └─────────────────────────┘  │
└───────────┼─────────────────────────────────────────────┘
            │ HTTP/REST (JWT Bearer)
┌───────────▼─────────────────────────────────────────────┐
│                  HTTP Server Layer                        │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Middleware Chain                                │    │
│  │  LoggingMiddleware → CORSMiddleware → AuthMiddle │    │
│  └──────────────────────┬──────────────────────────┘    │
│                         │                               │
│  ┌──────────┐  ┌────────▼───────┐  ┌────────────────┐  │
│  │ /api/auth│  │/api/library/*  │  │ /api/stream/*  │  │
│  │ register │  │ tracks, search │  │ Range streaming│  │
│  │ login    │  │ upload         │  │                │  │
│  └──────────┘  └───────┬────────┘  └───────┬────────┘  │
└──────────────────────┬─┴─────────────────┬─┴────────────┘
                       │                   │
┌──────────────────────▼───────────────────▼──────────────┐
│                  Internal Layer                           │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────┐  │
│  │ auth package│  │database pkg  │  │library package │  │
│  │ bcrypt      │  │ pgxpool      │  │ file scanner   │  │
│  │ JWT HS256   │  │ UserRepo     │  │ goroutine scan │  │
│  │ DBService   │  │ TrackRepo    │  │ 5-min ticker   │  │
│  └─────────────┘  └──────┬───────┘  └────────────────┘  │
└─────────────────────────┬┴──────────────────────────────┘
                          │
┌─────────────────────────▼──────────────────────────────┐
│               PostgreSQL Database                        │
│   users  │  tracks  │  playlists  │  playlist_tracks    │
└────────────────────────────────────────────────────────┘
```

---

## Summary Table

| Requirement | Status | Key Files |
|---|---|---|
| HTTP server + efficient routing | ✅ | `server.go`, `middleware.go` |
| bcrypt credential security | ✅ | `auth.go`, `db_service.go` |
| JWT authentication | ✅ | `jwt.go`, `middleware.go` |
| Concurrency / goroutines | ✅ | `server.go` (WaitGroup, context, ticker) |
| Async workflows | ✅ | `backgroundScanner` goroutine |
| Functional Web UI | ✅ | `web/index.html`, `web/app.js` |
| UI ↔ backend communication | ✅ | REST API + JWT, audio streaming |
| Modular codebase | ✅ | 7 packages, repository pattern, DI |
| PostgreSQL database | ✅ | `internal/database/` |
| Unit tests | ✅ | `internal/auth/auth_test.go` (4 tests) |
