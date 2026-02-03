# jobcli spec

## Goal

Build a single, clean, high-performance Go CLI that aggregates job postings from:

- LinkedIn
- Indeed
- Glassdoor
- ZipRecruiter
- Google Jobs

This replaces the existing Python `JobSpy` library conceptually, but:

- **Performance**: Leverages Go's native concurrency (Goroutines) to scrape multiple sites simultaneously without GIL blocking.
- **Portability**: Single static binary distribution (no Python environment/venv required).
- **Architecture**: Clean, interface-driven design using modern Go patterns.

## Non-goals

- A GUI or Web UI (this is strictly a CLI).
- Resume parsing or complex NLP matching (focus is on retrieval).
- Running a persistent server or REST API (use cron + CLI for automation).
- Bypassing strict CAPTCHAs that require human intervention (focus on public/guest modes).

## Language/runtime

- Go `1.25` (see `go.mod`)

## CLI framework

- `github.com/alecthomas/kong`
- Root command: `jobcli`
- Global flags:
  - `--color=auto|always|never` (default `auto`)
  - `--json` (JSON output to stdout; disables colors)
  - `--plain` (TSV output to stdout; stable/parseable; disables colors)
  - `--verbose` (enable debug logging to stderr)
  - `--version` (print version)

Notes:

- We run `SilenceUsage: true` and print errors ourselves (colored when possible).
- `NO_COLOR` is respected.

Environment:

- `JOBCLI_COLOR=auto|always|never` (default `auto`, overridden by `--color`)
- `JOBCLI_JSON=1` (default JSON output; overridden by flags)
- `JOBCLI_VERBOSE=1` (enable debug logs)

## Output (TTY-aware colors)

- `github.com/muesli/termenv` is used to detect rich TTY capabilities and render colored output.
- Colors are enabled when:
  - output is a rich terminal and `--color=auto`, and `NO_COLOR` is not set; or
  - `--color=always`
- Colors are disabled when:
  - `--color=never`; or
  - `--json` or `--plain` is set; or
  - `NO_COLOR` is set

Implementation: `internal/ui/ui.go`.

## Network & Anti-Bot Strategy

### TLS Fingerprinting

- Standard Go `net/http` is easily fingerprinted and blocked by WAFs (Cloudflare/Akamai).
- We use `github.com/bogdanfinn/tls-client` (or `fhttp`) to mimic Chrome/Firefox TLS Client Hellos.
- **User-Agent**: Rotated automatically per request from an internal list of modern browser headers.
- **Cookies**: Ephemeral cookie jars used per session/scraper instance.

Implementation: `internal/network/client.go`.

### Proxy Management (Rotator)

- Proxies are required for high-volume scraping to avoid 429/403 errors.
- **Input**:
  - Flag: `--proxies "http://u:p@host:port,http://..."`
  - File: `$(os.UserConfigDir())/jobcli/proxies.txt` (newline separated)
- **Strategy**:
  - Round-robin rotation.
  - Automatic temporary ban of proxies returning 403/429 status codes.
  - Supports `http`, `https`, `socks5`.

Implementation: `internal/network/rotator.go`.

## Config layout

- Base config dir: `$(os.UserConfigDir())/jobcli/`
- Files:
  - `config.json` (JSON; defaults for search params)
  - `proxies.txt` (Plain text; one proxy per line)
  - `cookies.json` (Optional; persistent session data if needed)

Environment:

- `JOBCLI_PROXIES` (Comma-separated list; overrides config file)
- `JOBCLI_DEFAULT_LOCATION="New York, NY"`
- `JOBCLI_DEFAULT_COUNTRY="usa"`
- `JOBCLI_DEFAULT_LIMIT=20`

Flag aliases:

- `--out` also accepts `--output`.
- `--file` also accepts `--output`.

## Commands (current + planned)

### Implemented

- `jobcli version`
- `jobcli config init` (writes default `config.json` and empty `proxies.txt`)
- `jobcli config path`

### Planned

