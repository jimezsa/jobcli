---
name: jobcli-cv-summary
description: Extract per-user machine-readable persona profiles and keyword banks from CV PDFs.
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

# CV Persona Builder (JSON-only)

Goal: produce one machine-readable persona JSON per user from CV PDFs.

Trigger: the user provides one or more `.pdf` CV files.

## 0) Inputs (Required)

Per CV, collect:

- `user_id`: lowercase slug `[a-z0-9_-]`
- `cv_pdf_path`: source CV path
- `default_location`: optional
- `default_country_code`: optional ISO-3166-1 alpha-2

## 1) Output Contract (Only JSON)

Write only these files:

- `profiles/<user_id>/resume.pdf`
- `profiles/<user_id>/persona_profile.json`

Do not generate `CVSUMMARY.md`.

## 2) Extract and Normalize Persona

From CV text, extract:

- `seniority_target`: Junior | Mid | Mid-Senior | Senior | Staff+
- `role_family`: array
- `must_have_skills`: array, max 12
- `preferred_skills`: array, max 15
- `excluded_areas`: array
- `work_mode`: Remote | Hybrid | Onsite | Flexible
- `location_constraints`: array
- `language_requirements`: array

## 3) Generate Keywords

Generate exactly 20 realistic job-position titles (2-5 words each):

- 10 English
- 10 original language (or English variants if primary language is English)

Also produce bucketed query sets:

- `core_role`
- `skill_intent`
- `domain_seniority`

Token rules:

- remove generic low-intent terms
- de-duplicate semantic equivalents
- use market-standard position titles only
- do not use skill-only/tool-only phrases (for example `Python`, `Kubernetes`)

## 4) Save JSON

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

## 5) Privacy Rules

Never include:

- name, email, phone, address
- employer/school names
- DOB, nationality, IDs

## 6) Validation

Confirm per user:

1. `profiles/<user_id>/resume.pdf` exists
2. `profiles/<user_id>/persona_profile.json` exists
3. JSON schema fields are present

If multiple users were provided, repeat steps 0-6 per user.

## Notes

- users must stay isolated by folder
- re-run only when CV changes
- next step: run `skills/jobcli-job-search/SKILL.md`
