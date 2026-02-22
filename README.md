# 🧑‍💻 JobCLI - Jobs in your terminal

Fast, single-binary job aggregation CLI written in Go. Scrapes multiple sites in parallel and exports results to table, CSV, TSV, JSON, or Markdown.

![JobCLI Demo](docs/assets/jobcli_x5.gif)

## Features

- Concurrent scraping across LinkedIn, Indeed, Glassdoor, ZipRecruiter, and Stepstone
- TLS fingerprinting via `tls-client` to reduce blocking
- Proxy rotation with temporary bans on 403/429 responses
- Seen-jobs workflow with JSON diff/update commands to avoid reprocessing old listings
- Human-friendly tables or machine-friendly exports
- Config + proxies stored in the user config directory

## Requirements

- Go 1.25

## Installation

Choose one method:

- macOS/Linux (Homebrew): `brew install jimezsa/tap/jobcli`
- Windows: download the latest `.zip` from [Releases](https://github.com/jimezsa/jobcli/releases), extract, and run `jobcli.exe` (or add it to `PATH`)
- Build from source:

```bash
git clone https://github.com/jimezsa/jobcli
cd jobcli
make
./jobcli --help
```

## Quick Start

```bash
# show overview and command list
jobcli

# search software engineer roles in Munich, Germany
jobcli search "software engineer" --location "Munich, Germany"  --limit 100 --hours 48

# search multiple queries in one run (comma-separated)
jobcli search "software engineer,hardware engineer,data scientist" --location "Munich, Germany" --limit 100 --hours 48

# load queries from JSON file
jobcli search --query-file queries.json --location "Munich, Germany" --limit 100 --hours 48

# search a single site last 48 hours
jobcli linkedin "chemical engineer" --location "Munich, Germany"  --limit 10 --hours 48

# output only unseen jobs using a seen-history JSON , update seen jobs (jobs_seen.json)
jobcli search "software engineer" --location "Munich, Germany" --limit 30 --hours 48 \
  --seen jobs_seen.json --seen-update --new-only --json --output jobs_new.json

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
- `jobcli search [<query>] [--query-file queries.json] [--location L] [--sites S] [--limit N] [--offset N]`
- `jobcli linkedin [<query>] [--query-file queries.json] ...`
- `jobcli indeed [<query>] [--query-file queries.json] ...`
- `jobcli glassdoor [<query>] [--query-file queries.json] ...`
- `jobcli ziprecruiter [<query>] [--query-file queries.json] ...`
- `jobcli stepstone [<query>] [--query-file queries.json] ...`
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

| Flag | Description |
| --- | --- |
| `--color=auto\|always\|never` | Color mode for CLI output. |
| `--json` | Write JSON to stdout and disable colors. |
| `--plain` | Write TSV to stdout and disable colors. |
| `--verbose` | Enable debug logging. |
| `--version` | Print version information. |

Search and site flags (`search`, `linkedin`, `indeed`, `glassdoor`, `ziprecruiter`, `stepstone`):

| Flag | Description |
| --- | --- |
| `--location` | Job location filter. |
| `--country` | Country code filter (used by Indeed/Glassdoor). |
| `--limit` | Maximum results fetched per query. |
| `--offset` | Pagination offset. |
| `--remote` | Remote-only roles. |
| `--job-type=fulltime\|parttime\|contract\|internship` | Job type filter. |
| `--hours` | Jobs posted in the last N hours. |
| `--format=csv\|json\|md` | Explicit output format override. |
| `--links=short\|full` | Table link rendering style. |
| `--output`, `-o` | Write primary output to a file. |
| `--out` | Alias for `--output`. |
| `--file` | Alias for `--output`. |
| `--proxies` | Comma-separated proxy URLs. |
| `--query-file` | Path to JSON queries file (top-level string array or object with `job_titles` string array). |
| `--seen` | Path to seen jobs JSON history file. |
| `--new-only` | Output only unseen jobs (`A - B`); requires `--seen`. |
| `--new-out` | Also write unseen jobs (`A - B`) to JSON; requires `--seen`. |
| `--seen-update` | Merge newly discovered unseen jobs into `--seen`; requires `--seen`. |
| `--sites` | Comma-separated site list for `search` only (default: `all`). |

Seen command flags:

| Command | Flag | Description |
| --- | --- | --- |
| `seen diff` | `--new` | Path to new jobs JSON (`A`). |
| `seen diff` | `--seen` | Path to seen jobs JSON (`B`); missing file is treated as empty. |
| `seen diff` | `--out` | Output path for unseen jobs JSON (`C = A - B`). |
| `seen diff` | `--stats` | Print comparison stats. |
| `seen update` | `--seen` | Path to seen jobs JSON (`B`); missing file is treated as empty. |
| `seen update` | `--input` | Input jobs JSON to merge into seen history. |
| `seen update` | `--out` | Output path for updated seen jobs JSON. |
| `seen update` | `--stats` | Print merge stats. |

Proxy command flags:

| Command | Flag | Description |
| --- | --- | --- |
| `proxies check` | `--target` | Target URL to validate proxies against (default: `https://www.google.com`). |
| `proxies check` | `--timeout` | Per-request timeout in seconds (default: `15`). |

Notes:

- `search` supports comma-separated positional query lists (max `10`), e.g. `"backend,platform,sre"`.
- `--query-file` accepts either `["backend","platform"]` or `{"job_titles":["backend","platform"]}`.
- Positional and file queries can be combined; positional entries are applied first, then deduped case-insensitively.

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

- **`skills/jobcli-cv-summary/SKILL.md`** (`jobcli-cv-summary`) — reads one or more CV PDFs, creates a privacy-safe persona summary, and writes 20 search keywords to `profiles/<user_id>/CVSUMMARY.md`.
- **`skills/jobcli-job-search/SKILL.md`** (`jobcli-job-search`) — reads `profiles/<user_id>/CVSUMMARY.md`, runs unseen-job searches per keyword, deduplicates, and ranks matches from 0.0 to 1.0.
- **`skills/pirate-motivator/SKILL.md`** (`pirate-motivator`) — generates short pirate-style motivational audio for job-hunt encouragement.

## Notes

- Scrapers are best-effort and may require selector updates as sites change.
- Heavy usage may require rotating proxies.

## Inspiration

- JobSpy: https://github.com/speedyapply/JobSpy
- gogcli: https://github.com/steipete/gogcli

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
