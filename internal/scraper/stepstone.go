package scraper

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/MrJJimenez/jobcli/internal/network"
	"github.com/PuerkitoBio/goquery"
)

const stepstonePageSize = 25

type Stepstone struct {
	client *network.Client
}

func NewStepstone(client *network.Client) *Stepstone {
	return &Stepstone{client: client}
}

func (s *Stepstone) Name() string {
	return SiteStepstone
}

func (s *Stepstone) Search(ctx context.Context, params models.SearchParams) ([]models.Job, error) {
	var jobs []models.Job
	limit := params.Limit

	page := stepstonePageFromOffset(params.Offset)
	skip := 0
	if params.Offset > 0 {
		skip = params.Offset % stepstonePageSize
	}

	seen := map[string]struct{}{}
	for {
		if limit > 0 && len(jobs) >= limit {
			break
		}

		searchURL := buildStepstoneURL(params, page)
		doc, err := fetchDocument(ctx, s.client, searchURL, map[string]string{
			"accept-language": "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7",
		})
		if err != nil {
			return nil, fmt.Errorf("stepstone: %w", err)
		}

		pageJobs := parseStepstoneJobs(doc)
		if len(pageJobs) == 0 {
			break
		}

		added := 0
		for _, job := range pageJobs {
			if skip > 0 {
				skip--
				continue
			}
			if params.Remote && !job.Remote {
				continue
			}
			if job.URL == "" {
				continue
			}
			if _, ok := seen[job.URL]; ok {
				continue
			}
			seen[job.URL] = struct{}{}
			jobs = append(jobs, job)
			added++
			if limit > 0 && len(jobs) >= limit {
				break
			}
		}

		if added == 0 {
			break
		}
		page++
	}

	return jobs, nil
}

func stepstonePageFromOffset(offset int) int {
	if offset <= 0 {
		return 1
	}
	return offset/stepstonePageSize + 1
}

func buildStepstoneURL(params models.SearchParams, page int) string {
	base := "https://www.stepstone.de/jobs"
	query := stepstoneSlug(params.Query)
	if query == "" {
		query = strings.ToLower(strings.TrimSpace(params.Query))
	}
	path := fmt.Sprintf("%s/%s", base, url.PathEscape(query))
	if params.Location != "" {
		location := stepstoneSlug(params.Location)
		if location != "" {
			path = fmt.Sprintf("%s/in-%s", path, url.PathEscape(location))
		}
	}
	if page > 1 {
		return fmt.Sprintf("%s?page=%d", path, page)
	}
	return path
}

func stepstoneSlug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func parseStepstoneJobs(doc *goquery.Document) []models.Job {
	jobs := parseJSONLDJobs(doc, SiteStepstone)
	jobs = append(jobs, parseStepstoneJobCards(doc)...)
	return dedupeJobs(jobs)
}

func parseStepstoneJobCards(doc *goquery.Document) []models.Job {
	var jobs []models.Job
	seen := map[string]struct{}{}

	doc.Find("a[href*='stellenangebote--']").Each(func(_ int, s *goquery.Selection) {
		href := strings.TrimSpace(s.AttrOr("href", ""))
		if href == "" {
			return
		}
		link := absoluteURL("https://www.stepstone.de", href)
		if link == "" {
			return
		}
		if _, ok := seen[link]; ok {
			return
		}

		title := cleanText(s.Text())
		if title == "" {
			return
		}

		card := stepstoneCardForAnchor(s)
		company, location, snippet, posted, remote := stepstoneParseCard(card, title)

		jobs = append(jobs, models.Job{
			Site:        SiteStepstone,
			Title:       title,
			Company:     company,
			Location:    location,
			URL:         link,
			Snippet:     snippet,
			PostedAtRaw: posted,
			Remote:      remote || isRemote(location, snippet),
		})
		seen[link] = struct{}{}
	})

	return jobs
}

func stepstoneCardForAnchor(s *goquery.Selection) *goquery.Selection {
	if s == nil {
		return nil
	}
	if card := s.Closest("article"); card.Length() > 0 {
		return card
	}
	if card := s.Closest("li"); card.Length() > 0 {
		return card
	}
	if card := s.Closest("section"); card.Length() > 0 {
		return card
	}
	if card := s.Closest("div"); card.Length() > 0 {
		return card
	}
	return s.Parent()
}

func stepstoneParseCard(card *goquery.Selection, title string) (string, string, string, string, bool) {
	if card == nil || card.Length() == 0 {
		return "", "", "", "", false
	}

	posted := stepstonePostedText(card)
	lines := stepstoneCardLines(card, title)

	remote := false
	for _, line := range lines {
		if stepstoneIsRemoteLine(line) {
			remote = true
			break
		}
	}

	candidates := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == title {
			continue
		}
		if stepstoneIsNoiseLine(line) || stepstoneIsRemoteLine(line) || stepstoneIsPostedLine(line) {
			continue
		}
		candidates = append(candidates, line)
	}

	var company, location, snippet string
	if len(candidates) > 0 {
		company = candidates[0]
	}
	if len(candidates) > 1 {
		location = candidates[1]
	}
	if len(candidates) > 2 {
		for _, line := range candidates[2:] {
			if len(line) >= 30 {
				snippet = line
				break
			}
		}
		if snippet == "" {
			snippet = candidates[2]
		}
	}

	if posted == "" {
		for _, line := range lines {
			if stepstoneIsPostedLine(line) {
				posted = line
				break
			}
		}
	}

	return company, location, snippet, posted, remote
}

func stepstoneCardLines(card *goquery.Selection, title string) []string {
	raw := card.Text()
	parts := strings.Split(raw, "\n")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		line := cleanText(part)
		if line == "" || line == title {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		out = append(out, line)
	}
	return out
}

func stepstonePostedText(card *goquery.Selection) string {
	if card == nil {
		return ""
	}
	if value := cleanText(card.Find("time").First().AttrOr("datetime", "")); value != "" {
		return value
	}
	if value := cleanText(card.Find("time").First().Text()); value != "" {
		return value
	}
	return ""
}

func stepstoneIsRemoteLine(line string) bool {
	value := strings.ToLower(line)
	return strings.Contains(value, "home-office") ||
		strings.Contains(value, "homeoffice") ||
		strings.Contains(value, "remote")
}

func stepstoneIsPostedLine(line string) bool {
	value := strings.ToLower(line)
	if strings.HasPrefix(value, "vor ") {
		return true
	}
	return value == "heute" || value == "gestern"
}

func stepstoneIsNoiseLine(line string) bool {
	value := strings.ToLower(line)
	switch value {
	case "gehalt", "gehalt anzeigen", "mehr", "neu", "top-job":
		return true
	}
	if strings.Contains(value, "gehalt anzeigen") {
		return true
	}
	if strings.Contains(value, "schnelle bewerbung") {
		return true
	}
	if strings.Contains(value, "anschreiben nicht erforderlich") {
		return true
	}
	return false
}
