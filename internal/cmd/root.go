package cmd

import (
	"github.com/alecthomas/kong"
	"github.com/jimezsa/jobcli/internal/scraper"
)

type CLI struct {
	Color   string `help:"Color output: auto, always, never." enum:"auto,always,never" default:"auto"`
	JSON    bool   `help:"JSON output to stdout; disables colors."`
	Plain   bool   `help:"TSV output to stdout; disables colors."`
	Verbose bool   `help:"Enable debug logging."`

	VersionFlag kong.VersionFlag `help:"Print version."`

	Version      VersionCmd `cmd:"" help:"Print version."`
	Config       ConfigCmd  `cmd:"" help:"Manage configuration."`
	Search       SearchCmd  `cmd:"" help:"Search job listings."`
	LinkedIn     SiteCmd    `cmd:"" name:"linkedin" help:"Search LinkedIn."`
	Indeed       SiteCmd    `cmd:"" name:"indeed" help:"Search Indeed."`
	Glassdoor    SiteCmd    `cmd:"" name:"glassdoor" help:"Search Glassdoor."`
	ZipRecruiter SiteCmd    `cmd:"" name:"ziprecruiter" help:"Search ZipRecruiter."`
	Google       SiteCmd    `cmd:"" name:"google" help:"Search Google Jobs."`
	Stepstone    SiteCmd    `cmd:"" name:"stepstone" help:"Search Stepstone."`
	Seen         SeenCmd    `cmd:"" help:"Seen jobs utilities."`
	Proxies      ProxiesCmd `cmd:"" help:"Proxy utilities."`
}

func NewCLI() *CLI {
	return &CLI{
		LinkedIn:     SiteCmd{Site: scraper.SiteLinkedIn},
		Indeed:       SiteCmd{Site: scraper.SiteIndeed},
		Glassdoor:    SiteCmd{Site: scraper.SiteGlassdoor},
		ZipRecruiter: SiteCmd{Site: scraper.SiteZipRecruiter},
		Google:       SiteCmd{Site: scraper.SiteGoogleJobs},
		Stepstone:    SiteCmd{Site: scraper.SiteStepstone},
	}
}
