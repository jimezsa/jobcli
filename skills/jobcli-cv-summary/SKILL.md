---
name: jobcli-cv-summary
description: Build minimal persona artifacts for strict YES/HIGH job filtering.
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

# CV Persona Builder (Minimal)

Goal: create `CVSUMMARY.md` and `persona_profile.json` from a CV.

Trigger: user provides one or more CV PDFs.

## Inputs
Per user:
- `user_id` (slug: `[a-z0-9_-]`)
- `cv_pdf_path`
- `default_location` (optional)
- `default_country_code` (optional, ISO-3166-1 alpha-2)

## Required Outputs
Write under `profiles/<user_id>/`:
- `resume.pdf`
- `CVSUMMARY.md`
- `persona_profile.json`

## Required Persona Fields
- `keywords_en` (array, exactly 6 realistic job titles in English)
- `keywords_local` (array, exactly 6 realistic job titles in country language)
- `target_roles` (array)
- `excluded_roles_or_domains` (array)
- `seniority_target` (`Junior|Mid|Mid-Senior|Senior|Staff+`)
- `must_have_skills` (array, max 12)
- `preferred_skills` (array, max 15)
- `work_mode` (`Remote|Hybrid|Onsite|Flexible`)
- `location_constraints` (array)
- `language_requirements` (array)

## CVSUMMARY.md Structure
Use this exact section layout with bullet lists:
- `# Persona Summary`
- `## Keywords (English)`
- `## Keywords (Local Language)`
- `## Target Roles`
- `## Excluded Roles/Domains`
- `## Seniority Target`
- `## Must-Have Skills`
- `## Preferred Skills`
- `## Work Mode`
- `## Location Constraints`
- `## Language Requirements`

## persona_profile.json Contract
Include these keys:
- `user_id`
- `default_location`
- `default_country_code`
- `keywords_en`
- `keywords_local`
- `target_roles`
- `excluded_roles_or_domains`
- `seniority_target`
- `must_have_skills`
- `preferred_skills`
- `work_mode`
- `location_constraints`
- `language_requirements`

## Rules
1. Keep content short, explicit, and machine-friendly.
2. Include hard exclusions to prevent cross-domain matches.
3. `## Keywords (English)` is required in `CVSUMMARY.md` and must contain exactly 6 keywords.
4. `## Keywords (Local Language)` is required in `CVSUMMARY.md` and must contain exactly 6 keywords in the country language.
5. Keywords are for job search only: each keyword must be a realistic job title used in job boards.
6. Do not include skills, tools, technologies, certifications, or generic terms as keywords.
7. Use one normalized job-title keyword per bullet and avoid duplicates inside each list.
8. Remove personal identifiers (name, email, phone, address, IDs, employer/school names).
9. Keep user data isolated under each `profiles/<user_id>/` directory.

## Validation
Per user, confirm:
1. `resume.pdf` exists.
2. `CVSUMMARY.md` exists and includes both required keyword sections.
3. `CVSUMMARY.md` contains exactly 6 bullets in `## Keywords (English)` and exactly 6 bullets in `## Keywords (Local Language)`.
4. Every keyword in both sections is a job title phrase (not a skill/tool/certification term).
5. `persona_profile.json` exists and has all required keys (including `keywords_en` and `keywords_local`).

Next skill: `skills/jobcli-job-search/SKILL.md`.
