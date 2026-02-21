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

# search multiple queries in one run (comma-separated)
./jobcli search "software engineer,hardware engineer,data scientist" --location "Munich, Germany" --country de --limit 25

# search a single site
./jobcli linkedin "platform engineer" --remote

# output to CSV
./jobcli search "sre" --format csv --output results.csv

# JSON output
./jobcli search "backend" --json

# output only unseen jobs (A-B)
./jobcli search "backend" --json --seen jobs_seen.json --new-only --output jobs_new.json

# update seen history with reviewed/ranked jobs
./jobcli seen update --seen jobs_seen.json --input jobs_new.json --out jobs_seen.json --stats

# or: auto-update the seen history directly during search
./jobcli search "backend" --json --seen jobs_seen.json --new-only --seen-update --output jobs_new.json
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
- `jobcli stepstone <query> ...`
- `jobcli seen diff --new A.json --seen B.json --out C.json [--stats]`
- `jobcli seen update --seen B.json --input C.json --out B.json [--stats]`
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
- `--limit` (maximum rows fetched per query; final merged output may exceed this when using comma-separated queries)
- `--offset`
- `--remote`
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

Notes:

- `search <query>` supports comma-separated query lists (max `10`), e.g. `"backend,platform,sre"`.
- If you use `--new-only --json --output jobs_new.json`, you usually don’t need `--new-out`.
- Use `--new-out` when you want to keep the primary output as "all jobs" (table/CSV/etc) but still persist unseen jobs for `jobcli seen update`.
- Use `--seen-update` if you want to mark newly discovered unseen jobs as "seen" immediately (no separate `jobcli seen update` step).

## Seen workflow

```bash
# Diff new results against history (C = A - B)
./jobcli seen diff --new jobs_a.json --seen jobs_seen.json --out jobs_c.json --stats

# Merge reviewed/ranked jobs into seen history
./jobcli seen update --seen jobs_seen.json --input jobs_c_ranked.json --out jobs_seen.json --stats
```

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
- Search errors are summarized on stderr after the search completes; use `--verbose` to include not-implemented scrapers.
- If output is hard to parse, use `--json` or `--plain`.
- If color output looks wrong, set `--color=never` or export `NO_COLOR=1`.
