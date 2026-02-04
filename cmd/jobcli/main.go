package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MrJJimenez/jobcli/internal/cmd"
	"github.com/MrJJimenez/jobcli/internal/config"
	"github.com/MrJJimenez/jobcli/internal/ui"
	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	if len(os.Args) == 1 {
		printOverview()
		return
	}

	cli := cmd.NewCLI()
	applyEnvDefaults(cli, os.Args[1:])
	versionString := buildVersion()

	parser, err := kong.New(cli,
		kong.Name("jobcli"),
		kong.Description("Job aggregation CLI."),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{"version": versionString},
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	kctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		fallbackUI := ui.New(os.Stdout, os.Stderr, ui.NormalizeColorMode(os.Getenv("JOBCLI_COLOR")), false)
		fallbackUI.Errorf("%v", err)
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	configDir, err := config.ConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	colorMode := ui.NormalizeColorMode(cli.Color)
	disableColor := cli.JSON || cli.Plain
	userInterface := ui.New(os.Stdout, os.Stderr, colorMode, disableColor)

	level := zerolog.InfoLevel
	if cli.Verbose {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	runCtx := &cmd.Context{
		Out:        os.Stdout,
		Err:        os.Stderr,
		UI:         userInterface,
		Config:     cfg,
		ConfigDir:  configDir,
		Logger:     logger,
		Verbose:    cli.Verbose,
		JSONOutput: cli.JSON,
		PlainText:  cli.Plain,
		Version:    versionString,
		ColorMode:  colorMode,
	}

	if cli.JSON && cli.Plain {
		userInterface.Errorf("cannot combine --json and --plain")
		os.Exit(1)
	}

	if err := kctx.Run(runCtx); err != nil {
		userInterface.Errorf("%v", err)
		os.Exit(1)
	}
}

func printOverview() {
	colorMode := ui.NormalizeColorMode(os.Getenv("JOBCLI_COLOR"))
	disableColor := envBool("JOBCLI_JSON")
	userInterface := ui.New(os.Stdout, os.Stderr, colorMode, disableColor)

	header := "🧑‍💻 JobCLI - Jobs in your terminal"
	description := "Fast, single-binary job aggregation CLI that scrapes multiple sites in parallel and exports results to table, CSV, TSV, JSON, or Markdown."

	fmt.Fprintln(os.Stdout, userInterface.LinkText(header))
	fmt.Fprintln(os.Stdout, description)
	fmt.Fprintln(os.Stdout)

	commands := []struct {
		Name string
		Desc string
	}{
		{Name: "search", Desc: "Search job listings."},
		{Name: "linkedin", Desc: "Search LinkedIn."},
		{Name: "indeed", Desc: "Search Indeed."},
		{Name: "glassdoor", Desc: "Search Glassdoor."},
		{Name: "ziprecruiter", Desc: "Search ZipRecruiter."},
		{Name: "google", Desc: "Search Google Jobs."},
		{Name: "config", Desc: "Manage configuration."},
		{Name: "proxies", Desc: "Proxy utilities."},
		{Name: "version", Desc: "Print version."},
		{Name: "help", Desc: "Show help."},
	}

	maxLen := 0
	for _, cmd := range commands {
		if len(cmd.Name) > maxLen {
			maxLen = len(cmd.Name)
		}
	}

	fmt.Fprintln(os.Stdout, "Commands:")
	for _, cmd := range commands {
		padding := strings.Repeat(" ", maxLen-len(cmd.Name))
		fmt.Fprintf(os.Stdout, "  %s%s  %s\n", userInterface.LinkText(cmd.Name), padding, cmd.Desc)
	}
	fmt.Fprintln(os.Stdout)

	fmt.Fprintln(os.Stdout, "Usage:")
	fmt.Fprintln(os.Stdout, "  jobcli search \"<query>\" [--location L] [--sites S] [--limit N] [--format csv|json|md]")
	fmt.Fprintln(os.Stdout, "  jobcli <site> \"<query>\" [flags]")
	fmt.Fprintln(os.Stdout)

	fmt.Fprintln(os.Stdout, "Example:")
	fmt.Fprintln(os.Stdout, "  jobcli search \"golang\" --location \"New York, NY\" --limit 25")
	fmt.Fprintln(os.Stdout, "  jobcli google \"data engineer\" --remote --format json")
	fmt.Fprintln(os.Stdout)

	fmt.Fprintln(os.Stdout, "Help:")
	fmt.Fprintln(os.Stdout, "  jobcli --help")
	fmt.Fprintln(os.Stdout, "  jobcli help")
}

func buildVersion() string {
	if commit == "" && date == "" {
		return version
	}
	if commit == "" {
		return fmt.Sprintf("%s (%s)", version, date)
	}
	if date == "" {
		return fmt.Sprintf("%s (%s)", version, commit)
	}
	return fmt.Sprintf("%s (%s, %s)", version, commit, date)
}

func applyEnvDefaults(cli *cmd.CLI, args []string) {
	hasJSON := hasFlag(args, "--json")
	hasPlain := hasFlag(args, "--plain")
	hasColor := hasFlag(args, "--color")
	hasVerbose := hasFlag(args, "--verbose")

	if !hasJSON && !hasPlain && envBool("JOBCLI_JSON") {
		cli.JSON = true
	}
	if !hasVerbose && envBool("JOBCLI_VERBOSE") {
		cli.Verbose = true
	}
	if !hasColor {
		if value := os.Getenv("JOBCLI_COLOR"); value != "" {
			cli.Color = value
		}
	}
}

func hasFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
		if strings.HasPrefix(arg, name+"=") {
			return true
		}
		if arg == "--" {
			return false
		}
	}
	return false
}

func envBool(key string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
