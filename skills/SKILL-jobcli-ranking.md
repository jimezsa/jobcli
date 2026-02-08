---
name: jobcli-job-search
description: Search and rank job listings against a persona using the JobCLI tool.
homepage: https://github.com/jimezsa/jobcli
metadata:
  {
    "openclaw":
      {
        "emoji": "💼",
        "os": ["linux", "darwin"],
        "requires": { "bins": ["jobcli"] },
        "install":
          [
            {
              "id": "source",
              "kind": "shell",
              "script": "git clone https://github.com/jimezsa/jobcli && cd jobcli && make && sudo mv jobcli /usr/local/bin/",
              "bins": ["jobcli"],
              "label": "Build JobCLI from source (requires Go 1.25)",
            },
            {
              "id": "release",
              "kind": "shell",
              "script": "curl -sL https://github.com/jimezsa/jobcli/releases/latest/download/jobcli_$(uname -s)_$(uname -m).tar.gz | tar xz && sudo mv jobcli /usr/local/bin/",
              "bins": ["jobcli"],
              "label": "Install JobCLI from GitHub release",
            },
          ],
      },
  }
---

# JobCLI — Daily Job Search & Persona Ranking

Daily (or on-demand) skill that reads the persona summary and keywords from
`CVSUMMARY.md`, runs targeted job searches with `jobcli`, deduplicates results,
ranks every listing against the persona, and presents a scored table.

> **Prerequisite:** `CVSUMMARY.md` must exist in the working directory.
> Generate it first with **SKILL-cv-summmary.md** (provide a PDF CV).

> **Trigger:** the user asks to search for jobs, or it is time for a daily
> search run.

---

## 1 — Read CVSUMMARY.md

Load `CVSUMMARY.md` from the current working directory. Extract:

- The **persona summary** (under `## Persona Summary`).
- All **10 keyword phrases** (under `## Search Keywords`).
- The **ranking criteria** (under `## Ranking Criteria`).

If `CVSUMMARY.md` does not exist, stop and tell the user to run the CV summary
skill first (provide a `.pdf` CV).

---

## 2 — JobCLI Quick Reference

`jobcli` is a single Go binary. First-time setup:

```bash
jobcli config init
```

### Basic search (all sites)

```bash
jobcli search "<query>" --location "<City, Country>" --limit <N>
```

### Single-site search

```bash
jobcli linkedin "<query>" --location "<City, Country>" --limit <N>
jobcli indeed   "<query>" --location "<City, Country>" --limit <N>
jobcli glassdoor "<query>" --location "<City, Country>" --limit <N>
jobcli google   "<query>" --location "<City, Country>" --limit <N>
jobcli stepstone "<query>" --location "<City, Country>" --limit <N>
jobcli ziprecruiter "<query>" --location "<City, Country>" --limit <N>
```

### Flags

| Flag            | Description                                      |
| --------------- | ------------------------------------------------ |
| `--location`    | City/region string (e.g., `"Berlin, Germany"`)   |
| `--country`     | Two-letter country code (`de`, `us`, `uk`, …)    |
| `--limit N`     | Max results to return (default varies by site)   |
| `--offset N`    | Skip first N results (pagination)                |
| `--job-type`    | `fulltime`, `parttime`, `contract`, `internship` |
| `--sites`       | Comma-separated site list (default `all`)        |
| `--format`      | `csv`, `json`, `md`                              |
| `--json`        | Shorthand for JSON output                        |
| `--output FILE` | Write results to a file instead of stdout        |
| `--proxies`     | Comma-separated proxy URLs for anti-bot rotation |
| `--hours N`     | Filter by posting age in hours                   |
| `--links`       | `short` or `full` URL display                    |

### Output formats

- Default: human-readable table (columns: site, title, company, url).
- `--json`: JSON array — best for programmatic processing.
- `--format csv`: CSV file.
- `--format md`: Markdown table.

### Dealing with blocks (403/429)

- Narrow to fewer sites: `--sites linkedin,google`.
- Add proxies: `--proxies "http://user:pass@host:port"`.
- Reduce `--limit` and paginate with `--offset`.

---

## 3 — Run Searches

For **each keyword** from `CVSUMMARY.md` (all 10), execute a search.
Ask the user for `--location` and `--country` if not already known.
Use `--json` output and save to a file for processing:

```bash
jobcli search "<keyword>" --location "<location>" --country "<code>" --limit 30 --json --output jobs_keyword_N.json
```

Run searches **sequentially** (not in parallel) to avoid triggering anti-bot
protections. If a search returns a 403 or empty result, retry once with a
narrower `--sites` subset (e.g., `--sites linkedin,google`), then skip and
note the failure.

---

## 4 — Aggregate & Deduplicate

Merge all `jobs_keyword_*.json` result files into a single job list.
Deduplicate by URL — if the same job appears from multiple keywords, keep one
entry and note which keywords matched it.

---

## 5 — Rank Against Persona

For each unique job, assign a **relevance score from 0.0 to 1.0** by comparing
the job's title and or description against the persona summary from `CVSUMMARY.md`.

Scoring dimensions (equal weight, 0.2 each):

| Dimension               | What to check                                                         |
| ----------------------- | --------------------------------------------------------------------- |
| **Title match**         | Does the job title align with the persona's target roles?             |
| **Skill overlap**       | How many of the persona's technical skills appear in the description? |
| **Domain fit**          | Is the industry/sector relevant to the persona's experience?          |
| **Seniority alignment** | Does the role's level (junior/mid/senior/lead) match?                 |
| **Language fit**        | Does the persona meet any stated language requirements?               |

Final score = average of the five dimensions.

---

## 6 — Present Results

Display the final ranked list as a **Markdown table** sorted by score
(descending):

| Score | Title | Company | Site | URL | Matched Keywords |
| ----- | ----- | ------- | ---- | --- | ---------------- |

- Show only jobs with score >= **0.7** by default.
- If the user asks, show all results or apply a custom threshold.
- Offer to save the ranked table to a file (e.g., `ranked_jobs.md`).

---

## 7 — Cleanup

After presenting results, delete the intermediate `jobs_keyword_*.json` files
to keep the working directory clean. Keep `CVSUMMARY.md` and the final
`ranked_jobs.md` (if saved).

---

## Notes

- **Privacy first:** never expose personal data from `CVSUMMARY.md` in search
  queries, logs, or output files.
- **Rate limits:** if many keywords fail with 403/429, sggest to use `--limit 10`
- **Iterative refinement:** the user may ask to adjust keywords or re-rank
  with different criteria — re-read `CVSUMMARY.md` and repeat from Step 3.
- **Daily runs:** on repeat runs, consider using `--hours 48` to fetch only
  jobs posted in the last 48 hours.
- Run `jobcli config init` once to create the default config directory.
