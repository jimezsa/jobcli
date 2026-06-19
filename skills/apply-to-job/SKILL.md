---
name: apply-to-job
description: End-to-end application workflow for a single job posting. Tailors the LaTeX CV to the job via `tailor-latex-cv-to-job`, then drives the application form in a real browser via `browser-use`, fills fields from a cached application profile, uploads the tailored PDF, stops on the review page with a screenshot and summary for human approval, and (only after explicit human approval in chat) clicks the final submit button itself.
---

# apply-to-job

Orchestrates one application end-to-end: tailor the CV, open the posting in a real browser, fill the form from cached profile data, upload the PDF, screenshot the filled form, get explicit human approval in chat, then click submit and record the confirmation.

## HARD RULES

1. **Click the final Submit / Send / Bewerben absenden button ONLY after explicit human approval in chat.** Approval counts when the human sends a clear go-ahead message — "apply", "submit", "send it", "go", "yes submit", or similar unambiguous approval — referring to the current pre-submit screenshot/summary. Without that approval, stop on the review page and wait. Never click submit unprompted, never interpret silence or an unrelated "ok" as approval. After clicking, take a confirmation screenshot and report the outcome.
2. **Never edit the master CV** at `cv/<source-cv>.tex`. All CV work happens inside the application workspace via `tailor-latex-cv-to-job`.
3. **Respect the user's scope filter.** Apply only to roles that match the user's `persona_querie.json` scope (e.g. Werkstudent / Working Student, Masterarbeit / Master Thesis, full-time, etc.). If the posting drifts from the configured scope at apply-time, abort and report.
4. **No cover letter unless the form explicitly requires one and blocks progress without it.** If forced, generate a short truthful one in the workspace and surface it for review before upload.
5. **No fabricated data.** If the form asks for a value not in `application_profile.json` and the human is not reachable, stop and ask. Never invent visa status, salary, GPA, certifications, or experience.
6. **Account creation still requires explicit human approval.** Submitting a filled-out form is reversible-ish via withdrawal; creating a new account is not, and tenant rules vary. Always ask before creating a new ATS account.

## Required Inputs

One of:
- A job URL (LinkedIn, company careers page, Greenhouse, Workday, etc.)
- An index or 0-based slot reference into `profiles/<user_id>/jobs_yes_high.json`
- A JSON snippet `{title, company, url}` pasted by the human

If only a URL is given, derive `title` and `company` from the page during step 1.

## Required Resources

- `profiles/<user_id>/application_profile.json` — cached candidate data. Auto-created by this skill if missing.
- `applications/APPLICATIONS.md` — **canonical applications log**. This is the skill's persistent memory of every application driven by this skill. Read it first, write to it at every status change. Schema is documented inside the file.
- `cv/<source-cv>.tex` — master CV (read-only).
- `skills/tailor-latex-cv-to-job/SKILL.md` — invoked for the CV tailoring stage.
- `skills/browser-use/SKILL.md` — invoked for the browser-driving stage.
- An authenticated browser profile for LinkedIn (see `skills/browser-use/SKILL.md`) when the URL is a `linkedin.com` posting.

## Workspace Layout

For every application, create:

```
applications/<YYYY-MM-DD>-<company-slug>-<role-slug>/
├── job.json              # {site, url, title, company, location, description, source}
├── job-description.txt   # plain text job description (handed to tailor skill)
├── CV_<Name>.tex         # tailored LaTeX (from tailor skill, named to match the master CV)
├── CV_<Name>.pdf         # compiled PDF (from tailor skill) — this is what gets uploaded to recruiters
├── ats-report.md         # ATS report (from tailor skill)
├── form-fill-log.md      # what fields were filled, what value, source
├── screenshots/
│   ├── 01-posting.png
│   ├── 02-form-page-N.png
│   └── 99-pre-submit.png # final pre-submit screenshot for human approval
├── cover-letter.md       # ONLY if the form requires one (rare)
└── STATUS.md             # one-line state: tailored | opened | filling | ready-to-submit | submitted | withdrawn
```

Slug = lowercase, hyphenated, ASCII-only, max 40 chars per segment.

## Workflow

### Step 0 — Check the applications log

Read `applications/APPLICATIONS.md` first. If a row already exists for this posting (match on `url`, or `company` + `role` if URLs differ across mirrors like LinkedIn vs. company portal):

- status is `submitted`, `interview`, or `offer` → STOP. Tell the human we already applied and link to the existing workspace.
- status is `ready-to-submit` → STOP. Surface the existing pre-submit screenshot and ask whether to resume or abandon.
- status is `withdrawn` or `rejected` → ASK the human whether to re-apply before proceeding.
- status is `tailored`, `opened`, or `filling` → resume the existing workspace instead of creating a new one.

If no matching row exists, continue to Step 1.

### Step 1 — Resolve the job

Resolve the input to a `(site, url, title, company, location, description)` tuple.

