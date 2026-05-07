#!/usr/bin/env python3
"""Check whether a tailored CV PDF is extractable and aligned with a job description."""

from __future__ import annotations

import argparse
import json
import re
import shutil
import subprocess
from collections import Counter
from pathlib import Path


STOPWORDS = {
    "about", "above", "across", "after", "again", "against", "also", "and", "any", "are",
    "because", "been", "being", "best", "both", "can", "company", "could", "day", "each",
    "excellent", "fast", "for", "from", "has", "have", "having", "into", "its", "job",
    "more", "must", "our", "over", "paced", "preferred", "required", "requirements",
    "responsibilities", "role", "should", "such", "team", "that", "the", "their", "this",
    "through", "using", "was", "will", "with", "work", "working", "you", "your",
}

TECH_TERMS = [
    "go", "golang", "python", "java", "javascript", "typescript", "node.js", "react",
    "angular", "vue", "next.js", "sql", "postgresql", "mysql", "mongodb", "redis",
    "elasticsearch", "aws", "azure", "gcp", "docker", "kubernetes", "terraform",
    "ansible", "linux", "git", "github", "gitlab", "ci/cd", "jenkins", "github actions",
    "rest", "graphql", "grpc", "api", "microservices", "distributed systems",
    "machine learning", "ai", "data engineering", "spark", "kafka", "airflow",
    "security", "oauth", "saml", "observability", "prometheus", "grafana",
    "agile", "scrum", "kanban", "leadership", "mentoring", "stakeholder management",
]

SECTION_NAMES = [
    "summary", "profile", "experience", "work experience", "professional experience",
    "skills", "technical skills", "projects", "education", "certifications", "languages",
]


def normalize(text: str) -> str:
    return re.sub(r"\s+", " ", text.lower()).strip()


def extract_pdf_text(pdf_path: Path) -> str:
    if shutil.which("pdftotext") is None:
        raise SystemExit("pdftotext is required but was not found on PATH")
    result = subprocess.run(
        ["pdftotext", "-layout", str(pdf_path), "-"],
        text=True,
        capture_output=True,
        check=False,
    )
    if result.returncode != 0:
        raise SystemExit(result.stderr.strip() or "pdftotext failed")
    return result.stdout


def token_counts(text: str) -> Counter[str]:
    tokens = re.findall(r"[a-zA-Z][a-zA-Z0-9+#.]{1,}", text.lower())
    return Counter(token for token in tokens if token not in STOPWORDS and len(token) > 2)


def contains_term(text: str, term: str) -> bool:
    pattern = r"(?<![a-zA-Z0-9+#.])" + re.escape(term.lower()) + r"(?![a-zA-Z0-9+#.])"
    return re.search(pattern, text.lower()) is not None


def job_terms(job_text: str, limit: int) -> list[str]:
    normalized_job = normalize(job_text)
    terms: list[str] = []
    for term in TECH_TERMS:
        if contains_term(normalized_job, term):
            terms.append(term)

    for token, _ in token_counts(job_text).most_common(limit * 2):
        if token not in terms:
            terms.append(token)
        if len(terms) >= limit:
            break
    return terms[:limit]


def detected_sections(cv_text: str) -> list[str]:
    lines = [normalize(line) for line in cv_text.splitlines()]
    found = []
    for section in SECTION_NAMES:
        if any(line == section or line.startswith(section + " ") for line in lines):
            found.append(section)
    return found


def contact_checks(cv_text: str) -> dict[str, bool]:
    return {
        "email": re.search(r"\b[\w.+-]+@[\w.-]+\.[a-zA-Z]{2,}\b", cv_text) is not None,
        "phone": re.search(r"(\+\d{1,3}[\s.-]?)?(?:\(?\d{2,4}\)?[\s.-]?){2,}\d{2,4}", cv_text) is not None,
        "url": re.search(r"\b(?:https?://|linkedin\.com|github\.com|www\.)\S+", cv_text, re.I) is not None,
    }


def build_report(pdf_path: Path, job_text: str, max_terms: int) -> dict[str, object]:
    cv_text = extract_pdf_text(pdf_path)
    normalized_cv = normalize(cv_text)
    terms = job_terms(job_text, max_terms) if job_text.strip() else []
    present = [term for term in terms if contains_term(normalized_cv, term)]
    missing = [term for term in terms if term not in present]
    words = re.findall(r"\b\w+\b", cv_text)
    contacts = contact_checks(cv_text)
    sections = detected_sections(cv_text)

    warnings = []
    if len(words) < 150:
        warnings.append("Extracted text is very short; PDF may be image-only or poorly parsed.")
    if not sections:
        warnings.append("No standard resume sections detected in extracted text.")
    if not contacts["email"]:
        warnings.append("No extractable email detected.")
    if terms and len(present) / len(terms) < 0.45:
        warnings.append("Low keyword coverage against the job description.")

    return {
        "pdf": str(pdf_path),
        "extractable_text": bool(cv_text.strip()),
        "word_count": len(words),
        "sections_detected": sections,
        "contact_detected": contacts,
        "keyword_coverage": round((len(present) / len(terms)) if terms else 0, 3),
        "terms_checked": terms,
        "terms_present": present,
        "terms_missing": missing,
        "warnings": warnings,
    }


def to_markdown(report: dict[str, object]) -> str:
    lines = [
        "# ATS Text Check",
        "",
        f"- PDF: `{report['pdf']}`",
        f"- Extractable text: `{report['extractable_text']}`",
        f"- Word count: `{report['word_count']}`",
        f"- Keyword coverage: `{report['keyword_coverage']}`",
        "",
        "## Sections Detected",
        "",
    ]
    sections = report["sections_detected"]
    lines.extend(f"- {section}" for section in sections) if sections else lines.append("- None")
    lines.extend(["", "## Contact Detected", ""])
    contacts = report["contact_detected"]
    for key, value in contacts.items():
        lines.append(f"- {key}: `{value}`")
    lines.extend(["", "## Present Terms", ""])
    present = report["terms_present"]
    lines.extend(f"- {term}" for term in present) if present else lines.append("- None")
    lines.extend(["", "## Missing Terms", ""])
    missing = report["terms_missing"]
    lines.extend(f"- {term}" for term in missing) if missing else lines.append("- None")
    lines.extend(["", "## Warnings", ""])
    warnings = report["warnings"]
    lines.extend(f"- {warning}" for warning in warnings) if warnings else lines.append("- None")
    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pdf", type=Path, required=True, help="Generated tailored CV PDF.")
    parser.add_argument("--job-text", type=Path, help="Plain-text job description file.")
    parser.add_argument("--out", type=Path, help="Report output path.")
    parser.add_argument("--format", choices=["json", "md"], default="json")
    parser.add_argument("--max-terms", type=int, default=60)
    args = parser.parse_args()

    pdf_path = args.pdf.expanduser().resolve()
    if not pdf_path.exists():
        raise SystemExit(f"PDF does not exist: {pdf_path}")
    job_text = ""
    if args.job_text:
        job_text = args.job_text.expanduser().read_text(encoding="utf-8", errors="replace")

    report = build_report(pdf_path, job_text, args.max_terms)
    output = to_markdown(report) if args.format == "md" else json.dumps(report, indent=2) + "\n"
    if args.out:
        out_path = args.out.expanduser()
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(output, encoding="utf-8")
    print(output, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
