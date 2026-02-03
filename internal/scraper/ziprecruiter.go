package scraper

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/MrJJimenez/jobcli/internal/network"
	"github.com/PuerkitoBio/goquery"
)

type ZipRecruiter struct {
	client *network.Client
}

func NewZipRecruiter(client *network.Client) *ZipRecruiter {
	return &ZipRecruiter{client: client}
}

func (z *ZipRecruiter) Name() string {
	return SiteZipRecruiter
}

func (z *ZipRecruiter) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	searchURL := buildZipRecruiterURL(params)
	doc, err := fetchDocument(ctx, z.client, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ziprecruiter: %w", err)
	}

	jobs := parseJSONLDJobs(doc, SiteZipRecruiter)
	jobs = append(jobs, parseZipRecruiterJobs(doc)...)
	jobs = dedupeJobs(jobs)

	if params.Remote {
		jobs = filterRemote(jobs)
	}
	if params.Limit > 0 && len(jobs) > params.Limit {
		jobs = jobs[:params.Limit]
	}
	return jobs, nil
}

func buildZipRecruiterURL(params models.SearchParams) string {
	values := url.Values{}
	values.Set("search", params.Query)
	if params.Location != "" {
		values.Set("location", params.Location)
	}
	if params.Offset > 0 {
		page := params.Offset/25 + 1
		values.Set("page", fmt.Sprintf("%d", page))
	}
	return fmt.Sprintf("https://www.ziprecruiter.com/jobs-search?%s", values.Encode())
}

func parseZipRecruiterJobs(doc *goquery.Document) []models.Job {
	var jobs []models.Job

	selectors := []string{
		"article.job_result",
		"article.job-result",
		"div.job_result",
		"div.job-result",
		"li.job_result",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			job := zipRecruiterJobFromCard(s)
			if job.Title == "" || job.URL == "" {
				return
			}
			jobs = append(jobs, job)
		})
		if len(jobs) > 0 {
			break
		}
	}

	return jobs
}

func zipRecruiterJobFromCard(s *goquery.Selection) models.Job {
	title := cleanText(firstText(s, "a.job_link", "a.job_link", "a.job_title", "a.jobLink", "a.t_job_link"))
	if title == "" {
		title = cleanText(firstText(s, "h2", "h3"))
	}

	company := cleanText(firstText(s, "a.t_org_link", "a.company_name", "span.company_name", "span.name"))
	location := cleanText(firstText(s, "div.location", "span.location", "span.job_location", "div.job_location"))
	snippet := cleanText(firstText(s, "div.job_snippet", "p.job_snippet", "div.snippet", "p"))

	link := firstAttr(s, "a.job_link", "href")
	if link == "" {
		link = firstAttr(s, "a.job_title", "href")
	}
	link = absoluteURL("https://www.ziprecruiter.com", link)

	return models.Job{
		Site:     SiteZipRecruiter,
		Title:    title,
		Company:  company,
		Location: location,
		URL:      link,
		Snippet:  snippet,
		Remote:   isRemote(location, snippet),
	}
}

func firstText(s *goquery.Selection, selectors ...string) string {
	for _, selector := range selectors {
		text := cleanText(s.Find(selector).First().Text())
		if text != "" {
			return text
		}
	}
	return ""
}

func firstAttr(s *goquery.Selection, selector string, attr string) string {
	return strings.TrimSpace(s.Find(selector).First().AttrOr(attr, ""))
}