- If the input is a `jobs_yes_high.json` index, load that record directly.
- If the input is a URL, open it with `browser-use --headed` (use the LinkedIn-authenticated session for `linkedin.com` URLs) and extract page text. If the description is long enough (> 400 chars), use it; otherwise ask the human to paste the description.
- Confirm the role matches the user's configured scope (see `persona_querie.json`). If the posting drifts, ABORT and report to the human.

Create the workspace folder and write `job.json` + `job-description.txt`. The helper `scripts/create_application_workspace.py` does this end-to-end from a `jobs_yes_high.json` index, an inline JSON snippet, or `--url --title --company` flags.

Immediately append a new row to `applications/APPLICATIONS.md` (newest at top of the Log section) with `status: tailored`, the workspace path, and the source. This is the moment the application enters the log — do not wait until Step 8.

### Step 2 — Tailor the CV

Invoke `skills/tailor-latex-cv-to-job/SKILL.md` against:

- source: `cv/<source-cv>.tex`
- job description: the saved `job-description.txt`
- output root: the current application workspace (use `--out-root` so the workspace is the tailor workspace too)
- tex name: pass `--tex-name CV_<Name>.tex` to `create_cv_workspace.py` so the artifacts end up as `CV_<Name>.tex` / `CV_<Name>.pdf`. **Never use a filename containing the word "tailored" for the uploaded PDF** — recruiters see the filename and a "tailored" tag reads as awkward. Match the master CV's naming pattern instead.

The tailor skill will produce `CV_<Name>.tex`, `CV_<Name>.pdf`, and `ats-report.md`. Verify the PDF exists and the ATS report has no critical parseability warnings. If the ATS report flags real gaps, surface them to the human in the final summary — do not paper them over with fabricated content.

Update `STATUS.md` to `tailored`. The log row from Step 1 already says `tailored`; no log update needed yet.

### Step 3 — Load the application profile

Read `profiles/<user_id>/application_profile.json`.

For each field listed in `runtime_prompts` that is currently null, ask the human in chat once for the value, write it back to `application_profile.json`, and only then proceed. Cache means we don't ask twice.

Visa status, salary expectation, etc. fall in this bucket — ask once, cache forever (until the human says otherwise).

### Step 4 — Open the posting in the browser

Use `browser-use` per `skills/browser-use/SKILL.md`. Always run headed so the human can see what's happening on a visible display.

For LinkedIn URLs use the persistent profile + the `linkedin` named session:

```bash
browser-use --headed --profile --session linkedin open "<url>"
```

For everything else use a per-application session name (e.g. `apply-<company-slug>`):

```bash
browser-use --headed --session "apply-<company-slug>" open "<url>"
```

Save `screenshots/01-posting.png`. Update `STATUS.md` to `opened` **and** update the row's `status` in `applications/APPLICATIONS.md` to `opened`.

### Step 5 — Navigate to the application form

Click the visible "Apply" / "Easy Apply" / "Bewerben" / "Jetzt bewerben" button. On LinkedIn many postings deep-link out to the company's ATS portal — follow the link.

If the site requires creating a new account, STOP and ask the human whether to proceed (account creation has external consequences and should be human-decided).

Update `STATUS.md` to `filling` **and** update the row's `status` in `applications/APPLICATIONS.md` to `filling`.

### Step 6 — Fill the form, page by page

For each form page:

1. Run `browser-use state` to list numbered interactive elements.
2. For each labelled field, decide the source of the value using `references/form-fill-hints.md` (label → profile path mapping). Common mappings:
   - `First name` / `Vorname` → `identity.first_name`
   - `Last name` / `Nachname` → `identity.last_name`
   - `Email` / `E-Mail` → `identity.email`
   - `Phone` / `Telefon` → `identity.phone`
   - `Address` / `Adresse` / `Straße` → `address.street`
   - `Postcode` / `PLZ` → `address.postal_code`
   - `City` / `Stadt` / `Wohnort` → `address.city`
   - `Country` / `Land` → `address.country`
   - `LinkedIn` → `links.linkedin`
   - `GitHub` / `Portfolio` / `Website` → `links.github` or `links.website`
   - `Visa` / `Work authorization` / `Arbeitserlaubnis` → `work_authorization.visa_status`
   - `Earliest start` / `Verfügbar ab` / `Eintrittsdatum` → `availability.earliest_start_date`
   - `Hours per week` / `Wochenstunden` → `availability.weekly_hours`
   - `Expected salary` / `Gehaltsvorstellung` → `compensation.expected_hourly_eur` or `expected_monthly_eur`
   - `English level` → `languages.english`
   - `German level` / `Deutschkenntnisse` → `languages.german`
3. Use `browser-use input <index> "value"` for text, `browser-use select <index> "option"` for dropdowns, `browser-use click <index>` for checkboxes.
4. For unknown fields, ask the human in chat. Do not guess.
5. For file upload fields, upload `CV_<Name>.pdf`. Use the appropriate `browser-use` upload command or simulate a click + file path. If the page requires a separate "cover letter" upload AND it is not optional, generate a short truthful cover letter to `cover-letter.md`, render it to PDF via `pandoc cover-letter.md -o cover-letter.pdf` and upload.
6. Append every action to `form-fill-log.md`:
   - `field | value | source (profile path | human-provided | derived) | element index`
