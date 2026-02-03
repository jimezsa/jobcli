package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/muesli/termenv"
)

type Format string

const (
	FormatTable    Format = "table"
	FormatCSV      Format = "csv"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "md"
	FormatTSV      Format = "tsv"
)

type WriteOptions struct {
	ColorEnabled bool
	Hyperlinks   bool
	LinkStyle    LinkStyle
}

type LinkStyle string

const (
	LinkStyleShort LinkStyle = "short"
	LinkStyleFull  LinkStyle = "full"
)

func WriteJobs(w io.Writer, jobs []models.Job, format Format, opts WriteOptions) error {
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
		return writeTable(w, jobs, opts)
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

func writeTable(w io.Writer, jobs []models.Job, opts WriteOptions) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(tableHeader(), "\t"))
	output := termenv.NewOutput(w)
	for _, job := range jobs {
		fmt.Fprintln(tw, strings.Join(tableRow(job, output, opts), "\t"))
	}
	return tw.Flush()
}

func writeMarkdown(w io.Writer, jobs []models.Job) error {
	if len(jobs) == 0 {
		_, err := fmt.Fprintln(w, "No results.")
		return err
	}
	for _, job := range jobs {
		urlLine := "  URL: -"
		if url := safe(job.URL); url != "" {
			urlLine = fmt.Sprintf("  URL: [Open listing](<%s>)", url)
		}
		lines := []string{
			fmt.Sprintf("- **%s** (%s)", safe(job.Title), safe(job.Company)),
			fmt.Sprintf("  Location: %s", safe(job.Location)),
			fmt.Sprintf("  Site: %s", safe(job.Site)),
			urlLine,
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

func tableHeader() []string {
	return []string{
		"site",
		"title",
		"company",
		"url",
	}
}

func tableRow(job models.Job, output *termenv.Output, opts WriteOptions) []string {
	const linkColor = "#87CEEB"

	url := safe(job.URL)
	displayURL := "-"
	if url != "" {
		displayURL = url
		if opts.LinkStyle == LinkStyleShort && opts.Hyperlinks {
			displayURL = shortURLLabel(url)
		}
		if opts.ColorEnabled {
			displayURL = output.String(displayURL).Foreground(output.Color(linkColor)).String()
		}
		if opts.Hyperlinks {
			displayURL = hyperlink(url, displayURL)
		}
	}
	return []string{
		safe(job.Site),
		safe(job.Title),
		safe(job.Company),
		displayURL,
	}
}

func hyperlink(url string, text string) string {
	const esc = "\x1b"
	return esc + "]8;;" + url + esc + "\\" + text + esc + "]8;;" + esc + "\\"
}

func shortURLLabel(raw string) string {
	const maxLen = 60
	label := strings.TrimSpace(raw)
	if parsed, err := url.Parse(raw); err == nil {
		host := strings.TrimPrefix(parsed.Host, "www.")
		if host != "" {
			label = host + parsed.Path
		}
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = raw
	}
	if len(label) > maxLen {
		label = label[:maxLen-3] + "..."
	}
	return label
}
