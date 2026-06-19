# Form Fill Hints

Label → `application_profile.json` path. Use this as a default mapping when scanning a form.
Match labels case-insensitively, trim trailing punctuation, ignore "*" and "(required)".
DE/EN are both common in German postings.

## Identity

| Label (EN)              | Label (DE)              | Profile path                  |
|-------------------------|-------------------------|-------------------------------|
| First name              | Vorname                 | identity.first_name           |
| Last name / Surname     | Nachname / Familienname | identity.last_name            |
| Full name               | Vollständiger Name      | identity.first_name + last_name |
| Preferred name / Nickname | Rufname               | identity.preferred_name       |
| Email                   | E-Mail / E-Mail-Adresse | identity.email                |
| Phone / Mobile          | Telefon / Telefonnummer / Handynummer | identity.phone  |

## Address

| Label (EN)              | Label (DE)              | Profile path                  |
|-------------------------|-------------------------|-------------------------------|
| Street / Address line 1 | Straße / Anschrift      | address.street                |
| Postal code / ZIP       | PLZ / Postleitzahl      | address.postal_code           |
| City                    | Stadt / Wohnort / Ort   | address.city                  |
| Country                 | Land                    | address.country               |

## Links

| Label (EN)              | Label (DE)              | Profile path                  |
|-------------------------|-------------------------|-------------------------------|
| LinkedIn / LinkedIn profile | LinkedIn-Profil     | links.linkedin                |
| GitHub                  | GitHub                  | links.github                  |
| Portfolio / Website     | Webseite / Portfolio    | links.website                 |

## Work authorization

| Label (EN)              | Label (DE)              | Profile path                  |
|-------------------------|-------------------------|-------------------------------|
| Authorized to work in <country>? | Arbeitserlaubnis vorhanden? | work_authorization.visa_status |
| Visa status / Residence permit | Aufenthaltstitel | work_authorization.visa_status |
| Need sponsorship?       | Visumssponsoring benötigt? | work_authorization.needs_sponsorship |

For boolean Y/N forms, derive from visa_status string when possible.

## Availability

| Label (EN)              | Label (DE)              | Profile path                  |
|-------------------------|-------------------------|-------------------------------|
| Earliest start date / Available from | Verfügbar ab / Eintrittsdatum / Startdatum | availability.earliest_start_date |
| Hours per week          | Wochenstunden / Arbeitszeit pro Woche | availability.weekly_hours |
| Duration               | Dauer                    | availability.duration_open_to  |

## Compensation

| Label (EN)              | Label (DE)              | Profile path                  |
|-------------------------|-------------------------|-------------------------------|
| Hourly rate expectation | Stundensatzvorstellung  | compensation.expected_hourly_eur |
| Salary expectation      | Gehaltsvorstellung      | compensation.expected_monthly_eur |

## Languages

| Label (EN)              | Label (DE)              | Profile path                  |
|-------------------------|-------------------------|-------------------------------|
| English level           | Englischkenntnisse      | languages.english             |
| German level            | Deutschkenntnisse       | languages.german              |

## Education

| Label (EN)              | Label (DE)              | Profile path                  |
|-------------------------|-------------------------|-------------------------------|
| Highest degree / Current degree | Aktueller Abschluss / Studiengang | education.degree |
| University              | Universität / Hochschule | education.university          |

## Demographics (optional fields)

Default answer: "Prefer not to say" / "Keine Angabe" unless the human supplied a value in `demographics_optional.*`.

## Common confusion points

- "Bewerbung als" (DE) = "Applying for" — usually the role title; copy from job posting, not the profile.
- "Wie haben Sie von uns erfahren?" / "How did you hear about us?" — answer LinkedIn unless told otherwise.
- "Notice period" / "Kündigungsfrist" — for student / Werkstudent roles typically "None" / "Keine".
- "Are you currently employed?" — student status: answer "Student" or "No, full-time student" depending on the form's options.
- "Diversity / EEO / Gender" sections — always optional in EU; default "Prefer not to say" unless the human said otherwise.

## Unknown fields

When a field has no mapping above:

1. Read its label literally.
2. Check if it's clearly derivable from existing profile data (e.g. "Name as on passport" = first + last).
3. Otherwise STOP, screenshot the field, and ask the human inline.
4. Cache the answer back into `application_profile.json` under a new path so future runs don't re-ask.
