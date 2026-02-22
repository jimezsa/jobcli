package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jimezsa/jobcli/internal/config"
	"github.com/jimezsa/jobcli/internal/export"
	"github.com/jimezsa/jobcli/internal/models"
	"github.com/jimezsa/jobcli/internal/network"
	"github.com/jimezsa/jobcli/internal/scraper"
	"github.com/jimezsa/jobcli/internal/seen"
	"github.com/jimezsa/jobcli/internal/ui"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog"
)

type SearchCmd struct {
	Query string `arg:"" optional:"" help:"Search query (comma-separated). Optional when --query-file is provided."`
	Sites string `help:"Comma-separated list of sites (default: all)." default:"all"`
	SearchOptions
}

type SiteCmd struct {
	Query string `arg:"" optional:"" help:"Search query (comma-separated). Optional when --query-file is provided."`
	SearchOptions
	Site string `kong:"-"`
}

type SearchOptions struct {
	Location   string `help:"Job location." env:"JOBCLI_DEFAULT_LOCATION"`
	Country    string `help:"Country code (Indeed/Glassdoor)." env:"JOBCLI_DEFAULT_COUNTRY"`
	Limit      int    `help:"Maximum results per query." env:"JOBCLI_DEFAULT_LIMIT"`
	Offset     int    `help:"Offset for pagination."`
	Remote     bool   `help:"Remote-only roles."`
	JobType    string `help:"Job type filter (fulltime, parttime, contract, internship)." enum:",fulltime,parttime,contract,internship" default:""`
	Hours      int    `help:"Jobs posted in the last N hours."`
	Format     string `help:"Output format: csv, json, md." enum:",csv,json,md" default:""`
	Links      string `help:"Table link display: short or full." enum:"short,full" default:"full"`
	Output     string `name:"output" short:"o" help:"Write output to a file."`
	Out        string `name:"out" help:"Alias for --output."`
	File       string `name:"file" help:"Alias for --output."`
	Proxies    string `help:"Comma-separated proxy URLs." env:"JOBCLI_PROXIES"`
	QueryFile  string `help:"Path to JSON file with queries or a full search profile."`
	Seen       string `help:"Path to seen jobs JSON file."`
	NewOnly    bool   `help:"Output only unseen jobs (requires --seen)."`
	NewOut     string `help:"Write unseen jobs JSON to a file (requires --seen)."`
	SeenUpdate bool   `help:"Update --seen history file by merging in newly discovered unseen jobs after search completes (requires --seen)."`
}

const maxQueries = 10

func (s *SearchCmd) Run(ctx *Context) error {
	return runSearch(ctx, s.Query, s.Sites, s.SearchOptions, true)
}

func (s *SiteCmd) Run(ctx *Context) error {
	return runSearch(ctx, s.Query, s.Site, s.SearchOptions, false)
}

