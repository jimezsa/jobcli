package scraper

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jimezsa/jobcli/internal/models"
	"github.com/jimezsa/jobcli/internal/network"
)

const linkedInPageSize = 10

var linkedInJobIDPattern = regexp.MustCompile(`\d{6,}`)

// LinkedIn uses a guest endpoint that returns HTML job cards.
type LinkedIn struct {
	client *network.Client
}

func NewLinkedIn(client *network.Client) *LinkedIn {
	return &LinkedIn{client: client}
}

func (l *LinkedIn) Name() string {
	return SiteLinkedIn
}

func (l *LinkedIn) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	var jobs []models.Job
	limit := params.Limit

	start := params.Offset
	for {
		if limit > 0 && len(jobs) >= limit {
			break
		}

		searchURL := buildLinkedInURL(params, start)
		doc, err := fetchDocument(ctx, l.client, searchURL, nil)
		if err != nil {
			return nil, fmt.Errorf("linkedin: %w", err)
		}

		pageJobs := parseLinkedInJobs(doc)
		if len(pageJobs) == 0 {
			break
		}

		for _, job := range pageJobs {
			if params.Remote && !job.Remote {
				continue
			}
			if job.Description == "" {
				job.Description = l.fetchLinkedInDescription(ctx, job.URL)
			}
			if job.Description == "" {
				job.Description = job.Snippet
			}
			jobs = append(jobs, job)
			if limit > 0 && len(jobs) >= limit {
				break
			}
		}

		start += linkedInPageSize
	}

	return jobs, nil
}

func buildLinkedInURL(params models.SearchParams, start int) string {
	values := url.Values{}
	values.Set("keywords", params.Query)
	if params.Location != "" {
		values.Set("location", params.Location)
	}
	values.Set("start", fmt.Sprintf("%d", start))
	if params.Hours > 0 {
		values.Set("f_TPR", fmt.Sprintf("r%d", params.Hours*3600))
	}
	return fmt.Sprintf("https://www.linkedin.com/jobs-guest/jobs/api/seeMoreJobPostings/search?%s", values.Encode())
}

func parseLinkedInJobs(doc *goquery.Document) []models.Job {
	var jobs []models.Job
	seen := map[string]struct{}{}

	doc.Find("li").Each(func(_ int, s *goquery.Selection) {
		link := attrFirst(s, "a.base-card__full-link", "href")
		link = strings.TrimSpace(link)
		if link == "" {
			return
		}

		title := cleanText(s.Find("h3.base-search-card__title").First().Text())
		company := cleanText(s.Find("h4.base-search-card__subtitle").First().Text())
		location := cleanText(s.Find("span.job-search-card__location").First().Text())
		snippet := cleanText(firstText(
			s,
			"div.job-search-card__snippet",
			"p.job-search-card__snippet",
		))
		postedRaw := cleanText(s.Find("time").First().AttrOr("datetime", ""))
		if postedRaw == "" {
			postedRaw = cleanText(s.Find("time").First().Text())
		}

		job := models.Job{
			Site:        SiteLinkedIn,
			Title:       title,
			Company:     company,
			Location:    location,
			URL:         link,
			Snippet:     snippet,
			PostedAtRaw: postedRaw,
			Remote:      isRemote(location, snippet),
		}
		if job.PostedAtRaw != "" {
			if ts, err := parsePostedAt(job.PostedAtRaw); err == nil {
				job.PostedAt = ts
			}
		}

		key := job.URL
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		jobs = append(jobs, job)
	})

	return jobs
}

func attrFirst(s *goquery.Selection, selector string, attr string) string {
	if s == nil {
		return ""
	}
	return s.Find(selector).First().AttrOr(attr, "")
}

func (l *LinkedIn) fetchLinkedInDescription(ctx context.Context, rawJobURL string) string {
	detailURL := linkedInDetailURL(rawJobURL)
	if detailURL == "" {
		return ""
	}

	doc, err := fetchDocument(ctx, l.client, detailURL, nil)
	if err != nil {
		return ""
	}
	return parseLinkedInDescription(doc)
}

func linkedInDetailURL(rawJobURL string) string {
	id := linkedInJobID(rawJobURL)
	if id == "" {
		return ""
	}
	return fmt.Sprintf("https://www.linkedin.com/jobs-guest/jobs/api/jobPosting/%s", id)
}

func linkedInJobID(rawJobURL string) string {
	rawJobURL = strings.TrimSpace(rawJobURL)
	if rawJobURL == "" {
		return ""
	}

	parsed, err := url.Parse(rawJobURL)
	if err != nil {
		return ""
	}

	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return ""
	}

	segments := strings.Split(path, "/")
	for i := 0; i < len(segments); i++ {
		if segments[i] != "view" || i+1 >= len(segments) {
			continue
		}
		if id := lastLinkedInIDMatch(segments[i+1]); id != "" {
			return id
		}
	}

	for _, segment := range segments {
		if id := lastLinkedInIDMatch(segment); id != "" {
			return id
		}
	}

	return ""
}

func lastLinkedInIDMatch(value string) string {
	matches := linkedInJobIDPattern.FindAllString(value, -1)
	if len(matches) == 0 {
		return ""
	}
	return matches[len(matches)-1]
}

func parseLinkedInDescription(doc *goquery.Document) string {
	if doc == nil {
		return ""
	}

	return firstText(
		doc.Selection,
		"div.show-more-less-html__markup",
		"section.show-more-less-html",
		"div.description__text",
		"div.jobs-description__content",
		"div.jobs-description-content__text",
	)
}