- `jobcli search <query> [--location L] [--sites S] [--limit N] [--offset N]`
  - **The core command.**
  - `--sites`: Comma-separated list (e.g., `linkedin,indeed`). Default: `all`.
  - `--remote`: Filter for remote-only jobs.
  - `--job-type`: Filter by `fulltime`, `parttime`, `contract`, `internship`.
  - `--hours`: Filter jobs posted in the last N hours.
  - `--country`: Subdomain for Indeed/Glassdoor (e.g., `uk`, `ca`).
  - `--format`: `csv|json|md` (defaults to `csv` or table depending on TTY).
- `jobcli linkedin <query> ...`
  - Direct access to LinkedIn scraper with site-specific flags (if any).
- `jobcli indeed <query> ...`
  - Direct access to Indeed scraper.
- `jobcli proxies check`
  - Validates the current proxy list against a target URL (e.g., `google.com`) and reports latency/success rate.

## Output formats

Default: Human-friendly tables (stdlib `text/tabwriter`) printed to `stdout`.

- **JSON**: `--json` dumps a struct array. Useful for piping to `jq`.
- **TSV**: `--plain` outputs stable tab-separated values.
- **CSV**: `--format=csv` writes standard CSV (default for file output).
- **Markdown**: `--format=md` writes a structured list (useful for copy-pasting into notes).

## Code layout

- `cmd/jobcli/main.go` — binary entrypoint
- `internal/cmd/*` — kong command structs
- `internal/ui/*` — color + printing
- `internal/config/*` — config paths + file parsing
- `internal/network/*` — TLS client wrapper + proxy rotator
- `internal/scraper/*` — Scraper interface & site implementations
  - `interface.go` (`Scraper` interface definition)
  - `linkedin.go`
  - `indeed.go`
  - `glassdoor.go`
- `internal/models/*` — Shared data structs (`Job`, `ScraperConfig`)
- `internal/export/*` — Writers for CSV and JSON

## Dependencies (Planned)

- `github.com/alecthomas/kong` (CLI framework)
- `github.com/PuerkitoBio/goquery` (HTML parsing)
- `github.com/bogdanfinn/tls-client` (TLS fingerprinting)
- `github.com/rs/zerolog` (High-performance logging)
- `github.com/muesli/termenv` (Rich terminal output)

## Formatting, linting, tests

### Formatting

Pinned tools, installed into local `.tools/` via `make tools`:

- `mvdan.cc/gofumpt@v0.7.0`
- `golang.org/x/tools/cmd/goimports@v0.38.0`
- `github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2`

Commands:

- `make fmt` — applies `goimports` + `gofumpt`
- `make fmt-check` — formats and fails if Go files or `go.mod/go.sum` change

### Lint

- `golangci-lint` with config in `.golangci.yml`
- `make lint`

### Tests

- stdlib `testing`
- `make test` — runs unit tests (parsers, utility functions)

### Integration tests (local only)

There is an opt-in integration test suite guarded by build tags (not run in CI). These hit the live sites.

- Requires:
  - Working internet connection (or proxies).
- Run:
  - `go test -tags=integration ./internal/scraper/...`
  - _Note:_ Flaky by definition due to anti-bot measures.

## CI (GitHub Actions)

Workflow: `.github/workflows/ci.yml`

- runs on push + PR
- uses `actions/setup-go` with `go-version-file: go.mod`
- runs:
  - `make tools`
  - `make fmt-check`
  - `go test ./...` (Unit tests only)
  - `golangci-lint`

## Next implementation steps

1.  **Skeleton**: Initialize `kong` structure, logging, and configuration loading.
2.  **Network Layer**: Implement `internal/network` with `tls-client` and proxy rotation logic.
3.  **Models**: Define the unified `Job` struct to handle data from all sources.
4.  **First Scraper (Indeed)**: Implement parsing logic and pagination for Indeed.
5.  **Concurrency**: Implement the worker pool pattern to handle multiple scrapers running in parallel.
6.  **Export**: Wire up the CSV exporter.