func runSearch(ctx *Context, query string, sitesArg string, opts SearchOptions, allowSitesOverride bool) error {
	queryConfig, err := loadQueryFileConfigIfProvided(opts.QueryFile)
	if err != nil {
		return err
	}

	providedFlags := collectProvidedFlags(os.Args[1:])
	opts, sitesArg, err = applyQueryFileDefaults(ctx, opts, sitesArg, queryConfig, allowSitesOverride, providedFlags)
	if err != nil {
		return err
	}

	if opts.NewOnly && strings.TrimSpace(opts.Seen) == "" {
		return fmt.Errorf("--new-only requires --seen")
	}
	if strings.TrimSpace(opts.NewOut) != "" && strings.TrimSpace(opts.Seen) == "" {
		return fmt.Errorf("--new-out requires --seen")
	}
	if opts.SeenUpdate && strings.TrimSpace(opts.Seen) == "" {
		return fmt.Errorf("--seen-update requires --seen")
	}

	queries, err := mergeAndNormalizeQueries(splitQueries(query), queryConfig.Queries)
	if err != nil {
		return err
	}

	cfg := ctx.Config
	baseParams := models.SearchParams{
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

	var (
		jobs     []models.Job
		failures []scraperFailure
	)
	for _, currentQuery := range queries {
		queryJobs, queryFailures, runErr := runScrapersForQuery(selected, baseParams, currentQuery)
		if runErr != nil {
			return runErr
		}
		queryJobs = limitJobs(queryJobs, baseParams.Limit)
		jobs = mergeUniqueJobs(jobs, queryJobs)
		failures = append(failures, queryFailures...)
	}

	sortJobsBySite(jobs)
	sortScraperFailures(failures)

	reportScraperFailures(ctx, failures)

	var unseenJobs []models.Job
	if strings.TrimSpace(opts.Seen) != "" {
		seenJobs, err := seen.ReadJobsAllowMissing(opts.Seen)
		if err != nil {
			return fmt.Errorf("read --seen: %w", err)
		}
		unseenJobs, _ = seen.Diff(jobs, seenJobs)
	}

	outputJobs := jobs
	if opts.NewOnly {
		outputJobs = unseenJobs
	}

	outputPath := resolveOutputPath(opts)
	if strings.TrimSpace(opts.NewOut) != "" && pathsEqual(outputPath, opts.NewOut) {
		return fmt.Errorf("--new-out path must differ from --output")
	}
	if strings.TrimSpace(opts.Seen) != "" && pathsEqual(outputPath, opts.Seen) {
		return fmt.Errorf("--output path must differ from --seen")
	}
	if strings.TrimSpace(opts.NewOut) != "" && pathsEqual(opts.NewOut, opts.Seen) {
		return fmt.Errorf("--new-out path must differ from --seen")
	}

	if strings.TrimSpace(opts.NewOut) != "" {
		if err := seen.WriteJobs(opts.NewOut, unseenJobs); err != nil {
			return fmt.Errorf("write --new-out: %w", err)
		}
	}

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
	if err := export.WriteJobs(writer, outputJobs, format, export.WriteOptions{
		ColorEnabled: colorEnabled,
		Hyperlinks:   hyperlinks,
		LinkStyle:    linkStyle,
	}); err != nil {
		return err
	}

	if opts.SeenUpdate && strings.TrimSpace(opts.Seen) != "" {
		if err := updateSeenHistory(opts.Seen, unseenJobs); err != nil {
			return err
		}
	}

	summaryJobs := jobs
	if strings.TrimSpace(opts.Seen) != "" {
		summaryJobs = unseenJobs
	}
	printSearchSummary(ctx, summaryJobs)

	return nil
}

func pathsEqual(a, b string) bool {
	if strings.TrimSpace(a) == "" || strings.TrimSpace(b) == "" {
		return false
	}
	absA, errA := filepath.Abs(a)
	absB, errB := filepath.Abs(b)
	if errA == nil && errB == nil {
		return absA == absB
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func updateSeenHistory(seenPath string, inputJobs []models.Job) error {
	seenJobs, err := seen.ReadJobsAllowMissing(seenPath)
	if err != nil {
		return fmt.Errorf("read --seen: %w", err)
	}

	mergedJobs, _ := seen.Merge(seenJobs, inputJobs)
	if err := seen.WriteJobs(seenPath, mergedJobs); err != nil {
		return fmt.Errorf("write --seen: %w", err)
	}

	return nil
}

func printSearchSummary(ctx *Context, jobs []models.Job) {
	if ctx == nil || ctx.Err == nil {
		return
	}
	_, _ = fmt.Fprintf(ctx.Err, "\n%s\n", formatSearchSummary(jobs))
}

func formatSearchSummary(jobs []models.Job) string {
	counts := countJobsBySite(jobs)
	if len(counts) == 0 {
		return "summary: new_jobs=0 by_site=none"
	}

	parts := make([]string, 0, len(counts))
	for _, count := range counts {
		parts = append(parts, fmt.Sprintf("%s:%d", count.site, count.total))
	}

	return fmt.Sprintf("summary: new_jobs=%d by_site=%s", len(jobs), strings.Join(parts, ", "))
}

type siteCount struct {
	site  string
	total int
}

func countJobsBySite(jobs []models.Job) []siteCount {
	siteTotals := make(map[string]int, len(jobs))
	for _, job := range jobs {
		site := strings.ToLower(strings.TrimSpace(job.Site))
		if site == "" {
			site = "unknown"
		}
		siteTotals[site]++
	}

	counts := make([]siteCount, 0, len(siteTotals))
	for site, total := range siteTotals {
		counts = append(counts, siteCount{site: site, total: total})
	}

	sort.SliceStable(counts, func(i, j int) bool {
		return counts[i].site < counts[j].site
	})
	return counts
}

type queryFileConfig struct {
	Queries []string
	Search  queryFileSearchOptions
	Global  queryFileGlobalOptions
}

type queryFileSearchOptions struct {
	Location   *string `json:"location"`
	Country    *string `json:"country"`
	Sites      *string `json:"sites"`
	Limit      *int    `json:"limit"`
	Offset     *int    `json:"offset"`
	Remote     *bool   `json:"remote"`
	JobType    *string `json:"job_type"`
	Hours      *int    `json:"hours"`
	Format     *string `json:"format"`
	Links      *string `json:"links"`
	Output     *string `json:"output"`
	Proxies    *string `json:"proxies"`
	Seen       *string `json:"seen"`
	NewOnly    *bool   `json:"new_only"`
	NewOut     *string `json:"new_out"`
	SeenUpdate *bool   `json:"seen_update"`
}

type queryFileGlobalOptions struct {
	JSON    *bool   `json:"json"`
	Plain   *bool   `json:"plain"`
	Color   *string `json:"color"`
	Verbose *bool   `json:"verbose"`
}

func applyQueryFileDefaults(ctx *Context, opts SearchOptions, sitesArg string, fileCfg queryFileConfig, allowSitesOverride bool, providedFlags map[string]bool) (SearchOptions, string, error) {
	cliProvided := func(names ...string) bool {
		return hasAnyProvidedFlag(providedFlags, names...)
	}

	if fileCfg.Search.Location != nil && !cliProvided("--location") {
		opts.Location = *fileCfg.Search.Location
	}
	if fileCfg.Search.Country != nil && !cliProvided("--country") {
		opts.Country = *fileCfg.Search.Country
	}
	if fileCfg.Search.Limit != nil && !cliProvided("--limit") {
		opts.Limit = *fileCfg.Search.Limit
	}
	if fileCfg.Search.Offset != nil && !cliProvided("--offset") {
		opts.Offset = *fileCfg.Search.Offset
	}
	if fileCfg.Search.Remote != nil && !cliProvided("--remote") {
		opts.Remote = *fileCfg.Search.Remote
	}
	if fileCfg.Search.JobType != nil && !cliProvided("--job-type") {
		opts.JobType = strings.ToLower(strings.TrimSpace(*fileCfg.Search.JobType))
	}
	if fileCfg.Search.Hours != nil && !cliProvided("--hours") {
		opts.Hours = *fileCfg.Search.Hours
	}
	if fileCfg.Search.Format != nil && !cliProvided("--format") {
		opts.Format = strings.ToLower(strings.TrimSpace(*fileCfg.Search.Format))
	}
	if fileCfg.Search.Links != nil && !cliProvided("--links") {
		opts.Links = strings.ToLower(strings.TrimSpace(*fileCfg.Search.Links))
	}
	if fileCfg.Search.Proxies != nil && !cliProvided("--proxies") {
		opts.Proxies = *fileCfg.Search.Proxies
	}
	if fileCfg.Search.Seen != nil && !cliProvided("--seen") {
		opts.Seen = *fileCfg.Search.Seen
	}
	if fileCfg.Search.NewOnly != nil && !cliProvided("--new-only") {
		opts.NewOnly = *fileCfg.Search.NewOnly
	}
	if fileCfg.Search.NewOut != nil && !cliProvided("--new-out") {
		opts.NewOut = *fileCfg.Search.NewOut
	}
	if fileCfg.Search.SeenUpdate != nil && !cliProvided("--seen-update") {
		opts.SeenUpdate = *fileCfg.Search.SeenUpdate
	}
	if fileCfg.Search.Output != nil && !cliProvided("--output", "--out", "--file", "-o") {
		opts.Output = *fileCfg.Search.Output
		opts.Out = ""
		opts.File = ""
	}
	if allowSitesOverride && fileCfg.Search.Sites != nil && !cliProvided("--sites") {
		sitesArg = strings.TrimSpace(*fileCfg.Search.Sites)
	}

	if ctx == nil {
		return opts, sitesArg, nil
	}

	if fileCfg.Global.JSON != nil && !cliProvided("--json") {
		ctx.JSONOutput = *fileCfg.Global.JSON
	}
	if fileCfg.Global.Plain != nil && !cliProvided("--plain") {
		ctx.PlainText = *fileCfg.Global.Plain
	}
	if fileCfg.Global.Verbose != nil && !cliProvided("--verbose") {
		ctx.Verbose = *fileCfg.Global.Verbose
	}
	if fileCfg.Global.Color != nil && !cliProvided("--color") {
		ctx.ColorMode = ui.NormalizeColorMode(*fileCfg.Global.Color)
	}

	if ctx.JSONOutput && ctx.PlainText {
		return SearchOptions{}, "", fmt.Errorf("cannot combine --json and --plain")
	}

	level := zerolog.InfoLevel
	if ctx.Verbose {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	if ctx.Out != nil && ctx.Err != nil {
		disableColor := ctx.JSONOutput || ctx.PlainText
		ctx.UI = ui.New(ctx.Out, ctx.Err, ctx.ColorMode, disableColor)
	}

	return opts, sitesArg, nil
}

func collectProvidedFlags(args []string) map[string]bool {
	flags := make(map[string]bool)
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if strings.HasPrefix(arg, "--") {
			flag := arg
			if idx := strings.Index(flag, "="); idx > 0 {
				flag = flag[:idx]
			}
			flags[flag] = true
			continue
		}
		if arg == "-o" {
			flags[arg] = true
		}
	}
	return flags
}

func hasAnyProvidedFlag(providedFlags map[string]bool, names ...string) bool {
	for _, name := range names {
		if providedFlags[name] {
			return true
		}
	}
	return false
}

func parseQueries(raw string) ([]string, error) {
	return mergeAndNormalizeQueries(splitQueries(raw), nil)
}

func resolveQueries(raw string, queryFile string) ([]string, error) {
	fileCfg, err := loadQueryFileConfigIfProvided(queryFile)
	if err != nil {
		return nil, err
	}
	return mergeAndNormalizeQueries(splitQueries(raw), fileCfg.Queries)
}

func loadQueryFileConfigIfProvided(queryFile string) (queryFileConfig, error) {
	if strings.TrimSpace(queryFile) == "" {
		return queryFileConfig{}, nil
	}
	return loadQueryFileConfig(queryFile)
}

func loadQueryFileConfig(path string) (queryFileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return queryFileConfig{}, fmt.Errorf("read --query-file %q: %w", path, err)
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return queryFileConfig{}, fmt.Errorf("invalid --query-file %q: expected top-level string array or object with \"job_titles\" string array", path)
	}

	switch trimmed[0] {
	case '[':
		var values []any
		if err := json.Unmarshal(trimmed, &values); err != nil {
			return queryFileConfig{}, fmt.Errorf("parse --query-file %q: %w", path, err)
		}
		queries, err := parseStringArray(values, path, "root array")
		if err != nil {
			return queryFileConfig{}, err
		}
		return queryFileConfig{Queries: queries}, nil
	case '{':
		cfg, err := parseQueryFileObject(path, trimmed)
		if err != nil {
			return queryFileConfig{}, err
		}
		if err := validateQueryFileConfig(path, cfg); err != nil {
			return queryFileConfig{}, err
		}
		return cfg, nil
	default:
		return queryFileConfig{}, fmt.Errorf("invalid --query-file %q: expected top-level string array or object with \"job_titles\" string array", path)
	}
}

func parseQueryFileObject(path string, data []byte) (queryFileConfig, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return queryFileConfig{}, fmt.Errorf("parse --query-file %q: %w", path, err)
	}

	allowedFields := map[string]struct{}{
		"job_titles":     {},
		"search_options": {},
		"global_options": {},
	}
	for field := range root {
		if _, ok := allowedFields[field]; ok {
			continue
		}
		return queryFileConfig{}, fmt.Errorf("invalid --query-file %q: unknown field %q", path, field)
	}

	cfg := queryFileConfig{}

	if rawTitles, ok := root["job_titles"]; ok {
		var titles []any
		if err := json.Unmarshal(rawTitles, &titles); err != nil {
			return queryFileConfig{}, fmt.Errorf("invalid --query-file %q: field \"job_titles\" must be an array of strings", path)
		}
		queries, err := parseStringArray(titles, path, "job_titles")
		if err != nil {
			return queryFileConfig{}, err
		}
		cfg.Queries = queries
	}

	if rawSearch, ok := root["search_options"]; ok {
		if err := unmarshalStrict(rawSearch, &cfg.Search); err != nil {
			return queryFileConfig{}, fmt.Errorf("invalid --query-file %q: field \"search_options\": %w", path, err)
		}
	}

	if rawGlobal, ok := root["global_options"]; ok {
		if err := unmarshalStrict(rawGlobal, &cfg.Global); err != nil {
			return queryFileConfig{}, fmt.Errorf("invalid --query-file %q: field \"global_options\": %w", path, err)
		}
	}

	return cfg, nil
}

func validateQueryFileConfig(path string, cfg queryFileConfig) error {
	if cfg.Search.JobType != nil {
		jobType := strings.ToLower(strings.TrimSpace(*cfg.Search.JobType))
		switch jobType {
		case "", "fulltime", "parttime", "contract", "internship":
		default:
			return fmt.Errorf("invalid --query-file %q: field \"search_options.job_type\" must be one of fulltime, parttime, contract, internship, or empty", path)
		}
	}

	if cfg.Search.Format != nil {
		format := strings.ToLower(strings.TrimSpace(*cfg.Search.Format))
		switch format {
		case "", "csv", "json", "md":
		default:
			return fmt.Errorf("invalid --query-file %q: field \"search_options.format\" must be one of csv, json, md, or empty", path)
		}
	}

	if cfg.Search.Links != nil {
		links := strings.ToLower(strings.TrimSpace(*cfg.Search.Links))
		if links != "short" && links != "full" {
			return fmt.Errorf("invalid --query-file %q: field \"search_options.links\" must be one of short or full", path)
		}
	}

	if cfg.Global.Color != nil {
		color := strings.ToLower(strings.TrimSpace(*cfg.Global.Color))
		switch color {
		case "auto", "always", "never":
		default:
			return fmt.Errorf("invalid --query-file %q: field \"global_options.color\" must be one of auto, always, or never", path)
		}
	}

	if cfg.Global.JSON != nil && cfg.Global.Plain != nil && *cfg.Global.JSON && *cfg.Global.Plain {
		return fmt.Errorf("invalid --query-file %q: global_options.json and global_options.plain cannot both be true", path)
	}

	return nil
}

func loadQueriesFromJSON(path string) ([]string, error) {
	cfg, err := loadQueryFileConfig(path)
	if err != nil {
		return nil, err
	}
	return cfg.Queries, nil
}

func unmarshalStrict(data []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("unexpected trailing JSON value")
	}
	return nil
}

