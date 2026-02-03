package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/MrJJimenez/jobcli/internal/models"
)

type Format string

const (
	FormatTable    Format = "table"
	FormatCSV      Format = "csv"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "md"
	FormatTSV      Format = "tsv"
)

func WriteJobs(w io.Writer, jobs []models.Job, format Format) error {
	switch format {
	case FormatJSON:
		return writeJSON(w, jobs)
	case FormatCSV:
		return writeCSV(w, jobs, ',')
	case FormatTSV:
		return writeCSV(w, jobs, '\t')
	case FormatMarkdown:
		return writeMarkdown(w, jobs)
	default:
		return writeTable(w, jobs)
	}
}

func writeJSON(w io.Writer, jobs []models.Job) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jobs)
}

func writeCSV(w io.Writer, jobs []models.Job, delim rune) error {
	writer := csv.NewWriter(w)
	writer.Comma = delim
	if err := writer.Write(csvHeader()); err != nil {
		return err
	}
	for _, job := range jobs {
		if err := writer.Write(csvRow(job)); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func writeTable(w io.Writer, jobs []models.Job) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(csvHeader(), "\t"))
	for _, job := range jobs {
		fmt.Fprintln(tw, strings.Join(csvRow(job), "\t"))
	}
	return tw.Flush()
}

func writeMarkdown(w io.Writer, jobs []models.Job) error {
	if len(jobs) == 0 {
		_, err := fmt.Fprintln(w, "No results.")
		return err
	}
	for _, job := range jobs {
		lines := []string{
			fmt.Sprintf("- **%s** (%s)", safe(job.Title), safe(job.Company)),
			fmt.Sprintf("  Location: %s", safe(job.Location)),
			fmt.Sprintf("  Site: %s", safe(job.Site)),
			fmt.Sprintf("  URL: %s", safe(job.URL)),
		}
		if job.Remote {
			lines = append(lines, "  Remote: yes")
		}
		if job.JobType != "" {
			lines = append(lines, fmt.Sprintf("  Type: %s", safe(job.JobType)))
		}
		if job.Salary != "" {
			lines = append(lines, fmt.Sprintf("  Salary: %s", safe(job.Salary)))
		}
		if !job.PostedAt.IsZero() {
			lines = append(lines, fmt.Sprintf("  Posted: %s", job.PostedAt.Format(time.RFC3339)))
		}
		if job.PostedAtRaw != "" {
			lines = append(lines, fmt.Sprintf("  Posted (raw): %s", safe(job.PostedAtRaw)))
		}
		if job.Snippet != "" {
			lines = append(lines, fmt.Sprintf("  Summary: %s", safe(job.Snippet)))
		}
		for _, line := range lines {
			if _, err := fmt.Fprintln(w, line); err != nil {
				return err
			}
		}
	}
	return nil
}

func csvHeader() []string {
	return []string{
		"site",
		"title",
		"company",
		"location",
		"url",
		"remote",
		"job_type",
		"salary",
		"snippet",
		"posted_at",
		"posted_at_raw",
	}
}

func csvRow(job models.Job) []string {
	posted := ""
	if !job.PostedAt.IsZero() {
		posted = job.PostedAt.Format(time.RFC3339)
	}
	return []string{
		job.Site,
		job.Title,
		job.Company,
		job.Location,
		job.URL,
		boolString(job.Remote),
		job.JobType,
		job.Salary,
		job.Snippet,
		posted,
		job.PostedAtRaw,
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func safe(value string) string {
	return strings.TrimSpace(value)
}
