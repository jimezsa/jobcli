# jobcli usage

This document describes how to build, configure, and use `jobcli`.

## Build

```bash
make
# or:
# go build -o jobcli ./cmd/jobcli
```

## Quick start

```bash
# initialize config and proxies
./jobcli config init

# search all sites
./jobcli search "golang" --location "New York, NY" --limit 25

# search software engineer roles in Munich, Germany
./jobcli search "software engineer" --location "Munich, Germany" --country de --limit 25

# search a single site
./jobcli linkedin "platform engineer" --remote

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
- `jobcli proxies check`

## Global flags

- `--color=auto|always|never`
- `--json`
- `--plain`
- `--verbose`
- `--version`

## Search flags

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

## Output formats

- Default: table when stdout is a TTY, CSV otherwise (table columns: site/title/company/url; URL is blue)
- `--json`: JSON array
- `--plain`: TSV
- `--format=csv|json|md`: explicit format override
- `--links=short|full`: table URL display (default `full`, `short` only applies when terminal hyperlinks are supported)

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

## Proxy usage

When running without proxies, some sites may return 403/429. You can either narrow the sites you hit, or provide proxies.

```bash
# narrow to a single site
./jobcli search "software engineer" --sites linkedin --location "Munich, Germany" --country de --limit 10

# use proxies inline
./jobcli search "software engineer" --location "Munich, Germany" --country de --proxies "http://user:pass@host:port,http://host2:port"
```

To validate proxies:

```bash
./jobcli proxies check --target "https://www.google.com" --timeout 15
```

## Troubleshooting

- If you see `http 403`, try fewer sites or use proxies.
- If output is hard to parse, use `--json` or `--plain`.
- If color output looks wrong, set `--color=never` or export `NO_COLOR=1`.
