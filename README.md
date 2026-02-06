# 🧑‍💻 JobCLI - Jobs in your terminal

![GitHub Repo Banner](docs/assets/jobcli.png)

Fast, single-binary job aggregation CLI written in Go. Scrapes multiple sites in parallel and exports results to table, CSV, TSV, JSON, or Markdown.

## Features

- Concurrent scraping across LinkedIn, Indeed, Glassdoor, ZipRecruiter, Google Jobs, and Stepstone
- TLS fingerprinting via `tls-client` to reduce blocking
- Proxy rotation with temporary bans on 403/429 responses
- Human-friendly tables or machine-friendly exports
- Config + proxies stored in the user config directory

## Requirements

- Go 1.25

## Install

### Build from source

```bash
git clone https://github.com/jimezsa/jobcli
cd jobcli
make
# or:
# go build -o jobcli ./cmd/jobcli
```

Run:

```bash
./jobcli --help
```

## Quick Start

```bash
# show overview and command list
./jobcli

# initialize config and proxies
./jobcli config init

# search all sites
./jobcli search "golang" --location "New York, NY" --limit 25

# search software engineer roles in Munich, Germany
./jobcli search "software engineer" --location "Munich, Germany" --country de --limit 25

# search a single site
./jobcli linkedin "platform engineer" --remote

# avoid 403s by narrowing sites or providing proxies
./jobcli search "software engineer" --sites linkedin --location "Munich, Germany" --country de --limit 10
./jobcli search "software engineer" --location "Munich, Germany" --country de --proxies "http://user:pass@host:port,http://host2:port"

# output to CSV
./jobcli search "sre" --format csv --output results.csv

# JSON output
./jobcli search "backend" --json
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
- `--remote`
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
./jobcli proxies check --target "https://www.google.com" --timeout 15
```

## Notes

- Scrapers are best-effort and may require selector updates as sites change.
- Heavy usage may require rotating proxies.

## Inspiration

- JobSpy: https://github.com/speedyapply/JobSpy
- gogcli: https://github.com/steipete/gogcli

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
