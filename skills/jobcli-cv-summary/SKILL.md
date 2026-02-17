---
name: jobcli-cv-summary
description: Build persona summary + JSON profile files for strict YES/NO job filtering.
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

# CV Persona Builder (Summary + JSON)

Goal: produce a concise `CVSUMMARY.md` that can drive binary job filtering.

Trigger: the user provides one or more `.pdf` CV files.

## 0) Inputs (Required)

Per CV, collect:

- `user_id`: lowercase slug `[a-z0-9_-]`
- `cv_pdf_path`: source CV path
- `default_location`: optional
- `default_country_code`: optional ISO-3166-1 alpha-2

## 1) Output Contract (Required)

Write these files per user:

- `profiles/<user_id>/resume.pdf`
- `profiles/<user_id>/CVSUMMARY.md`
- `profiles/<user_id>/persona_profile.json`

`CVSUMMARY.md` is the primary input for `jobcli-job-search`.

## 2) Extract and Normalize Persona

From CV text, extract:

1. `target_roles` (allowed role families/titles)
2. `excluded_roles_or_domains` (explicit negatives)
3. `seniority_target`: Junior | Mid | Mid-Senior | Senior | Staff+
4. `must_have_skills`: array, max 12
5. `preferred_skills`: array, max 15
6. `work_mode`: Remote | Hybrid | Onsite | Flexible
7. `location_constraints`: array
8. `language_requirements`: array

## 3) Write Persona Summary Markdown

Create `profiles/<user_id>/CVSUMMARY.md` with this structure:

```md
# Persona Summary

## Target Roles
- <role_family_or_title_1>
- <role_family_or_title_2>

## Excluded Roles/Domains
- <excluded_role_or_domain_1>
- <excluded_role_or_domain_2>

## Seniority Target
- <Junior|Mid|Mid-Senior|Senior|Staff+>

## Must-Have Skills
- <must_have_skill_1>
- <must_have_skill_2>

## Preferred Skills
- <preferred_skill_1>
- <preferred_skill_2>

## Work Mode
- <Remote|Hybrid|Onsite|Flexible>

## Location Constraints
- <location_constraint_1>
- <location_constraint_2>

## Language Requirements
- <language_requirement_1>
- <language_requirement_2>
```

Rules:

1. Keep it short and explicit.
2. Include hard exclusions to prevent wrong-domain matches.
3. Use bullet lists only for machine-friendly parsing.
4. No personal identifiers.

## 4) Save JSON Companion

Create `profiles/<user_id>/persona_profile.json` with the same semantic fields:

```jsonc
{
  "user_id": "<user_id>",
  "default_location": "<value-or-Unknown>",
  "default_country_code": "<value-or-Unknown>",
  "target_roles": ["<role_family_or_title_1>", "<role_family_or_title_2>"],
  "excluded_roles_or_domains": ["<excluded_role_or_domain_1>", "<excluded_role_or_domain_2>"],
  "seniority_target": "<Junior|Mid|Mid-Senior|Senior|Staff+>",
  "must_have_skills": ["<must_have_skill_1>", "<must_have_skill_2>"],
  "preferred_skills": ["<preferred_skill_1>", "<preferred_skill_2>"],
  "work_mode": "<Remote|Hybrid|Onsite|Flexible>",
  "location_constraints": ["<location_constraint_1>", "<location_constraint_2>"],
  "language_requirements": ["<language_requirement_1>", "<language_requirement_2>"]
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
2. `profiles/<user_id>/CVSUMMARY.md` exists
3. `profiles/<user_id>/persona_profile.json` exists
4. summary includes target + excluded + seniority + must-have + work-mode + location
5. JSON schema fields are present

If multiple users were provided, repeat steps 0-6 per user.

## Notes

- users must stay isolated by folder
- re-run only when CV changes
- next step: run `skills/jobcli-job-search/SKILL.md`
