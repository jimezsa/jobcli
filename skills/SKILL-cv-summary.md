---
name: jobcli-cv-summary
description: Extract an anonymous persona summary and search keywords from a CV PDF.
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

One-time (or on-CV-update) skill that reads a PDF résumé, produces a
privacy-safe persona summary, generates job-search keywords, and writes
everything to `CVSUMMARY.md`. That file is then consumed by the companion
skill **SKILL-jobsearch.md** for daily job searching and ranking.

> **Trigger:** the user provides a `.pdf` CV/resume file.

---

## 1 — Read the CV

Use your PDF reading capability to ingest the full text of the provided CV.

---

## 2 — Generate the Ultra-Compact Persona Summary

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

Produce **exactly 10 keyword phrases** that will be used as `jobcli search`
queries. Each keyword phrase should represent a **realistic job title** that
matches the persona's skills and experience level.

Generate them in two groups:

1. **English keywords** (5 phrases) — use standard English job-market titles
   (e.g., "Senior Backend Engineer", "DevOps Team Lead").
2. **Original-language keywords** (5 phrases) — translate or adapt the same
   intent into the persona's primary working language if it is not English.
   If the persona's language **is** English, generate 5 alternative/synonym
   English titles instead (e.g., "Software Developer" vs "Software Engineer").

Each keyword should be 2–5 words and suitable for direct use in
`jobcli search "<keyword>"`.

---

## 4 — Save to CVSUMMARY.md

Create (or overwrite) a file named **`CVSUMMARY.md`** in the current working
directory with the following structure:

```markdown
# CV Summary

## Persona Summary

<the ultra-compact persona summary>

## Search Keywords

### English

1. <keyword 1>
2. <keyword 2>
3. <keyword 3>
4. <keyword 4>
5. <keyword 5>

### Original Language (<language name>)

1. <keyword 1>
2. <keyword 2>
3. <keyword 3>
4. <keyword 4>
5. <keyword 5>

## Ranking Criteria

Use the persona summary above to score each job from 0.0 to 1.0 based on:
- Title match (does the job title align with the persona's level and domain?)
- Skill overlap (how many of the persona's core skills appear in the description?)
- Domain fit (is the industry/sector relevant?)
- Seniority alignment (junior/mid/senior match)
- Language requirements (does the persona meet them?)
```

Confirm the file was written and show its contents to the user.

---

## Notes

- **Privacy first:** never expose personal data from the CV in the summary,
  keywords, or any output file.
- **Re-run trigger:** run this skill again only when the user provides a new
  or updated CV.
- **Next step:** once `CVSUMMARY.md` exists, use **SKILL-jobsearch.md** to
  search for jobs and rank them against the persona.
