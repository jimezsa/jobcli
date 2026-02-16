---
name: jobcli-cv-summary
description: Extract per-user anonymous persona summaries, machine-readable persona profiles, and keyword banks from CV PDFs.
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

# CV Summary Generator for JobCLI (v2)

Generate a privacy-safe persona package per user from CV PDFs.

Outputs from this skill are consumed by `skills/jobcli-job-search/SKILL.md`.

Trigger: the user provides one or more `.pdf` CV files.

## 0) Multi-User Inputs (Required)

For each CV, collect:

- `user_id` (required): lowercase slug, only `[a-z0-9_-]`
- `cv_pdf_path` (required): source CV path
- `default_location` (optional but recommended): city/region or `Remote`
- `default_country_code` (optional but recommended): ISO-3166-1 alpha-2 (for example `US`, `MX`, `ES`)

Storage convention:

- `profiles/<user_id>/resume.pdf`
- `profiles/<user_id>/CVSUMMARY.md`
- `profiles/<user_id>/persona_profile.json`

Never mix users in one shared summary/profile file.

## 1) Read CV Text (Per User)

Read the CV PDF and extract text needed for:

- skills
- role level
- domains
- languages
- work-mode/location preference

Do not copy raw CV text into outputs.

## 2) Write Persona Summary (Human-Readable)

Create `## Persona Summary` in max 120 words:

- years of experience and seniority
- core stack and tooling
- domain/industry focus
- education/certification category (no school names)
- language proficiency
- target role/work mode

Privacy rules (strict):

- never include name, email, phone, address
- never include employer, school, or personal identifiers
- never include DOB, nationality, ID data

## 3) Build Persona Profile v2 (Machine-Readable)

Extract structured fields:

- `seniority_target`: Junior | Mid | Mid-Senior | Senior | Staff+
- `role_family`: list (for example `["Backend","Platform"]`)
- `must_have_skills`: list (max 12)
- `preferred_skills`: list (max 15)
- `excluded_areas`: list (for example `["Frontend-only","QA-only"]`)
- `work_mode`: Remote | Hybrid | Onsite | Flexible
- `location_constraints`: list of countries/regions/cities
- `language_requirements`: list with level
- `compensation_floor`: optional object `{amount,currency}`

## 4) Generate Search Keyword Bank

Produce exactly 20 realistic title queries, each 2-5 words:

1. English: 10 queries
2. Original language (or English alternatives): 10 queries

Also classify each keyword with an intent bucket:

- `core_role`
- `skill_intent`
- `domain_seniority`

Token-efficiency rule:

- no generic low-intent queries (for example `Engineer`, `Developer`)
- de-duplicate semantic equivalents

## 5) Save Outputs

Copy source CV:

- `cv_pdf_path` -> `profiles/<user_id>/resume.pdf` (overwrite allowed)

Create `profiles/<user_id>/CVSUMMARY.md`:

```markdown
# CV Summary

## User Context

- User ID: <user_id>
- Default Location: <default_location or Unknown>
- Default Country Code: <default_country_code or Unknown>

## Persona Summary

<max-120-word summary>

## Persona Profile v2

- Seniority Target: <...>
- Role Family: <...>
- Must-Have Skills: <...>
- Preferred Skills: <...>
- Excluded Areas: <...>
- Work Mode: <...>
- Location Constraints: <...>
- Language Requirements: <...>
- Compensation Floor: <optional>

## Search Keywords

### English

1. <keyword>
2. <keyword>
3. <keyword>
4. <keyword>
5. <keyword>
6. <keyword>
7. <keyword>
8. <keyword>
9. <keyword>
10. <keyword>

### Original Language (<language name>)

1. <keyword>
2. <keyword>
3. <keyword>
4. <keyword>
5. <keyword>
6. <keyword>
7. <keyword>
8. <keyword>
9. <keyword>
10. <keyword>
```

Create `profiles/<user_id>/persona_profile.json`:

```json
{
  "user_id": "<user_id>",
  "default_location": "<value-or-Unknown>",
  "default_country_code": "<value-or-Unknown>",
  "seniority_target": "Mid-Senior",
  "role_family": ["Backend", "Platform"],
  "must_have_skills": ["Go", "SQL", "APIs", "Docker"],
  "preferred_skills": ["Kubernetes", "AWS", "Terraform"],
  "excluded_areas": ["Frontend-only", "QA-only"],
  "work_mode": "Remote",
  "location_constraints": ["US", "Canada"],
  "language_requirements": ["English (Fluent)"],
  "compensation_floor": { "amount": 130000, "currency": "USD" },
  "keywords": {
    "english": ["...10 items..."],
    "original_language": ["...10 items..."],
    "bucketed": {
      "core_role": ["..."],
      "skill_intent": ["..."],
      "domain_seniority": ["..."]
    }
  }
}
```

## 6) Validation Checklist

Before finishing, confirm:

1. `profiles/<user_id>/resume.pdf` exists
2. `profiles/<user_id>/CVSUMMARY.md` exists
3. `profiles/<user_id>/persona_profile.json` exists
4. summary is privacy-safe
5. keywords are 2-5 words and de-duplicated

Show the generated summary to the user (not raw CV text).

If multiple users were provided, run steps 1-6 per user.

## Notes

- privacy first: never expose personal data from CV content
- isolation rule: each user has its own folder
- re-run this skill only when that user CV changes
- next step: run `skills/jobcli-job-search/SKILL.md` for search and ranking
