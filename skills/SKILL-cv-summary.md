---
name: jobcli-cv-summary
description: Extract per-user anonymous persona summaries and search keywords from CV PDFs.
homepage: https://github.com/jimezsa/jobcli
metadata:
  {
    "openclaw":
      {
        "emoji": "📄",
        "os": ["linux", "darwin"],
        "requires": { "bins": [] },
        "install": [],
      },
  }
---

# CV Summary Generator for JobCLI

One-time (or on-CV-update) skill that reads one or more PDF resumes, produces a
privacy-safe persona summary per user, generates job-search keywords, and writes
everything to a user-scoped `CVSUMMARY.md`. These files are consumed by
**SKILL-jobcli-ranking.md** for per-user job searching and ranking.

> **Trigger:** the user provides one or more `.pdf` CV/resume files.

---

## 0 — Multi-User Inputs (Required)

For each CV, collect a stable user identifier before processing:

- `user_id` (required): lowercase slug, only `[a-z0-9_-]` (examples: `ana`, `user_02`, `candidate-sf`)
- `cv_pdf_path` (required): path to that user's CV PDF
- `default_location` (optional but recommended): city/region or `Remote`
- `default_country_code` (optional but recommended): ISO-3166-1 alpha-2 (for example `US`, `MX`, `ES`)

Storage convention (must be followed):

- `profiles/<user_id>/CVSUMMARY.md`

Never store multiple users in one shared `CVSUMMARY.md`.

---

## 1 — Read the CV (Per User)

Use your PDF reading capability to ingest the full text of that user's CV.

---

## 2 — Generate the Ultra-Compact Persona Summary (Per User)

Produce a **concise professional summary** (max 120 words) that captures:

- Years of experience and seniority level.
- Core technical skills, tools, and frameworks.
- Domain expertise and industry sectors.
- Relevant certifications or education level (degree field only, no institution names).
- Language proficiency (spoken languages and fluency levels).
- Target role type (e.g., full-time, contract, remote preference).

**Privacy rules — NEVER include:**

- Full name, email, phone number, or physical address.
- Employer names, university names, or any other personally identifiable information.
- Dates of birth, nationalities, or ID numbers.

---

## 3 — Generate Search Keywords

Produce **exactly 20 keyword phrases** that will be used as `jobcli search`
queries. Each keyword phrase should represent a **realistic job title** that
matches the persona's skills and experience level.

Generate them in two groups:

1. **English keywords** (10 phrases) — use standard English job-market titles
   (e.g., "Senior Backend Engineer", "DevOps Team Lead").
2. **Original-language keywords** (10 phrases) — translate or adapt the same
   intent into the persona's primary working language if it is not English.
   If the persona's language **is** English, generate 5 alternative/synonym
   English titles instead (e.g., "Software Developer" vs "Software Engineer").

Each keyword should be 2–5 words and suitable for direct use in
`jobcli search "<keyword>"`.

---

## 4 — Save to `profiles/<user_id>/CVSUMMARY.md`

Create (or overwrite) a file named **`profiles/<user_id>/CVSUMMARY.md`** with
the following structure:

```markdown
# CV Summary

## User Context

- User ID: <user_id>
- Default Location: <default_location or Unknown>
- Default Country Code: <default_country_code or Unknown>

## Persona Summary

<the ultra-compact persona summary>

## Search Keywords

### English

1. <keyword 1>
2. <keyword 2>
3. <keyword 3>
4. <keyword 4>
5. <keyword 5>
6. <keyword 6>
7. <keyword 7>
8. <keyword 8>
9. <keyword 9>
10. <keyword 10>

### Original Language (<language name>)

1. <keyword 1>
2. <keyword 2>
3. <keyword 3>
4. <keyword 4>
5. <keyword 5>
6. <keyword 6>
7. <keyword 7>
8. <keyword 8>
9. <keyword 9>
10. <keyword 10>

## Ranking Criteria

Use the persona summary above to score each job from 0.0 to 1.0 based on:

- Title match (does the job title align with the persona's level and domain?)
- Skill overlap (how many of the persona's core skills appear in the description?)
- Domain fit (is the industry/sector relevant?)
- Seniority alignment (junior/mid/senior match)
- Language requirements (does the persona meet them?)
```

Confirm the file was written and show its contents to the user.

If multiple users were provided, repeat Steps 1-4 for each `(user_id, cv_pdf_path)` pair.

---

## Notes

- **Privacy first:** never expose personal data from the CV in the summary,
  keywords, or any output file.
- **Isolation rule:** each user must map to exactly one folder under `profiles/`;
  never merge summaries across users.
- **Re-run trigger:** run this skill again only when the user provides a new
  or updated CV for that specific `user_id`.
- **Next step:** once `profiles/<user_id>/CVSUMMARY.md` exists, use
  **SKILL-jobcli-ranking.md** to search and rank jobs for that same `user_id`.
