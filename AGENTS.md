# Agents Guide: Project Structure

This document gives an LLM (or any contributor) a quick map of how the project is organized and where to make changes.

## Top-Level Layout

- `cmd/jobcli/` — binary entrypoint (CLI wiring + runtime initialization).
- `internal/cmd/` — kong command structs and command handlers.
- `internal/config/` — config paths, defaults, init, and proxy loading.
- `internal/ui/` — color handling and user-facing output helpers.
- `internal/network/` — TLS client wrapper + proxy rotator.
- `internal/scraper/` — scraper interface and site implementations.
- `internal/models/` — shared data types (`Job`, `SearchParams`, etc.).
- `internal/export/` — output writers (table, CSV/TSV, JSON, Markdown).
- `docs/` — specs and usage documentation.

## Entry Flow

1. `cmd/jobcli/main.go` initializes kong, loads config, sets up logging/UI.
2. CLI commands in `internal/cmd/` run with a shared `*cmd.Context`.
3. `search` command builds `models.SearchParams`, creates network clients,
   then runs multiple scrapers concurrently.
4. Results are normalized into `[]models.Job` and sent to `export.WriteJobs`.

## Key Files (By Area)

CLI:

- `cmd/jobcli/main.go`
- `internal/cmd/root.go`
- `internal/cmd/search.go`
- `internal/cmd/config.go`
- `internal/cmd/proxies.go`

Config:

- `internal/config/config.go`

Network:

- `internal/network/client.go`
- `internal/network/rotator.go`

Scrapers:

- `internal/scraper/interface.go`
- `internal/scraper/registry.go`
- `internal/scraper/indeed.go`
- `internal/scraper/linkedin.go`
- `internal/scraper/glassdoor.go`
- `internal/scraper/ziprecruiter.go`
- `internal/scraper/google.go`
- `internal/scraper/common.go`

Output:

- `internal/export/export.go`

Models:

- `internal/models/job.go`
- `internal/models/search.go`

## Adding a New Scraper

1. Implement the `Scraper` interface in `internal/scraper/<site>.go`.
2. Register it in `internal/scraper/registry.go`.
3. Add a site command in `internal/cmd/root.go` if you want a direct entrypoint.
4. Add tests in `internal/scraper/<site>_test.go`.

## Where to Fix 403s / Anti-Bot Issues

- `internal/network/client.go` — TLS client profile, cookies, proxy handling.
- `internal/scraper/common.go` — shared HTTP headers and document fetching.
- Site-specific scrapers — add warm-up requests or custom headers.

## Testing

- Unit tests live in `internal/scraper/*_test.go`.
- Run all tests: `go test ./...`
- Lint: `make lint`
- Format: `make fmt` or `make fmt-check`

## Agent Guidelines

- Keep CLI output stable and backward compatible unless the user requests changes.
- Mirror existing kong flag patterns in `internal/cmd/` when adding new options.
- Prefer shared helpers in `internal/network/` and `internal/scraper/common.go` over per-site duplication.
- Avoid introducing new dependencies without confirming with the user.
- Update `docs/` when user-facing behavior or flags change.
- Use `gofmt` or `make fmt` on modified Go files.
- Run `go test ./...` and `make lint` when practical; report if not run.
