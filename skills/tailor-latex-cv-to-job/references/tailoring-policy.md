# Tailoring Policy

Use this reference when deciding what may be changed in a CV for a target job.

## Core Rule

Optimize alignment, not truth. Every tailored claim must be supported by the source CV or by explicit user-provided facts.

ATS systems vary by vendor and employer. This workflow improves parseability and keyword alignment, but it cannot guarantee passing a filter.

## Allowed Edits

- Reorder existing skills so job-required skills appear earlier.
- Rephrase bullets to use the job posting's language when the underlying experience is already present.
- Expand abbreviations and add synonyms, such as `Amazon Web Services (AWS)`, when the CV already supports the tool.
- Emphasize relevant responsibilities in the summary/profile.
- Reorder projects or selected bullets when chronology and meaning remain honest.
- Remove, compress, or de-emphasize less relevant content to preserve length.
- Add a skills subsection only from technologies already present in the CV or explicitly supplied by the user.
- Make contact details text-readable if they were icon-only.

## Forbidden Edits

- Do not invent employers, titles, dates, degrees, certifications, tools, languages, metrics, publications, clearances, visas, locations, or authorization status.
- Do not claim professional use of a tool that appears only as a job requirement.
- Do not add hidden text, white text, zero-size text, metadata keyword stuffing, or off-page keyword dumps.
- Do not alter seniority, education, or certification claims to satisfy a posting.
- Do not change chronology to hide gaps or imply roles that did not exist.
- Do not translate the CV unless the user asks.

## ATS Readability Checklist

- PDF text extracts with `pdftotext`.
- Name, email, phone, LinkedIn, GitHub, and portfolio links are selectable text where possible.
- Section names use common labels, such as `Summary`, `Experience`, `Skills`, `Projects`, `Education`, `Certifications`, and `Languages`.
- The reading order is coherent in extracted text.
- Important skills are not embedded only in images, icons, headers, footers, or decorative sidebars.
- Tables, multi-column layouts, and custom positioning do not scramble the extracted order. If they do, report the risk and offer an ATS-first version.
- Bullet points use plain text glyphs that extract cleanly.

## Keyword Coverage Guidance

Prioritize required skills and responsibilities over generic job-posting language. A missing keyword is fixable only when the source CV supports it.

Treat these as high-value matches:

- exact technology names
- common acronyms and expansions
- target role title terms
- domain terms repeated in required responsibilities
- certifications or methodologies explicitly requested by the job

Treat these as low-value terms:

- company values
- generic words such as motivated, collaborative, fast-paced, stakeholder, team, and excellent
- requirements the candidate does not meet

## Reporting Gaps

When a requirement is not supported, report it clearly:

```text
Gap: Job asks for Kubernetes. The source CV does not show Kubernetes experience, so it was not added.
```

When a term is supported but absent, revise naturally:

```text
Fixable: Source CV mentions Docker-based deployments. The job asks for containerization, so the deployment bullet can include "containerized services with Docker."
```
