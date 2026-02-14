# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
make build          # Build binary → ./jobcli
make test           # Run all tests (go test ./...)
make lint           # Run golangci-lint
make fmt            # Format code (goimports + gofumpt)
make fmt-check      # Format + fail if diff exists (used in CI)
make tools          # Install dev tools (gofumpt, goimports, golangci-lint) into .tools/
```

Run a single test: `go test ./internal/seen/ -run TestDiff`

CI runs: fmt-check → test → lint.

## Architecture

Go CLI tool that scrapes job listings from 6 sites (LinkedIn, Indeed, Glassdoor, ZipRecruiter, Google Jobs, Stepstone) concurrently and exports results in multiple formats.

**Execution flow:** `cmd/jobcli/main.go` (kong CLI init) → `internal/cmd/` (command handlers) → `internal/scraper/` (concurrent scraping) → `internal/export/` (output formatting).

**Key packages:**
- `internal/cmd/` — Kong command structs. All commands receive a shared `*cmd.Context`. The search command orchestrates scraping with goroutines + channels.
- `internal/scraper/` — Each site implements the `Scraper` interface (`Name()` + `Search()`). Registered in `registry.go`. Shared HTTP/parsing helpers in `common.go`.
- `internal/network/` — TLS client wrapper (Chrome_120 fingerprint via `tls-client`) with proxy rotation. `rotator.go` temporarily bans proxies on 403/429.
- `internal/seen/` — Deduplication logic. Keys jobs by `title::company` (normalized), falls back to `url::URL`. Diff (unseen) and Merge operations.
- `internal/export/` — Output writers: table (TTY), CSV (non-TTY default), JSON, TSV, Markdown.
- `internal/models/` — `Job` (13 fields) and `SearchParams` structs.
- `internal/config/` — Loads from `~/.config/jobcli/config.json` (JSON5). Env vars override config; CLI flags override env vars.

## Conventions

- Use existing kong flag patterns in `internal/cmd/` when adding new options.
- Reuse shared helpers in `internal/scraper/common.go` and `internal/network/` rather than duplicating per-site.
- Keep CLI output stable and backward compatible.
- Update `docs/` when user-facing behavior or flags change.
- Adding a new scraper: implement `Scraper` interface → register in `registry.go` → add site command in `root.go` → add tests.
- Avoid new dependencies without user confirmation.
