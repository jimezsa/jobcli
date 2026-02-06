package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/jimezsa/jobcli/internal/models"
	"github.com/jimezsa/jobcli/internal/network"
)

func fetchDocument(ctx context.Context, client *network.Client, target string, headers map[string]string) (*goquery.Document, error) {
	req, err := fhttp.NewRequestWithContext(ctx, fhttp.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}

	applyHeaders(req, headers)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func applyHeaders(req *fhttp.Request, headers map[string]string) {
	if headers == nil {
		headers = map[string]string{}
	}
	if _, ok := headers["accept"]; !ok {
		headers["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	}
	if _, ok := headers["accept-language"]; !ok {
		headers["accept-language"] = "en-US,en;q=0.9"
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func cleanText(value string) string {
	value = html.UnescapeString(value)
	return strings.Join(strings.Fields(value), " ")
}

func absoluteURL(base string, href string) string {
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return href
	}
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return baseURL.ResolveReference(ref).String()
}

func parsePostedAt(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("empty")
	}
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02T15:04:05-0700",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, value); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %s", value)
}

func parseJSONLDJobs(doc *goquery.Document, site string) []models.Job {
	var jobs []models.Job
	seen := map[string]struct{}{}

	doc.Find("script[type='application/ld+json']").Each(func(_ int, s *goquery.Selection) {
		raw := strings.TrimSpace(s.Text())
		if raw == "" {
			return
		}

		data, err := decodeJSONLD(raw)
		if err != nil {
			return
		}

		for _, job := range extractJobsFromJSONLD(data, site) {
			key := job.URL
			if key == "" {
				key = strings.ToLower(job.Title + "|" + job.Company + "|" + job.Location)
			}
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			jobs = append(jobs, job)
		}
	})

	return jobs
}

func decodeJSONLD(raw string) (any, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "<!--")
	raw = strings.TrimSuffix(raw, "-->")
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, "\u2028", "")
	raw = strings.ReplaceAll(raw, "\u2029", "")

	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func extractJobsFromJSONLD(data any, site string) []models.Job {
	var jobs []models.Job

	switch value := data.(type) {
	case []any:
		for _, item := range value {
			jobs = append(jobs, extractJobsFromJSONLD(item, site)...)
		}
	case map[string]any:
		if typ := strings.ToLower(stringValue(value["@type"], value["type"])); typ != "" {
			switch typ {
			case "jobposting":
				jobs = append(jobs, jobFromJobPosting(value, site))
				return jobs
			case "itemlist":
				jobs = append(jobs, jobsFromItemList(value, site)...)
			}
		}
		if graph, ok := value["@graph"]; ok {
			jobs = append(jobs, extractJobsFromJSONLD(graph, site)...)
		}
		if main, ok := value["mainEntity"]; ok {
			jobs = append(jobs, extractJobsFromJSONLD(main, site)...)
		}
	}

	return jobs
}

func jobsFromItemList(value map[string]any, site string) []models.Job {
	items, ok := value["itemListElement"]
	if !ok {
		return nil
	}

	var jobs []models.Job
	switch list := items.(type) {
	case []any:
		for _, item := range list {
			jobs = append(jobs, extractJobsFromJSONLD(item, site)...)
		}
	case map[string]any:
		jobs = append(jobs, extractJobsFromJSONLD(list, site)...)
	}
	return jobs
}

func jobFromJobPosting(value map[string]any, site string) models.Job {
	job := models.Job{Site: site}
	job.Title = stringValue(value["title"], value["name"])
	job.Company = stringValue(mapValue(value["hiringOrganization"], "name"))
	job.URL = stringValue(value["url"], value["@id"])
	job.JobType = stringValue(value["employmentType"])
	job.Salary = salaryFromJSONLD(value["baseSalary"])
	job.PostedAtRaw = stringValue(value["datePosted"])
	if job.PostedAtRaw != "" {
		if ts, err := parsePostedAt(job.PostedAtRaw); err == nil {
			job.PostedAt = ts
		}
	}
	job.Location = locationFromJSONLD(value["jobLocation"])
	job.Snippet = truncate(cleanText(stringValue(value["description"])), 240)
	job.Remote = strings.Contains(strings.ToLower(job.Location), "remote")
	return job
}

func salaryFromJSONLD(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case map[string]any:
		if amount := mapValue(v["value"], "value"); amount != nil {
			return stringValue(amount)
		}
		if amount := mapValue(v["value"], "minValue"); amount != nil {
			max := mapValue(v["value"], "maxValue")
			currency := stringValue(v["currency"])
			minStr := stringValue(amount)
			maxStr := stringValue(max)
			if maxStr != "" {
				return strings.TrimSpace(minStr + " - " + maxStr + " " + currency)
			}
			return strings.TrimSpace(minStr + " " + currency)
		}
	case string:
		return v
	}
	return ""
}

func locationFromJSONLD(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case []any:
		var parts []string
		for _, item := range v {
			loc := locationFromJSONLD(item)
			if loc != "" {
				parts = append(parts, loc)
			}
		}
		return strings.Join(parts, "; ")
	case map[string]any:
		address := v["address"]
		if addressMap, ok := address.(map[string]any); ok {
			return joinAddress(addressMap)
		}
		return joinAddress(v)
	case string:
		return v
	}

	return ""
}

func joinAddress(value map[string]any) string {
	parts := []string{
		stringValue(value["streetAddress"]),
		stringValue(value["addressLocality"]),
		stringValue(value["addressRegion"]),
		stringValue(value["postalCode"]),
		stringValue(value["addressCountry"]),
	}
	var cleaned []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		cleaned = append(cleaned, part)
	}
	return strings.Join(cleaned, ", ")
}

func stringValue(values ...any) string {
	for _, value := range values {
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		case float64:
			return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
		case int:
			return fmt.Sprintf("%d", v)
		case int64:
			return fmt.Sprintf("%d", v)
		case json.Number:
			return v.String()
		case fmt.Stringer:
			if v.String() != "" {
				return strings.TrimSpace(v.String())
			}
		case map[string]any:
			if name := stringValue(v["name"]); name != "" {
				return name
			}
		}
	}
	return ""
}

func mapValue(value any, key string) any {
	if value == nil {
		return nil
	}
	m, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	return m[key]
}

func truncate(value string, max int) string {
	if max <= 0 {
		return value
	}
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return strings.TrimSpace(value[:max]) + "..."
}

func filterRemote(jobs []models.Job) []models.Job {
	filtered := jobs[:0]
	for _, job := range jobs {
		if job.Remote {
			filtered = append(filtered, job)
		}
	}
	return filtered
}

func dedupeJobs(jobs []models.Job) []models.Job {
	seen := map[string]struct{}{}
	out := make([]models.Job, 0, len(jobs))
	for _, job := range jobs {
		key := job.URL
		if key == "" {
			key = strings.ToLower(job.Title + "|" + job.Company + "|" + job.Location)
		}
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, job)
	}
	return out
}