7. Take `screenshots/02-form-page-<N>.png` after each page is filled, before clicking Next/Weiter.

### Step 7 — Stop on the review page and request approval

When the page is the final review / confirmation page (look for buttons labelled `Submit application`, `Send application`, `Bewerbung absenden`, `Abschicken`, `Submit`):

1. Do **NOT** click submit yet.
2. `browser-use screenshot screenshots/99-pre-submit.png`.
3. Update `STATUS.md` to `ready-to-submit` **and** update the row's `status` in `applications/APPLICATIONS.md` to `ready-to-submit`.
4. Surface the screenshot to the human via the agent's preferred file-delivery mechanism.
5. Tell the human in chat:
   - workspace path
   - one-paragraph summary of the fields filled, the CV uploaded, and any honest gaps from the ATS report or fallbacks used
   - the explicit ask: reply **"apply" / "submit" / "go"** to have the agent click submit itself, **"wait"** to leave it parked, **"change X"** to edit before submitting, or **"withdraw"** to abandon.

### Step 8 — Click submit on approval

When the human replies with explicit approval ("apply" / "submit" / "send it" / "go" / "yes submit" / equivalent):

1. Re-focus the browser session: `browser-use --session "<name>" state` to verify the review page is still the active page and the submit button is still present. If the page has navigated away or timed out, take a screenshot, report it, and stop — do not blindly re-fill.
2. Identify the submit button by label (`Submit`, `Submit application`, `Send application`, `Apply`, `Bewerben`, `Bewerbung absenden`, `Abschicken`) using `browser-use state`. If multiple buttons match, prefer the one at the bottom of the form and screenshot first to confirm.
3. `browser-use click <index>` on the submit button.
4. Wait briefly for the confirmation page to load. Take `screenshots/100-post-submit.png`.
5. Update `STATUS.md` to `submitted` **and** update the row's `status` in `applications/APPLICATIONS.md` to `submitted`, with the submit timestamp in `notes`.
6. Surface `screenshots/100-post-submit.png` to the human so the confirmation page is visible.
7. Close the named browser session: `browser-use --session "<name>" close`.

If the click does not produce a confirmation page (validation error toast, network error, captcha), screenshot the result, report it honestly, leave status as `ready-to-submit`, and ask the human what to do. Do not retry blindly or fabricate a confirmation.

If the human's reply is **"wait"**, leave status at `ready-to-submit` and stop.
If the reply is **"change <field>"**, edit the indicated field via `browser-use` and return to Step 7.
If the reply is **"withdraw"**, update status to `withdrawn`, close the browser session, and stop.

### Step 9 — Follow-up updates

If the human later reports a follow-up (rejection email, interview invite, offer), update only the log row: bump `status` to `rejected` / `interview` / `offer` and add a short dated note to the `notes` column.

## Helper Script

`scripts/create_application_workspace.py` creates the workspace folder + `job.json` + `job-description.txt` from either a `jobs_yes_high.json` index or a JSON snippet. See the script's `--help` for usage.

## Done Criteria

- Workspace exists with tailored CV PDF, ATS report, screenshots, and `form-fill-log.md`.
- `STATUS.md` ends in `ready-to-submit`, `submitted`, or `withdrawn`.
- `applications/APPLICATIONS.md` has a row for this posting and its `status` field matches `STATUS.md`.
- `99-pre-submit.png` was surfaced to the human for approval.
- If the human approved, `100-post-submit.png` (the confirmation page) was captured and surfaced too.
- `application_profile.json` has been updated with any newly answered runtime prompts.
- Final summary lists: company, role, workspace path, fields the human supplied at runtime, any honest gaps from the ATS report, and clearly states whether the agent clicked submit on the human's approval (with timestamp), is parked at `ready-to-submit`, or was withdrawn.

## Failure Modes & Recovery

- **Posting requires login on a non-LinkedIn careers site** → ask the human to authenticate once in the visible browser, then continue from `browser-use state`.
- **Form spans 5+ pages with custom widgets** → continue field-by-field. If a custom widget (e.g. drag-drop) breaks, screenshot, ask the human to interact directly in the visible browser, then resume.
- **PDF upload silently fails** → check the page state for an error toast. Re-upload. If still failing, ask the human to upload manually in the open browser.
- **The "apply" button leads to an external ATS the agent doesn't know** → run `browser-use state` on the new page and resume from Step 6. Most ATSes (Greenhouse, Lever, SmartRecruiters, Workday, vendor-hosted portals) follow the same field-by-field pattern.
- **Browser session crashes mid-fill** → reopen with the same `--session` name, navigate back to the form, and resume from the last logged field in `form-fill-log.md`.
