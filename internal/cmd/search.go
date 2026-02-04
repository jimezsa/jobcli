package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MrJJimenez/jobcli/internal/config"
	"github.com/MrJJimenez/jobcli/internal/export"
	"github.com/MrJJimenez/jobcli/internal/models"
	"github.com/MrJJimenez/jobcli/internal/network"
	"github.com/MrJJimenez/jobcli/internal/scraper"
	"github.com/muesli/termenv"
)

type SearchCmd struct {
	Query string `arg:"" required:"" help:"Search query."`
	Sites string `help:"Comma-separated list of sites (default: all)." default:"all"`
	SearchOptions
}

type SiteCmd struct {
	Query string `arg:"" required:"" help:"Search query."`
	SearchOptions
	Site string `kong:"-"`
}

type SearchOptions struct {
	Location string `help:"Job location." env:"JOBCLI_DEFAULT_LOCATION"`
	Country  string `help:"Country code (Indeed/Glassdoor)." env:"JOBCLI_DEFAULT_COUNTRY"`
	Limit    int    `help:"Maximum results." env:"JOBCLI_DEFAULT_LIMIT"`
	Offset   int    `help:"Offset for pagination."`
	Remote   bool   `help:"Remote-only roles."`
	JobType  string `help:"Job type filter (fulltime, parttime, contract, internship)." enum:",fulltime,parttime,contract,internship" default:""`
	Hours    int    `help:"Jobs posted in the last N hours."`
	Format   string `help:"Output format: csv, json, md." enum:",csv,json,md" default:""`
	Links    string `help:"Table link display: short or full." enum:"short,full" default:"full"`
	Output   string `name:"output" short:"o" help:"Write output to a file."`
	Out      string `name:"out" help:"Alias for --output."`
	File     string `name:"file" help:"Alias for --output."`
	Proxies  string `help:"Comma-separated proxy URLs." env:"JOBCLI_PROXIES"`
}

func (s *SearchCmd) Run(ctx *Context) error {
	return runSearch(ctx, s.Query, s.Sites, s.SearchOptions)
}

func (s *SiteCmd) Run(ctx *Context) error {
	return runSearch(ctx, s.Query, s.Site, s.SearchOptions)
}

func runSearch(ctx *Context, query string, sitesArg string, opts SearchOptions) error {
	cfg := ctx.Config
	params := models.SearchParams{
		Query:    query,
		Location: firstNonEmpty(opts.Location, cfg.DefaultLocation),
		Country:  firstNonEmpty(opts.Country, cfg.DefaultCountry),
		Limit:    defaultInt(opts.Limit, cfg.DefaultLimit),
		Offset:   opts.Offset,
		Remote:   opts.Remote,
		JobType:  opts.JobType,
		Hours:    opts.Hours,
	}

	proxies, err := config.LoadProxies(opts.Proxies)
	if err != nil {
		return err
	}

	var rotator *network.Rotator
	if len(proxies) > 0 {
		rotator, err = network.NewRotator(proxies, 10*time.Minute)
		if err != nil {
			return err
		}
	}

	registry, err := scraper.Registry(rotator)
	if err != nil {
		return err
	}
	selected, err := selectScrapers(registry, sitesArg)
	if err != nil {
		return err
	}

	stopIndicator := startSearchIndicator(ctx)
	if stopIndicator != nil {
		defer stopIndicator()
	}

	jobs, err := runScrapers(ctx, selected, params)
	if err != nil {
		return err
	}

	if params.Limit > 0 && len(jobs) > params.Limit {
		jobs = jobs[:params.Limit]
	}

	outputPath := resolveOutputPath(opts)
	format, err := resolveFormat(ctx, opts, outputPath)
	if err != nil {
		return err
	}

	writer := ctx.Out
	var file *os.File
	if outputPath != "" {
		file, err = os.Create(outputPath)
		if err != nil {
			return err
		}
		defer file.Close()
		writer = file
	}

	colorEnabled := ctx.UI != nil && ctx.UI.ColorEnabled
	hyperlinks := colorEnabled && isTTY(writer)
	linkStyle := export.LinkStyleShort
	if strings.EqualFold(opts.Links, string(export.LinkStyleFull)) {
		linkStyle = export.LinkStyleFull
	}
	return export.WriteJobs(writer, jobs, format, export.WriteOptions{
		ColorEnabled: colorEnabled,
		Hyperlinks:   hyperlinks,
		LinkStyle:    linkStyle,
	})
}

