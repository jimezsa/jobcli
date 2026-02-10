# 🧑‍💻 JobCLI - Jobs in your terminal

![GitHub Repo Banner](docs/assets/jobcli.png)

Fast, single-binary job aggregation CLI written in Go. Scrapes multiple sites in parallel and exports results to table, CSV, TSV, JSON, or Markdown.

![JobCLI Demo](docs/assets/jobcli_x5.gif)

## Features

- Concurrent scraping across LinkedIn, Indeed, Glassdoor, ZipRecruiter, Google Jobs, and Stepstone
- TLS fingerprinting via `tls-client` to reduce blocking
- Proxy rotation with temporary bans on 403/429 responses
- Seen-jobs workflow with JSON diff/update commands to avoid reprocessing old listings
- Human-friendly tables or machine-friendly exports
- Config + proxies stored in the user config directory

## Requirements

- Go 1.25

## Installation

### Homebrew (macOS/Linux)

```bash
brew install jimezsa/tap/jobcli
```

### Windows

1. Download the latest `.zip` for your architecture from the [Releases](https://github.com/jimezsa/jobcli/releases) page:
   - `jobcli_<version>_windows_amd64.zip` for 64-bit Intel/AMD
   - `jobcli_<version>_windows_arm64.zip` for ARM64
2. Extract the `.zip` file
3. Move `jobcli.exe` to a directory in your `PATH`, or run it directly:

```powershell
.\jobcli.exe --help
```

### Build from source

```bash
git clone https://github.com/jimezsa/jobcli
cd jobcli
make
./jobcli
```

Run:

```bash
jobcli --help
```

## Quick Start

```bash
# show overview and command list
jobcli

# search software engineer roles in Munich, Germany
jobcli search "software engineer" --location "Munich, Germany"  --limit 100

# search a single site last 48 hours
jobcli linkedin "chemical engineer" --location "Munich, Germany"  --limit 10 --hours 48

# search a single site
jobcli stepstone "hardware engineer" --location "Munich, Germany"  --limit 100

# output only unseen jobs using a seen-history JSON
jobcli search "software engineer" --location "Munich, Germany" --limit 30 \
  --seen jobs_seen.json --new-only --json --output jobs_new.json

# update seen-history after reviewing/ranking new jobs
jobcli seen update --seen jobs_seen.json --input jobs_new.json --out jobs_seen.json --stats

# avoid 403s by narrowing sites or providing proxies
jobcli search "software engineer" --sites linkedin --location "Munich, Germany" --country de --limit 10
jobcli search "software engineer" --location "Munich, Germany" --country de --proxies "http://user:pass@host:port,http://host2:port"


```

## Commands

- `jobcli version`
- `jobcli config init`
- `jobcli config path`
- `jobcli search <query> [--location L] [--sites S] [--limit N] [--offset N]`
- `jobcli linkedin <query> ...`
- `jobcli indeed <query> ...`
- `jobcli glassdoor <query> ...`
- `jobcli ziprecruiter <query> ...`
- `jobcli google <query> ...`
- `jobcli stepstone <query> ...`
- `jobcli seen diff --new A.json --seen B.json --out C.json [--stats]`
- `jobcli seen update --seen B.json --input C.json --out B.json [--stats]`
- `jobcli proxies check`

## Output Formats

- Default: table when stdout is a TTY, CSV otherwise (columns: site/title/company/url; URL is blue)
- `--json`: JSON array
- `--plain`: TSV
- `--format=csv|json|md`: explicit format override

## Flags

Global flags:

- `--color=auto|always|never`
- `--json`
- `--plain`
- `--verbose`
- `--version`

Search flags:

- `--location`
- `--sites` (comma-separated list; default `all`)
- `--limit`
- `--offset`
- `--job-type=fulltime|parttime|contract|internship`
- `--hours`
- `--country`
- `--format=csv|json|md`
- `--links=short|full`
- `--output` (aliases: `--out`, `--file`) (write the primary output to a file)
- `--proxies` (comma-separated URLs)
- `--seen` (path to seen jobs JSON history)
- `--new-only` (output only unseen jobs; requires `--seen`)
- `--new-out` (also write unseen jobs (`A - B`) to a JSON file; requires `--seen`)
- `--seen-update` (update `--seen` by merging in newly discovered unseen jobs after the search completes; requires `--seen`)

Seen flags:

- `--stats` (print diff/merge stats to stdout)

## Seen Jobs Workflow

Use this when you want only fresh jobs in recurring runs.

```bash
# 1) scrape and keep only unseen jobs (C = A - B)
jobcli search "hardware engineer" --location "Munich, Germany" --limit 30 \
  --seen jobs_seen.json --new-only --json --output jobs_new.json

# If you want to auto-mark new jobs as "seen" in the same run (no separate
# `jobcli seen update` step), add --seen-update:
# jobcli search "hardware engineer" --location "Munich, Germany" --limit 30 \
#   --seen jobs_seen.json --new-only --seen-update --json --output jobs_new.json

# If you want to keep the primary output as "all jobs" (table/CSV/etc) but still
# write unseen jobs to JSON, use --new-out (no --new-only needed):
# jobcli search "hardware engineer" --location "Munich, Germany" --limit 30 \
#   --seen jobs_seen.json --format csv --output jobs_all.csv --new-out jobs_new.json

# 2) (optional) rank/review jobs_new.json with your own tooling

# 3) persist accepted/new jobs back into seen history
jobcli seen update --seen jobs_seen.json --input jobs_new.json --out jobs_seen.json --stats
```

## Config

Config directory:

```
$(os.UserConfigDir())/jobcli/
```

Files:

- `config.json`
- `proxies.txt`
- `cookies.json` (optional)

Environment variables:

- `JOBCLI_COLOR=auto|always|never`
- `JOBCLI_JSON=1`
- `JOBCLI_VERBOSE=1`
- `JOBCLI_PROXIES=...`
- `JOBCLI_DEFAULT_LOCATION="New York, NY"`
- `JOBCLI_DEFAULT_COUNTRY="usa"`
- `JOBCLI_DEFAULT_LIMIT=20`

## Proxy Checking

```bash
jobcli proxies check --target "https://www.google.com" --timeout 15
```

## AI Agent Skills

The `skills/` directory contains ready-made skill files that let AI coding agents (OpenClaw, Cursor, Codex, etc.) use JobCLI on your behalf:

- **SKILL-cv-summary.md** — reads a PDF CV, extracts an anonymous persona summary and 10 search keywords, and saves them to `CVSUMMARY.md`. Run once or whenever you update your CV.
- **SKILL-jobcli-ranking.md** — reads `CVSUMMARY.md`, runs jobcli searches for each keyword, deduplicates results, ranks every listing 0–1 against your persona, and presents a scored table. Designed for daily use.

## Notes

- Scrapers are best-effort and may require selector updates as sites change.
- Heavy usage may require rotating proxies.

## Inspiration

- JobSpy: https://github.com/speedyapply/JobSpy
- gogcli: https://github.com/steipete/gogcli

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