func splitQueries(raw string) []string {
	parts := strings.Split(raw, ",")
	queries := make([]string, 0, len(parts))

	for _, part := range parts {
		query := strings.TrimSpace(part)
		if query == "" {
			continue
		}
		queries = append(queries, query)
	}

	return queries
}

func mergeAndNormalizeQueries(primary []string, secondary []string) ([]string, error) {
	queries := make([]string, 0, len(primary)+len(secondary))
	seenQueries := make(map[string]struct{}, len(primary)+len(secondary))

	appendUnique := func(rawQuery string) {
		query := strings.TrimSpace(rawQuery)
		if query == "" {
			return
		}
		normalized := strings.ToLower(query)
		if _, exists := seenQueries[normalized]; exists {
			return
		}
		seenQueries[normalized] = struct{}{}
		queries = append(queries, query)
	}

	for _, query := range primary {
		appendUnique(query)
	}
	for _, query := range secondary {
		appendUnique(query)
	}

	if len(queries) == 0 {
		return nil, fmt.Errorf("at least one non-empty query is required")
	}
	if len(queries) > maxQueries {
		return nil, fmt.Errorf("too many queries: max %d", maxQueries)
	}

	return queries, nil
}

func parseStringArray(values []any, path string, fieldName string) ([]string, error) {
	queries := make([]string, 0, len(values))
	for idx, rawValue := range values {
		query, ok := rawValue.(string)
		if !ok {
			return nil, fmt.Errorf("invalid --query-file %q: %s[%d] must be a string", path, fieldName, idx)
		}
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		queries = append(queries, query)
	}
	return queries, nil
}