func runScrapers(ctx *Context, scrapers []scraper.Scraper, params models.SearchParams) ([]models.Job, error) {
	var (
		wg      sync.WaitGroup
		results = make(chan scraperResult, len(scrapers))
	)

	for _, sc := range scrapers {
		wg.Add(1)
		go func(sc scraper.Scraper) {
			defer wg.Done()
			jobs, err := sc.Search(context.Background(), params)
			results <- scraperResult{site: sc.Name(), jobs: jobs, err: err}
		}(sc)
	}

	wg.Wait()
	close(results)

	var all []models.Job
	for res := range results {
		if res.err != nil {
			if errors.Is(res.err, scraper.ErrNotImplemented) {
				if ctx.Verbose {
					ctx.UI.Warnf("%s scraper: %v", res.site, res.err)
				}
				continue
			}
			ctx.UI.Warnf("%s scraper error: %v", res.site, res.err)
			continue
		}
		all = append(all, res.jobs...)
	}

	sort.SliceStable(all, func(i, j int) bool {
		return strings.ToLower(all[i].Site) < strings.ToLower(all[j].Site)
	})

	return all, nil
}

type scraperResult struct {
	site string
	jobs []models.Job
	err  error
}

func resolveOutputPath(opts SearchOptions) string {
	if opts.Output != "" {
		return opts.Output
	}
	if opts.Out != "" {
		return opts.Out
	}
	return opts.File
}

func resolveFormat(ctx *Context, opts SearchOptions, outputPath string) (export.Format, error) {
	if outputPath != "" {
		if opts.Format == "" {
			return export.FormatCSV, nil
		}
		return parseFormat(opts.Format)
	}

	if ctx.JSONOutput {
		return export.FormatJSON, nil
	}
	if ctx.PlainText {
		return export.FormatTSV, nil
	}
	if opts.Format != "" {
		return parseFormat(opts.Format)
	}
	if isTTY(ctx.Out) {
		return export.FormatTable, nil
	}
	return export.FormatCSV, nil
}

func parseFormat(value string) (export.Format, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "csv":
		return export.FormatCSV, nil
	case "json":
		return export.FormatJSON, nil
	case "md", "markdown":
		return export.FormatMarkdown, nil
	case "tsv":
		return export.FormatTSV, nil
	case "table", "":
		return export.FormatTable, nil
	default:
		return "", fmt.Errorf("unknown format: %s", value)
	}
}

func selectScrapers(registry map[string]scraper.Scraper, sitesArg string) ([]scraper.Scraper, error) {
	requested := scraper.NormalizeSites(strings.Split(sitesArg, ","))
	if len(requested) == 0 || (len(requested) == 1 && requested[0] == "all") {
		requested = make([]string, 0, len(registry))
		for site := range registry {
			requested = append(requested, site)
		}
	}

	requested = expandAliases(requested)

	selected := make([]scraper.Scraper, 0, len(requested))
	for _, site := range requested {
		sc, ok := registry[site]
		if !ok {
			return nil, fmt.Errorf("unknown site: %s", site)
		}
		selected = append(selected, sc)
	}

	return selected, nil
}

func expandAliases(sites []string) []string {
	out := make([]string, 0, len(sites))
	for _, site := range sites {
		switch site {
		case "googlejobs", "google-jobs", "googlejobs.com":
			out = append(out, scraper.SiteGoogleJobs)
		case "zip", "zip-recruiter":
			out = append(out, scraper.SiteZipRecruiter)
		default:
			out = append(out, site)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func isTTY(out io.Writer) bool {
	output := termenv.NewOutput(out)
	return output.ColorProfile() != termenv.Ascii
}

func startSearchIndicator(ctx *Context) func() {
	if ctx == nil || ctx.Err == nil || ctx.UI == nil {
		return nil
	}
	if !isTTY(ctx.Err) {
		return nil
	}

	done := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		start := time.Now()
		frames := []string{"|", "/", "-", "\\"}
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		index := 0

		for {
			select {
			case <-done:
				fmt.Fprint(ctx.Err, "\r\033[2K")
				return
			case <-ticker.C:
				seconds := int(time.Since(start).Seconds())
				frame := frames[index%len(frames)]
				fmt.Fprintf(ctx.Err, "\r\033[2KSearching... %ds %s", seconds, frame)
				index++
			}
		}
	}()

	return func() {
		close(done)
		<-stopped
	}
}
