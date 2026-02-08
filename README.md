# 🧑‍💻 JobCLI - Jobs in your terminal

![GitHub Repo Banner](docs/assets/jobcli.png)

Fast, single-binary job aggregation CLI written in Go. Scrapes multiple sites in parallel and exports results to table, CSV, TSV, JSON, or Markdown.

![JobCLI Demo](docs/assets/jobcli_x5.gif)

## Features

- Concurrent scraping across LinkedIn, Indeed, Glassdoor, ZipRecruiter, Google Jobs, and Stepstone
- TLS fingerprinting via `tls-client` to reduce blocking
- Proxy rotation with temporary bans on 403/429 responses
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
- `--output` (aliases: `--out`, `--file`)
- `--proxies` (comma-separated URLs)

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