func runScrapersForQuery(scrapers []scraper.Scraper, base models.SearchParams, query string) ([]models.Job, []scraperFailure, error) {
	params := base
	params.Query = query
	return runScrapers(scrapers, params)
}

func mergeUniqueJobs(existing []models.Job, incoming []models.Job) []models.Job {
	if len(incoming) == 0 {
		return existing
	}

	keys := make(map[string]struct{}, len(existing)+len(incoming))
	merged := make([]models.Job, 0, len(existing)+len(incoming))

	for _, job := range existing {
		merged = append(merged, job)
		key, ok := seen.Key(job)
		if !ok {
			continue
		}
		keys[key] = struct{}{}
	}

	for _, job := range incoming {
		key, ok := seen.Key(job)
		if !ok {
			merged = append(merged, job)
			continue
		}
		if _, exists := keys[key]; exists {
			continue
		}
		merged = append(merged, job)
	}

	return merged
}

func limitJobs(jobs []models.Job, limit int) []models.Job {
	if limit <= 0 || len(jobs) <= limit {
		return jobs
	}
	return jobs[:limit]
}

func runScrapers(scrapers []scraper.Scraper, params models.SearchParams) ([]models.Job, []scraperFailure, error) {
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

	var (
		all      []models.Job
		failures []scraperFailure
	)
	for res := range results {
		if res.err != nil {
			failures = append(failures, scraperFailure{
				site:           res.site,
				err:            res.err,
				notImplemented: errors.Is(res.err, scraper.ErrNotImplemented),
			})
			continue
		}
		all = append(all, res.jobs...)
	}

	sortJobsBySite(all)
	sortScraperFailures(failures)

	return all, failures, nil
}

