# gtmpc Makefile — Build targets for the client-server music player
# Usage:
#   make tui-client   — Build the TUI client binary
#   make server       — Build the server binary (if exists)
#   make web          — Build the React web frontend
#   make web-dev      — Run React dev server with API proxy
#   make build-all    — Build everything
#   make clean        — Remove build artifacts

.PHONY: all tui-client server web web-dev web-install build-all clean lint

# Go output directory
BIN := bin

# Default target
all: build-all

# ── Go binaries ────────────────────────────────────────────────────────────────

$(BIN):
	mkdir -p $(BIN)

tui-client: | $(BIN)
	go build -o $(BIN)/gtmpc-client ./cmd/client

server: | $(BIN)
	@if [ -d "./cmd/server" ]; then \
		go build -o $(BIN)/gtmpc-server ./cmd/server; \
	else \
		echo "No cmd/server found — skipping."; \
	fi

# ── React frontend ─────────────────────────────────────────────────────────────

web-install:
	cd web && npm install

web:
	cd web && npm run build

web-dev:
	cd web && npm run dev

# ── Combined targets ───────────────────────────────────────────────────────────

build-all: tui-client server web

# ── Code quality ──────────────────────────────────────────────────────────────

lint:
	go vet ./...
	@echo "Go vet passed."

# ── Cleanup ───────────────────────────────────────────────────────────────────

clean:
	rm -rf $(BIN)
	rm -rf web/dist
	@echo "Clean done."

# ── Help ──────────────────────────────────────────────────────────────────────

help:
	@echo "Available targets:"
	@echo "  make tui-client   Build the TUI client (bin/gtmpc-client)"
	@echo "  make server       Build the server binary if present (bin/gtmpc-server)"
	@echo "  make web          Build React frontend (web/dist)"
	@echo "  make web-dev      Start React dev server with API proxy"
	@echo "  make web-install  Install npm dependencies"
	@echo "  make build-all    Build everything"
	@echo "  make lint         Run go vet"
	@echo "  make clean        Remove bin/ and web/dist/"
