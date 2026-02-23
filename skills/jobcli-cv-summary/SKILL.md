---
name: jobcli-cv-summary
description: Build a single JobCLI query/persona JSON from a CV. Use when a user provides a CV and downstream steps must consume `profiles/<user_id>/persona_querie.json` via `jobcli --query-file` and the discriminator script.
---

# CV Persona JSON Builder

Goal: create one canonical JSON persona artifact from a CV.

Do not create `CVSUMMARY.md`.

## Inputs
- `user_id` (slug: `[a-z0-9_-]`)
- `cv_pdf_path`
- `default_location` (optional)
- `default_country_code` (optional, ISO-3166-1 alpha-2)

## Required Outputs
Write under `profiles/<user_id>/`:
- `resume.pdf`
- `persona_querie.json`

## `persona_querie.json` Contract
Use this structure (same `--query-file` format as README/docs, plus persona data from the old markdown file):

```json
{
  "job_titles": [
    "Software Engineer",
    "Backend Engineer",
    "Platform Engineer",
    "DevOps Engineer",
    "Site Reliability Engineer",
    "Cloud Engineer",
    "Softwareentwickler",
    "Backend Entwickler",
    "Plattformingenieur",
    "DevOps Ingenieur",
    "SRE Ingenieur",
    "Cloud Ingenieur"
  ],
  "search_options": {
    "location": "Munich, Germany",
    "country": "de",
    "sites": "all",
    "limit": 20,
    "hours": 48,
    "seen": "profiles/<user_id>/jobs_seen.json",
    "seen_update": true,
    "new_only": true,
    "output": "profiles/<user_id>/jobs_new_all.json"
  },
  "global_options": {
    "json": true,
    "plain": false,
    "color": "auto",
    "verbose": false
  },
  "persona": {
    "user_id": "<user_id>",
    "default_location": "Munich, Germany",
    "default_country_code": "de",
    "keywords_en": ["Software Engineer", "Backend Engineer", "Platform Engineer", "DevOps Engineer", "Site Reliability Engineer", "Cloud Engineer"],
    "keywords_local": ["Softwareentwickler", "Backend Entwickler", "Plattformingenieur", "DevOps Ingenieur", "SRE Ingenieur", "Cloud Ingenieur"],
    "target_roles": ["Software Engineer", "Backend Engineer"],
    "excluded_roles_or_domains": ["Internship", "Working Student"],
    "seniority_target": "Mid-Senior",
    "must_have_skills": ["Go", "APIs", "Distributed Systems"],
    "preferred_skills": ["Kubernetes", "AWS", "CI/CD"],
    "work_mode": "Flexible",
    "location_constraints": ["Germany", "EU Remote"],
    "language_requirements": ["English B2+", "German B1+"]
  }
}
```

## Rules
1. Keep content short, explicit, and machine-friendly in JSON only.
2. Keep `job_titles` as realistic job titles only.
3. `job_titles` must include all persona keywords that were previously stored in markdown keyword sections.
4. Keep both `persona.keywords_en` and `persona.keywords_local`, with exactly 6 items each.
5. Include hard exclusions to prevent cross-domain matches.
6. Always include these default exclusions for non-student users:
   - Werkstudent
   - Intern
   - Internship
   - Praktikant
   - Student Assistant
   - Working Student
   - HiWi
   Add them to `persona.excluded_roles_or_domains` unless the user explicitly wants student positions.
7. Do not include skills, tools, technologies, certifications, or generic terms in `job_titles`.
8. Remove personal identifiers (name, email, phone, address, IDs, employer/school names).
9. Keep user data isolated under each `profiles/<user_id>/` directory.

## Validation
Per user, confirm:
1. `resume.pdf` exists.
2. `persona_querie.json` exists and is valid JSON.
3. `persona_querie.json` contains `job_titles`, `search_options`, `global_options`, and `persona`.
4. `job_titles` include all entries from `persona.keywords_en` and `persona.keywords_local` (case-insensitive set containment).
5. Every keyword in `job_titles`, `persona.keywords_en`, and `persona.keywords_local` is a realistic job title phrase.

Next skill: `skills/jobcli-job-search/SKILL.md`.