func sortJobsBySite(jobs []models.Job) {
	sort.SliceStable(jobs, func(i, j int) bool {
		return strings.ToLower(jobs[i].Site) < strings.ToLower(jobs[j].Site)
	})
}

func sortScraperFailures(failures []scraperFailure) {
	sort.SliceStable(failures, func(i, j int) bool {
		return strings.ToLower(failures[i].site) < strings.ToLower(failures[j].site)
	})
}

type scraperResult struct {
	site string
	jobs []models.Job
	err  error
}

type scraperFailure struct {
	site           string
	err            error
	notImplemented bool
}

func reportScraperFailures(ctx *Context, failures []scraperFailure) {
	if ctx == nil || ctx.UI == nil {
		return
	}
	if !ctx.Verbose {
		return
	}

	if len(failures) == 0 {
		return
	}

	ctx.UI.Warnf("\nScraper errors:")
	for _, failure := range failures {
		ctx.UI.Warnf("  %s: %v", failure.site, failure.err)
	}
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
		if ctx.JSONOutput {
			return export.FormatJSON, nil
		}
		if ctx.PlainText {
			return export.FormatTSV, nil
		}
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
		case "zip", "zip-recruiter":
			out = append(out, scraper.SiteZipRecruiter)
		case "stepstone.de", "stepstone-de":
			out = append(out, scraper.SiteStepstone)
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
